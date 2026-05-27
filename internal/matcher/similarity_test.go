package matcher

import (
	"math"
	"testing"
)

// closeEnough checks if two floats are equal within a small margin of error (delta)
func closeEnough(a, b, delta float64) bool {
	return math.Abs(a-b) < delta
}

func TestJaroWinkler(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected float64
	}{
		{"MARTHA", "MARHTA", 0.9611},  // Winkler boost applied
		{"DIXON", "DICKSONX", 0.8133}, // Shared prefix "DI" gives boost
		{"JELLYFISH", "SMELLYFISH", 0.8963},
		{"", "", 1.0},
		{"A", "", 0.0},
		{"", "A", 0.0},
		{"A", "A", 1.0},
	}

	for _, tt := range tests {
		result := JaroWinkler(tt.s1, tt.s2)
		if !closeEnough(result, tt.expected, 0.001) {
			t.Errorf("JaroWinkler(%q, %q) = %f; want %f", tt.s1, tt.s2, result, tt.expected)
		}
	}
}

func TestTokenSortJaroWinkler(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected float64
	}{
		{"Go Software Engineer", "Software Engineer in Go", 1.0}, // "in" will be filtered
		{"Senior Product Manager", "Product Manager", 0.9363},
		{"Google Cloud Platform", "Platform Google Cloud", 1.0},
		{"Sr. SWE", "Senior Software Engineer", 1.0},
		{"Staff Dev", "Developer", 1.0},
		{"VP of Eng", "Vice President Engineer", 1.0},
	}

	for _, tt := range tests {
		result := TokenSortJaroWinkler(tt.s1, tt.s2)
		if !closeEnough(result, tt.expected, 0.001) {
			t.Errorf("TokenSortJaroWinkler(%q, %q) = %f; want %f", tt.s1, tt.s2, result, tt.expected)
		}
	}
}
