// Package fether
package fetcher

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type RemoteOKJob struct {
	Slug        string   `json:"slug"`
	ID          string   `json:"id"`
	Epoch       int      `json:"epoch"`
	Date        string   `json:"date"`
	Company     string   `json:"company"`
	CompanyLogo string   `json:"company_logo"`
	Position    string   `json:"position"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	Location    string   `json:"location"`
	ApplyURL    string   `json:"apply_url"`
	SalaryMin   int      `json:"salary_min"`
	SalaryMax   int      `json:"salary_max"`
	Logo        string   `json:"logo"`
	URL         string   `json:"url"`
}

func FetchRemoteOKJobs(url string) ([]RemoteOKJob, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "JobRadar-Fetcher/1.0 (contact: sam@samnodier.com)")

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network err: %w", err)
	}

	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			log.Printf("warning: failed to close response body: %v", closeErr)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned status: %d", res.StatusCode)
	}

	var allItems []RemoteOKJob
	if err := json.NewDecoder(res.Body).Decode(&allItems); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	// Only the regal disclaimer which means the list is empty
	if len(allItems) <= 1 {
		return nil, nil
	}

	// Remove the first item (legal disclaimer)
	if len(allItems) > 0 {
		return allItems[1:], nil
	}
	return allItems, nil
}
