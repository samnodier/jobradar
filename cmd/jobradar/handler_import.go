package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samnodier/jobradar/internal/convert"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/extraction"
	"github.com/samnodier/jobradar/internal/fetcher"
	"github.com/samnodier/jobradar/internal/htmlx"
	"github.com/samnodier/jobradar/internal/httpx"
	"github.com/samnodier/jobradar/internal/llm"
)

type URLRequest struct {
	URL string `json:"url"`
}

type jobConfirmImportRequest struct {
	Industry       *string  `json:"industry"`
	Category       *string  `json:"category"`
	CompanyStage   *string  `json:"company_stage"`
	EmploymentType *string  `json:"employment_type"`
	SourceURL      string   `json:"source_url"`
	Title          string   `json:"title"`
	CompanyName    string   `json:"company_name"`
	Description    string   `json:"description"`
	SalaryMin      *int32   `json:"salary_min"`
	SalaryMax      *int32   `json:"salary_max"`
	Currency       *string  `json:"currency"`
	JobLocation    *string  `json:"job_location"`
	IsRemote       *bool    `json:"is_remote"`
	Skills         []string `json:"skills"`
	LogoURL        *string  `json:"logo_url"`
}

func (cfg *apiConfig) handlerExtractJob(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB
	ctx, cancel := context.WithTimeout(r.Context(), 75*time.Second)
	defer cancel()
	// User helper function to get the UserID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	// Decode the body
	var requestURL URLRequest
	if err := json.NewDecoder(r.Body).Decode(&requestURL); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate the URL
	parsedURL, err := url.Parse(requestURL.URL)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid url")
		return
	}
	// Gate the scheme to http/https
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		httpx.RespondError(w, http.StatusBadRequest, "only http and https URLs are supported")
		return
	}
	// Check if the URL contains the domain name or the API
	if parsedURL.Host == "" {
		httpx.RespondError(w, http.StatusBadRequest, "url must have a host")
		return
	}

	data, err := fetcher.Fetch(ctx, parsedURL.String())
	if err != nil {
		httpx.RespondError(w, http.StatusUnprocessableEntity, "could not fetch that URL")
		return
	}

	text, err := htmlx.HTMLToText(string(data))
	if err != nil {
		log.Printf("handleExtractJob: failed to parse page text for user %s: Err: %v", userID, err)
		httpx.RespondError(w, http.StatusInternalServerError, "unable to parse the page text")
		return
	}

	draft, structured := extraction.Page(string(data))
	if structured {
		httpx.RespondJSON(w, http.StatusOK, draft)
		return
	}
	// Every provider the user has a key for, in priority order
	keys, err := cfg.selectProviderKeys(ctx, userID)
	if err != nil {
		if errors.Is(err, errNoAPIKey) {
			httpx.RespondJSON(w, http.StatusOK, draft)
			return
		}
		log.Printf("handleExtractJob: failed to load api keys for user %s: %v", userID, err)
		httpx.RespondError(w, http.StatusInternalServerError, "failed to load your API key")
		return
	}

	// Try each provider in turn; fall back to the next one on any failure
	// (down, rate-limited, revoked key). extractErr holds the last failure.
	var result llm.ExtractionResult
	var extractErr error
	for _, pk := range keys {
		var llmExtractor llm.Extractor
		llmExtractor, extractErr = cfg.newExtractor(ctx, pk.provider, pk.apiKey)
		if extractErr != nil {
			log.Printf("handleExtractJob: failed to create %s extractor for user %s: %v", pk.provider, userID, extractErr)
			continue
		}
		result, extractErr = llmExtractor.Extract(ctx, llm.ExtractionInput{
			PageText: llm.LimitText(text, 24000),
		})
		if extractErr == nil {
			break
		}
		log.Printf("handleExtractJob: %s extract failed for URL %s user %s (trying next provider if any): %v", pk.provider, parsedURL.Hostname(), userID, extractErr)
	}
	if extractErr != nil {
		httpx.RespondJSON(w, http.StatusOK, draft)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, result)
}

func (cfg *apiConfig) handlerImportConfirm(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	ctx := r.Context()

	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	var req jobConfirmImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate the URL
	parsedURL, err := url.Parse(req.SourceURL)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid url")
		return
	}
	// Gate the scheme to http/https
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		httpx.RespondError(w, http.StatusBadRequest, "only http and https URLs are supported")
		return
	}
	// Check if the URL contains the domain name or the API
	if parsedURL.Host == "" {
		httpx.RespondError(w, http.StatusBadRequest, "url must have a host")
		return
	}

	// Trim spaces on non-empty fields
	jobTitle := strings.TrimSpace(req.Title)
	if jobTitle == "" {
		httpx.RespondError(w, http.StatusBadRequest, "title cannot be empty")
		return
	}
	companyName := strings.TrimSpace(req.CompanyName)
	if companyName == "" {
		httpx.RespondError(w, http.StatusBadRequest, "company name cannot be empty")
		return
	}
	jobDescription := strings.TrimSpace(req.Description)
	if jobDescription == "" {
		httpx.RespondError(w, http.StatusBadRequest, "description cannot be empty")
		return
	}

	host := parsedURL.Host // This can have www. prefix
	jobSource := strings.TrimPrefix(host, "www.")
	externalID := parsedURL.String()

	params := database.CreateImportedJobParams{
		Industry: req.Industry, Category: req.Category, CompanyStage: req.CompanyStage, EmploymentType: req.EmploymentType,
		ExternalID:      externalID,
		JobSource:       jobSource,
		Title:           jobTitle,
		CompanyName:     companyName,
		Description:     &jobDescription,
		SourceUrl:       parsedURL.String(),
		SalaryMin:       req.SalaryMin,
		SalaryMax:       req.SalaryMax,
		Currency:        req.Currency,
		JobLocation:     req.JobLocation,
		IsRemote:        req.IsRemote,
		Skills:          req.Skills,
		LogoUrl:         req.LogoURL,
		CreatedByUserID: convert.ToPgUUID(userID),
	}
	createdJob, err := cfg.db.CreateImportedJob(ctx, params)
	if err != nil {
		// check for postgres-specific errors
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "user account not found")
				return
			}
		}
		log.Printf("handlerImportConfirm: creating or updating the imported job failed for user %s: %v", userID, err)
		httpx.RespondError(w, http.StatusInternalServerError, "failed to create job")
		return
	}

	// The jobs trigger has already recorded durable matching work.

	httpx.RespondJSON(w, http.StatusCreated, createdJob)
}
