// Package sources adapts external job APIs to one provider-independent record.
package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Job struct {
	ID, Source, Title, Company, Description, URL, Location, Currency, Logo string
	Category, EmploymentType                                               string
	SalaryMin, SalaryMax                                                   *int32
	Remote                                                                 *bool
	Skills                                                                 []string
	PostedAt                                                               *time.Time
}

type Source interface {
	Name() string
	Fetch(context.Context) ([]Job, error)
}

func getJSON(ctx context.Context, client *http.Client, endpoint string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("invalid source endpoint")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Jobradar/1.0")
	resp, err := client.Do(req)
	// Never wrap URL errors: Adzuna authenticates in the query string.
	if err != nil {
		return fmt.Errorf("source request failed (network or timeout)")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("source returned HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20+1))
	if err != nil {
		return fmt.Errorf("read source response")
	}
	if len(data) > 16<<20 {
		return fmt.Errorf("source response exceeds 16 MiB")
	}
	if err = json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("invalid source JSON: %w", err)
	}
	return nil
}

func validURL(s string) bool {
	u, e := url.Parse(s)
	return e == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}
func (j Job) Valid() bool {
	return j.ID != "" && strings.TrimSpace(j.Title) != "" && strings.TrimSpace(j.Company) != "" && validURL(j.URL)
}
func parseDate(s string) *time.Time {
	t, e := time.Parse(time.RFC3339, s)
	if e != nil {
		return nil
	}
	return &t
}
