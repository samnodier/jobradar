// Package fetcher
package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxBytes = 5 * 1024 * 1024

func Fetch(ctx context.Context, url string) ([]byte, error) {
	// Build a context-aware req so that a slow fetch dies when the request that triggered it dies
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Fetch: failed to build request: %w", err)
	}

	req.Header.Set("User-Agent", "JobRadar-Fetcher/1.0 (contact: sam@samnodier.com)")

	client := safeClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Fetch: request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Check the response status code. Anything outside 2xx is a failed fetch
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Fetch: server returned unsuccessful status: %d", resp.StatusCode)
	}

	// Read the header and reject if it isn't HTML-ish
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		return nil, fmt.Errorf("Fetch: error reading response: %w", err)
	}

	// Wrap the response body in a limited reader
	// We add +1 byte to the limit so that we can detect if the file was actually larger than 5MB
	limitedBody := io.LimitReader(resp.Body, maxBytes+1)

	data, err := io.ReadAll(limitedBody)
	if err != nil {
		return nil, fmt.Errorf("Fetch: error reading response: %w", err)
	}

	// Check if the file exceeded our 5MB limit
	if len(data) > maxBytes {
		return nil, fmt.Errorf("Fetch: response exceeded the 5MB limit")
	}

	return data, nil
}
