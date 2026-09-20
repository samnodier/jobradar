package sources

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Greenhouse struct {
	Board   string
	Client  *http.Client
	BaseURL string
}

func (g *Greenhouse) Name() string { return "greenhouse:" + g.Board }
func (g *Greenhouse) Fetch(ctx context.Context) ([]Job, error) {
	base := g.BaseURL
	if base == "" {
		base = "https://boards-api.greenhouse.io/v1/boards"
	}
	client := g.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	var result struct {
		Jobs []struct {
			ID       int64  `json:"id"`
			Title    string `json:"title"`
			Content  string `json:"content"`
			URL      string `json:"absolute_url"`
			Company  string `json:"company_name"`
			Location struct {
				Name string `json:"name"`
			} `json:"location"`
		} `json:"jobs"`
	}
	if err := getJSON(ctx, client, fmt.Sprintf("%s/%s/jobs?content=true", base, url.PathEscape(g.Board)), &result); err != nil {
		return nil, err
	}
	jobs := []Job{}
	for _, r := range result.Jobs {
		company := r.Company
		if company == "" {
			company = g.Board
		}
		j := Job{ID: strconv.FormatInt(r.ID, 10), Source: g.Name(), Title: r.Title, Company: company, Description: r.Content, URL: r.URL, Location: r.Location.Name, Skills: []string{}}
		if j.Valid() {
			jobs = append(jobs, j)
		}
	}
	return jobs, nil
}
