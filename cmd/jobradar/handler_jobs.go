package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samnodier/jobradar/internal/auth"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
)

func (cfg *apiConfig) handlerJobsGet(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	userID := uuid.Nil
	if ok {
		userID = session.UserID
	}

	jobs, err := cfg.db.GetJobs(r.Context(), userID)
	if err != nil {
		log.Printf("Error fetching jobs: %v", err)
		httpx.RespondError(w, http.StatusInternalServerError, "Couldn't fetch jobs")
		return
	}
	httpx.RespondJSON(w, http.StatusOK, jobs)
}

func (cfg *apiConfig) handlerJobGetByID(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	userID := uuid.Nil
	if ok {
		userID = session.UserID
	}

	jobIDString := chi.URLParam(r, "jobID")
	jobID, err := uuid.Parse(jobIDString)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "Invalid job ID format")
		return
	}

	job, err := cfg.db.GetJobByID(r.Context(), database.GetJobByIDParams{
		UserID: userID,
		ID:     jobID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "job not found")
			return
		}
		log.Printf("Error fetching job %v: %v", jobID, err)

		httpx.RespondError(w, http.StatusInternalServerError, "failed to fetch the job")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, job)
}

func (cfg *apiConfig) handlerJobStatsGet(w http.ResponseWriter, r *http.Request) {
	stats, err := cfg.db.GetJobStats(r.Context())
	if err != nil {
		log.Printf("Error fetching job stats: %v", err)
		httpx.RespondError(w, http.StatusInternalServerError, "Couldn't fetch job stats")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, stats)
}
