package sources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJobicyPreservesGeographyAndSkipsInvalidDuplicates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("count") != "200" {
			t.Error("missing bounded count")
		}
		w.Write([]byte(`{"jobs":[{"id":1,"jobTitle":"Engineer","companyName":"Example","jobGeo":"USA","jobIndustry":["Engineering"],"jobType":["full-time"],"url":"https://jobicy.com/jobs/1","pubDate":"2026-09-22 12:00:00","salaryMin":30,"salaryPeriod":"hourly"},{"id":1,"jobTitle":"Duplicate","companyName":"Example","url":"https://jobicy.com/jobs/1"},{"id":2,"jobTitle":"Bad URL","companyName":"Example","url":"javascript:alert(1)"},{"id":0,"jobTitle":"No ID","companyName":"Example","url":"https://jobicy.com/jobs/0"}]}`))
	}))
	defer server.Close()
	jobs, err := (&Jobicy{BaseURL: server.URL}).Fetch(context.Background())
	if err != nil || len(jobs) != 1 {
		t.Fatalf("%+v %v", jobs, err)
	}
	j := jobs[0]
	if j.Location != "USA" || j.Source != "jobicy" || j.Remote == nil || !*j.Remote || j.PostedAt == nil || j.Category != "Engineering" || j.SalaryMin != nil {
		t.Fatalf("incorrect normalization: %+v", j)
	}
}

func TestMyJobMagPreservesFieldsAndFiltersExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<?xml version="1.0" encoding="iso-8859-1"?><rss><channel>
<item><guid>1</guid><position>Engineer</position><company>Example</company><location>Nairobi</location><industry>ICT</industry><link>https://www.myjobmag.co.ke/job/1</link><pubDate>Tue, 22 Sep 2026 16:00:58 GMT</pubDate><description><![CDATA[Build <b>Go</b> services]]></description></item>
<item><guid>1</guid><position>Duplicate</position><company>Example</company><link>https://www.myjobmag.co.ke/job/1</link></item>
<item><guid>2</guid><position>Expired</position><company>Example</company><link>https://www.myjobmag.co.ke/job/2</link><expiryDate>2000-01-01</expiryDate></item>
<item><guid>3</guid><position>Missing company</position><link>https://www.myjobmag.co.ke/job/3</link></item>
</channel></rss>`))
	}))
	defer server.Close()
	jobs, err := (&MyJobMag{BaseURL: server.URL}).Fetch(context.Background())
	if err != nil || len(jobs) != 1 {
		t.Fatalf("%+v %v", jobs, err)
	}
	j := jobs[0]
	if j.Location != "Nairobi" || j.Remote != nil || j.PostedAt == nil || j.Category != "ICT" || j.Description != "Build <b>Go</b> services" {
		t.Fatalf("incorrect normalization: %+v", j)
	}
}

func TestPublicFeedsErrorsAndEmptyResponses(t *testing.T) {
	for _, kind := range []string{"jobicy", "myjobmag"} {
		for _, tc := range []struct {
			name, body string
			status     int
			wantErr    bool
		}{
			{"unavailable", "unavailable", 503, true},
			{"malformed", "not a feed", 200, true},
			{"empty", map[string]string{"jobicy": `{"jobs":[]}`, "myjobmag": `<rss><channel/></rss>`}[kind], 200, false},
			{"oversized", strings.Repeat("x", (16<<20)+1), 200, true},
		} {
			t.Run(kind+"/"+tc.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.status); w.Write([]byte(tc.body)) }))
				defer server.Close()
				var source Source = &Jobicy{BaseURL: server.URL}
				if kind == "myjobmag" {
					source = &MyJobMag{BaseURL: server.URL}
				}
				jobs, err := source.Fetch(context.Background())
				if (err != nil) != tc.wantErr || (!tc.wantErr && len(jobs) != 0) {
					t.Fatalf("jobs=%d err=%v", len(jobs), err)
				}
			})
		}
	}
}

func TestMyJobMagRejectsInvalidRSSStructure(t *testing.T) {
	for _, body := range []string{`<rss/>`, `<html><channel/></html>`, `<rss><channel/></rss><rss/>`, `<rss><channel/></rss>unexpected`} {
		t.Run(body, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) }))
			defer server.Close()
			if _, err := (&MyJobMag{BaseURL: server.URL}).Fetch(context.Background()); err == nil {
				t.Fatal("invalid feed reported as a successful empty fetch")
			}
		})
	}
}

func TestJobicyUsesBoundedCursorPagination(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls > 1 && r.URL.Query().Get("cursor") == "" {
			t.Error("cursor lost")
		}
		w.Write([]byte(`{"nextCursor":"next-token","jobs":[{"id":1,"jobTitle":"Engineer","companyName":"Example","url":"https://jobicy.com/jobs/1"}]}`))
	}))
	defer server.Close()
	jobs, err := (&Jobicy{BaseURL: server.URL}).Fetch(context.Background())
	if err != nil || len(jobs) != 1 || calls != 2 {
		t.Fatalf("dedup/repeated cursor: jobs=%d calls=%d err=%v", len(jobs), calls, err)
	}
}
