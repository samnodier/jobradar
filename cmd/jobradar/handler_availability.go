package main

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/fetcher"
	"github.com/samnodier/jobradar/internal/httpx"
	"net/http"
	"time"
)

func (cfg *apiConfig) checkJobAvailability(ctx context.Context, jobID uuid.UUID, sourceURL string) error {
	// Daily caching bounds requests. Transport failures remain unknown rather than
	// turning a bot-blocked or temporarily offline employer into a closed job.
	old, err := cfg.db.GetJobAvailability(ctx, jobID)
	if err == nil && time.Since(old.CheckedAt.Time) < 24*time.Hour {
		return nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	callCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	status, code := fetcher.CheckAvailability(callCtx, sourceURL)
	return cfg.db.RecordJobAvailability(ctx, database.RecordJobAvailabilityParams{JobID: jobID, Status: status, ResponseCode: int32(code)})
}
func (cfg *apiConfig) handlerJobAvailability(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}
	jobID, err := uuid.Parse(chi.URLParam(r, "jobID"))
	if err != nil {
		httpx.RespondError(w, 400, "invalid job ID")
		return
	}
	job, err := cfg.db.GetJobByID(r.Context(), database.GetJobByIDParams{UserID: userID, ID: jobID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, 404, "job not found")
		} else {
			httpx.RespondError(w, 500, "could not load job")
		}
		return
	}
	if r.Method == http.MethodPost {
		if err := cfg.checkJobAvailability(r.Context(), jobID, job.SourceUrl); err != nil {
			httpx.RespondError(w, 500, "could not record availability")
			return
		}
	}
	result, err := cfg.db.GetJobAvailability(r.Context(), jobID)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.RespondJSON(w, 200, map[string]any{"status": "unchecked"})
		return
	}
	if err != nil {
		httpx.RespondError(w, 500, "could not load availability")
		return
	}
	httpx.RespondJSON(w, 200, result)
}
func (cfg *apiConfig) filterDigestAvailability(ctx context.Context, rows []database.GetDigestCandidatesRow) ([]database.GetDigestCandidatesRow, error) {
	kept := make([]database.GetDigestCandidatesRow, 0, len(rows))
	for _, job := range rows {
		if err := cfg.checkJobAvailability(ctx, job.ID, job.SourceUrl); err != nil {
			return nil, err
		}
		result, err := cfg.db.GetJobAvailability(ctx, job.ID)
		if err != nil {
			return nil, err
		}
		if result.Status != "unavailable" {
			kept = append(kept, job)
		}
	}
	return kept, nil
}
