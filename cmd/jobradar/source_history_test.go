package main

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/testutil"
)

func TestSourceDailyObservations(t *testing.T) {
	pool := testutil.Pool(t)
	ctx := context.Background()
	q := database.New(pool)
	name := "history-" + uuid.NewString()
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM service_health WHERE service_name=$1", name) })
	// The real scraper uses this upsert. Every invocation records one observation.
	for _, state := range []string{"healthy", "degraded", "healthy"} {
		secret := "provider-secret-must-not-leak"
		if _, err := q.UpdateServiceHealth(ctx, database.UpdateServiceHealthParams{ServiceName: name, ServiceStatus: state, LastError: &secret}); err != nil {
			t.Fatal(err)
		}
	}
	var successes, failures int
	if err := pool.QueryRow(ctx, "SELECT successes,failures FROM source_health_daily WHERE service_name=$1", name).Scan(&successes, &failures); err != nil {
		t.Fatal(err)
	}
	if successes != 2 || failures != 1 {
		t.Fatalf("counts=%d/%d", successes, failures)
	}
	// Updating descriptive metadata alone must not count as another fetch.
	if _, err := pool.Exec(ctx, "UPDATE service_health SET last_error=NULL WHERE service_name=$1", name); err != nil {
		t.Fatal(err)
	}
	rows, err := q.ListSourceHealthDaily(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range rows {
		if r.ServiceName == name {
			found = true
			if r.Successes != 2 || r.Failures != 1 || r.Day != time.Now().UTC().Format("2006-01-02") {
				t.Fatalf("bad daily observation: %+v", r)
			}
		}
	}
	if !found {
		t.Fatal("missing source history")
	}
	// History is bounded; the next real observation prunes dates outside 30 UTC days.
	if _, err := pool.Exec(ctx, "INSERT INTO source_health_daily(service_name,day,successes) VALUES($1,(NOW() AT TIME ZONE 'UTC')::date - 40,1)", name); err != nil {
		t.Fatal(err)
	}
	if _, err := q.UpdateServiceHealth(ctx, database.UpdateServiceHealthParams{ServiceName: name, ServiceStatus: "healthy"}); err != nil {
		t.Fatal(err)
	}
	var old int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM source_health_daily WHERE service_name=$1 AND day < (NOW() AT TIME ZONE 'UTC')::date - 29", name).Scan(&old); err != nil || old != 0 {
		t.Fatalf("old history %d %v", old, err)
	}
	if addr := os.Getenv("TEST_REDIS_ADDR"); addr != "" {
		rdb := redis.NewClient(&redis.Options{Addr: addr, DB: 2})
		defer rdb.Close()
		if _, err := pool.Exec(ctx, "UPDATE service_health SET last_error='provider-secret-must-not-leak' WHERE service_name=$1", name); err != nil {
			t.Fatal(err)
		}
		cfg := &apiConfig{db: q, pool: pool, rdb: rdb}
		res := httptest.NewRecorder()
		cfg.handlerSystemStatus(res, httptest.NewRequest("GET", "/api/system/status", nil))
		if res.Code != 200 {
			t.Fatalf("status=%d %s", res.Code, res.Body.String())
		}
		if strings.Contains(res.Body.String(), "provider-secret") {
			t.Fatal("raw provider error leaked")
		}
		var body struct {
			Sources []struct {
				Name    string `json:"name"`
				History []struct {
					Day       string `json:"day"`
					Successes int    `json:"successes"`
					Failures  int    `json:"failures"`
				} `json:"history"`
				LastChecked time.Time `json:"last_checked"`
			} `json:"sources"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		found = false
		for _, source := range body.Sources {
			if source.Name == name {
				found = true
				if len(source.History) != 1 || source.History[0].Successes != 3 || source.LastChecked.IsZero() {
					t.Fatalf("bad API history %+v", source)
				}
			}
		}
		if !found {
			t.Fatal("missing API source")
		}
	}
}
