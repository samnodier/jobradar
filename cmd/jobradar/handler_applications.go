package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samnodier/jobradar/internal/auth"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
)

func (cfg *apiConfig) handlerApplicationGetByID(w http.ResponseWriter, r *http.Request) {
	// Extract the session and then the user ID
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "Session not found in context")
		return
	}

	userID := session.UserID

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
	// Extract the session and then the user ID
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "Session not found in context")
		return
	}

	userID := session.UserID
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
	// Extract the session and then the user ID
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "Session not found in context")
		return
	}

	userID := session.UserID

	type applicationCreateRequest struct {
		JobID  uuid.UUID `json:"job_id"`
		Status string    `json:"status"`
	}

	var req applicationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	application, err := cfg.db.CreateApplication(r.Context(), database.CreateApplicationParams{
		UserID: userID,
		JobID:  req.JobID,
		Status: req.Status,
	})
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to create the application")
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, application)
}
