package auth

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/samnodier/jobradar/internal/convert"
	"github.com/samnodier/jobradar/internal/httpx"
)

// UserResponse for the frontend communication
type UserResponse struct {
	ID                string `json:"id"`
	Username          string `json:"username"`
	Email             string `json:"email"`
	FullName          string `json:"full_name"`
	AvatarURL         string `json:"avatar_url"`
	Phone             string `json:"phone"`
	UserLocation      string `json:"user_location"`
	WebsiteURL        string `json:"website_url"`
	LinkedInURL       string `json:"linkedin_url"`
	GitHubURL         string `json:"github_url"`
	Headline          string `json:"headline"`
	UserSummary       string `json:"user_summary"`
	Availability      string `json:"availability"`
	MinSalary         int32  `json:"min_salary"`
	MaxSalary         int32  `json:"max_salary"`
	SalaryCurrency    string `json:"salary_currency"`
	YearsOfExperience int32  `json:"years_of_experience"`
	IsAdmin           bool   `json:"is_admin"`
	CreatedAt         string `json:"created_at"`
}

func (h *Handler) HandlerUserGet(w http.ResponseWriter, r *http.Request) {
	// 1. Extract session from the context (validated by @RequireAuth)
	session, ok := SessionFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "Session not found in context")
		return
	}

	// 2. Query the database using the userid stored in the session
	user, err := h.db.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusUnauthorized, "User no longer exists")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "Failed to fetch user profile")
		return
	}

	// 3. Map the database model to our UserResponse
	resp := UserResponse{
		ID:                user.ID.String(),
		Email:             user.Email,
		Username:          user.Username,
		FullName:          convert.FromNullString(user.FullName),
		AvatarURL:         convert.FromNullString(user.AvatarUrl),
		Phone:             convert.FromNullString(user.Phone),
		UserLocation:      convert.FromNullString(user.UserLocation),
		WebsiteURL:        convert.FromNullString(user.WebsiteUrl),
		LinkedInURL:       convert.FromNullString(user.LinkedinUrl),
		GitHubURL:         convert.FromNullString(user.GithubUrl),
		Headline:          convert.FromNullString(user.Headline),
		UserSummary:       convert.FromNullString(user.UserSummary),
		Availability:      convert.FromNullString(user.Availability),
		MinSalary:         convert.FromNullInt32(user.MinSalary),
		MaxSalary:         convert.FromNullInt32(user.MaxSalary),
		SalaryCurrency:    convert.FromNullString(user.SalaryCurrency),
		YearsOfExperience: convert.FromNullInt32(user.YearsOfExperience),
		IsAdmin:           user.IsAdmin != nil && *user.IsAdmin,
		CreatedAt:         convert.FromNullTime(user.CreatedAt),
	}

	httpx.RespondJSON(w, http.StatusOK, resp)
}
