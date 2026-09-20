package main

import (
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

func configuredSources() ([]scheduledSource, error) {
	interval := 6 * time.Hour
	if s := os.Getenv("SCRAPE_INTERVAL"); s != "" {
		v, e := time.ParseDuration(s)
		if e != nil || v < time.Hour {
			return nil, fmt.Errorf("SCRAPE_INTERVAL must be at least 1h")
		}
		interval = v
	}
	list := []scheduledSource{}
	if os.Getenv("REMOTEOK_ENABLED") != "false" {
		list = append(list, scheduledSource{&sources.RemoteOK{}, interval})
	}
	appID, key := os.Getenv("ADZUNA_APP_ID"), os.Getenv("ADZUNA_APP_KEY")
	if appID != "" || key != "" {
		country := os.Getenv("ADZUNA_COUNTRY")
		queries := splitCSV(os.Getenv("ADZUNA_QUERIES"))
		pages := 1
		if s := os.Getenv("ADZUNA_PAGES"); s != "" {
			var e error
			pages, e = strconv.Atoi(s)
			if e != nil {
				return nil, fmt.Errorf("ADZUNA_PAGES must be an integer")
			}
		}
		if appID == "" || key == "" || !regexp.MustCompile(`^[a-z]{2}$`).MatchString(country) || len(queries) == 0 || len(queries) > 5 || pages < 1 || pages > 3 {
			return nil, fmt.Errorf("Adzuna requires both credentials, a two-letter country, 1–5 queries, and 1–3 pages")
		}
		list = append(list, scheduledSource{&sources.Adzuna{AppID: appID, AppKey: key, Country: country, Queries: queries, Pages: pages, Location: os.Getenv("ADZUNA_LOCATION")}, 12 * time.Hour})
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
