package main

import (
	"context"
	"encoding/json"
	"fmt"
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

func TestJobPageBoundaries(t *testing.T) {
	pool := testutil.Pool(t)
	ctx := context.Background()
	marker := uuid.NewString()
	_, err := pool.Exec(ctx, `INSERT INTO jobs(external_id,job_source,title,company_name,source_url) SELECT $1 || '-' || n,'remoteok',$1,'Example','https://example.test' FROM generate_series(1,101) n`, marker)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM jobs WHERE title=$1", marker) })
	cfg := &apiConfig{db: database.New(pool)}
	for _, tc := range []struct {
		limit, offset, want int
		more                string
	}{
		{50, 0, 50, "true"}, {50, 50, 50, "true"}, {50, 100, 1, "false"}, {100, 0, 100, "true"}, {200, 0, 101, "false"}, {500, 0, 101, "false"}, {50, 101, 0, "false"},
	} {
		res := httptest.NewRecorder()
		cfg.handlerJobsGet(res, httptest.NewRequest("GET", fmt.Sprintf("/api/jobs?source=remoteok&q=%s&limit=%d&offset=%d", marker, tc.limit, tc.offset), nil))
		var rows []database.GetJobsRow
		if res.Code != 200 {
			t.Fatalf("status %d: %s", res.Code, res.Body.String())
		}
		if err := json.Unmarshal(res.Body.Bytes(), &rows); err != nil {
			t.Fatal(err)
		}
		if len(rows) != tc.want || res.Header().Get("X-Has-More") != tc.more {
			t.Fatalf("%+v: rows=%d more=%s", tc, len(rows), res.Header().Get("X-Has-More"))
		}
	}
	// A full last page must not advertise an empty next page.
	res := httptest.NewRecorder()
	cfg.handlerJobsGet(res, httptest.NewRequest("GET", "/api/jobs?q="+marker+"&limit=101", nil))
	if res.Header().Get("X-Has-More") != "false" {
		t.Fatal("full final page has next")
	}
}

func TestJobPageRejectsInvalidLimit(t *testing.T) {
	for _, value := range []string{"0", "501", "-1", "nope"} {
		res := httptest.NewRecorder()
		(&apiConfig{}).handlerJobsGet(res, httptest.NewRequest("GET", "/api/jobs?limit="+value, nil))
		if res.Code != 400 {
			t.Fatalf("%s: %d", value, res.Code)
		}
	}
}
