package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
)

func (cfg *apiConfig) handlerApplicationGetByID(w http.ResponseWriter, r *http.Request) {
	// Use helper function to get userID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	applicationIDString := chi.URLParam(r, "applicationID")
	applicationID, err := uuid.Parse(applicationIDString)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	application, err := cfg.db.GetApplicationByID(r.Context(), database.GetApplicationByIDParams{
		ID:     applicationID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "application not found")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "failed to retrieve application")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, application)
}

func (cfg *apiConfig) handlerApplicationsGet(w http.ResponseWriter, r *http.Request) {
	// Use helper function to get userID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	applications, err := cfg.db.GetApplicationsByUserID(r.Context(), userID)
	if err != nil {
		// Check for pgx "no rows" err
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondJSON(w, http.StatusOK, []database.GetApplicationsByUserIDRow{})
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "failed to retrieve applications")
		return
	}

	if applications == nil {
		applications = []database.GetApplicationsByUserIDRow{}
	}

	httpx.RespondJSON(w, http.StatusOK, applications)
}

func (cfg *apiConfig) handlerApplicationCreate(w http.ResponseWriter, r *http.Request) {
	// Use helper function to get userID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	type applicationCreateRequest struct {
		JobID             uuid.UUID `json:"job_id"`
		ApplicationStatus string    `json:"application_status"`
	}

	var req applicationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	application, err := cfg.db.CreateApplication(r.Context(), database.CreateApplicationParams{
		UserID:            userID,
		JobID:             req.JobID,
		ApplicationStatus: req.ApplicationStatus,
	})
	if err != nil {
		// check for "no rows" first (e.g. if query uses RETURNING and finds nothing)
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "job not found")
			return
		}

		// check for postgres-specific errors
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				httpx.RespondError(w, http.StatusConflict, "you have already tracked this application")
				return
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "job not found")
				return
			}
		}
		httpx.RespondError(w, http.StatusInternalServerError, "failed to create the application")
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, application)
}

// Update the application
func (cfg *apiConfig) handlerApplicationUpdate(w http.ResponseWriter, r *http.Request) {
	// Use helper function to get userID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	// Get the application id from url and parse it
	appIDStr := chi.URLParam(r, "id")
	applicationID, err := uuid.Parse(appIDStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid application id format")
		return
	}

	// Decode the JSON body
	type applicationUpdateRequest struct {
		ApplicationNotes  *string    `json:"notes"`
		ApplicationStatus *string    `json:"application_status"`
		AppliedAt         *time.Time `json:"applied_at"`
		ClearAppliedAt    bool       `json:"clear_applied_at"`
		FollowUpAt        *time.Time `json:"follow_up_at"`
		ClearFollowUpAt   bool       `json:"clear_follow_up_at"`
	}

	var req applicationUpdateRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	application, err := cfg.db.UpdateApplication(r.Context(), database.UpdateApplicationParams{
		ID:                applicationID,
		UserID:            userID,
		Notes:             req.ApplicationNotes,
		ApplicationStatus: req.ApplicationStatus,
		ClearAppliedAt:    &req.ClearAppliedAt,
		AppliedAt:         req.AppliedAt,
		ClearFollowUpAt:   &req.ClearFollowUpAt,
		FollowUpAt:        req.FollowUpAt,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "application not found")
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			httpx.RespondError(w, http.StatusBadRequest, "invalid application status")
			return
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to update application")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, application)
}
