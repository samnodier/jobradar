package sources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdzunaNormalizesAndDeduplicates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jobs/us/search/1" || r.URL.Query().Get("app_key") != "secret" || r.URL.Query().Get("what") != "Go Developer" {
			t.Errorf("unexpected request path or parameters")
		}
		w.Write([]byte(`{"results":[{"id":"1","title":"Go Developer","description":"Backend work","redirect_url":"https://example.test/1","created":"2026-09-20T00:00:00Z","company":{"display_name":"Example"},"location":{"display_name":"New York"},"salary_min":100000},{"id":"1","title":"Go Developer","redirect_url":"https://example.test/1","company":{"display_name":"Example"}}]}`))
	}))
	defer server.Close()
	source := &Adzuna{AppID: "id", AppKey: "secret", Country: "us", Queries: []string{"Go Developer"}, Pages: 1, BaseURL: server.URL}
	jobs, err := source.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0].Currency != "USD" || jobs[0].Remote != nil || jobs[0].PostedAt == nil || *jobs[0].SalaryMin != 100000 {
		t.Fatalf("bad normalization: %+v", jobs)
	}
}
func TestSourceErrorDoesNotLeakCredentialURL(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := getJSON(ctx, http.DefaultClient, "https://example.test?app_key=supersecret", new(any))
	if err == nil || strings.Contains(err.Error(), "supersecret") {
		t.Fatalf("credential leak: %v", err)
	}
}
func TestGreenhouseContentAndUnknownRemote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("content") != "true" {
			t.Error("description not requested")
		}
		w.Write([]byte(`{"jobs":[{"id":12,"title":"Engineer","content":"<p>Go</p>","absolute_url":"https://example.test/job","location":{"name":"Remote - US"}}]}`))
	}))
	defer server.Close()
	jobs, err := (&Greenhouse{Board: "example", BaseURL: server.URL}).Fetch(context.Background())
	if err != nil || len(jobs) != 1 || jobs[0].Remote != nil {
		t.Fatalf("unexpected jobs: %+v %v", jobs, err)
	}
}
