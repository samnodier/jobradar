package auth

import (
	"net/http"
)

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		_ = h.deleteSession(r.Context(), cookie.Value)
	}

	h.clearSessionCookie(w)
	http.Redirect(w, r, h.cfg.AppBaseURL+"/login", http.StatusFound)
}
