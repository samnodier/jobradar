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
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
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
		ID:        user.ID.String(),
		Email:     user.Email,
		Username:  convert.FromNullString(user.Username),
		Name:      convert.FromNullString(user.FullName),
		AvatarURL: convert.FromNullString(user.AvatarUrl),
	}

	httpx.RespondJSON(w, http.StatusOK, resp)
}
