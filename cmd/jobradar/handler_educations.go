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

type educationCreateRequest struct {
	InstitutionName string  `json:"institution_name"`
	DegreeType      *string `json:"degree_type"`
	DegreeName      *string `json:"degree_name"`
	FieldOfStudy    *string `json:"field_of_study"`
	StartDate       string  `json:"start_date"`
	EndDate         *string `json:"end_date"`
	IsCurrent       *bool   `json:"is_current"`
	Description     *string `json:"description"`
}

type educationUpdateRequest struct {
	InstitutionName *string `json:"institution_name"`
	DegreeType      *string `json:"degree_type"`
	DegreeName      *string `json:"degree_name"`
	FieldOfStudy    *string `json:"field_of_study"`
	StartDate       *string `json:"start_date"`
	EndDate         *string `json:"end_date"`
	IsCurrent       *bool   `json:"is_current"`
	Description     *string `json:"description"`
}

func (cfg *apiConfig) handlerCreateEducation(w http.ResponseWriter, r *http.Request) {
	// User helper function to get the UserID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	var req educationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	startDate, err := convert.ToDateRequired(req.StartDate)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid start date: "+err.Error())
		return
	}

	var endDateStr string
	if req.EndDate != nil {
		endDateStr = *req.EndDate
	}
	endDate, err := convert.ToDateOptional(endDateStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid end date: "+err.Error())
		return
	}

	education, err := cfg.db.CreateUserEducation(r.Context(), database.CreateUserEducationParams{
		UserID:          userID,
		InstitutionName: req.InstitutionName,
		DegreeType:      req.DegreeType,
		DegreeName:      req.DegreeName,
		FieldOfStudy:    req.FieldOfStudy,
		StartDate:       startDate,
		EndDate:         endDate,
		IsCurrent:       req.IsCurrent,
		Description:     req.Description,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "education not found or unauthorized")
			return
		}

		// check for postgres-specific errors
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				httpx.RespondError(w, http.StatusConflict, "an education with the details already exists")
				return
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "user account not found")
				return
			}
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to create education")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, education)
}

func (cfg *apiConfig) handlerGetEducations(w http.ResponseWriter, r *http.Request) {
	// User helper function to get the UserID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	educations, err := cfg.db.GetEducationsByUserID(r.Context(), userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to retrieve educations")
		return
	}
	if educations == nil {
		educations = []database.GetEducationsByUserIDRow{}
	}

	httpx.RespondJSON(w, http.StatusOK, educations)
}

func (cfg *apiConfig) handlerUpdateEducation(w http.ResponseWriter, r *http.Request) {
	// User helper function to get the UserID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	// Get the education ID from the URL
	edIDStr := chi.URLParam(r, "id")
	educationID, err := uuid.Parse(edIDStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid education ID format")
		return
	}

	var req educationUpdateRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var startDateStr string
	if req.StartDate != nil {
		startDateStr = *req.StartDate
	}

	// If is_current switched off, end_date must be provided
	if req.IsCurrent != nil && !*req.IsCurrent && (req.EndDate == nil || *req.EndDate == "") {
		httpx.RespondError(w, http.StatusBadRequest, "end_date is required and is_current is false")
		return
	}

	startDate, err := convert.ToDateOptional(startDateStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid start date: "+err.Error())
		return
	}

	var endDateStr string
	if req.EndDate != nil {
		endDateStr = *req.EndDate
	}
	endDate, err := convert.ToDateOptional(endDateStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid end date: "+err.Error())
		return
	}

	education, err := cfg.db.UpdateUserEducation(r.Context(), database.UpdateUserEducationParams{
		ID:              educationID,
		UserID:          userID,
		InstitutionName: req.InstitutionName,
		DegreeType:      req.DegreeType,
		DegreeName:      req.DegreeName,
		FieldOfStudy:    req.FieldOfStudy,
		StartDate:       startDate,
		IsCurrent:       req.IsCurrent,
		EndDate:         endDate,
		Description:     req.Description,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "education not found or unauthorized")
			return
		}

		// check for postgres-specific errors
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				httpx.RespondError(w, http.StatusConflict, "an education with the details already exists")
				return
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "user account not found")
				return
			}
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to update education")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, education)
}

func (cfg *apiConfig) handlerDeleteEducation(w http.ResponseWriter, r *http.Request) {
	// User helper function to get the UserID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	// Get the educationID from the url
	expIDStr := chi.URLParam(r, "id")
	educationID, err := uuid.Parse(expIDStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid education ID format")
		return
	}

	err = cfg.db.DeleteUserEducation(r.Context(), database.DeleteUserEducationParams{
		ID:     educationID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "education not found")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "failed to delete education")
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 — success, no body
}
