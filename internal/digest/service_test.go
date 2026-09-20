package digest

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samnodier/jobradar/internal/mailer"
	"github.com/samnodier/jobradar/internal/testutil"
)

type fakeSender struct {
	mu       sync.Mutex
	calls    int
	keys     []string
	messages []mailer.Message
	err      error
	safe     bool
}

func (f *fakeSender) RetrySafe() bool { return f.safe }
func (f *fakeSender) Send(_ context.Context, key string, m mailer.Message) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.keys = append(f.keys, key)
	f.messages = append(f.messages, m)
	return "provider-id", f.err
}
func setup(t *testing.T) (*Service, *fakeSender, uuid.UUID, *pgxpool.Pool) {
	t.Helper()
	p := testutil.Pool(t)
	ctx := context.Background()
	user, job := uuid.New(), uuid.New()
	email := user.String() + "@example.test"
	for _, stmt := range []struct {
		sql  string
		args []any
	}{
		{"INSERT INTO users(id,email,username)VALUES($1,$2::text,$2::text)", []any{user, email}},
		{"INSERT INTO jobs(id,external_id,job_source,title,company_name,source_url,created_by_user_id)VALUES($1,$2,'manual','<script>alert(1)</script> Engineer','Example','https://example.test/job',$3)", []any{job, job.String(), user}},
		{"INSERT INTO user_job_matches(user_id,job_id,match_score)VALUES($1,$2,90)", []any{user, job}},
		{"DELETE FROM pending_matches WHERE user_id=$1", []any{user}},
	} {
		if _, err := p.Exec(ctx, stmt.sql, stmt.args...); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { p.Exec(ctx, "DELETE FROM users WHERE id=$1", user) })
	sender := &fakeSender{safe: true}
	service := &Service{Pool: p, Config: Config{Enabled: true, Recipient: email, From: "jobs@example.test", Provider: "resend", Location: time.UTC, Hour: 8, MinScore: 50, MaxJobs: 10}, Sender: sender}
	return service, sender, user, p
}
func TestDigestPreviewAndDailyDeduplication(t *testing.T) {
	s, f, user, p := setup(t)
	ctx := context.Background()
	preview, err := s.Preview(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(preview, "<script>") {
		t.Fatal("unescaped job content")
	}
	if f.calls != 0 {
		t.Fatal("preview sent mail")
	}
	now := time.Now()
	for i := 0; i < 2; i++ {
		if err = s.Tick(ctx, now, true); err != nil {
			t.Fatal(err)
		}
	}
	if f.calls != 1 {
		t.Fatalf("expected one send, got %d", f.calls)
	}
	if err = s.Tick(ctx, now.Add(24*time.Hour), true); err != nil {
		t.Fatal(err)
	}
	if f.calls != 1 {
		t.Fatal("same job sent on next day")
	}
	var count int
	if err = p.QueryRow(ctx, "SELECT count(*) FROM email_digests WHERE user_id=$1 AND status='sent'", user).Scan(&count); err != nil || count != 1 {
		t.Fatal(count, err)
	}
}
func TestDigestRetriesExactPersistedPayload(t *testing.T) {
	s, f, user, p := setup(t)
	ctx := context.Background()
	f.err = &mailer.SendError{Code: 503, Ambiguous: true}
	if err := s.Tick(ctx, time.Now(), true); err == nil {
		t.Fatal("expected provider error")
	}
	if _, err := p.Exec(ctx, "UPDATE email_digests SET next_attempt_at=now()-interval '1 second' WHERE user_id=$1", user); err != nil {
		t.Fatal(err)
	}
	f.err = nil
	if err := s.Tick(ctx, time.Now(), true); err != nil {
		t.Fatal(err)
	}
	if f.calls != 2 || f.keys[0] != f.keys[1] || f.messages[0].HTML != f.messages[1].HTML {
		t.Fatal("retry changed identity or content")
	}
}
func TestNonIdempotentAmbiguousSendIsNotRepeated(t *testing.T) {
	s, f, user, p := setup(t)
	ctx := context.Background()
	f.safe = false
	f.err = &mailer.SendError{Ambiguous: true}
	if err := s.Tick(ctx, time.Now(), true); err == nil {
		t.Fatal("expected error")
	}
	f.err = nil
	if err := s.Tick(ctx, time.Now(), true); err != nil {
		t.Fatal(err)
	}
	if f.calls != 1 {
		t.Fatal("unsafe resend")
	}
	var status string
	if err := p.QueryRow(ctx, "SELECT status FROM email_digests WHERE user_id=$1", user).Scan(&status); err != nil || status != "uncertain" {
		t.Fatal(status, err)
	}
}
func TestDigestHonorsOptOut(t *testing.T) {
	s, f, user, p := setup(t)
	ctx := context.Background()
	p.Exec(ctx, "UPDATE users SET notify_jobs=false WHERE id=$1", user)
	if err := s.Tick(ctx, time.Now(), true); err != nil {
		t.Fatal(err)
	}
	if f.calls != 0 {
		t.Fatal("opted-out user received email")
	}
}
func TestDigestExpiredRetryWindowNeedsReview(t *testing.T) {
	s, f, user, p := setup(t)
	ctx := context.Background()
	f.err = errors.New("temporary failure")
	s.Tick(ctx, time.Now(), true)
	p.Exec(ctx, "UPDATE email_digests SET first_attempt_at=now()-interval '24 hours',next_attempt_at=now()-interval '1 second' WHERE user_id=$1", user)
	f.err = nil
	if err := s.Tick(ctx, time.Now(), true); err != nil {
		t.Fatal(err)
	}
	if f.calls != 1 {
		t.Fatal("resent outside idempotency window")
	}
}
