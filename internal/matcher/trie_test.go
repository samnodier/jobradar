package matcher

import (
	"testing"
)

func TestSkillMatcher(t *testing.T) {
	skills := []string{"Go", "Vue.js", "PostgreSQL", "C++", ".NET", "Google Cloud"}
	sm := NewSkillMatcher(skills)

	tests := []struct {
		text     string
		expected []string
	}{
		{
			text:     "We are looking for a Go developer who knows Vue.js and PostgreSQL.",
			expected: []string{"Go", "PostgreSQL", "Vue.js"},
		},

		{
			text:     "Google is a search engine, outgoing person.",
			expected: []string{}, // "Go" shouldn't match inside "Google" or "outgoing"
		},

		{
			text:     "Experience with C++ and .NET is required.",
			expected: []string{".NET", "C++"},
		},
		{
			text:     "I have experience with Google Cloud platform.",
			expected: []string{"Google Cloud"},
		},
	}

	for _, tt := range tests {
		result := sm.FindSkills(tt.text)
		if len(result) != len(tt.expected) {
			t.Errorf("FindSkills(%q) = %v; want %v", tt.text, result, tt.expected)
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("FindSkills(%q) = %v; want %v", tt.text, result, tt.expected)
				break
			}
		}
	}
}
