package main

import (
	"context"
	"github.com/google/uuid"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/testutil"
	"net/url"
	"testing"
	"time"
)

func TestExportArgumentsAndDST(t *testing.T) {
	p, e := parseExportQuery(url.Values{"date": {"2026-03-08"}, "min_score": {"70"}}, "America/New_York", 90)
	if e != nil || p.MinScore != 70 || p.EndAt.Sub(*p.StartAt) != 23*time.Hour {
		t.Fatalf("date or override wrong: %+v %v", p, e)
	}
	for _, v := range []url.Values{{"min_score": {"101"}}, {"limit": {"0"}}, {"timezone": {"invalid"}}, {"cursor": {"bad"}}, {"date": {"2026-02-30"}}, {"user_id": {"other"}}, {"min_score": {"60", "70"}}} {
		if _, e := parseExportQuery(v, "UTC", 60); e == nil {
			t.Fatalf("accepted invalid args: %v", v)
		}
	}
}
func TestExportOwnershipThresholdAndPagination(t *testing.T) {
	p := testutil.Pool(t)
	ctx := context.Background()
	q := database.New(p)
	owner, other := uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{owner, other} {
		if _, err := p.Exec(ctx, "INSERT INTO users(id,email,username)VALUES($1,$2::text,$2::text)", id, id.String()+"@example.test"); err != nil {
			t.Fatal(err)
		}
	}
	defer p.Exec(ctx, "DELETE FROM users WHERE id=ANY($1)", []uuid.UUID{owner, other})
	// All are private fixtures: two eligible, one low score, one ineligible, one owned by another user.
	ids := []uuid.UUID{}
	for i := 0; i < 5; i++ {
		id := uuid.New()
		ids = append(ids, id)
		uid := owner
		if i == 4 {
			uid = other
		}
		if _, err := p.Exec(ctx, "INSERT INTO jobs(id,external_id,job_source,title,company_name,source_url,created_by_user_id)VALUES($1,$2,'test-export','Engineer','Example',$2,$3)", id, id.String(), uid); err != nil {
			t.Fatal(err)
		}
		score := 80
		if i == 2 {
			score = 69
		}
		if _, err := p.Exec(ctx, "INSERT INTO user_job_matches(user_id,job_id,match_score,preference_eligible)VALUES($1,$2,$3,$4)", owner, id, score, i != 3); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := p.Exec(ctx, "DELETE FROM pending_matches WHERE user_id=$1", owner); err != nil {
		t.Fatal(err)
	}
	args, e := parseExportQuery(url.Values{"min_score": {"70"}, "date": {"today"}, "source": {"test-export"}}, "UTC", 95)
	if e != nil {
		t.Fatal(e)
	}
	args.UserID = owner
	args.PageLimit = 1
	first, e := q.ExportJobs(ctx, args)
	if e != nil || len(first) != 1 {
		t.Fatalf("first page: %v %v", first, e)
	}
	args.BeforeScore = *first[0].MatchScore
	args.BeforeID = first[0].ID
	second, e := q.ExportJobs(ctx, args)
	if e != nil || len(second) != 1 || second[0].ID == first[0].ID {
		t.Fatalf("second page: %v %v", second, e)
	}
	args.BeforeScore = *second[0].MatchScore
	args.BeforeID = second[0].ID
	last, e := q.ExportJobs(ctx, args)
	if e != nil || len(last) != 0 {
		t.Fatalf("leaked ineligible, low or private jobs: %v %v", last, e)
	}
	args.BeforeScore = 101
	args.EligibleOnly = false
	args.PageLimit = 20
	rows, e := q.ExportJobs(ctx, args)
	if e != nil || len(rows) != 3 {
		t.Fatalf("eligibility override: %d %v", len(rows), e)
	}
	// Pending recomputation is excluded even when eligibility is overridden.
	if err := q.ScheduleUserMatches(ctx, owner); err != nil {
		t.Fatal(err)
	}
	rows, e = q.ExportJobs(ctx, args)
	if e != nil || len(rows) != 0 {
		t.Fatal("exported stale matches", e)
	}
}
