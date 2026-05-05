package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/samnodier/jobradar/internal/auth"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
)

func (cfg *apiConfig) handlerJobSave(w http.ResponseWriter, r *http.Request) {
	// Extract the session and then the user ID
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "Session not found in context")
		return
	}

	userID := session.UserID

	type saveJobRequest struct {
		JobID uuid.UUID `json:"job_id"`
	}

	var req saveJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	savedJob, err := cfg.db.SaveJob(r.Context(), database.SaveJobParams{
		UserID: userID,
		JobID:  req.JobID,
	})
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to save the job")
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, savedJob)
}
