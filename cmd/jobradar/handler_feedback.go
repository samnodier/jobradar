package main

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
	"net/http"
)

func (cfg *apiConfig) handlerJobFeedback(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}
	jobID, err := uuid.Parse(chi.URLParam(r, "jobID"))
	if err != nil {
		httpx.RespondError(w, 400, "invalid job ID")
		return
	}
	// GET/DELETE also authorize the job, preventing private-job existence leaks.
	if _, err := cfg.db.GetJobByID(r.Context(), database.GetJobByIDParams{UserID: userID, ID: jobID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, 404, "job not found")
		} else {
			httpx.RespondError(w, 500, "could not load job")
		}
		return
	}
	sentiment := int16(0)
	switch r.Method {
	case http.MethodGet:
		sentiment, err = cfg.db.GetJobFeedback(r.Context(), database.GetJobFeedbackParams{UserID: userID, JobID: jobID})
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
		}
	case http.MethodDelete:
		err = cfg.db.ClearJobFeedback(r.Context(), database.ClearJobFeedbackParams{UserID: userID, JobID: jobID})
	case http.MethodPut:
		var body struct {
			Sentiment int16 `json:"sentiment"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		if json.NewDecoder(r.Body).Decode(&body) != nil || (body.Sentiment != -1 && body.Sentiment != 1) {
			httpx.RespondError(w, 400, "sentiment must be -1 or 1")
			return
		}
		sentiment, err = cfg.db.SetJobFeedback(r.Context(), database.SetJobFeedbackParams{UserID: userID, JobID: jobID, Sentiment: body.Sentiment})
	}
	if err != nil {
		httpx.RespondError(w, 500, "could not save feedback")
		return
	}
	httpx.RespondJSON(w, 200, map[string]any{"sentiment": sentiment})
}
