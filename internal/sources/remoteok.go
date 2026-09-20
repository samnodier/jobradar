package sources

import (
	"context"
	"github.com/samnodier/jobradar/internal/fetcher"
	"net/http"
	"time"
)

type RemoteOK struct {
	Client *http.Client
	URL    string
}

func (r *RemoteOK) Name() string { return "remoteok" }
func (r *RemoteOK) Fetch(ctx context.Context) ([]Job, error) {
	endpoint := r.URL
	if endpoint == "" {
		endpoint = "https://remoteok.com/api"
	}
	client := r.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	var records []fetcher.RemoteOKJob
	if err := getJSON(ctx, client, endpoint, &records); err != nil {
		return nil, err
	}
	jobs := []Job{}
	remote := true
	for _, r := range records {
		j := Job{ID: r.ID, Source: "remoteok", Title: r.Position, Company: r.Company, Description: r.Description, URL: r.URL, Location: r.Location, Logo: r.Logo, Skills: r.Tags, Remote: &remote, PostedAt: parseDate(r.Date)}
		if r.SalaryMin > 0 {
			v := int32(r.SalaryMin)
			j.SalaryMin = &v
		}
		if r.SalaryMax > 0 {
			v := int32(r.SalaryMax)
			j.SalaryMax = &v
		}
		if j.Valid() {
			jobs = append(jobs, j)
		}
	}
	return jobs, nil
}
