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

type projectCreateRequest struct {
	Title         string  `json:"title"`
	RoleTitle     *string `json:"role_title"`
	Description   *string `json:"description"`
	Impact        *string `json:"impact"`
	ProjectURL    *string `json:"project_url"`
	RepositoryURL *string `json:"repository_url"`
	StartDate     string  `json:"start_date"`
	EndDate       *string `json:"end_date"`
	IsCurrent     *bool   `json:"is_current"`
	IsFeatured    *bool   `json:"is_featured"`
}

type projectUpdateRequest struct {
	Title         *string `json:"title"`
	RoleTitle     *string `json:"role_title"`
	Description   *string `json:"description"`
	Impact        *string `json:"impact"`
	ProjectURL    *string `json:"project_url"`
	RepositoryURL *string `json:"repository_url"`
	StartDate     *string `json:"start_date"`
	EndDate       *string `json:"end_date"`
	IsCurrent     *bool   `json:"is_current"`
	IsFeatured    *bool   `json:"is_featured"`
}

func (cfg *apiConfig) handlerCreateProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	var req projectCreateRequest
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

	project, err := cfg.db.CreateUserProject(r.Context(), database.CreateUserProjectParams{
		UserID:        userID,
		Title:         req.Title,
		RoleTitle:     req.RoleTitle,
		Description:   req.Description,
		Impact:        req.Impact,
		ProjectUrl:    req.ProjectURL,
		RepositoryUrl: req.RepositoryURL,
		StartDate:     startDate,
		EndDate:       endDate,
		IsCurrent:     req.IsCurrent,
		IsFeatured:    req.IsFeatured,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "project not found or unauthorized")
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				httpx.RespondError(w, http.StatusConflict, "a project with the details exist")
				return
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "user account not found")
				return
			}
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to create project")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, project)
}

func (cfg *apiConfig) handlerGetProjects(w http.ResponseWriter, r *http.Request) {
	// User helper function to get the UserID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	projects, err := cfg.db.GetProjectsByUserID(r.Context(), userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to retrieve projects")
		return
	}

	if projects == nil {
		projects = []database.UserProject{}
	}

	httpx.RespondJSON(w, http.StatusOK, projects)
}

func (cfg *apiConfig) handlerUpdateProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid project ID format")
		return
	}

	var req projectUpdateRequest
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

	if req.IsCurrent != nil && !*req.IsCurrent && (req.EndDate == nil || *req.EndDate == "") {
		httpx.RespondError(w, http.StatusBadRequest, "end_date is required and is_current is false")
		return
	}

	project, err := cfg.db.UpdateUserProject(r.Context(), database.UpdateUserProjectParams{
		ID:            projectID,
		UserID:        userID,
		Title:         req.Title,
		RoleTitle:     req.RoleTitle,
		Description:   req.Description,
		Impact:        req.Impact,
		ProjectUrl:    req.ProjectURL,
		RepositoryUrl: req.RepositoryURL,
		StartDate:     startDate,
		EndDate:       endDate,
		IsCurrent:     req.IsCurrent,
		IsFeatured:    req.IsFeatured,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "project not found or unauthorized")
			return
		}

		// check for postgres-specific errors
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				httpx.RespondError(w, http.StatusConflict, "a project with the details already exists")
				return
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "user account not found")
				return
			}
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to update project")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, project)
}

func (cfg *apiConfig) handlerDeleteProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid project ID format")
		return
	}

	err = cfg.db.DeleteUserProject(r.Context(), database.DeleteUserProjectParams{
		ID:     projectID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "project not found")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "failed to delete project")
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 — success, no body
}
