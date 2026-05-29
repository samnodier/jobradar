package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
	"github.com/samnodier/jobradar/internal/stringutils"
)

type SkillPreference struct {
	SkillName       string  `json:"skill_name"`
	SkillCategory   *string `json:"skill_category"`
	Proficiency     *string `json:"proficiency"`
	YearsExperience *int32  `json:"years_experience"`
	IsFeatured      *bool   `json:"is_featured"`
	EndorsedByAi    *bool   `json:"endorsed_by_ai"`
	DisplayOrder    *int32  `json:"display_order"`
}

type DesiredLocationPreference struct {
	LocationName string `json:"location_name"`
	IsRemote     *bool  `json:"is_remote"`
	Priority     *int32 `json:"priority"`
}

type DesiredRolePreference struct {
	RoleTitle string `json:"role_title"`
	Priority  *int32 `json:"priority"`
}

type PreferencesResponse struct {
	MinSalary              *int32                      `json:"min_salary"`
	MaxSalary              *int32                      `json:"max_salary"`
	SalaryCurrency         *string                     `json:"salary_currency"`
	PreferredJobTypes      []string                    `json:"preferred_job_types"`
	PreferredIndustries    []string                    `json:"preferred_industries"`
	CompanyStagePreference []string                    `json:"company_stage_preference"`
	NotifyJobs             *bool                       `json:"notify_jobs"`
	Skills                 []SkillPreference           `json:"skills"`
	DesiredLocations       []DesiredLocationPreference `json:"desired_locations"`
	DesiredRoles           []DesiredRolePreference     `json:"desired_roles"`
}

type preferencesPatchRequest struct {
	MinSalary              *int32                       `json:"min_salary"`
	MaxSalary              *int32                       `json:"max_salary"`
	SalaryCurrency         *string                      `json:"salary_currency"`
	PreferredJobTypes      *[]string                    `json:"preferred_job_types"`
	PreferredIndustries    *[]string                    `json:"preferred_industries"`
	CompanyStagePreference *[]string                    `json:"company_stage_preference"`
	NotifyJobs             *bool                        `json:"notify_jobs"`
	Skills                 *[]SkillPreference           `json:"skills"`
	DesiredLocations       *[]DesiredLocationPreference `json:"desired_locations"`
	DesiredRoles           *[]DesiredRolePreference     `json:"desired_roles"`
}

// GET /api/profile/preferences
func (cfg *apiConfig) handlerGetPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	ctx := r.Context()

	// 1. Get the user for basic preferences
	user, err := cfg.db.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {

			httpx.RespondError(w, http.StatusNotFound, "user not found")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "failed to get the user details")
		return
	}

	// 2. Fetch the desire locations
	locations, err := cfg.db.GetUserDesiredLocations(ctx, userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to get desired locations")
		return
	}

	// 3. Fetch desired roles
	roles, err := cfg.db.GetUserDesiredRoles(ctx, userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to get desired roles")
		return
	}

	// 4. Fetch skills
	skills, err := cfg.db.GetUserSkillsByUserID(ctx, userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to get user skills")
		return
	}

	// Map db rows to responses
	respSkills := make([]SkillPreference, len(skills))
	for i, s := range skills {
		respSkills[i] = SkillPreference{
			SkillName:       s.SkillName,
			SkillCategory:   s.SkillCategory,
			Proficiency:     s.Proficiency,
			YearsExperience: s.YearsExperience,
			IsFeatured:      s.IsFeatured,
			EndorsedByAi:    s.EndorsedByAi,
			DisplayOrder:    s.DisplayOrder,
		}
	}

	respLocations := make([]DesiredLocationPreference, len(locations))
	for i, loc := range locations {
		respLocations[i] = DesiredLocationPreference{
			LocationName: loc.LocationName,
			IsRemote:     loc.IsRemote,
			Priority:     loc.Priority,
		}
	}

	respRoles := make([]DesiredRolePreference, len(roles))
	for i, role := range roles {
		respRoles[i] = DesiredRolePreference{
			RoleTitle: role.RoleTitle,
			Priority:  role.Priority,
		}
	}
	preferredJobTypes := user.PreferredJobTypes
	if preferredJobTypes == nil {
		preferredJobTypes = []string{}
	}
	preferredIndustries := user.PreferredIndustries
	if preferredIndustries == nil {
		preferredIndustries = []string{}
	}
	companyStagePreference := user.CompanyStagePreference
	if companyStagePreference == nil {
		companyStagePreference = []string{}
	}

	preferencesResponse := PreferencesResponse{
		MinSalary:              user.MinSalary,
		MaxSalary:              user.MaxSalary,
		SalaryCurrency:         user.SalaryCurrency,
		PreferredJobTypes:      preferredJobTypes,
		PreferredIndustries:    preferredIndustries,
		CompanyStagePreference: companyStagePreference,
		NotifyJobs:             user.NotifyJobs,
		Skills:                 respSkills,
		DesiredLocations:       respLocations,
		DesiredRoles:           respRoles,
	}

	httpx.RespondJSON(w, http.StatusOK, preferencesResponse)
}

