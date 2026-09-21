package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
)

type exportCursor struct {
	Score int32     `json:"score"`
	ID    uuid.UUID `json:"id"`
}

func parseExportQuery(v url.Values, timezone string, threshold int32) (database.ExportJobsParams, error) {
	p := database.ExportJobsParams{MinScore: threshold, EligibleOnly: true, DateField: "added", BeforeScore: 101, PageLimit: 51}
	allowed := map[string]bool{}
	for _, k := range []string{"min_score", "limit", "date", "timezone", "date_field", "source", "location", "eligible_only", "exclude_applied", "remote_only", "include_expired", "cursor"} {
		allowed[k] = true
	}
	for k, values := range v {
		if !allowed[k] || len(values) != 1 {
			return p, fmt.Errorf("unknown or repeated parameter: %s", k)
		}
	}
	if s := v.Get("min_score"); s != "" {
		n, e := strconv.Atoi(s)
		if e != nil || n < 0 || n > 100 {
			return p, fmt.Errorf("min_score must be 0–100")
		}
		p.MinScore = int32(n)
	}
	if s := v.Get("limit"); s != "" {
		n, e := strconv.Atoi(s)
		if e != nil || n < 1 || n > 100 {
			return p, fmt.Errorf("limit must be 1–100")
		}
		p.PageLimit = int32(n + 1)
	}
	for key, dst := range map[string]*bool{"eligible_only": &p.EligibleOnly, "exclude_applied": &p.ExcludeApplied, "remote_only": &p.RemoteOnly, "include_expired": &p.IncludeExpired} {
		if s := v.Get(key); s != "" {
			b, e := strconv.ParseBool(s)
			if e != nil {
				return p, fmt.Errorf("%s must be true or false", key)
			}
			*dst = b
		}
	}
	p.Source = v.Get("source")
	p.Location = v.Get("location")
	if len(p.Source) > 100 || len(p.Location) > 200 {
		return p, fmt.Errorf("source/location is too long")
	}
	if s := v.Get("date_field"); s != "" {
		p.DateField = s
	}
	if p.DateField != "added" && p.DateField != "posted" && p.DateField != "matched" {
		return p, fmt.Errorf("date_field must be added, posted or matched")
	}
	if s := v.Get("timezone"); s != "" {
		timezone = s
	}
	zone, e := time.LoadLocation(timezone)
	if e != nil {
		return p, fmt.Errorf("invalid timezone")
	}
	if s := v.Get("date"); s != "" {
		if s == "today" {
			s = time.Now().In(zone).Format("2006-01-02")
		}
		start, e := time.ParseInLocation("2006-01-02", s, zone)
		if e != nil {
			return p, fmt.Errorf("date must be today or YYYY-MM-DD")
		}
		end := start.AddDate(0, 0, 1)
		p.StartAt = &start
		p.EndAt = &end
	}
	if s := v.Get("cursor"); s != "" {
		if len(s) > 256 {
			return p, fmt.Errorf("invalid cursor")
		}
		b, e := base64.RawURLEncoding.DecodeString(s)
		var c exportCursor
		if e != nil || json.Unmarshal(b, &c) != nil || c.Score < 0 || c.Score > 100 || c.ID == uuid.Nil {
			return p, fmt.Errorf("invalid cursor")
		}
		p.BeforeScore = c.Score
		p.BeforeID = c.ID
	}
	return p, nil
}

func (cfg *apiConfig) handlerExportJobs(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}
	settings, err := cfg.db.GetMatchSettings(r.Context(), userID)
	if err != nil {
		httpx.RespondError(w, 500, "could not load profile settings")
		return
	}
	params, err := parseExportQuery(r.URL.Query(), settings.DigestTimezone, settings.MinScore)
	if err != nil {
		httpx.RespondError(w, 400, err.Error())
		return
	}
	params.UserID = userID
	jobs, err := cfg.db.ExportJobs(r.Context(), params)
	if err != nil {
		httpx.RespondError(w, 500, "could not export jobs")
		return
	}
	if jobs == nil {
		jobs = []database.ExportJobsRow{}
	}
	next := ""
	limit := int(params.PageLimit) - 1
	if len(jobs) > limit {
		jobs = jobs[:limit]
		last := jobs[len(jobs)-1]
		b, _ := json.Marshal(exportCursor{Score: *last.MatchScore, ID: last.ID})
		next = base64.RawURLEncoding.EncodeToString(b)
	}
	httpx.RespondJSON(w, 200, map[string]any{"jobs": jobs, "next_cursor": next, "min_score": params.MinScore, "date_field": params.DateField, "start_at": params.StartAt, "end_at": params.EndAt, "eligible_only": params.EligibleOnly})
}
