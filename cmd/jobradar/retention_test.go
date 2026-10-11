package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/sources"
	"github.com/samnodier/jobradar/internal/testutil"
)

func TestJobRetentionFromEnv(t *testing.T) {
	for _, tc := range []struct {
		raw     string
		want    time.Duration
		wantErr bool
	}{
		{"", 90 * 24 * time.Hour, false},
		{"0", 0, false},
		{"30", 30 * 24 * time.Hour, false},
		{"-1", 0, true},
		{"three months", 0, true},
	} {
		t.Setenv("JOB_RETENTION_DAYS", tc.raw)
		got, err := jobRetentionFromEnv()
		if (err != nil) != tc.wantErr || got != tc.want {
			t.Fatalf("JOB_RETENTION_DAYS=%q: got %v, %v", tc.raw, got, err)
		}
	}
}

// Fixtures use dates around the year 2000 with a 2000-01-01 cutoff, so the shared test
// database's other rows (all far newer) are never affected by this test's deletes.
func TestDeleteExpiredJobsKeepsAppliedSavedAndImportedJobs(t *testing.T) {
	pool := testutil.Pool(t)
	ctx := context.Background()
	user := uuid.New()
	if _, err := pool.Exec(ctx, "INSERT INTO users(id,email,username) VALUES($1,$2::text,$2::text)", user, user.String()+"@example.test"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM users WHERE id=$1", user) })

	type fixture struct {
		name                string
		posted, created     string
		private, saved, app bool
		wantKept            bool
	}
	fixtures := []fixture{
		{name: "old public", created: "1999-06-01", wantKept: false},
		{name: "old posting first seen recently", posted: "1999-06-01", created: "2001-06-01", wantKept: false},
		{name: "old public #3 (exercises batching)", created: "1999-07-01", wantKept: false},
		{name: "recent public", created: "2001-06-01", wantKept: true},
		{name: "old but applied to", created: "1999-06-01", app: true, wantKept: true},
		{name: "old but saved", created: "1999-06-01", saved: true, wantKept: true},
		{name: "old private import", created: "1999-06-01", private: true, wantKept: true},
	}
	ids := make([]uuid.UUID, len(fixtures))
	for i, f := range fixtures {
		ids[i] = uuid.New()
		var owner *uuid.UUID
		if f.private {
			owner = &user
		}
		var posted *string
		if f.posted != "" {
			posted = &f.posted
		}
		if _, err := pool.Exec(ctx, `INSERT INTO jobs(id, external_id, job_source, title, company_name, source_url, created_by_user_id, posted_at, created_at)
			VALUES($1,$2,'retention-test',$3,'Example','https://example.test/'||$2,$4,$5::timestamptz,$6::timestamptz)`,
			ids[i], ids[i].String(), f.name, owner, posted, f.created); err != nil {
			t.Fatal(f.name, err)
		}
		if f.saved {
			if _, err := pool.Exec(ctx, "INSERT INTO saved_jobs(user_id, job_id) VALUES($1,$2)", user, ids[i]); err != nil {
				t.Fatal(err)
			}
		}
		if f.app {
			if _, err := pool.Exec(ctx, "INSERT INTO applications(user_id, job_id, application_status) VALUES($1,$2,'applied')", user, ids[i]); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM jobs WHERE id=ANY($1)", ids) })

	cutoff := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	removed, err := deleteExpiredJobs(ctx, database.New(pool), cutoff, 2)
	if err != nil {
		t.Fatal(err)
	}
	if removed != 3 {
		t.Fatalf("removed %d jobs, want 3", removed)
	}
	for i, f := range fixtures {
		var exists bool
		if err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM jobs WHERE id=$1)", ids[i]).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if exists != f.wantKept {
			t.Errorf("%s: kept=%v, want %v", f.name, exists, f.wantKept)
		}
	}
}

type staticSource struct {
	name string
	jobs []sources.Job
}

func (s staticSource) Name() string                                 { return s.name }
func (s staticSource) Fetch(context.Context) ([]sources.Job, error) { return s.jobs, nil }

func TestScrapeSkipsPostingsOlderThanRetention(t *testing.T) {
	pool := testutil.Pool(t)
	ctx := context.Background()
	name := "retention-src-" + uuid.NewString()[:8]
	old := time.Now().AddDate(-1, 0, 0)
	fresh := time.Now().Add(-time.Hour)
	src := staticSource{name: name, jobs: []sources.Job{
		{ID: "old", Source: name, Title: "Old posting", Company: "Example", URL: "https://example.test/old", PostedAt: &old},
		{ID: "fresh", Source: name, Title: "Fresh posting", Company: "Example", URL: "https://example.test/fresh", PostedAt: &fresh},
		{ID: "undated", Source: name, Title: "Undated posting", Company: "Example", URL: "https://example.test/undated"},
	}}
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM jobs WHERE job_source=$1", name)
		pool.Exec(ctx, "DELETE FROM source_health_daily WHERE service_name=$1", name)
		pool.Exec(ctx, "DELETE FROM service_health WHERE service_name=$1", name)
	})
	cfg := &apiConfig{db: database.New(pool), jobRetention: 90 * 24 * time.Hour}
	if err := cfg.scrapeSource(ctx, src); err != nil {
		t.Fatal(err)
	}
	rows, err := pool.Query(ctx, "SELECT external_id FROM jobs WHERE job_source=$1 ORDER BY external_id", name)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		got = append(got, id)
	}
	if len(got) != 2 || got[0] != "fresh" || got[1] != "undated" {
		t.Fatalf("stored %v, want [fresh undated]: postings older than retention must not be re-inserted", got)
	}
}
