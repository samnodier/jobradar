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
	"sync"
	"testing"
	"time"
)

func TestPreferencePatchClearingAndConcurrentUpdates(t *testing.T) {
	pool := testutil.Pool(t)
	ctx := context.Background()
	q := database.New(pool)
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR required")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr, DB: 2})
	defer rdb.Close()
	id := uuid.New()
	email := id.String() + "@example.test"
	if _, err := pool.Exec(ctx, "INSERT INTO users(id,email,username,headline,min_salary,max_salary)VALUES($1,$2::text,$2::text,'Keep my profile',100,200)", id, email); err != nil {
		t.Fatal(err)
	}
	defer pool.Exec(ctx, "DELETE FROM users WHERE id=$1", id)
	token := uuid.NewString()
	data, _ := json.Marshal(auth.Session{UserID: id, Email: email, CreatedAt: time.Now().Unix()})
	if err := rdb.Set(ctx, "session:"+token, data, time.Minute).Err(); err != nil {
		t.Fatal(err)
	}
	defer rdb.Del(ctx, "session:"+token)
	h, err := auth.NewHandler(auth.AuthConfig{AppBaseURL: "http://example.test"}, rdb, q, pool)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &apiConfig{pool: pool, db: q}
	router := chi.NewRouter()
	router.Use(h.RequireAuth)
	router.Patch("/preferences", cfg.handlerUpdatePreferences)
	patch := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("PATCH", "/preferences", strings.NewReader(body))
		req.AddCookie(&http.Cookie{Name: "session_id", Value: token})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}
	w := patch(`{"min_salary":null,"max_salary":null}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	user, err := q.GetUserByID(ctx, id)
	if err != nil || user.MinSalary != nil || user.MaxSalary != nil || user.Headline == nil || *user.Headline != "Keep my profile" {
		t.Fatal("nullable update corrupted profile", err)
	}
	if patch(`{"min_salary":100,"max_salary":50}`).Code != 400 {
		t.Fatal("invalid range accepted")
	}
	if patch(`{"min_salary":"invalid"}`).Code != 400 {
		t.Fatal("invalid type accepted")
	}
	var wg sync.WaitGroup
	for _, body := range []string{`{"preferred_job_types":["Full-time","Part-time","Contract"]}`, `{"preferred_industries":["SaaS"]}`, `{"desired_locations":[{"location_name":"Nairobi, Kenya"}]}`} {
		wg.Add(1)
		go func(body string) {
			defer wg.Done()
			w := patch(body)
			if w.Code != 200 {
				t.Errorf("patch failed %d %s", w.Code, w.Body.String())
			}
		}(body)
	}
	wg.Wait()
	user, err = q.GetUserByID(ctx, id)
	if err != nil || len(user.PreferredJobTypes) != 3 || len(user.PreferredIndustries) != 1 {
		t.Fatal("concurrent partial update lost fields", err)
	}
	locations, err := q.GetUserDesiredLocations(ctx, id)
	if err != nil || len(locations) != 1 || locations[0].LocationName != "Nairobi, Kenya" {
		t.Fatal("location changed", err)
	}
	// A rejected list must roll back deletions performed earlier in the transaction.
	if patch(`{"desired_locations":[{"location_name":""}]}`).Code != 400 {
		t.Fatal("empty location accepted")
	}
	locations, err = q.GetUserDesiredLocations(ctx, id)
	if err != nil || len(locations) != 1 {
		t.Fatal("failed patch deleted prior locations", err)
	}
}
