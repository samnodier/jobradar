package matcher

import (
	"testing"
)

func TestMatchJob(t *testing.T) {
	desiredRoles := []string{"Go Developer", "Backend Engineer"}
	userSkills := map[string]float64{"Go": 1.0, "PostgreSQL": 0.8, "Vue.js": 0.5, "Docker": 0.5}
	jobSkills := []string{"Go", "SQL", "React"}
	userExperiences := []string{
		"Built scalable backend systems and web applications with Go and Vue.",
		"Optimized relational databases like PostgreSQL and Docker container deployments.",
	}

	tests := []struct {
		jobTitle        string
		jobDesc         string
		expectedSkipped bool
		expectHighMatch bool // Expect score > 50 if true, else score < 50
		expectScoreMax  int
	}{
		{
			jobTitle:        "Senior Go Software Engineer",
			jobDesc:         "Looking for a backend  engineer who works with Go, PostgreSQL, and Docker containers.",
			expectedSkipped: false,
			expectHighMatch: true,
		},
		{
			jobTitle:        "Fullstack Developer",
			jobDesc:         "Looking for a fullstack developer who has experience with Go, SQL, React, PostgreSQL, docker, and AWS.",
			expectedSkipped: false,
			expectHighMatch: true,
			expectScoreMax:  100,
		},

		{
			jobTitle: "Creative Director",
			jobDesc:  "Looking for an art designer with experience leading marketing teams.", expectedSkipped: true, // Title similarity is too low, skips early
			expectHighMatch: false,
		},
	}

	for _, tt := range tests {
		res := MatchJob(tt.jobTitle, tt.jobDesc, desiredRoles, userSkills, jobSkills, userExperiences, 0.55)
		runMatchAssertions(t, tt.jobTitle, res, tt.expectedSkipped, tt.expectHighMatch, tt.expectScoreMax)
	}
}

func runMatchAssertions(t *testing.T, jobTitle string, res MatchResult, expectedSkipped, expectHighMatch bool, expectScoreMax int) {
	t.Helper()
	if res.Skipped != expectedSkipped {
		t.Errorf("MatchJob(%q) Skipped = %v; want %v", jobTitle, res.Skipped, expectedSkipped)
	}

	if !res.Skipped {
		if expectHighMatch && res.Score <= 50 {
			t.Errorf("MatchJob(%q) Score = %d; expected a high match (>50)", jobTitle, res.Score)
		}
		if !expectHighMatch && res.Score > 50 {
			t.Errorf("MatchJob(%q) Score = %d; expected a low match (<=50)", jobTitle, res.Score)
		}
		if expectScoreMax > 0 && res.Score > expectScoreMax {
			t.Errorf("MatchJob(%q) Score = %d; exceeded max of %d", jobTitle, res.Score, expectScoreMax)
		}
	} else {
		if res.Score != 0 {
			t.Errorf("MatchJob(%q) expected Score = 0 when skipped; got %d", jobTitle, res.Score)
		}
	}
}

// A user with skills and experience but no desired roles must still get
// matches — the title dimension drops out and the score renormalizes over
// skills + experience.
func TestMatchJobNoDesiredRoles(t *testing.T) {
	userSkills := map[string]float64{"Go": 1.0, "PostgreSQL": 0.8}
	jobSkills := []string{"Go", "PostgreSQL"}
	userExperiences := []string{"Built backend systems with Go and PostgreSQL."}

	res := MatchJob(
		"Backend Engineer",
		"Backend role working with Go and PostgreSQL.",
		nil, userSkills, jobSkills, userExperiences, 0.55,
	)
	if res.Skipped {
		t.Fatal("expected job not to be skipped when user has skills/experience but no desired roles")
	}
	if res.Score <= 0 {
		t.Errorf("expected a positive score from skills+experience, got %d", res.Score)
	}
	if res.TitleScore != 0 {
		t.Errorf("expected TitleScore 0 with no desired roles, got %f", res.TitleScore)
	}
}

// A profile with no roles, no skills, and no experiences has zero signal:
// the only honest answer is a skip, never a fabricated score (and never a
// divide-by-zero).
func TestMatchJobNoSignalAtAll(t *testing.T) {
	res := MatchJob("Backend Engineer", "Any description.", nil, nil, nil, nil, 0.55)
	if !res.Skipped {
		t.Errorf("expected skip when the profile has no roles, skills, or experiences; got score %d", res.Score)
	}
}

// titleThreshold 0 disables the early-exit gate (used for user-imported
// jobs): a title matching none of the desired roles must be scored honestly,
// not skipped.
func TestMatchJobZeroThresholdBypassesGate(t *testing.T) {
	desiredRoles := []string{"Go Developer"}
	userSkills := map[string]float64{"Go": 1.0}

	res := MatchJob(
		"Creative Director",
		"Leading a marketing team; some Go tooling involved.",
		desiredRoles, userSkills, []string{"Go"}, nil, 0,
	)
	if res.Skipped {
		t.Fatal("expected no skip with titleThreshold 0")
	}
}
