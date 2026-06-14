package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samnodier/jobradar/internal/convert"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/fetcher"
	"github.com/samnodier/jobradar/internal/htmlx"
	"github.com/samnodier/jobradar/internal/httpx"
	"github.com/samnodier/jobradar/internal/llm"
	"github.com/samnodier/jobradar/internal/queue"
)

type URLRequest struct {
	URL string `json:"url"`
}

type jobConfirmImportRequest struct {
	SourceURL   string   `json:"source_url"`
	Title       string   `json:"title"`
	CompanyName string   `json:"company_name"`
	Description string   `json:"description"`
	SalaryMin   *int32   `json:"salary_min"`
	SalaryMax   *int32   `json:"salary_max"`
	Currency    *string  `json:"currency"`
	JobLocation *string  `json:"job_location"`
	IsRemote    *bool    `json:"is_remote"`
	Skills      []string `json:"skills"`
	LogoURL     *string  `json:"logo_url"`
}

func (cfg *apiConfig) handlerExtractJob(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB
	ctx := r.Context()
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

	// Get the user Gemini Key
	geminiKeyPointer, err := cfg.db.GetGeminiKeyByUserID(ctx, userID)
	if err != nil {
		log.Printf("handleExtractJob: failed to fetch gemini key for user: %s. Err: %v", userID, err)
		httpx.RespondError(w, http.StatusInternalServerError, "can't fetch the gemini key")
		return
	}
	if geminiKeyPointer.EncryptedGeminiApiKey == nil {
		httpx.RespondError(w, http.StatusUnprocessableEntity, "no Gemini API key configured - add one in Settings")
		return
	}

	geminiKey, err := cfg.crypto.Decrypt(*geminiKeyPointer.EncryptedGeminiApiKey)
	if err != nil {
		log.Printf("handleExtractJob: failed to decrypt gemini key for user %s: %v", userID, err)
		httpx.RespondError(w, http.StatusInternalServerError, "failed to decrypt the gemini key")
		return
	}

	llmExtractor, err := cfg.newExtractor(ctx, geminiKey)
	if err != nil {
		log.Printf("handleExtractJob: failed to create extractor for user %s: %v", userID, err)
		httpx.RespondError(w, http.StatusInternalServerError, "failed to create extractor")
		return
	}

	result, err := llmExtractor.Extract(ctx, llm.ExtractionInput{
		PageText: text,
	})
	if err != nil {
		// The permanent error (bad key -> 401) is the user's problem (422)
		if errors.Is(err, llm.ErrPermanent) {
			log.Printf("handleExtractJob: permanent failure for URL %s user %s, dropping: %v", parsedURL, userID, err)
			httpx.RespondError(w, http.StatusUnprocessableEntity, "failed to extract the job details from the url")
			return // drop it, don't retry
		} else {
			// Transient error are a service's problem
			// We should try again
			log.Printf("handleExtractJob: transient failure for URL %s user %s, retrying: %v", parsedURL, userID, err)
			httpx.RespondError(w, http.StatusBadGateway, "service unavailable")
			return
		}
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

	payload := MatchJobPayload{
		JobID:  createdJob.ID.String(),
		UserID: userID.String(),
	}
	matchJobPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal match job payload: %v", err)
		return
	}
	// Push matching task to queue
	err = cfg.queue.Enqueue(cfg.rootCtx, &queue.Job{
		ID:       createdJob.ID.String() + "-" + userID.String(),
		Type:     queue.JobMatchJob,
		Payload:  matchJobPayload,
		MaxRetry: 3,
	})
	if err != nil {
		log.Printf("failed to enqueue matching job for %s: %v", createdJob.ID, err)
	}

	httpx.RespondJSON(w, http.StatusCreated, createdJob)
}
