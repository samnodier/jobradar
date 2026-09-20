package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/queue"
	"github.com/samnodier/jobradar/internal/testutil"
)

func TestPrivateJobAuthorizationAndDurableMatching(t *testing.T) {
	p := testutil.Pool(t)
	ctx := context.Background()
	q := database.New(p)
	owner, other, job := uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{owner, other} {
		_, e := p.Exec(ctx, "INSERT INTO users(id,email,username) VALUES($1,$2::text,$2::text)", id, id.String()+"@example.test")
		if e != nil {
			t.Fatal(e)
		}
	}
	t.Cleanup(func() { p.Exec(ctx, "DELETE FROM users WHERE id=ANY($1)", []uuid.UUID{owner, other}) })
	_, err := p.Exec(ctx, "INSERT INTO jobs(id,external_id,job_source,title,company_name,source_url,created_by_user_id) VALUES($1,'private','manual','Go Developer','Example','https://example.test/job',$2)", job, owner)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = q.SaveJob(ctx, database.SaveJobParams{UserID: other, ID: job}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("private save must be rejected: %v", err)
	}
	if _, err = q.CreateApplication(ctx, database.CreateApplicationParams{UserID: other, ID: job, ApplicationStatus: "applied"}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("private application must be rejected: %v", err)
	}
	if _, err = q.SaveJob(ctx, database.SaveJobParams{UserID: owner, ID: job}); err != nil {
		t.Fatal(err)
	}
	for _, user := range []uuid.UUID{owner, other} {
		feed, err := q.GetJobs(ctx, database.GetJobsParams{UserID: user, Limit: 1, Search: "Go Developer", Filter: "saved"})
		if err != nil {
			t.Fatal(err)
		}
		want := 0
		if user == owner {
			want = 1
		}
		if len(feed) != want {
			t.Fatalf("filtered feed authorization: got %d, want %d", len(feed), want)
		}
	}
	rev, err := q.GetPendingMatchRevision(ctx, database.GetPendingMatchRevisionParams{UserID: owner, JobID: job})
	if err != nil {
		t.Fatal("job insert must atomically create work:", err)
	}
	if _, err = q.GetPendingMatchRevision(ctx, database.GetPendingMatchRevisionParams{UserID: other, JobID: job}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatal("private work leaked to another user")
	}
	if err = q.ScheduleUserMatches(ctx, owner); err != nil {
		t.Fatal(err)
	}
	if err = q.CompletePendingMatch(ctx, database.CompletePendingMatchParams{UserID: owner, JobID: job, Revision: rev}); err != nil {
		t.Fatal(err)
	}
	next, err := q.GetPendingMatchRevision(ctx, database.GetPendingMatchRevisionParams{UserID: owner, JobID: job})
	if err != nil || next <= rev {
		t.Fatal("old acknowledgement must not delete newer work")
	}
	// Simulate delivery followed by process death: an expired delivery lease becomes eligible again.
	_, err = p.Exec(ctx, "UPDATE pending_matches SET enqueued_at=now()-interval '6 minutes' WHERE user_id=$1", owner)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := q.ListPendingMatches(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range rows {
		if r.UserID == owner && r.JobID == job {
			found = true
		}
	}
	if !found {
		t.Fatal("lost delivery was not recovered")
	}
}
func TestDurableMatchEndToEnd(t *testing.T) {
	p := testutil.Pool(t)
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR required")
	}
	ctx := context.Background()
	q := database.New(p)
	rdb := redis.NewClient(&redis.Options{Addr: addr, DB: 2})
	defer rdb.Close()
	name := "jobradar:test:" + uuid.NewString()
	rq := queue.NewNamedRedisQueue(rdb, name)
	t.Cleanup(func() { rdb.Del(ctx, name) })
	id, job := uuid.New(), uuid.New()
	_, err := p.Exec(ctx, "INSERT INTO users(id,email,username)VALUES($1,$2::text,$2::text)", id, id.String()+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { p.Exec(ctx, "DELETE FROM users WHERE id=$1", id) })
	if _, err = p.Exec(ctx, "INSERT INTO user_desired_roles(user_id,role_title)VALUES($1,'Go Developer')", id); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Exec(ctx, "INSERT INTO jobs(id,external_id,job_source,title,company_name,source_url,created_by_user_id)VALUES($1,$2,'manual','Go Developer','Example','https://example.test/work',$3)", job, job.String(), id); err != nil {
		t.Fatal(err)
	}
	cfg := &apiConfig{pool: p, db: q, queue: rq, enrichQueue: rq, aiMatchThreshold: 60}
	if err = cfg.dispatchPendingMatches(ctx); err != nil {
		t.Fatal(err)
	}
	revision, err := q.GetPendingMatchRevision(ctx, database.GetPendingMatchRevisionParams{UserID: id, JobID: job})
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(MatchJobPayload{JobID: job.String(), UserID: id.String(), Revision: revision})
	task := &queue.Job{Payload: body}
	if err = cfg.handleDurableMatch(ctx, task); err != nil {
		t.Fatal(err)
	}
	if err = cfg.handleDurableMatch(ctx, task); err != nil {
		t.Fatal("duplicate delivery must be harmless:", err)
	}
	notifications, err := q.ListMatchNotifications(ctx, id)
	if err != nil || len(notifications) != 1 {
		t.Fatalf("expected one notification after duplicate delivery: %v %v", notifications, err)
	}
	count, err := q.ReadMatchNotification(ctx, database.ReadMatchNotificationParams{UserID: uuid.New(), ID: notifications[0].ID})
	if err != nil || count != 0 {
		t.Fatal("another user could mark notification read", err)
	}
	var score int
	if err = p.QueryRow(ctx, "SELECT match_score FROM user_job_matches WHERE user_id=$1 AND job_id=$2", id, job).Scan(&score); err != nil || score != 100 {
		t.Fatalf("expected no-AI score 100, got %d: %v", score, err)
	}
	if _, err = q.GetPendingMatchRevision(ctx, database.GetPendingMatchRevisionParams{UserID: id, JobID: job}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatal("completed work was not acknowledged")
	}
	if _, err = p.Exec(ctx, "INSERT INTO user_experiences(user_id,company_name,role_title,start_date)VALUES($1,'Example','Developer','2020-01-01')", id); err != nil {
		t.Fatal(err)
	}
	if _, err = q.GetPendingMatchRevision(ctx, database.GetPendingMatchRevisionParams{UserID: id, JobID: job}); err != nil {
		t.Fatal("experience edit did not schedule rematching")
	}
}
func TestUnknownAPIRouteDoesNotServeSPA(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/missing", nil)
	w := httptest.NewRecorder()
	spaHandler(t.TempDir())(w, r)
	if w.Code != 404 {
		t.Fatal(w.Code)
	}
}
