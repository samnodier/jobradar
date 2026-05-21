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

type certificationCreateRequest struct {
	CertificationName   string  `json:"certification_name"`
	IssuingOrganization string  `json:"issuing_organization"`
	IssueDate           *string `json:"issue_date"`
	ExpirationDate      *string `json:"expiration_date"`
	DoesNotExpire       *bool   `json:"does_not_expire"`
	CredentialID        *string `json:"credential_id"`
	CredentialURL       *string `json:"credential_url"`
	IsInProgress        *bool   `json:"is_in_progress"`
}

type certificationUpdateRequest struct {
	CertificationName   *string `json:"certification_name"`
	IssuingOrganization *string `json:"issuing_organization"`
	IssueDate           *string `json:"issue_date"`
	ExpirationDate      *string `json:"expiration_date"`
	DoesNotExpire       *bool   `json:"does_not_expire"`
	CredentialID        *string `json:"credential_id"`
	CredentialURL       *string `json:"credential_url"`
	IsInProgress        *bool   `json:"is_in_progress"`
}

type languageCreateRequest struct {
	UserLanguage string  `json:"user_language"`
	Proficiency  *string `json:"proficiency"`
}

type languageUpdateRequest struct {
	UserLanguage string  `json:"user_language"`
	Proficiency  *string `json:"proficiency"`
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

func (cfg *apiConfig) handlerCreateCertification(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	var req certificationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var issueDateStr string
	if req.IssueDate != nil {
		issueDateStr = *req.IssueDate
	}
	issueDate, err := convert.ToDateOptional(issueDateStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid issuing date: "+err.Error())
		return
	}

	var expirationDateStr string
	if req.ExpirationDate != nil {
		expirationDateStr = *req.ExpirationDate
	}
	expirationDate, err := convert.ToDateOptional(expirationDateStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid expiration date: "+err.Error())
		return
	}

	certification, err := cfg.db.CreateUserCertification(r.Context(), database.CreateUserCertificationParams{
		UserID:              userID,
		CertificationName:   req.CertificationName,
		IssuingOrganization: req.IssuingOrganization,
		IssueDate:           issueDate,
		ExpirationDate:      expirationDate,
		DoesNotExpire:       req.DoesNotExpire,
		CredentialID:        req.CredentialID,
		CredentialUrl:       req.CredentialURL,
		IsInProgress:        req.IsInProgress,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "certification not found or unauthorized")
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				httpx.RespondError(w, http.StatusConflict, "a certification with the details exist")
				return
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "user account not found")
				return
			}
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to create certification")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, certification)
}

func (cfg *apiConfig) handlerGetCertifications(w http.ResponseWriter, r *http.Request) {
	// User helper function to get the UserID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	certifications, err := cfg.db.GetCertificationsByUserID(r.Context(), userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to retrieve certifications")
		return
	}

	if certifications == nil {
		certifications = []database.GetCertificationsByUserIDRow{}
	}

	httpx.RespondJSON(w, http.StatusOK, certifications)
}

func (cfg *apiConfig) handlerUpdateCertification(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	certificationIDStr := chi.URLParam(r, "id")
	certificationID, err := uuid.Parse(certificationIDStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid certification ID format")
		return
	}

	var req certificationUpdateRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var issueDateStr string
	if req.IssueDate != nil {
		issueDateStr = *req.IssueDate
	}
	issueDate, err := convert.ToDateOptional(issueDateStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid issuing date: "+err.Error())
		return
	}

	var expirationDateStr string
	if req.ExpirationDate != nil {
		expirationDateStr = *req.ExpirationDate
	}
	expirationDate, err := convert.ToDateOptional(expirationDateStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid expiration date: "+err.Error())
		return
	}

	if req.DoesNotExpire != nil && !*req.DoesNotExpire && (req.ExpirationDate == nil || *req.ExpirationDate == "") {
		httpx.RespondError(w, http.StatusBadRequest, "expiration_date is required and does_not_expire is false")
		return
	}

	certification, err := cfg.db.UpdateUserCertification(r.Context(), database.UpdateUserCertificationParams{
		ID:                  certificationID,
		UserID:              userID,
		CertificationName:   req.CertificationName,
		IssuingOrganization: req.IssuingOrganization,
		IssueDate:           issueDate,
		ExpirationDate:      expirationDate,
		DoesNotExpire:       req.DoesNotExpire,
		CredentialID:        req.CredentialID,
		CredentialUrl:       req.CredentialURL,
		IsInProgress:        req.IsInProgress,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "certification not found or unauthorized")
			return
		}

		// check for postgres-specific errors
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				httpx.RespondError(w, http.StatusConflict, "a certification with the details already exists")
				return
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "user account not found")
				return
			}
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to update certification")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, certification)
}

func (cfg *apiConfig) handlerDeleteCertification(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	certificationIDStr := chi.URLParam(r, "id")
	certificationID, err := uuid.Parse(certificationIDStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid certification ID format")
		return
	}

	err = cfg.db.DeleteUserCertification(r.Context(), database.DeleteUserCertificationParams{
		ID:     certificationID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "certification not found")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "failed to delete certification")
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 — success, no body
}

func (cfg *apiConfig) handlerCreateLanguage(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	var req languageCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	language, err := cfg.db.CreateUserLanguage(r.Context(), database.CreateUserLanguageParams{
		UserID:       userID,
		UserLanguage: req.UserLanguage,
		Proficiency:  req.Proficiency,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "langauge not found or unauthorized")
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				httpx.RespondError(w, http.StatusConflict, "a language with the details exist")
				return
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "user account not found")
				return
			}
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to create language")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, language)
}

func (cfg *apiConfig) handlerGetLanguages(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	languages, err := cfg.db.GetLanguagesByUserID(r.Context(), userID)
	if err != nil {

		httpx.RespondError(w, http.StatusInternalServerError, "failed to retrieve langauges")
		return
	}

	if languages == nil {
		languages = []database.UserLanguage{}
	}

	httpx.RespondJSON(w, http.StatusOK, languages)
}

func (cfg *apiConfig) handlerUpdateLanguage(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	var req languageUpdateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	language, err := cfg.db.UpdateUserLanguage(r.Context(), database.UpdateUserLanguageParams{
		UserID:       userID,
		UserLanguage: req.UserLanguage,
		Proficiency:  req.Proficiency,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "language not found or unauthorized")
			return
		}

		// check for postgres-specific errors
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				httpx.RespondError(w, http.StatusConflict, "a language with the details already exists")
				return
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "user account not found")
				return
			}
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to update language")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, language)
}

func (cfg *apiConfig) handlerDeleteLanguage(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	userLanguage := chi.URLParam(r, "id")

	err := cfg.db.DeleteUserLanguage(r.Context(), database.DeleteUserLanguageParams{
		UserID:       userID,
		UserLanguage: userLanguage,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "language not found")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "failed to delete language")
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 — success, no body
}
