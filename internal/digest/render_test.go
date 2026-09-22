package digest

import (
	"github.com/samnodier/jobradar/internal/database"
	"strings"
	"testing"
)

func TestEmailUsesTablesAndEscapesData(t *testing.T) {
	score := int32(78)
	_, html, err := Render([]database.GetDigestCandidatesRow{{Title: "<script>alert(1)</script>", CompanyName: "Mind Detect", SourceUrl: "javascript:alert(1)", MatchScore: &score}}, RenderOptions{BaseURL: "https://jobs.example.test", Schedule: "07:00 · America/New_York"})
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"<script", "display:flex", "display:grid", "javascript:", "Quick apply", "Actively hiring", "example.com/jobs"} {
		if strings.Contains(html, bad) {
			t.Fatalf("unsafe or fabricated content %q", bad)
		}
	}
	for _, want := range []string{`role="presentation"`, "MD", "78% match", "https://jobs.example.test/profile", "07:00", "&lt;script&gt;"} {
		if !strings.Contains(html, want) {
			t.Fatalf("missing %q", want)
		}
	}
}
