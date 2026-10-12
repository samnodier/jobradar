package main

import (
	"encoding/json"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
	"net/http"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

func (cfg *apiConfig) handlerMatchSettingsGet(w http.ResponseWriter, r *http.Request) {
	id, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}
	settings, err := cfg.db.GetMatchSettings(r.Context(), id)
	if err != nil {
		httpx.RespondError(w, 500, "could not load matching settings")
		return
	}
	httpx.RespondJSON(w, 200, settings)
}
func (cfg *apiConfig) handlerMatchSettingsPut(w http.ResponseWriter, r *http.Request) {
	id, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	var input database.UpsertMatchSettingsParams
	if json.NewDecoder(r.Body).Decode(&input) != nil {
		httpx.RespondError(w, 400, "invalid matching settings")
		return
	}
	input.SearchLocation = strings.TrimSpace(input.SearchLocation)
	if utf8.RuneCountInString(input.SearchLocation) > 100 {
		httpx.RespondError(w, 400, "search location must be at most 100 characters")
		return
	}
	input.UserID = id // Ownership always comes from the authenticated session.
	if input.MinScore < 0 || input.MinScore > 100 || input.DigestMaxJobs < 1 || input.DigestMaxJobs > 25 || !slices.Contains([]string{"any", "remote", "onsite"}, input.WorkMode) {
		httpx.RespondError(w, 400, "invalid threshold, work mode or digest size")
		return
	}
	if !slices.Contains([]string{"at", "au", "be", "br", "ca", "ch", "de", "es", "fr", "gb", "in", "it", "mx", "nl", "nz", "pl", "sg", "us", "za"}, input.SearchCountry) {
		httpx.RespondError(w, 400, "choose a supported Adzuna search country")
		return
	}
	if _, err := time.LoadLocation(input.DigestTimezone); err != nil {
		httpx.RespondError(w, 400, "invalid timezone")
		return
	}
	if len(input.DigestHours) < 1 || len(input.DigestHours) > 2 {
		httpx.RespondError(w, 400, "choose one or two delivery hours")
		return
	}
	slices.Sort(input.DigestHours)
	for i, h := range input.DigestHours {
		if h < 0 || h > 23 || (i > 0 && input.DigestHours[i-1] == h) {
			httpx.RespondError(w, 400, "delivery hours must be distinct values from 0 to 23")
			return
		}
	}
	if err := cfg.db.UpsertMatchSettings(r.Context(), input); err != nil {
		httpx.RespondError(w, 500, "could not save matching settings")
		return
	}
	cfg.handlerMatchSettingsGet(w, r)
}
