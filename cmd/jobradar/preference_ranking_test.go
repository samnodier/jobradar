package main

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/queue"
	"github.com/samnodier/jobradar/internal/testutil"
	"testing"
)

func TestProfileChangesRecomputeStoredRanking(t *testing.T) {
	p := testutil.Pool(t)
	ctx := context.Background()
	q := database.New(p)
	user, job := uuid.New(), uuid.New()
	if _, e := p.Exec(ctx, "INSERT INTO users(id,email,username,preferred_industries)VALUES($1,$2::text,$2::text,ARRAY['Finance'])", user, user.String()+"@example.test"); e != nil {
		t.Fatal(e)
	}
	defer p.Exec(ctx, "DELETE FROM users WHERE id=$1", user)
	for _, statement := range []string{"INSERT INTO user_desired_roles(user_id,role_title)VALUES($1,'Go Developer')", "INSERT INTO user_desired_locations(user_id,location_name)VALUES($1,'Nairobi')"} {
		if _, e := p.Exec(ctx, statement, user); e != nil {
			t.Fatal(e)
		}
	}
	if _, e := p.Exec(ctx, "INSERT INTO jobs(id,external_id,job_source,title,company_name,source_url,created_by_user_id,job_location,category)VALUES($1,$2,'manual','Go Developer','Example',$2,$3,'Nairobi','Finance Jobs')", job, job.String(), user); e != nil {
		t.Fatal(e)
	}
	cfg := &apiConfig{db: q, pool: p}
	body, _ := json.Marshal(MatchJobPayload{JobID: job.String(), UserID: user.String()})
	run := func() database.GetJobByIDRow {
		t.Helper()
		if e := cfg.handleMatchJob(ctx, &queue.Job{Payload: body}); e != nil {
			t.Fatal(e)
		}
		row, e := q.GetJobByID(ctx, database.GetJobByIDParams{UserID: user, ID: job})
		if e != nil {
			t.Fatal(e)
		}
		return row
	}
	first := run()
	if first.MatchScore == nil || (first.PreferenceEligible == nil || !*first.PreferenceEligible) {
		t.Fatal("matching profile should qualify")
	}
	rev, e := q.GetPendingMatchRevision(ctx, database.GetPendingMatchRevisionParams{UserID: user, JobID: job})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = p.Exec(ctx, "UPDATE users SET preferred_industries=ARRAY['Hospitality'] WHERE id=$1", user); e != nil {
		t.Fatal(e)
	}
	next, e := q.GetPendingMatchRevision(ctx, database.GetPendingMatchRevisionParams{UserID: user, JobID: job})
	if e != nil || next <= rev {
		t.Fatal("industry change did not invalidate scores", e)
	}
	changed := run()
	if (changed.PreferenceEligible == nil || *changed.PreferenceEligible) || *changed.MatchScore >= *first.MatchScore {
		t.Fatal("industry must change stored rank and eligibility")
	}
}
