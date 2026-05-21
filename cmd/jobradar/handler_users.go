package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
)

type userUpdateRequest struct {
	Username               *string  `json:"username"`
	FullName               *string  `json:"full_name"`
	Phone                  *string  `json:"phone"`
	UserLocation           *string  `json:"user_location"`
	WebsiteURL             *string  `json:"website_url"`
	LinkedInURL            *string  `json:"linkedin_url"`
	GitHubURL              *string  `json:"github_url"`
	Headline               *string  `json:"headline"`
	UserSummary            *string  `json:"user_summary"`
	Availability           *string  `json:"availability"`
	MinSalary              *int32   `json:"min_salary"`
	MaxSalary              *int32   `json:"max_salary"`
	SalaryCurrency         *string  `json:"salary_currency"`
	YearsOfExperience      *int32   `json:"years_of_experience"`
	PreferredJobTypes      []string `json:"preferred_job_types"`
	PreferredIndustries    []string `json:"preferred_industries"`
	CompanyStagePreference []string `json:"company_stage_preference"`
	NotifyJobs             *bool    `json:"notify_jobs"`
}

func (cfg *apiConfig) handlerUserGet(w http.ResponseWriter, r *http.Request) {
	// Get the user
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	user, err := cfg.db.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {

			httpx.RespondError(w, http.StatusNotFound, "user not found")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "failed to retrieve user")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, user)
}

func (cfg *apiConfig) handlerUserUpdate(w http.ResponseWriter, r *http.Request) {
	// Get the user
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	var req userUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == nil || strings.TrimSpace(*req.Username) == "" {
		httpx.RespondError(w, http.StatusBadRequest, "username is required")
		return
	}

	if req.FullName == nil || strings.TrimSpace(*req.FullName) == "" {
		httpx.RespondError(w, http.StatusBadRequest, "full name is required")
		return
	}

	user, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:                     userID,
		Username:               *req.Username,
		FullName:               req.FullName,
		Phone:                  req.Phone,
		UserLocation:           req.UserLocation,
		WebsiteUrl:             req.WebsiteURL,
		LinkedinUrl:            req.LinkedInURL,
		GithubUrl:              req.GitHubURL,
		Headline:               req.Headline,
		UserSummary:            req.UserSummary,
		Availability:           req.Availability,
		MinSalary:              req.MinSalary,
		MaxSalary:              req.MaxSalary,
		SalaryCurrency:         req.SalaryCurrency,
		YearsOfExperience:      req.YearsOfExperience,
		PreferredJobTypes:      req.PreferredJobTypes,
		PreferredIndustries:    req.PreferredIndustries,
		CompanyStagePreference: req.CompanyStagePreference,
		NotifyJobs:             req.NotifyJobs,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "user not found")
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			httpx.RespondError(w, http.StatusBadRequest, "username already taken")
			return
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, user)
}

func (cfg *apiConfig) HandleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	// 1. Extract session from the context (validated by @RequireAuth)
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	// 2. Get the session cookie from the request
	cookie, err := r.Cookie("session_id")
	if err != nil {
		httpx.RespondError(w, http.StatusUnauthorized, "Session not found")
		return
	}

	// Delete the user from the database
	err = cfg.db.DeleteUserByID(r.Context(), userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	// Delete the Redis session
	_ = cfg.rdb.Del(r.Context(), "session:"+cookie.Value)

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cfg.IsProduction,
		SameSite: http.SameSiteLaxMode,
	})

	// Respond with success
	httpx.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "account deleted successfully",
	})
}
