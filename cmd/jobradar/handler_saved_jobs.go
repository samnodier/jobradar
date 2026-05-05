package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
)

type saveJobRequest struct {
	JobID uuid.UUID `json:"job_id"`
}

func (cfg *apiConfig) handlerJobSave(w http.ResponseWriter, r *http.Request) {
	// Use helper function to get userID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	var req saveJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.JobID == uuid.Nil {
		httpx.RespondError(w, http.StatusBadRequest, "job_id is required")
		return
	}

	savedJob, err := cfg.db.SaveJob(r.Context(), database.SaveJobParams{
		UserID: userID,
		JobID:  req.JobID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation — already saved
				httpx.RespondError(w, http.StatusConflict, "job already saved")
				return
			case "23503": // foreign_key_violation — job doesn't exist
				httpx.RespondError(w, http.StatusNotFound, "job not found")
				return
			}
		}
		httpx.RespondError(w, http.StatusInternalServerError, "failed to save the job")
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, savedJob)
}

func (cfg *apiConfig) handlerJobUnsave(w http.ResponseWriter, r *http.Request) {
	// Use helper function to get userID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	var req saveJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.JobID == uuid.Nil {
		httpx.RespondError(w, http.StatusBadRequest, "job_id is required")
		return
	}

	// Unsave the job
	err := cfg.db.UnSaveJob(r.Context(), database.UnSaveJobParams{
		JobID:  req.JobID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "saved job not found")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "failed to unsave the job")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]string{"message": "job unsaved"})
}
