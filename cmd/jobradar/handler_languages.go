package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
)

type languageCreateRequest struct {
	UserLanguage string  `json:"user_language"`
	Proficiency  *string `json:"proficiency"`
}

type languageUpdateRequest struct {
	UserLanguage string  `json:"user_language"`
	Proficiency  *string `json:"proficiency"`
}

func (cfg *apiConfig) handlerCreateLanguage(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	var req languageCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	language, err := cfg.db.CreateUserLanguage(r.Context(), database.CreateUserLanguageParams{
		UserID:       userID,
		UserLanguage: req.UserLanguage,
		Proficiency:  req.Proficiency,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "langauge not found or unauthorized")
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				httpx.RespondError(w, http.StatusConflict, "a language with the details exist")
				return
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "user account not found")
				return
			}
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to create language")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, language)
}

func (cfg *apiConfig) handlerGetLanguages(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	languages, err := cfg.db.GetLanguagesByUserID(r.Context(), userID)
	if err != nil {

		httpx.RespondError(w, http.StatusInternalServerError, "failed to retrieve langauges")
		return
	}

	if languages == nil {
		languages = []database.UserLanguage{}
	}

	httpx.RespondJSON(w, http.StatusOK, languages)
}

func (cfg *apiConfig) handlerUpdateLanguage(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	var req languageUpdateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	language, err := cfg.db.UpdateUserLanguage(r.Context(), database.UpdateUserLanguageParams{
		UserID:       userID,
		UserLanguage: req.UserLanguage,
		Proficiency:  req.Proficiency,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "language not found or unauthorized")
			return
		}

		// check for postgres-specific errors
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				httpx.RespondError(w, http.StatusConflict, "a language with the details already exists")
				return
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "user account not found")
				return
			}
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to update language")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, language)
}

func (cfg *apiConfig) handlerDeleteLanguage(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	userLanguage := chi.URLParam(r, "id")

	err := cfg.db.DeleteUserLanguage(r.Context(), database.DeleteUserLanguageParams{
		UserID:       userID,
		UserLanguage: userLanguage,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "language not found")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "failed to delete language")
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 — success, no body
}
