package main

import (
	"context"
	"github.com/google/uuid"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/testutil"
	"testing"
)

func TestJobScoreSortBeforePagination(t *testing.T) {
	p := testutil.Pool(t)
	ctx := context.Background()
	user := uuid.New()
	if _, err := p.Exec(ctx, "INSERT INTO users(id,email,username) VALUES($1,$2::text,$2::text)", user, user.String()+"@example.test"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { p.Exec(ctx, "DELETE FROM users WHERE id=$1", user) })
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	for i, id := range ids {
		if _, err := p.Exec(ctx, "INSERT INTO jobs(id,external_id,job_source,title,company_name,source_url,created_by_user_id,created_at) VALUES($1,$2,'manual',$3,'Example',$4,$5,now()+$6*interval '1 hour')", id, id.String(), user.String(), "https://example.test/"+id.String(), user, i); err != nil {
			t.Fatal(err)
		}
		if i < 2 {
			if _, err := p.Exec(ctx, "INSERT INTO user_job_matches(user_id,job_id,match_score) VALUES($1,$2,$3)", user, id, []int{80, 30}[i]); err != nil {
				t.Fatal(err)
			}
		}
	}
	if _, err := p.Exec(ctx, "DELETE FROM pending_matches WHERE user_id=$1", user); err != nil {
		t.Fatal(err)
	}
	q := database.New(p)
	for order, want := range map[string][]uuid.UUID{"score_desc": {ids[0], ids[1], ids[2]}, "score_asc": {ids[1], ids[0], ids[2]}, "newest": {ids[2], ids[1], ids[0]}, "oldest": ids} {
		for offset, id := range want {
			rows, err := q.GetJobs(ctx, database.GetJobsParams{UserID: user, Search: user.String(), Filter: "all", SortOrder: order, Limit: 1, Offset: int32(offset)})
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != 1 || rows[0].ID != id {
				t.Fatalf("%s offset %d: %+v", order, offset, rows)
			}
		}
	}
	// Feed membership depends only on score, even with failed eligibility or pending rematching.
	if _, err := p.Exec(ctx, "UPDATE user_job_matches SET preference_eligible=false WHERE user_id=$1", user); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Exec(ctx, "INSERT INTO pending_matches(user_id,job_id) VALUES($1,$2) ON CONFLICT DO NOTHING", user, ids[0]); err != nil {
		t.Fatal(err)
	}
	rows, err := q.GetJobs(ctx, database.GetJobsParams{UserID: user, Search: user.String(), Filter: "matched", SortOrder: "score_desc", Limit: 50})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != ids[0] {
		t.Fatalf("threshold filter: %+v", rows)
	}
}
