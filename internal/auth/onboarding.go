// Package auth...
package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/samnodier/jobradar/internal/convert"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/dbx"
	"github.com/samnodier/jobradar/internal/httpx"
	"github.com/samnodier/jobradar/internal/stringutils"
)

func (h *Handler) HandleOnboardingGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Read token
	token := r.URL.Query().Get("token")
	if token == "" {
		redirectWithError(w, r, "/login", "missing_token")
		return
	}

	// Fetch pending signup
	pending, err := h.getPendingSignup(ctx, token)
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, redis.Nil) {
		redirectWithError(w, r, "/login", "missing_token")
		return
	}
	if err != nil {
		redirectWithError(w, r, "/login", "server_error")
		return
	}

	// Parse request body
	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"github_id":          pending.GitHubID,
		"name":               pending.Name,
		"email":              pending.Email,
		"suggested_username": stringutils.GenerateUsername(pending.Email),
		"avatar_url":         pending.AvatarURL,
	})
}

func (h *Handler) HandleOnboardingComplete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get the token
	token := r.URL.Query().Get("token")
	if token == "" {
		redirectWithError(w, r, "/login", "missing_token")
		return
	}

	// Fetch pending signup
	pending, err := h.getPendingSignup(ctx, token)
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, redis.Nil) {
		redirectWithError(w, r, "/login", "invalid_or_expired_token")
		return
	}
	if err != nil {
		redirectWithError(w, r, "/login", "server_error")
		return
	}

	// Parse request body
	type request struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		redirectWithError(w, r, "/login", "invalid_request")
		return
	}

	if req.Email == "" || req.Name == "" || req.Username == "" {
		redirectWithError(w, r, "/login", "invalid_request")
		return
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		redirectWithError(w, r, "/login", "server_error")
		return
	}
	defer tx.Rollback(ctx)

	qtx := h.db.WithTx(tx)

	// Pre-validate the username
	if !stringutils.IsValidUsername(req.Username) {
		httpx.RespondError(w, http.StatusBadRequest, "Username must be 3-20 characters and alphanumeric")
		return
	}

	// Attempt to create the local user
	user, err := qtx.CreateUser(ctx, database.CreateUserParams{
		Email:     req.Email,
		Username:  convert.ToNullString(req.Username),
		FullName:  convert.ToNullString(req.Name),
		AvatarUrl: convert.ToNullString(pending.AvatarURL),
	})
	if err != nil {
		// Duplicate Error
		if dbx.IsUniqueViolation(err) {
			httpx.RespondError(w, http.StatusConflict, "That username is already taken")
			return
		}

		redirectWithError(w, r, "/login", "create_user_failed")
		return
	}

	// Create the AuthProvider account
	_, err = qtx.CreateUserAccount(ctx, database.CreateUserAccountParams{
		UserID:         user.ID,
		AuthProvider:   "github",
		AuthProviderID: strconv.FormatInt(pending.GitHubID, 10),
		AccessToken:    convert.ToNullString(pending.AccessToken),
	})
	if err != nil {
		redirectWithError(w, r, "/login", "create_account_failed")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		redirectWithError(w, r, "/login", "commit_failed")
		return
	}

	// Delete pending signup
	if err := h.deletePendingSignup(ctx, token); err != nil {
		redirectWithError(w, r, "/login", "cleanup_failed")
		return
	}

	sessionID, err := h.createSession(ctx, user.ID, user.Email)
	if err != nil {
		redirectWithError(w, r, "/login", "session_create_failed")
		return
	}

	h.setSessionCookie(w, sessionID)
	http.Redirect(w, r, h.cfg.AppBaseURL, http.StatusFound)
}
