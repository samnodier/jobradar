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

type experienceCreateRequest struct {
	UserID         uuid.UUID               `json:"user_id"`
	CompanyName    string                  `json:"company_name"`
	CompanyURL     *string                 `json:"company_url"`
	RoleTitle      string                  `json:"role_title"`
	ExpLocation    *string                 `json:"exp_location"`
	Industry       *string                 `json:"industry"`
	EmploymentType *string                 `json:"employment_type"`
	Description    *string                 `json:"description"`
	Achievements   []string                `json:"achievements"`
	StartDate      string                  `json:"start_date"`
	EndDate        *string                 `json:"end_date"`
	IsCurrent      *bool                   `json:"is_current"`
	Skills         []struct{ Name string } `json:"skills"`
}

type experienceUpdateRequest struct {
	UserID         uuid.UUID               `json:"user_id"`
	CompanyName    *string                 `json:"company_name"`
	CompanyURL     *string                 `json:"company_url"`
	RoleTitle      *string                 `json:"role_title"`
	ExpLocation    *string                 `json:"exp_location"`
	Industry       *string                 `json:"industry"`
	EmploymentType *string                 `json:"employment_type"`
	Description    *string                 `json:"description"`
	Achievements   []string                `json:"achievements"`
	StartDate      *string                 `json:"start_date"`
	EndDate        *string                 `json:"end_date"`
	IsCurrent      *bool                   `json:"is_current"`
	Skills         []struct{ Name string } `json:"skills"`
}

func (cfg *apiConfig) handlerCreateExperience(w http.ResponseWriter, r *http.Request) {
	// User helper function to get the UserID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
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

	var endDateStr string
	if req.EndDate != nil {
		endDateStr = *req.EndDate
	}
	endDate, err := convert.ToDateOptional(endDateStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid end date: "+err.Error())
		return
	}

	experience, err := cfg.db.CreateUserExperience(r.Context(), database.CreateUserExperienceParams{
		UserID:         userID,
		CompanyName:    req.CompanyName,
		CompanyUrl:     req.CompanyURL,
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
			httpx.RespondError(w, http.StatusNotFound, "experience not found or unauthorized")
			return
		}

		// check for postgres-specific errors
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				httpx.RespondError(w, http.StatusConflict, "an experience with the details already exists")
				return
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "user account not found")
				return
			}
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to create experience")
		return
	}

	// After creating the experience, attach the skills
	for _, skill := range req.Skills {
		skillID, err := cfg.db.GetOrCreateSkill(r.Context(), skill.Name)
		if err != nil {
			continue
		}
		_, err = cfg.db.AddSkillToExperience(r.Context(), database.AddSkillToExperienceParams{
			ExperienceID: experience.ID,
			SkillID:      skillID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httpx.RespondError(w, http.StatusNotFound, "experience or skill not found or unauthorized")
				return
			}

			// check for postgres-specific errors
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case "23505":
					continue // Skill already linked
				case "23503":
					httpx.RespondError(w, http.StatusBadRequest, "user account not found")
					return
				}
			}

			httpx.RespondError(w, http.StatusInternalServerError, "failed to update experience and skill union")
			return
		}
	}

	httpx.RespondJSON(w, http.StatusCreated, experience)
}

func (cfg *apiConfig) handlerUpdateExperience(w http.ResponseWriter, r *http.Request) {
	// User helper function to get the UserID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	// Get the experience ID from the URL
	expIDStr := chi.URLParam(r, "id")
	experienceID, err := uuid.Parse(expIDStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid experience ID format")
		return
	}

	var req experienceUpdateRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var startDateStr string
	if req.StartDate != nil {
		startDateStr = *req.StartDate
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

	experience, err := cfg.db.UpdateUserExperience(r.Context(), database.UpdateUserExperienceParams{
		ID:             experienceID,
		UserID:         userID,
		CompanyName:    req.CompanyName,
		CompanyUrl:     req.CompanyURL,
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
			httpx.RespondError(w, http.StatusNotFound, "experience not found or unauthorized")
			return
		}

		// check for postgres-specific errors
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				httpx.RespondError(w, http.StatusConflict, "an experience with the details already exists")
				return
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "user account not found")
				return
			}
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to update experience")
		return
	}

	// Sync skills: wipe old links, re-attach from payload
	if err := cfg.db.DeleteSkillsByExperienceID(r.Context(), experience.ID); err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to sync skills")
		return
	}
	// After creating the experience, attach the skills
	for _, skill := range req.Skills {
		skillID, err := cfg.db.GetOrCreateSkill(r.Context(), skill.Name)
		if err != nil {
			continue
		}
		_, err = cfg.db.AddSkillToExperience(r.Context(), database.AddSkillToExperienceParams{
			ExperienceID: experience.ID,
			SkillID:      skillID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httpx.RespondError(w, http.StatusNotFound, "experience or skill not found or unauthorized")
				return
			}

			// check for postgres-specific errors
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case "23505":
					httpx.RespondError(w, http.StatusConflict, "an experience and skill with the details already exists")
					return
				case "23503":
					httpx.RespondError(w, http.StatusBadRequest, "user account not found")
					return
				}
			}

			httpx.RespondError(w, http.StatusInternalServerError, "failed to update experience and skill union")
			return
		}
	}

	httpx.RespondJSON(w, http.StatusOK, experience)
}

func (cfg *apiConfig) handlerGetExperiences(w http.ResponseWriter, r *http.Request) {
	// User helper function to get the UserID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	experiences, err := cfg.db.GetExperiencesByUserID(r.Context(), userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to retrieve experience")
		return
	}
	if experiences == nil {
		experiences = []database.GetExperiencesByUserIDRow{}
	}

	httpx.RespondJSON(w, http.StatusOK, experiences)
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
