package matcher

import "testing"

func TestPreferenceRankingAndEligibility(t *testing.T) {
	p := Preferences{Locations: []string{"Nairobi"}, Industries: []string{"Finance"}, CompanyStages: []string{"Startup"}, IncludeUnknown: true}
	good := JobFacts{Location: "Nairobi, Kenya", Category: "Finance Jobs", CompanyStage: "Startup"}
	high, eligible, _ := PreferenceScore(70, p, good)
	if !eligible || high <= 70 {
		t.Fatalf("matching preferences should boost rank: %d %v", high, eligible)
	}
	bad := good
	bad.Location = "London"
	low, eligible, _ := PreferenceScore(70, p, bad)
	if eligible || low >= high {
		t.Fatalf("location must affect both rank and eligibility: %d", low)
	}
	bad = good
	bad.Category = "Hospitality Jobs"
	low, eligible, _ = PreferenceScore(70, p, bad)
	if eligible || low >= high {
		t.Fatal("category mismatch must lower rank and reject")
	}
	score, eligible, notes := PreferenceScore(70, p, JobFacts{})
	if score != 70 || !eligible || len(notes) != 3 {
		t.Fatal("unknown evidence must not fabricate a score")
	}
	p.IncludeUnknown = false
	_, eligible, _ = PreferenceScore(70, p, JobFacts{})
	if eligible {
		t.Fatal("strict unknown policy must reject missing metadata")
	}
}
