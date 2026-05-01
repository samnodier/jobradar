package main

import (
	"net/http"

	"github.com/samnodier/jobradar/internal/auth"
	"github.com/samnodier/jobradar/internal/httpx"
)

func (cfg *apiConfig) HandleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	// 1. Extract session from the context (validated by @RequireAuth)
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "Session not found in context")
		return
	}

	// 2. Get the session cookie from the request
	cookie, err := r.Cookie("session_id")
	if err != nil {
		httpx.RespondError(w, http.StatusUnauthorized, "Session not found")
		return
	}

	// Delete the user from the database
	userID := session.UserID
	err = cfg.db.DeleteUserByID(r.Context(), userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	// Delete the Redis session
	_ = cfg.rdb.Del(r.Context(), "session:"+cookie.Value)

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cfg.IsProduction,
		SameSite: http.SameSiteLaxMode,
	})

	// Respond with success	
	httpx.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "account deleted successfully",
	})
}
