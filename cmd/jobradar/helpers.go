package main

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/samnodier/jobradar/internal/auth"
	"github.com/samnodier/jobradar/internal/httpx"
)

// getUserIDFromRequest extracts the authenticated user's ID from the request context.
// It writes the appropriate error response and returns false if extraction fails.
func getUserIDFromRequest(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "unauthorized: session not found in context")
		return uuid.UUID{}, false
	}
	return session.UserID, true
}
