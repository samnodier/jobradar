package matcher

import "testing"

func TestPreferenceGates(t *testing.T) {
	remote := true
	min := int32(50000)
	low := int32(40000)
	cases := []struct {
		name string
		p    Preferences
		j    JobFacts
		want bool
	}{
		{"unknown included", Preferences{WorkMode: "remote", IncludeUnknown: true}, JobFacts{}, true},
		{"unknown strict", Preferences{WorkMode: "remote"}, JobFacts{}, false},
		{"location boundary", Preferences{Locations: []string{"US"}}, JobFacts{Location: "Australia"}, false},
		{"Nairobi", Preferences{Locations: []string{"Nairobi"}}, JobFacts{Location: "Nairobi, Kenya"}, true},
		{"remote restricted geography", Preferences{WorkMode: "remote", Locations: []string{"Kenya"}}, JobFacts{Remote: &remote, Location: "United States only"}, false},
		{"below salary", Preferences{MinSalary: &min, Currency: "USD"}, JobFacts{Currency: "USD", SalaryMax: &low}, false},
		{"currency unknown", Preferences{MinSalary: &min, Currency: "USD", IncludeUnknown: true}, JobFacts{Currency: "KES", SalaryMax: &low}, true},
		{"employment punctuation", Preferences{JobTypes: []string{"full_time"}}, JobFacts{JobType: "Full-time"}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, _ := CheckPreferences(c.p, c.j)
			if got != c.want {
				t.Fatalf("got %v want %v", got, c.want)
			}
		})
	}
}
