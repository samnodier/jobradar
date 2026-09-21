package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samnodier/jobradar/internal/auth"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
)

func (cfg *apiConfig) handlerReadTokens(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		tokens, err := cfg.db.ListReadTokens(r.Context(), userID)
		if err != nil {
			httpx.RespondError(w, 500, "could not load access tokens")
			return
		}
		if tokens == nil {
			tokens = []database.ListReadTokensRow{}
		}
		httpx.RespondJSON(w, 200, tokens)
	case http.MethodPost:
		var input struct {
			Name        string `json:"name"`
			ExpiresDays int    `json:"expires_days"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		if json.NewDecoder(r.Body).Decode(&input) != nil {
			httpx.RespondError(w, 400, "invalid token request")
			return
		}
		input.Name = strings.TrimSpace(input.Name)
		if input.ExpiresDays == 0 {
			input.ExpiresDays = 90
		}
		if len(input.Name) < 1 || len(input.Name) > 80 || input.ExpiresDays < 1 || input.ExpiresDays > 365 {
			httpx.RespondError(w, 400, "name must be 1–80 bytes and expires_days 1–365")
			return
		}
		token, hash, err := auth.NewReadToken()
		if err != nil {
			httpx.RespondError(w, 500, "could not create access token")
			return
		}
		row, err := cfg.db.CreateReadToken(r.Context(), database.CreateReadTokenParams{UserID: userID, Name: input.Name, TokenHash: hash, ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Duration(input.ExpiresDays) * 24 * time.Hour), Valid: true}})
		if err != nil {
			httpx.RespondError(w, 500, "could not save access token")
			return
		}
		httpx.RespondJSON(w, 201, map[string]any{"token": token, "access": row, "scope": "jobs and your AI context (read only)"})
	case http.MethodDelete:
		id, err := uuid.Parse(chi.URLParam(r, "tokenID"))
		if err != nil {
			httpx.RespondError(w, 400, "invalid token ID")
			return
		}
		n, err := cfg.db.RevokeReadToken(r.Context(), database.RevokeReadTokenParams{ID: id, UserID: userID})
		if err != nil {
			httpx.RespondError(w, 500, "could not revoke token")
			return
		}
		if n == 0 {
			httpx.RespondError(w, 404, "active token not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