// PATCH /api/profile/preferences
func (cfg *apiConfig) handlerUpdatePreferences(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	var req preferencesPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := r.Context()

	// begin transaction
	tx, err := cfg.pool.Begin(ctx)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Printf("unexpected transaction rollback error: %v", err)
		}
	}()

	// Create transaction-scoped queries
	qtx := cfg.db.WithTx(tx)

	// Fetch the current user data to server as fallback values
	user, err := qtx.GetUserByID(ctx, userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to fetch the user data")
		return
	}

	// merge and update the main preferences
	minSalary := user.MinSalary
	if req.MinSalary != nil {
		minSalary = req.MinSalary
	}
	maxSalary := user.MaxSalary
	if req.MaxSalary != nil {
		maxSalary = req.MaxSalary
	}
	salaryCurrency := user.SalaryCurrency
	if req.SalaryCurrency != nil {
		salaryCurrency = req.SalaryCurrency
	}
	preferredJobTypes := user.PreferredJobTypes
	if req.PreferredJobTypes != nil {
		preferredJobTypes = stringutils.SanitizeStringSlice(*req.PreferredJobTypes)
	}
	preferredIndustries := user.PreferredIndustries
	if req.PreferredIndustries != nil {
		preferredIndustries = stringutils.SanitizeStringSlice(*req.PreferredIndustries)
	}
	companyStagePreference := user.CompanyStagePreference
	if req.CompanyStagePreference != nil {
		companyStagePreference = stringutils.SanitizeStringSlice(*req.CompanyStagePreference)
	}
	notifyJobs := user.NotifyJobs
	if req.NotifyJobs != nil {
		notifyJobs = req.NotifyJobs
	}

	// validate salaries
	if minSalary != nil && *minSalary < 0 {
		httpx.RespondError(w, http.StatusBadRequest, "minimum salary cannot be negative")
		return
	}
	if maxSalary != nil && *maxSalary < 0 {
		httpx.RespondError(w, http.StatusBadRequest, "maximum salary cannot be negative")
		return
	}
	if minSalary != nil && maxSalary != nil && *maxSalary <
		*minSalary {
		httpx.RespondError(w, http.StatusBadRequest, "maximum salary cannot be less than minimum salary")
		return
	}
	_, err = qtx.UpdateUser(ctx, database.UpdateUserParams{
		ID:                     userID,
		Username:               user.Username,
		FullName:               user.FullName,
		Phone:                  user.Phone,
		UserLocation:           user.UserLocation,
		WebsiteUrl:             user.WebsiteUrl,
		LinkedinUrl:            user.LinkedinUrl,
		GithubUrl:              user.GithubUrl,
		Headline:               user.Headline,
		UserSummary:            user.UserSummary,
		Availability:           user.Availability,
		MinSalary:              minSalary,
		MaxSalary:              maxSalary,
		SalaryCurrency:         salaryCurrency,
		YearsOfExperience:      user.YearsOfExperience,
		PreferredJobTypes:      preferredJobTypes,
		PreferredIndustries:    preferredIndustries,
		CompanyStagePreference: companyStagePreference,
		NotifyJobs:             notifyJobs,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "user not found")
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			httpx.RespondError(w, http.StatusBadRequest, "username already taken")
			return
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	// wiping and re-inserting skills
	if req.Skills != nil {
		err = qtx.ClearUserSkills(ctx, userID)
		if err != nil {
			httpx.RespondError(w, http.StatusInternalServerError, "failed to clear skills")
			return
		}

		seenSkills := make(map[string]bool)
		var savedIdx int32 = 0

		for _, skillPref := range *req.Skills {
			trimmedName := strings.TrimSpace(skillPref.SkillName)
			if trimmedName == "" {
				httpx.RespondError(w, http.StatusBadRequest, "skill name cannot be empty")
				return
			}
			lowerName := strings.ToLower(trimmedName)
			if seenSkills[lowerName] {
				continue // Skip duplicate
			}
			seenSkills[lowerName] = true

			// 1. resolve or create the skill name to get the skill UUID
			skillID, skillErr := qtx.GetOrCreateSkill(ctx, skillPref.SkillName)
			if skillErr != nil {
				httpx.RespondError(w, http.StatusInternalServerError, "failed to resolve skill")
				return
			}

			// 2. map indices to display order and write association
			displayOrder := savedIdx
			_, err = qtx.AddSkillToUser(ctx, database.AddSkillToUserParams{
				UserID:          userID,
				SkillID:         skillID,
				Proficiency:     skillPref.Proficiency,
				YearsExperience: skillPref.YearsExperience,
				IsFeatured:      skillPref.IsFeatured,
				EndorsedByAi:    skillPref.EndorsedByAi,
				DisplayOrder:    &displayOrder,
			})
			if err != nil {
				httpx.RespondError(w, http.StatusInternalServerError, "failed to save user skill")
				return
			}
			savedIdx++
		}
	}

	// wiping and re-inserting desired locations
	if req.DesiredLocations != nil {
		err = qtx.ClearUserDesiredLocations(ctx, userID)
		if err != nil {
			httpx.RespondError(w, http.StatusInternalServerError, "failed to clear user desired locations")
			return
		}

		seenLocs := make(map[string]bool)
		for _, desiredLoc := range *req.DesiredLocations {
			trimmedLoc := strings.TrimSpace(desiredLoc.LocationName)
			if trimmedLoc == "" {
				httpx.RespondError(w, http.StatusBadRequest, "location name cannot be empty")
				return
			}
			lowerLoc := strings.ToLower(trimmedLoc)
			if seenLocs[lowerLoc] {
				continue
			}
			seenLocs[lowerLoc] = true

			_, err = qtx.AddDesiredLocation(ctx, database.AddDesiredLocationParams{
				UserID:       userID,
				LocationName: trimmedLoc,
				IsRemote:     desiredLoc.IsRemote,
				Priority:     desiredLoc.Priority,
			})
			if err != nil {
				httpx.RespondError(w, http.StatusInternalServerError, "failed to save user desired location")
				return
			}
		}
	}

	// Wiping and inserting desired roles
	if req.DesiredRoles != nil {
		err = qtx.ClearUserDesiredRoles(ctx, userID)
		if err != nil {
			httpx.RespondError(w, http.StatusInternalServerError, "failed to clear user desired roles")
			return
		}

		seenRoles := make(map[string]bool)
		for _, desiredRole := range *req.DesiredRoles {
			trimmedRole := strings.TrimSpace(desiredRole.RoleTitle)
			if trimmedRole == "" {
				httpx.RespondError(w, http.StatusBadRequest, "role title cannot be empty")
				return
			}
			lowerRole := strings.ToLower(trimmedRole)
			if seenRoles[lowerRole] {
				continue // skip duplicate
			}
			seenRoles[lowerRole] = true

			_, err = qtx.AddDesiredRole(ctx, database.AddDesiredRoleParams{
				UserID:    userID,
				RoleTitle: trimmedRole,
				Priority:  desiredRole.Priority,
			})
			if err != nil {
				httpx.RespondError(w, http.StatusInternalServerError, "failed to save user desired role")
				return
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		// log the actual detailed error to server stderr for debugging purposes
		log.Printf("Error committing preferences transation for the user %s: %v", userID, err)
		httpx.RespondError(w, http.StatusInternalServerError, "failed to save the changes")
		return
	}

	if req.Skills != nil || req.DesiredRoles != nil {
		err := cfg.enqueueMatchJobsForUser(cfg.rootCtx, userID)
		if err != nil {
			log.Printf("Error matching the jobs for the user %s: %v", userID, err)
		}
	}

	cfg.handlerGetPreferences(w, r)
}
