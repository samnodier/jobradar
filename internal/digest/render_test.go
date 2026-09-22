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

func TestEmailCompanyLogoAndFallback(t *testing.T) {
	for _, value := range []string{"https://cdn.example.test/logo.png", "javascript:alert(1)", "http://cdn.example.test/logo.png", ""} {
		_, html, err := Render([]database.GetDigestCandidatesRow{{CompanyName: "Mind Detect", LogoUrl: &value}})
		if err != nil {
			t.Fatal(err)
		}
		hasLogo := strings.Contains(html, `src="https://cdn.example.test/logo.png"`)
		if hasLogo != (strings.HasPrefix(value, "https://")) {
			t.Fatalf("incorrect logo for %q", value)
		}
		if !strings.Contains(html, "MD") {
			t.Fatal("missing initials fallback/alt")
		}
	}
}
