package main

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/testutil"
	"testing"
)

func TestFeedbackOwnershipAndDurableRematching(t *testing.T) {
	p := testutil.Pool(t)
	ctx := context.Background()
	q := database.New(p)
	owner, other, job := uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{owner, other} {
		if _, err := p.Exec(ctx, "INSERT INTO users(id,email,username) VALUES($1,$2::text,$2::text)", id, id.String()+"@example.test"); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { p.Exec(ctx, "DELETE FROM users WHERE id=ANY($1)", []uuid.UUID{owner, other}) })
	if _, err := p.Exec(ctx, "INSERT INTO jobs(id,external_id,job_source,title,company_name,source_url,created_by_user_id) VALUES($1,'feedback','manual','Go Developer','Example','https://example.test/feedback',$2)", job, owner); err != nil {
		t.Fatal(err)
	}
	if _, err := q.SetJobFeedback(ctx, database.SetJobFeedbackParams{UserID: other, JobID: job, Sentiment: 1}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("private feedback must fail: %v", err)
	}
	before, err := q.GetPendingMatchRevision(ctx, database.GetPendingMatchRevisionParams{UserID: owner, JobID: job})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := q.SetJobFeedback(ctx, database.SetJobFeedbackParams{UserID: owner, JobID: job, Sentiment: -1}); err != nil {
		t.Fatal(err)
	}
	after, err := q.GetPendingMatchRevision(ctx, database.GetPendingMatchRevisionParams{UserID: owner, JobID: job})
	if err != nil || after <= before {
		t.Fatal("vote must atomically schedule rematching", err)
	}
	if err := q.ClearJobFeedback(ctx, database.ClearJobFeedbackParams{UserID: other, JobID: job}); err != nil {
		t.Fatal(err)
	}
	if vote, err := q.GetJobFeedback(ctx, database.GetJobFeedbackParams{UserID: owner, JobID: job}); err != nil || vote != -1 {
		t.Fatal("other user cleared owner vote", err)
	}
	if err := q.ClearJobFeedback(ctx, database.ClearJobFeedbackParams{UserID: owner, JobID: job}); err != nil {
		t.Fatal(err)
	}
	if _, err := q.GetJobFeedback(ctx, database.GetJobFeedbackParams{UserID: owner, JobID: job}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatal("clear failed", err)
	}
}
