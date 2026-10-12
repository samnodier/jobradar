package main

import (
	"strings"
	"testing"
)

func TestSourceFailureDoesNotExposeRawError(t *testing.T) {
	raw := "Adzuna page 1: source returned HTTP 503 secret=never-expose"
	result := sourceFailureSummary("degraded", &raw)
	if !strings.Contains(result, "temporarily unavailable") || strings.Contains(result, "secret") {
		t.Fatal(result)
	}
	if result := sourceFailureSummary("healthy", &raw); result != "" {
		t.Fatal("stale failure exposed on healthy result")
	}
	raw = "private backend error token=secret"
	if strings.Contains(sourceFailureSummary("degraded", &raw), "secret") {
		t.Fatal("raw details exposed")
	}
}
