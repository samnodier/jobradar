package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"

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

	limit, offset := 50, 0
	if v := r.URL.Query().Get("limit"); v != "" {
		n, e := strconv.Atoi(v)
		if e != nil || n < 1 || n > 100 {
			httpx.RespondError(w, 400, "limit must be 1–100")
			return
		}
		limit = n
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		n, e := strconv.Atoi(v)
		if e != nil || n < 0 || n > 100000 {
			httpx.RespondError(w, 400, "invalid offset")
			return
		}
		offset = n
	}
	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = "all"
	}
	switch filter {
	case "all", "saved", "applied", "remote", "today", "matched":
	default:
		httpx.RespondError(w, 400, "invalid filter")
		return
	}
	jobs, err := cfg.db.GetJobs(r.Context(), database.GetJobsParams{UserID: userID, Limit: int32(limit), Offset: int32(offset), Search: r.URL.Query().Get("q"), Filter: filter, Oldest: r.URL.Query().Get("order") == "oldest"})
	if err != nil {
		log.Printf("Error fetching jobs: %v", err)
		httpx.RespondError(w, http.StatusInternalServerError, "Couldn't fetch jobs")
		return
	}
	if jobs == nil {
		jobs = []database.GetJobsRow{}
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
