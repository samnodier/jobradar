package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samnodier/jobradar/internal/convert"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
)

func (cfg *apiConfig) handlerCreateExperience(w http.ResponseWriter, r *http.Request) {
	// User helper function to get the UserID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	type experienceCreateRequest struct {
		UserID         uuid.UUID `json:"user_id"`
		CompanyName    string    `json:"company_name"`
		CompanyUrl     *string   `json:"company_url"`
		RoleTitle      string    `json:"role_title"`
		ExpLocation    *string   `json:"exp_location"`
		Industry       *string   `json:"industry"`
		EmploymentType *string   `json:"employment_type"`
		Description    *string   `json:"description"`
		Achievements   []string  `json:"achievements"`
		StartDate      string    `json:"start_date"`
		EndDate        string    `json:"end_date"`
		IsCurrent      *bool     `json:"is_current"`
	}

	var req experienceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	startDate, err := convert.ToDateRequired(req.StartDate)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid start date: "+err.Error())
		return
	}

	endDate, err := convert.ToDateOptional(req.EndDate)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid end date: "+err.Error())
		return
	}

	experience, err := cfg.db.CreateUserExperience(r.Context(), database.CreateUserExperienceParams{
		UserID:         userID,
		CompanyName:    req.CompanyName,
		CompanyUrl:     req.CompanyUrl,
		RoleTitle:      req.RoleTitle,
		ExpLocation:    req.ExpLocation,
		Industry:       req.Industry,
		EmploymentType: req.EmploymentType,
		Description:    req.Description,
		Achievements:   req.Achievements,
		StartDate:      startDate,
		EndDate:        endDate,
		IsCurrent:      req.IsCurrent,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "user not found")
			return
		}

		// check for postgres-specific errors
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				httpx.RespondError(w, http.StatusConflict, "you have already recorded this experience")
				return
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "user not found")
				return
			}
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to create exprience")
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, experience)
}

func (cfg *apiConfig) handlerGetExperiences(w http.ResponseWriter, r *http.Request) {
	// User helper function to get the UserID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	experience, err := cfg.db.GetExperiencesByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "experience not found")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "failed to retrieve experience")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, experience)
}

func (cfg *apiConfig) handlerDeleteExperience(w http.ResponseWriter, r *http.Request) {
	// User helper function to get the UserID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	// Get the experienceID from the url
	expIDStr := chi.URLParam(r, "id")
	experienceID, err := uuid.Parse(expIDStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid experience ID format")
		return
	}

	err = cfg.db.DeleteUserExperience(r.Context(), database.DeleteUserExperienceParams{
		ID:     experienceID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "experience not found")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "failed to delete experience")
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 — success, no body
}
