package main

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/samnodier/jobradar/internal/auth"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/testutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAIContextAuthorizationAndSafeExport(t *testing.T) {
	p := testutil.Pool(t)
	ctx := context.Background()
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR required")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr, DB: 2})
	defer rdb.Close()
	q := database.New(p)
	owner, other, job := uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{owner, other} {
		if _, e := p.Exec(ctx, "INSERT INTO users(id,email,username,headline)VALUES($1,$2::text,$2::text,'Backend engineer')", id, id.String()+"@example.test"); e != nil {
			t.Fatal(e)
		}
	}
	t.Cleanup(func() { p.Exec(ctx, "DELETE FROM users WHERE id=ANY($1)", []uuid.UUID{owner, other}) })
	if _, e := p.Exec(ctx, "INSERT INTO jobs(id,external_id,job_source,title,company_name,description,source_url,created_by_user_id)VALUES($1,$2,'manual','Go Developer','Example','<p>Build APIs</p>','https://example.test/job',$3)", job, job.String(), owner); e != nil {
		t.Fatal(e)
	}
	h, e := auth.NewHandler(auth.AuthConfig{GitHub: auth.GitHubOAuthConfig{ClientID: "test", ClientSecret: "test", RedirectURL: "http://example.test"}, AppBaseURL: "http://example.test", SessionTTL: time.Hour}, rdb, q, p)
	if e != nil {
		t.Fatal(e)
	}
	cfg := &apiConfig{db: q, pool: p}
	router := chi.NewRouter()
	router.Use(h.RequireAuth)
	router.Get("/jobs/{jobID}/ai-context", cfg.handlerAIContext)
	router.Put("/profile/matching", cfg.handlerMatchSettingsPut)
	for _, test := range []struct {
		id   uuid.UUID
		want int
	}{{owner, 200}, {other, 404}, {uuid.Nil, 401}} {
		req := httptest.NewRequest("GET", "/jobs/"+job.String()+"/ai-context", nil)
		if test.id != uuid.Nil {
			token := uuid.NewString()
			b, _ := json.Marshal(auth.Session{UserID: test.id, Email: test.id.String() + "@example.test"})
			rdb.Set(ctx, "session:"+token, b, time.Minute)
			defer rdb.Del(ctx, "session:"+token)
			req.AddCookie(&http.Cookie{Name: "session_id", Value: token})
		}
		if test.id == owner {
			for _, body := range []string{
				`{"min_score":101,"work_mode":"remote","digest_max_jobs":10}`,
				`{"min_score":70,"work_mode":"remote","digest_max_jobs":10,"search_country":"us","digest_timezone":"Invalid/Zone","digest_hours":[8,18]}`,
				`{"min_score":70,"work_mode":"remote","digest_max_jobs":10,"search_country":"us","digest_timezone":"UTC","digest_hours":[8,8]}`,
			} {
				invalid := httptest.NewRequest("PUT", "/profile/matching", strings.NewReader(body))
				invalid.Header.Set("Cookie", req.Header.Get("Cookie"))
				response := httptest.NewRecorder()
				router.ServeHTTP(response, invalid)
				if response.Code != 400 {
					t.Fatal("invalid settings accepted", response.Code, response.Body.String())
				}
			}
			valid := httptest.NewRequest("PUT", "/profile/matching", strings.NewReader(`{"user_id":"`+other.String()+`","min_score":70,"work_mode":"remote","digest_max_jobs":10,"search_country":"us","digest_timezone":"Africa/Nairobi","digest_hours":[8,18]}`))
			valid.Header.Set("Cookie", req.Header.Get("Cookie"))
			response := httptest.NewRecorder()
			router.ServeHTTP(response, valid)
			if response.Code != 200 || !strings.Contains(response.Body.String(), owner.String()) {
				t.Fatal("settings ownership failed", response.Body.String())
			}
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != test.want {
			t.Fatal(w.Code, w.Body.String())
		}
		if test.want == 200 {
			body := w.Body.String()
			if !strings.Contains(body, "Backend engineer") || !strings.Contains(body, "Build APIs") || strings.Contains(body, "<p>") || strings.Contains(body, "@example.test") {
				t.Fatal("export missing facts or exposing contact/HTML", body)
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("export cacheable")
			}
		}
	}
}
