package stringutils

import (
	"testing"
)

func TestSanitizeStrict(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "strips html tags",
			input:    "<b>Software Engineer</b>",
			expected: "Software Engineer",
		},
		{
			name:     "unescapes html entities",
			input:    "R&amp;D Engineer",
			expected: "R&D Engineer",
		},
		{
			name:     "unescapes html entities",
			input:    "We&rsquo;re hiring",
			expected: "We’re hiring",
		},
		{
			name:     "collapses extra whitespace",
			input:    "Senior   Frontend    Engineer",
			expected: "Senior Frontend Engineer",
		},
		{
			name:     "trims surrounding whitespace",
			input:    "  Engineer  ",
			expected: "Engineer",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "handles plain text unchanged",
			input:    "Go Engineer",
			expected: "Go Engineer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeStrict(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeStrict(%q)\n  got:  %q\n  want: %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizeDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "preserves html tags",

			input:    "<ul><li>Build things</li></ul>",
			expected: "<ul><li>Build things</li></ul>",
		},
		{
			name:     "unescapes html entities",
			input:    "We&rsquo;re hiring",
			expected: "We’re hiring", // Smart apostrophe U+2019
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "trims surrounding whitespace",
			input:    "  About the role  ",
			expected: "About the role",
		},
		{
			name:     "collapses multiple newlines",
			input:    "Line 1\n\n\n\nLine 2",
			expected: "Line 1\n\nLine 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeDescription(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeDescription(%q)\n got: %q\n want: %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGenerateUsername(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "extracts username from email",
			email:    "sam@example.com",
			expected: "sam",
		},
		{
			name:     "handles subdomain email",
			email:    "sam.nodier@company.co.uk",
			expected: "sam.nodier",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateUsername(tt.email)
			if got != tt.expected {
				t.Errorf("GenerateUsername(%q)\n  got:  %q\n  want: %q", tt.email, got, tt.expected)
			}
		})
	}
}

func TestGenerateUsernameEmpty(t *testing.T) {
	// empty email should generate a fallback username
	got := GenerateUsername("")
	if len(got) < 5 {
		t.Errorf("GenerateUsername(\"\") returned too short: %q", got)
	}
	if got[:5] != "user_" {
		t.Errorf("GenerateUsername(\"\") should start with 'user_', got: %q", got)
	}
}

func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		valid    bool
	}{
		{"valid username", "sam_123", true},
		{"valid all letters", "samuel", true},
		{"too short", "ab", false},
		{"too long", "this_username_is_way_too_long_123", false},
		{"has spaces", "sam nodier", false},
		{"has hyphen", "sam-nodier", false},
		{"has special chars", "sam@nodier", false},
		{"exactly 3 chars", "sam", true},
		{"exactly 20 chars", "sam_123_sam_123_sam_", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidUsername(tt.username)
			if got != tt.valid {
				t.Errorf("IsValidUsername(%q) = %v, want %v", tt.username, got, tt.valid)
			}
		})
	}
}

func TestSanitizeStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "trims, removes empty and duplicates",
			input:    []string{"  Go ", "", "go", " Vue.js ", "Go", "  "},
			expected: []string{"Go", "Vue.js"},
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: []string{},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeStringSlice(tt.input)
			if len(got) != len(tt.expected) {
				t.Fatalf("SanitizeStringSlice(%v) returned slice of length %d, want %d (got: %v)", tt.input, len(got), len(tt.expected), got)
			}
			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("SanitizeStringSlice(%v) at index %d got %q, want %q", tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestFixMojibake(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal ASCII String",
			input:    "Software Engineer - New York",
			expected: "Software Engineer - New York",
		},
		{
			name:     "Simple Mojibake Word",
			input:    "à¤°à¥ˆà¤ªà¥€à¤¡à¥‹",
			expected: "रैपीडो",
		},
		{
			name:     "Mojibake Emoji",
			input:    "ðŸš€ ðŸ’°",
			expected: "🚀 💰",
		},
		{
			// FixMojibake only reverses encoding corruption — HTML stripping is SanitizeStrict's job
			name:     "More Mojibake Sentence",
			input:    "Vous Ãªtes Ã  la recherche d'un <strong>job Ã©tudiant</strong>? Devenez professeur avec Voscours.",
			expected: "Vous êtes à la recherche d'un <strong>job étudiant</strong>? Devenez professeur avec Voscours.",
		},
		{
			// Covers Devanagari + emoji + checkmark in one pass using only bytes defined in Windows-1252.
			// The full Rapido payload was removed: it contains virama (U+094D, byte 0x8D) and short-u
			// matra (U+0941, byte 0x81) whose original bytes are undefined in Windows-1252, making
			// the round-trip encoding impossible.
			name:     "Mixed Devanagari emoji and ASCII",
			input:    "Gorakhpur âœ… à¤°à¤¾à¤® ðŸ“¢",
			expected: "Gorakhpur ✅ राम 📢",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FixMojibake(tc.input)

			if result != tc.expected {
				t.Errorf("\nExpected:\n%s\nGot:\n%s", tc.expected, result)
			}
		})
	}
}
