package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/samnodier/jobradar/internal/httpx"
)

// NewReadToken returns a high-entropy bearer secret and its storage hash.
func NewReadToken() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token := "jr_" + hex.EncodeToString(b)
	return token, HashReadToken(token), nil
}
func HashReadToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

// RequireReadAccess belongs ONLY on explicitly registered read-only export routes.
// Tokens never authenticate mutations or browser-session management endpoints.
func (h *Handler) RequireReadAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			httpx.RespondError(w, 405, "read-only access")
			return
		}
		header := r.Header.Get("Authorization")
		if header == "" {
			h.RequireAuth(next).ServeHTTP(w, r)
			return
		}
		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || len(parts[1]) != 67 || !strings.HasPrefix(parts[1], "jr_") {
			httpx.RespondError(w, 401, "invalid access token")
			return
		}
		user, err := h.db.AuthenticateReadToken(r.Context(), HashReadToken(parts[1]))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httpx.RespondError(w, 401, "invalid or expired access token")
			} else {
				httpx.RespondError(w, 503, "authentication unavailable")
			}
			return
		}
		if h.cfg.PersonalEmail != "" && !strings.EqualFold(user.Email, h.cfg.PersonalEmail) {
			httpx.RespondError(w, 403, "this is a personal Jobradar instance")
			return
		}
		ctx := context.WithValue(r.Context(), sessionKey, &Session{UserID: user.ID, Email: user.Email})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
