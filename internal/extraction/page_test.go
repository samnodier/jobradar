package extraction

import "testing"

func TestStructuredPosting(t *testing.T) {
	draft, structured := Page(`<script type="application/ld+json">{"@graph":[{"@type":"JobPosting","title":"Go Developer","description":"<p>Build APIs</p>","hiringOrganization":{"name":"Example"},"jobLocationType":"TELECOMMUTE","applicantLocationRequirements":{"name":"Kenya"},"jobLocation":{"address":{"addressLocality":"Nairobi","addressCountry":"KE"}}}]}</script>`)
	if !structured || draft.Title != "Go Developer" || draft.CompanyName != "Example" || draft.IsRemote == nil || !*draft.IsRemote || draft.JobLocation == nil || *draft.JobLocation != "Nairobi, KE; Kenya" {
		t.Fatalf("%+v", draft)
	}
}
func TestPlainPageDoesNotInventFacts(t *testing.T) {
	draft, structured := Page(`<title>Backend role</title><script>secretScript()</script><p>Build APIs in Nairobi</p>`)
	if structured || draft.Title != "Backend role" || draft.CompanyName != "" || draft.IsRemote != nil || draft.SalaryMin != nil {
		t.Fatalf("%+v", draft)
	}
}
