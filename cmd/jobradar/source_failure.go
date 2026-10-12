package main

import "strings"

// Summaries expose only known categories, never provider URLs, query strings,
// credentials, response bodies, database details or arbitrary raw errors.
func sourceFailureSummary(status string, raw *string) string {
	if status == "healthy" || status == "success" || raw == nil {
		return ""
	}
	value := *raw
	for _, code := range []string{"500", "502", "503", "504"} {
		if strings.Contains(value, "HTTP "+code) {
			return "The job provider is temporarily unavailable (HTTP " + code + "). Existing jobs are kept; collection will retry on its schedule."
		}
	}
	if strings.Contains(value, "HTTP 429") {
		return "The provider rate-limited collection. Wait for the next scheduled attempt."
	}
	if strings.Contains(value, "HTTP 401") || strings.Contains(value, "HTTP 403") {
		return "The provider refused access. Check source credentials and access requirements on the server."
	}
	if strings.Contains(value, "network or timeout") {
		return "The source could not be reached before the timeout. Existing jobs are kept."
	}
	if strings.Contains(value, "add desired roles") {
		return "Add desired roles in Profile preferences before Adzuna can search."
	}
	return "The latest collection did not complete. Existing jobs are kept. Inspect the server collection log for details."
}
