package matcher

import "strings"

type Preferences struct {
	WorkMode                  string
	IncludeUnknown            bool
	Locations, JobTypes       []string
	Industries, CompanyStages []string
	MinSalary, MaxSalary      *int32
	Currency                  string
}
type JobFacts struct {
	Remote                           *bool
	Location, JobType, Currency      string
	Industry, Category, CompanyStage string
	SalaryMin, SalaryMax             *int32
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
			if p.MinSalary != nil && j.SalaryMax == nil && (j.SalaryMin == nil || *j.SalaryMin < *p.MinSalary) {
				unknown("Salary upper bound is unknown")
			}
			if p.MaxSalary != nil && j.SalaryMin == nil && (j.SalaryMax == nil || *j.SalaryMax > *p.MaxSalary) {
				unknown("Salary lower bound is unknown")
			}
			if p.MinSalary != nil && j.SalaryMax != nil && *j.SalaryMax < *p.MinSalary {
				reject("Salary range is below your minimum")
			}
			if p.MaxSalary != nil && j.SalaryMin != nil && *j.SalaryMin > *p.MaxSalary {
				reject("Salary range is above your maximum")
			}
		}
	}
	for _, field := range []struct {
		name, value string
		choices     []string
	}{
		{"Industry/category", strings.TrimSpace(j.Industry + " " + j.Category), p.Industries},
		{"Company stage", j.CompanyStage, p.CompanyStages},
	} {
		if len(field.choices) == 0 {
			continue
		}
		if strings.TrimSpace(field.value) == "" {
			unknown(field.name + " is not specified")
			continue
		}
		matched := false
		for _, choice := range field.choices {
			if normalize(choice) != "" && containsPhrase(field.value, choice) {
				matched = true
			}
		}
		if !matched {
			reject(field.name + " does not match your preference")
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

// PreferenceScore combines the existing role/skill/experience score (weight 80)
// with each configured, known preference (weight 5). Missing evidence has no
// numeric weight; IncludeUnknown independently controls eligibility.
func PreferenceScore(base int, p Preferences, j JobFacts) (int, bool, []string) {
	eligible, notes := CheckPreferences(p, j)
	parts := []Preferences{}
	if p.WorkMode != "" && p.WorkMode != "any" && j.Remote != nil {
		parts = append(parts, Preferences{WorkMode: p.WorkMode})
	}
	if len(p.Locations) > 0 && strings.TrimSpace(j.Location) != "" {
		parts = append(parts, Preferences{Locations: p.Locations})
	}
	if len(p.JobTypes) > 0 && j.JobType != "" {
		parts = append(parts, Preferences{JobTypes: p.JobTypes})
	}
	if (p.MinSalary != nil || p.MaxSalary != nil) && p.Currency != "" && strings.EqualFold(p.Currency, j.Currency) && j.SalaryMin != nil && j.SalaryMax != nil {
		parts = append(parts, Preferences{MinSalary: p.MinSalary, MaxSalary: p.MaxSalary, Currency: p.Currency})
	}
	if len(p.Industries) > 0 && strings.TrimSpace(j.Industry+j.Category) != "" {
		parts = append(parts, Preferences{Industries: p.Industries})
	}
	if len(p.CompanyStages) > 0 && j.CompanyStage != "" {
		parts = append(parts, Preferences{CompanyStages: p.CompanyStages})
	}
	total, weight := base*80, 80
	for _, part := range parts {
		part.IncludeUnknown = true
		matches, _ := CheckPreferences(part, j)
		if matches {
			total += 500
		}
		weight += 5
	}
	return total / weight, eligible, notes
}
