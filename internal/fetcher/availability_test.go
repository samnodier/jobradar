package fetcher

import (
	"context"
	"testing"
)

func TestAvailabilityEvidence(t *testing.T) {
	for code, want := range map[int]string{200: "reachable", 204: "reachable", 404: "unavailable", 410: "unavailable", 403: "unknown", 429: "unknown", 500: "unknown"} {
		if got := availabilityFromStatus(code); got != want {
			t.Fatalf("%d: %s", code, got)
		}
	}
	for _, url := range []string{"http://127.0.0.1/private", "http://169.254.169.254/latest/meta-data/", "file:///etc/passwd", "https://user:password@example.com/job"} {
		if status, code := CheckAvailability(context.Background(), url); status != "unknown" || code != 0 {
			t.Fatalf("unsafe URL accepted: %s", url)
		}
	}
}
