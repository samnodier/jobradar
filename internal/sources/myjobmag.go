package sources

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

// MyJobMag consumes the detailed Kenya RSS feed linked by /feeds/.
type MyJobMag struct {
	Country string // ke (default) or za; official public feed, never arbitrary market URLs
	Client  *http.Client
	BaseURL string
}

func (s *MyJobMag) Name() string {
	if s.Country == "za" {
		return "myjobmag-za"
	}
	return "myjobmag"
}
func (s *MyJobMag) Fetch(ctx context.Context) ([]Job, error) {
	if s.Country != "" && s.Country != "ke" && s.Country != "za" {
		return nil, fmt.Errorf("unsupported MyJobMag country")
	}
	endpoint := s.BaseURL
	if endpoint == "" {
		endpoint = "https://www.myjobmag.co.ke/jobsxml_by_categories.xml"
		if s.Country == "za" {
			endpoint = "https://www.myjobmag.co.za/jobsxml_by_categories.xml"
		}
	}
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid source endpoint")
	}
	req.Header.Set("User-Agent", "Jobradar/1.0")
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("source request failed (network or timeout)")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("source returned HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20+1))
	if err != nil {
		return nil, fmt.Errorf("read source response")
	}
	if len(data) > 16<<20 {
		return nil, fmt.Errorf("source response exceeds 16 MiB")
	}
	var feed struct {
		XMLName xml.Name `xml:"rss"`
		Channel *struct {
			Items []struct {
				ID          string `xml:"guid"`
				Title       string `xml:"position"`
				Company     string `xml:"company"`
				Location    string `xml:"location"`
				URL         string `xml:"link"`
				Description string `xml:"description"`
				Category    string `xml:"industry"`
				Date        string `xml:"pubDate"`
				Expiry      string `xml:"expiryDate"`
			} `xml:"item"`
		} `xml:"channel"`
	}
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	decoder.CharsetReader = charset.NewReaderLabel
	if err := decoder.Decode(&feed); err != nil {
		return nil, fmt.Errorf("invalid source XML: %w", err)
	}
	if feed.Channel == nil {
		return nil, fmt.Errorf("invalid source RSS: missing channel")
	}
	// Decode consumes one root element; reject additional roots or malformed tails.
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("invalid source XML: %w", err)
		}
		switch token := token.(type) {
		case xml.Comment, xml.ProcInst:
			continue
		case xml.CharData:
			if strings.TrimSpace(string(token)) == "" {
				continue
			}
		}
		return nil, fmt.Errorf("invalid source XML: content after RSS")
	}
	jobs := []Job{}
	seen := map[string]bool{}
	for _, r := range feed.Channel.Items {
		if expiry := feedDate(r.Expiry); expiry != nil && expiry.Before(time.Now()) {
			continue
		}
		id := r.ID
		if id == "" {
			id = r.URL
		}
		j := Job{ID: id, Source: s.Name(), Title: r.Title, Company: r.Company, Location: r.Location, URL: r.URL, Description: r.Description, Category: r.Category, PostedAt: feedDate(r.Date), Skills: []string{}}
		// Do not infer remote status or country from the board's audience.
		if j.Valid() && !seen[id] {
			jobs = append(jobs, j)
			seen[id] = true
		}
	}
	return jobs, nil
}

func feedDate(value string) *time.Time {
	for _, layout := range []string{time.RFC1123Z, time.RFC1123, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, strings.TrimSpace(value)); err == nil {
			return &t
		}
	}
	return nil
}
