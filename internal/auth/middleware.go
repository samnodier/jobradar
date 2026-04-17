package auth

import (
	"context"
	"net/http"

	"github.com/samnodier/jobradar/internal/httpx"
)

func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			httpx.RespondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		session, err := h.getSession(r.Context(), cookie.Value)
		if err != nil {
			httpx.RespondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		ctx := context.WithValue(r.Context(), sessionKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
