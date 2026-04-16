package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/samnodier/jobradar/internal/httpx"
)

func (cfg *apiConfig) handlerJobsGet(w http.ResponseWriter, r *http.Request) {
	jobs, err := cfg.db.GetJobs(r.Context())
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Couldn't fetch jobs")
		return
	}
	httpx.RespondJSON(w, http.StatusOK, jobs)
}

func (cfg *apiConfig) handlerJobGetByID(w http.ResponseWriter, r *http.Request) {
	jobIDString := chi.URLParam(r, "jobID")
	jobID, err := uuid.Parse(jobIDString)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "Invalid job ID format")
		return
	}

	job, err := cfg.db.GetJobByID(r.Context(), jobID)
	if err != nil {
		httpx.RespondError(w, http.StatusNotFound, "Job not found")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, job)
}
