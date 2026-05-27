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
