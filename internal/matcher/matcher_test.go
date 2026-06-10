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

		if res.Skipped != tt.expectedSkipped {
			t.Errorf("MatchJob(%q) Skipped = %v; want %v", tt.jobTitle, res.Skipped, tt.expectedSkipped)
		}

		if !res.Skipped {
			if tt.expectHighMatch && res.Score <= 50 {
				t.Errorf("MatchJob(%q) Score = %d; expected a high match (>50)", tt.jobTitle, res.Score)
			}
			if !tt.expectHighMatch && res.Score > 50 {
				t.Errorf("MatchJob(%q) Score = %d; expected a low match (<=50)", tt.jobTitle, res.Score)
			}
			if tt.expectScoreMax > 0 && res.Score > tt.expectScoreMax {
				t.Errorf("MatchJob(%q) Score = %d; exceeded max of %d", tt.jobTitle, res.Score, tt.expectScoreMax)
			}
		} else {
			if res.Score != 0 {
				t.Errorf("MatchJob(%q) expected Score = 0 when skipped; got %d", tt.jobTitle, res.Score)
			}
		}
	}
}
