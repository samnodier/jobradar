package sources

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Adzuna struct {
	AppID, AppKey, Country, Location string
	Queries                          []string
	Pages                            int
	BaseURL                          string
	Client                           *http.Client
}

func (a *Adzuna) Name() string { return "adzuna" }
func (a *Adzuna) Fetch(ctx context.Context) ([]Job, error) {
	base := a.BaseURL
	if base == "" {
		base = "https://api.adzuna.com/v1/api"
	}
	client := a.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	if a.AppID == "" || a.AppKey == "" {
		return nil, fmt.Errorf("Adzuna credentials required")
	}
	if len(a.Queries) == 0 || len(a.Queries) > 5 || a.Pages < 1 || a.Pages > 3 {
		return nil, fmt.Errorf("Adzuna requires 1–5 queries and 1–3 pages")
	}
	jobs := []Job{}
	seen := map[string]bool{}
	requests := 0
	for _, query := range a.Queries {
		for page := 1; page <= a.Pages; page++ {
			// Stay below the published 25 requests/minute, including pagination.
			if requests > 0 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(3 * time.Second):
				}
			}
			requests++
			params := url.Values{"app_id": {a.AppID}, "app_key": {a.AppKey}, "what": {query}, "where": {a.Location}, "results_per_page": {"50"}, "sort_by": {"date"}, "max_days_old": {"30"}, "content-type": {"application/json"}}
			endpoint := fmt.Sprintf("%s/jobs/%s/search/%d?%s", base, url.PathEscape(a.Country), page, params.Encode())
			var result struct {
				Results []struct {
					Category struct {
						Label string `json:"label"`
					} `json:"category"`
					ContractTime string `json:"contract_time"`
					ID           string `json:"id"`
					Title        string `json:"title"`
					Description  string `json:"description"`
					URL          string `json:"redirect_url"`
					Created      string `json:"created"`
					Company      struct {
						Name string `json:"display_name"`
					} `json:"company"`
					Location struct {
						Name string `json:"display_name"`
					} `json:"location"`
					Min *float64 `json:"salary_min"`
					Max *float64 `json:"salary_max"`
				} `json:"results"`
			}
			if err := getJSON(ctx, client, endpoint, &result); err != nil {
				return nil, fmt.Errorf("Adzuna page %s: %w", strconv.Itoa(page), err)
			}
			for _, r := range result.Results {
				if seen[r.ID] {
					continue
				}
				seen[r.ID] = true
				currency := map[string]string{"us": "USD", "gb": "GBP", "ca": "CAD", "au": "AUD", "nz": "NZD", "de": "EUR", "fr": "EUR", "nl": "EUR", "at": "EUR", "it": "EUR", "es": "EUR", "be": "EUR", "ch": "CHF", "in": "INR", "pl": "PLN", "br": "BRL", "za": "ZAR", "sg": "SGD"}[strings.ToLower(a.Country)]
				j := Job{Category: r.Category.Label, EmploymentType: strings.ReplaceAll(r.ContractTime, "_", "-"), ID: r.ID, Source: a.Name(), Title: r.Title, Company: r.Company.Name, Description: r.Description, URL: r.URL, Location: r.Location.Name, Currency: currency, PostedAt: parseDate(r.Created), Skills: []string{}}
				if r.Min != nil && *r.Min >= 0 && *r.Min < 2147483647 {
					v := int32(*r.Min)
					j.SalaryMin = &v
				}
				if r.Max != nil && *r.Max >= 0 && *r.Max < 2147483647 {
					v := int32(*r.Max)
					j.SalaryMax = &v
				}
				// A keyword hit is not proof of remote eligibility; preserve unknown.
				if j.Valid() {
					jobs = append(jobs, j)
				}
			}
			if len(result.Results) < 50 {
				break
			}
		}
	}
	return jobs, nil
}
