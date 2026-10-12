package main

import (
	"context"
	"fmt"
	"github.com/samnodier/jobradar/internal/sources"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type scheduledSource struct {
	source   sources.Source
	interval time.Duration
}

func (cfg *apiConfig) configuredSources() ([]scheduledSource, error) {
	interval := 6 * time.Hour
	if s := os.Getenv("SCRAPE_INTERVAL"); s != "" {
		v, e := time.ParseDuration(s)
		if e != nil || v < time.Hour {
			return nil, fmt.Errorf("SCRAPE_INTERVAL must be at least 1h")
		}
		interval = v
	}
	list := []scheduledSource{}
	// Optional public feeds run at most four times daily, regardless of other sources.
	feedInterval := max(interval, 6*time.Hour)
	if os.Getenv("JOBICY_ENABLED") == "true" {
		list = append(list, scheduledSource{&sources.Jobicy{}, feedInterval})
	}
	if os.Getenv("MYJOBMAG_ENABLED") == "true" {
		list = append(list, scheduledSource{&sources.MyJobMag{}, feedInterval})
	}
	if os.Getenv("MYJOBMAG_ZA_ENABLED") == "true" {
		list = append(list, scheduledSource{&sources.MyJobMag{Country: "za"}, feedInterval})
	}
	if os.Getenv("REMOTEOK_ENABLED") != "false" {
		list = append(list, scheduledSource{&sources.RemoteOK{}, interval})
	}
	appID, key := os.Getenv("ADZUNA_APP_ID"), os.Getenv("ADZUNA_APP_KEY")
	if appID != "" || key != "" {
		pages := 1
		if s := os.Getenv("ADZUNA_PAGES"); s != "" {
			var e error
			pages, e = strconv.Atoi(s)
			if e != nil {
				return nil, fmt.Errorf("ADZUNA_PAGES must be an integer")
			}
		}
		if appID == "" || key == "" || pages < 1 || pages > 3 {
			return nil, fmt.Errorf("Adzuna requires both credentials and 1–3 pages")
		}
		list = append(list, scheduledSource{&profileAdzuna{cfg: cfg, appID: appID, key: key, pages: pages}, 12 * time.Hour})
	}
	for _, board := range splitCSV(os.Getenv("GREENHOUSE_BOARDS")) {
		if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(board) {
			return nil, fmt.Errorf("invalid Greenhouse board slug")
		}
		list = append(list, scheduledSource{&sources.Greenhouse{Board: board}, interval})
	}
	return list, nil
}
func splitCSV(s string) []string {
	out := []string{}
	for _, v := range strings.Split(s, ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// Read profile values on each scheduled fetch, so edits need no server restart.
type profileAdzuna struct {
	cfg        *apiConfig
	appID, key string
	pages      int
}

func (a *profileAdzuna) Name() string { return "adzuna" }
func (a *profileAdzuna) Fetch(ctx context.Context) ([]sources.Job, error) {
	users, err := a.cfg.db.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	owner := os.Getenv("PERSONAL_EMAIL")
	selected := -1
	for i, id := range users {
		u, err := a.cfg.db.GetUserByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if owner != "" && strings.EqualFold(u.Email, owner) {
			selected = i
			break
		}
	}
	if owner == "" && len(users) == 1 {
		selected = 0
	}
	if selected < 0 {
		return nil, fmt.Errorf("Adzuna: sign in and configure PERSONAL_EMAIL when more than one account exists")
	}
	user := users[selected]
	roles, err := a.cfg.db.GetUserDesiredRoles(ctx, user)
	if err != nil {
		return nil, err
	}
	queries := []string{}
	for _, r := range roles {
		if len(queries) < 5 {
			queries = append(queries, r.RoleTitle)
		}
	}
	if len(queries) == 0 {
		return nil, fmt.Errorf("Adzuna: add desired roles in Profile Preferences")
	}
	settings, err := a.cfg.db.GetMatchSettings(ctx, user)
	if err != nil {
		return nil, err
	}
	// A desired location belongs to MATCHING, not the selected provider market.
	// Nairobi passed to the US endpoint silently produced empty searches. Use a
	// separate explicit market location; blank means search the selected country.
	location := strings.TrimSpace(settings.SearchLocation)
	source := sources.Adzuna{AppID: a.appID, AppKey: a.key, Country: settings.SearchCountry, Queries: queries, Pages: a.pages, Location: location}
	return source.Fetch(ctx)
}

// A read-only probe can validate integrations without starting collection or email.
func checkPublicFeeds(ctx context.Context) error {
	for _, source := range []sources.Source{&sources.Jobicy{}, &sources.MyJobMag{}, &sources.MyJobMag{Country: "za"}} {
		callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		jobs, err := source.Fetch(callCtx)
		cancel()
		if err != nil {
			return fmt.Errorf("%s: %w", source.Name(), err)
		}
		fmt.Printf("%s: %d normalized jobs (read-only, none saved)\n", source.Name(), len(jobs))
	}
	return nil
}
