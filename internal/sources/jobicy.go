package sources

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Jobicy retains the free API's canonical listing URLs and geographic restrictions.
type Jobicy struct {
	Client  *http.Client
	BaseURL string
}

func (*Jobicy) Name() string { return "jobicy" }
func (s *Jobicy) Fetch(ctx context.Context) ([]Job, error) {
	endpoint := s.BaseURL
	if endpoint == "" {
		endpoint = "https://jobicy.com/api/v2/remote-jobs"
	}
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	var response struct {
		Jobs []struct {
			ID          int64    `json:"id"`
			URL         string   `json:"url"`
			Title       string   `json:"jobTitle"`
			Company     string   `json:"companyName"`
			Description string   `json:"jobDescription"`
			Location    string   `json:"jobGeo"`
			Category    []string `json:"jobIndustry"`
			Type        []string `json:"jobType"`
			Date        string   `json:"pubDate"`
		} `json:"jobs"`
	}
	if err := getJSON(ctx, client, endpoint+"?count=200", &response); err != nil {
		return nil, err
	}
	jobs := []Job{}
	seen := map[string]bool{}
	for _, r := range response.Jobs {
		remote := true
		date := parseDate(r.Date)
		if date == nil {
			if t, err := time.Parse("2006-01-02 15:04:05", r.Date); err == nil {
				date = &t
			}
		}
		j := Job{ID: strconv.FormatInt(r.ID, 10), Source: s.Name(), Title: r.Title, Company: r.Company, Description: r.Description, URL: r.URL, Location: r.Location, Category: strings.Join(r.Category, "; "), EmploymentType: strings.Join(r.Type, "; "), Remote: &remote, PostedAt: date, Skills: []string{}}
		// Salary intervals vary; don't compare hourly amounts to annual preferences.
		if r.ID > 0 && j.Valid() && !seen[j.ID] {
			jobs = append(jobs, j)
			seen[j.ID] = true
		}
	}
	return jobs, nil
}
