package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
	"net/http"
)

func (cfg *apiConfig) handlerNotifications(w http.ResponseWriter, r *http.Request) {
	id, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}
	rows, err := cfg.db.ListMatchNotifications(r.Context(), id)
	if err != nil {
		httpx.RespondError(w, 500, "could not load notifications")
		return
	}
	if rows == nil {
		rows = []database.ListMatchNotificationsRow{}
	}
	httpx.RespondJSON(w, 200, rows)
}
func (cfg *apiConfig) handlerNotificationRead(w http.ResponseWriter, r *http.Request) {
	user, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "notificationID"))
	if err != nil {
		httpx.RespondError(w, 400, "invalid notification")
		return
	}
	count, err := cfg.db.ReadMatchNotification(r.Context(), database.ReadMatchNotificationParams{UserID: user, ID: id})
	if err != nil {
		httpx.RespondError(w, 500, "could not mark read")
		return
	}
	if count == 0 {
		httpx.RespondError(w, 404, "notification not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
