package matcher

import (
	"testing"
)

func TestExperienceMatch(t *testing.T) {
	expTexts := []string{
		"Developed backend services using Go and PostgreSQL database.", "Optimized SQL queries and designed RESTful API endpoints.",
	}

	tests := []struct {
		jobDesc  string
		expected float64
	}{
		{
			jobDesc: "Looking for a Go developer who writes SQL and works with PostgreSQL backend.", expected: 0.2857, // 4 matching words ("go", "sql", "postgresql", "backend") / 14 unique experience words
		},
		{
			jobDesc: "Looking for a creative UI designer with Figma skills.", expected: 0.0, // No overlapping words
		},
	}

	for _, tt := range tests {
		result := ExperienceMatch(expTexts,
			tt.jobDesc)
		if !closeEnough(result, tt.expected, 0.001) {
			t.Errorf("ExperienceMatch() = %f; want %f", result, tt.expected)
		}
	}
}

// Regression: a fuller resume used to rank worse because the union of all experiences was
// the denominator. Adding an unrelated role must never lower the experience score.
func TestBestExperienceMatchIgnoresUnrelatedExperience(t *testing.T) {
	job := "Backend engineer building Go APIs, PostgreSQL schemas and Redis queues for web applications"
	developer := "Web Developer at Example building Go APIs and PostgreSQL backed web applications"
	marketing := "Social Media Manager at Print Shop running email campaigns, audience engagement and brand awareness content"

	alone := BestExperienceMatch([]string{developer}, job)
	withUnrelated := BestExperienceMatch([]string{developer, marketing}, job)
	if alone == 0 {
		t.Fatal("expected the developer experience to overlap the job")
	}
	if withUnrelated != alone {
		t.Fatalf("adding an unrelated experience changed the score: %.3f -> %.3f", alone, withUnrelated)
	}
	if union := ExperienceMatch([]string{developer, marketing}, job); union >= alone {
		t.Fatalf("expected the old union ratio to be diluted (%.3f) below the best single match (%.3f)", union, alone)
	}
	if got := BestExperienceMatch(nil, job); got != 0 {
		t.Fatalf("no experiences should score 0, got %.3f", got)
	}
}

func TestMatchJobUnrelatedExperienceDoesNotLowerScore(t *testing.T) {
	roles := []string{"Backend Engineer"}
	desc := "Backend engineer building Go APIs, PostgreSQL schemas and Redis queues for web applications"
	developer := "Web Developer at Example building Go APIs and PostgreSQL backed web applications"
	marketing := "Social Media Manager at Print Shop running email campaigns, audience engagement and brand awareness content"
	base := MatchJob("Backend Engineer", desc, roles, nil, nil, []string{developer}, 0)
	fuller := MatchJob("Backend Engineer", desc, roles, nil, nil, []string{developer, marketing}, 0)
	if fuller.Score < base.Score {
		t.Fatalf("fuller resume scored lower: %d < %d", fuller.Score, base.Score)
	}
}
