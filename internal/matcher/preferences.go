package matcher

import "strings"

type Preferences struct {
	WorkMode             string
	IncludeUnknown       bool
	Locations, JobTypes  []string
	MinSalary, MaxSalary *int32
	Currency             string
}
type JobFacts struct {
	Remote                      *bool
	Location, JobType, Currency string
	SalaryMin, SalaryMax        *int32
}

// Eligibility is a separate gate, never a fabricated AI score or a claim of work authorization.
func CheckPreferences(p Preferences, j JobFacts) (bool, []string) {
	eligible := true
	notes := []string{}
	unknown := func(message string) {
		notes = append(notes, message)
		if !p.IncludeUnknown {
			eligible = false
		}
	}
	reject := func(message string) { notes = append(notes, message); eligible = false }
	if p.WorkMode != "any" && p.WorkMode != "" {
		if j.Remote == nil {
			unknown("Work mode is not specified")
		} else if (p.WorkMode == "remote") != *j.Remote {
			reject("Work mode does not match your preference")
		}
	}
	locations := []string{}
	for _, v := range p.Locations {
		if v != "" && !strings.EqualFold(v, "remote") && !strings.EqualFold(v, "on-site") && !strings.EqualFold(v, "hybrid") {
			locations = append(locations, v)
		}
	}
	if len(locations) > 0 {
		if strings.TrimSpace(j.Location) == "" {
			unknown("Job location is not specified")
		} else {
			matched := false
			for _, v := range locations {
				if containsPhrase(j.Location, v) {
					matched = true
				}
			}
			if !matched {
				reject("Listed location does not match your desired locations")
			}
		}
	}
	if len(p.JobTypes) > 0 {
		if j.JobType == "" {
			unknown("Employment type is not specified")
		} else {
			match := false
			for _, v := range p.JobTypes {
				if normalize(v) == normalize(j.JobType) {
					match = true
				}
			}
			if !match {
				reject("Employment type does not match your preference")
			}
		}
	}
	if p.MinSalary != nil || p.MaxSalary != nil {
		if j.Currency == "" || p.Currency == "" || !strings.EqualFold(j.Currency, p.Currency) || (j.SalaryMin == nil && j.SalaryMax == nil) {
			unknown("Salary is missing or uses a different currency; no currency conversion was applied")
		} else {
			if p.MinSalary != nil && j.SalaryMax != nil && *j.SalaryMax < *p.MinSalary {
				reject("Salary range is below your minimum")
			}
			if p.MaxSalary != nil && j.SalaryMin != nil && *j.SalaryMin > *p.MaxSalary {
				reject("Salary range is above your maximum")
			}
		}
	}
	return eligible, notes
}
func normalize(s string) string {
	return strings.Join(strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r > 127) }), " ")
}
func containsPhrase(text, phrase string) bool {
	return strings.Contains(" "+normalize(text)+" ", " "+normalize(phrase)+" ")
}
