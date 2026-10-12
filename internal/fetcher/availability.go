package fetcher

import (
	"context"
	"net/http"
	"net/url"
)

// CheckAvailability uses the same connection-level SSRF protection as imports,
// including redirect targets. Reachable is not evidence that hiring is still open.
func CheckAvailability(ctx context.Context, rawURL string) (string, int) {
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Hostname() == "" || parsed.User != nil {
		return "unknown", 0
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "unknown", 0
	}
	req.Header.Set("User-Agent", "Jobradar/1.0")
	resp, err := safeClient().Do(req)
	if err != nil {
		return "unknown", 0
	}
	defer resp.Body.Close()
	return availabilityFromStatus(resp.StatusCode), resp.StatusCode
}
func availabilityFromStatus(code int) string {
	if code == 404 || code == 410 {
		return "unavailable"
	}
	if code >= 200 && code < 300 {
		return "reachable"
	}
	return "unknown"
}
