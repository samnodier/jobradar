package main

import (
	"testing"
	"time"
)

func TestPublicFeedsRequireOptInAndConservativeSchedule(t *testing.T) {
	for _, key := range []string{"JOBICY_ENABLED", "MYJOBMAG_ENABLED", "ADZUNA_APP_ID", "ADZUNA_APP_KEY", "GREENHOUSE_BOARDS"} {
		t.Setenv(key, "")
	}
	t.Setenv("REMOTEOK_ENABLED", "false")
	t.Setenv("SCRAPE_INTERVAL", "1h")
	cfg := &apiConfig{}
	list, err := cfg.configuredSources()
	if err != nil || len(list) != 0 {
		t.Fatalf("optional sources enabled by default: %v %v", list, err)
	}
	t.Setenv("JOBICY_ENABLED", "true")
	t.Setenv("MYJOBMAG_ENABLED", "true")
	list, err = cfg.configuredSources()
	if err != nil || len(list) != 2 {
		t.Fatalf("missing opt-in sources: %v %v", list, err)
	}
	for _, source := range list {
		if source.interval != 6*time.Hour {
			t.Errorf("%s interval %s", source.source.Name(), source.interval)
		}
	}
	t.Setenv("SCRAPE_INTERVAL", "12h")
	list, err = cfg.configuredSources()
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range list {
		if source.interval != 12*time.Hour {
			t.Errorf("%s ignored slower interval", source.source.Name())
		}
	}
}
