package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerJobsGet(w http.ResponseWriter, r *http.Request) {
	jobs, err := cfg.db.GetJobs(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't fetch jobs")
		return
	}
	respondWithJSON(w, http.StatusOK, jobs)
}

func (cfg *apiConfig) handlerJobGetByID(w http.ResponseWriter, r *http.Request) {
	jobIDString := chi.URLParam(r, "jobID")
	jobID, err := uuid.Parse(jobIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid job ID format")
		return
	}

	job, err := cfg.db.GetJobByID(r.Context(), jobID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Job not found")
		return
	}

	respondWithJSON(w, http.StatusOK, job)
}
