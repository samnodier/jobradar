package main

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/testutil"
)

func TestJobSourceFilterRejectsOversizedInput(t *testing.T) {
	cfg := &apiConfig{}
	res := httptest.NewRecorder()
	cfg.handlerJobsGet(res, httptest.NewRequest("GET", "/api/jobs?source="+strings.Repeat("a", 256), nil))
	if res.Code != 400 {
		t.Fatalf("got %d, want 400", res.Code)
	}
}

func TestJobSourceFilterBeforePagination(t *testing.T) {
	pool := testutil.Pool(t)
	ctx := context.Background()
	q := database.New(pool)
	marker := uuid.NewString()
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	for i, source := range []string{"remoteok", "adzuna", "remoteok"} {
		_, err := pool.Exec(ctx, `INSERT INTO jobs(id, external_id, job_source, title, company_name, source_url, created_at) VALUES($1,$2,$3,$4,'Example','https://example.test',$5::timestamptz)`, ids[i], ids[i].String(), source, marker, []string{"2026-01-03T00:00:00Z", "2026-01-02T00:00:00Z", "2026-01-01T00:00:00Z"}[i])
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM jobs WHERE id=ANY($1)", ids) })
	cfg := &apiConfig{db: q}
	for _, tc := range []struct {
		query string
		want  uuid.UUID
	}{
		{"source=ADZUNA&offset=0", ids[1]},
		{"source=remoteok&offset=1", ids[2]},
	} {
		res := httptest.NewRecorder()
		cfg.handlerJobsGet(res, httptest.NewRequest("GET", "/api/jobs?limit=1&q="+marker+"&"+tc.query, nil))
		if res.Code != 200 {
			t.Fatalf("request failed: %d %s", res.Code, res.Body.String())
		}
		var rows []database.GetJobsRow
		if err := json.Unmarshal(res.Body.Bytes(), &rows); err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 || rows[0].ID != tc.want {
			t.Fatalf("source filter was not applied before pagination: %+v", rows)
		}
	}
}
