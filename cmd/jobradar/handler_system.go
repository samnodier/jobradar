package main

import (
	"context"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (cfg *apiConfig) handlerDigestStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}
	user, err := cfg.db.GetUserByID(r.Context(), id)
	if err != nil {
		httpx.RespondError(w, 500, "could not load user")
		return
	}
	settings, err := cfg.db.GetMatchSettings(r.Context(), id)
	if err != nil {
		httpx.RespondError(w, 500, "could not load matching settings")
		return
	}
	enabled := cfg.digest.Config.Enabled && settings.EmailEnabled && user.NotifyJobs != nil && *user.NotifyJobs && strings.EqualFold(user.Email, cfg.digest.Config.Recipient)
	history, err := cfg.db.ListDigestHistory(r.Context(), id)
	if err != nil {
		httpx.RespondError(w, 500, "could not load digest history")
		return
	}
	if history == nil {
		history = []database.ListDigestHistoryRow{}
	}
	httpx.RespondJSON(w, 200, map[string]any{"enabled": enabled, "hours": settings.DigestHours, "timezone": settings.DigestTimezone, "min_score": settings.MinScore, "history": history})
}
func (cfg *apiConfig) handlerDigestPreview(w http.ResponseWriter, r *http.Request) {
	id, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}
	user, err := cfg.db.GetUserByID(r.Context(), id)
	if err != nil || !strings.EqualFold(user.Email, cfg.digest.Config.Recipient) {
		httpx.RespondError(w, 403, "personal digest is not configured for this account")
		return
	}
	html, err := cfg.digest.Preview(r.Context())
	if err != nil {
		httpx.RespondError(w, 500, "could not build digest preview")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; img-src https:; base-uri 'none'; frame-ancestors 'self'")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(html))
}
func (cfg *apiConfig) handlerSystemStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	dbOK := cfg.pool.Ping(ctx) == nil
	redisOK := cfg.rdb.Ping(ctx).Err() == nil
	rows, err := cfg.db.ListServiceHealth(ctx)
	if err != nil {
		httpx.RespondError(w, 503, "database unavailable")
		return
	}
	if rows == nil {
		rows = []database.ServiceHealth{}
	}
	var pending int
	_ = cfg.pool.QueryRow(ctx, "SELECT count(*) FROM pending_matches").Scan(&pending)
	// No raw provider errors or credentials are exposed by the health endpoint.
	type sourceStatus struct {
		Name        string     `json:"name"`
		Status      string     `json:"status"`
		LastSuccess *time.Time `json:"last_success"`
	}
	sources := []sourceStatus{}
	for _, row := range rows {
		sources = append(sources, sourceStatus{row.ServiceName, row.ServiceStatus, row.LastSuccessAt})
	}
	httpx.RespondJSON(w, 200, map[string]any{"database": dbOK, "redis": redisOK, "pending_matches": pending, "sources": sources})
}
func spaHandler(directory string) http.HandlerFunc {
	if directory == "" {
		directory = "frontend/dist"
	}
	files := http.FileServer(http.Dir(directory))
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/auth/") {
			http.NotFound(w, r)
			return
		}
		name := filepath.Join(directory, filepath.FromSlash(strings.TrimPrefix(filepath.Clean("/"+r.URL.Path), "/")))
		if info, err := os.Stat(name); err == nil && !info.IsDir() {
			if strings.HasPrefix(r.URL.Path, "/assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			files.ServeHTTP(w, r)
			return
		}
		if strings.Contains(filepath.Base(r.URL.Path), ".") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, filepath.Join(directory, "index.html"))
	}
}
