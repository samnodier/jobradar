package auth

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/testutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPasswordLifecycleAndSessionInvalidation(t *testing.T) {
	p := testutil.Pool(t)
	ctx := context.Background()
	q := database.New(p)
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR required")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr, DB: 3})
	defer rdb.Close()
	email := uuid.NewString() + "@example.test"
	defer p.Exec(ctx, "DELETE FROM users WHERE email=$1", email)
	token, err := IssuePasswordSetup(ctx, p, email)
	if err != nil {
		t.Fatal(err)
	}
	h, err := NewHandler(AuthConfig{AppBaseURL: "http://example.test", SessionTTL: time.Hour}, rdb, q, p)
	if err != nil {
		t.Fatal(err)
	}
	router := h.Routes()
	ip := uuid.NewString() + ":1234"
	send := func(method, path string, body any, cookie *http.Cookie, origin string) *httptest.ResponseRecorder {
		t.Helper()
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(method, path, strings.NewReader(string(b)))
		req.RemoteAddr = ip
		req.Header.Set("Content-Type", "application/json")
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		if cookie != nil {
			req.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}
	password := "a long original password"
	w := send("POST", "/password/setup", map[string]string{"token": token, "password": password}, nil, "http://evil.test")
	if w.Code != 403 {
		t.Fatal("CSRF accepted", w.Code)
	}
	w = send("POST", "/password/setup", map[string]string{"token": token, "password": password}, nil, "http://example.test")
	if w.Code != 200 {
		t.Fatal("setup failed", w.Code, w.Body.String())
	}
	cookie := w.Result().Cookies()[0]
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatal("unsafe cookie")
	}
	if send("POST", "/password/setup", map[string]string{"token": token, "password": password}, nil, "").Code != 400 {
		t.Fatal("setup replay accepted")
	}
	if send("GET", "/users/me", nil, cookie, "").Code != 200 {
		t.Fatal("session not usable")
	}
	if send("POST", "/password/login", map[string]string{"email": email, "password": "wrong"}, nil, "").Code != 401 {
		t.Fatal("wrong password accepted")
	}
	w = send("POST", "/password/login", map[string]string{"email": strings.ToUpper(email), "password": password}, nil, "")
	if w.Code != 200 {
		t.Fatal("case-insensitive login failed", w.Code, w.Body.String())
	}
	second := w.Result().Cookies()[0]
	w = send("PUT", "/password", map[string]string{"current_password": "wrong", "password": "a different long password"}, second, "")
	if w.Code != 401 {
		t.Fatal("current password not checked")
	}
	w = send("PUT", "/password", map[string]string{"current_password": password, "password": "a different long password"}, second, "")
	if w.Code != 200 {
		t.Fatal("change failed", w.Code, w.Body.String())
	}
	third := w.Result().Cookies()[0]
	if send("GET", "/users/me", nil, cookie, "").Code != 401 || send("GET", "/users/me", nil, second, "").Code != 401 {
		t.Fatal("old sessions survived")
	}
	if send("GET", "/users/me", nil, third, "").Code != 200 {
		t.Fatal("new session invalid")
	}
	// Password hashes must never leak through current-user responses.
	if strings.Contains(send("GET", "/users/me", nil, third, "").Body.String(), "argon2") {
		t.Fatal("credential leak")
	}
	for i := 0; i < 12; i++ {
		w = send("POST", "/password/login", map[string]string{"email": email, "password": "wrong"}, nil, "")
	}
	if w.Code != 429 {
		t.Fatal("throttling missing", w.Code)
	}
	// Expiry is enforced at redemption, and issuing a replacement invalidates older links.
	old, err := IssuePasswordSetup(ctx, p, email)
	if err != nil {
		t.Fatal(err)
	}
	fresh, err := IssuePasswordSetup(ctx, p, email)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = q.GetPasswordSetupToken(ctx, HashReadToken(old)); err == nil {
		t.Fatal("old setup link remained valid")
	}
	if _, err = p.Exec(ctx, "UPDATE password_setup_tokens SET expires_at=now()-interval '1 minute' WHERE token_hash=$1", HashReadToken(fresh)); err != nil {
		t.Fatal(err)
	}
	if send("POST", "/password/setup", map[string]string{"token": fresh, "password": password}, nil, "").Code != 400 {
		t.Fatal("expired setup accepted")
	}
}

func TestLoginSettingsGitHubConnection(t *testing.T) {
	p := testutil.Pool(t)
	ctx := context.Background()
	user, other := uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{user, other} {
		if _, err := p.Exec(ctx, "INSERT INTO users(id,email,username) VALUES($1,$2::text,$2::text)", id, id.String()+"@example.test"); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { p.Exec(ctx, "DELETE FROM users WHERE id=ANY($1)", []uuid.UUID{user, other}) })
	h := &Handler{db: database.New(p)}
	check := func(want bool) {
		t.Helper()
		req := httptest.NewRequest("GET", "/password", nil)
		req = req.WithContext(context.WithValue(ctx, sessionKey, &Session{UserID: user, Email: "test@example.test"}))
		w := httptest.NewRecorder()
		h.handlePasswordStatus(w, req)
		var body struct {
			Connected bool `json:"github_connected"`
		}
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Connected != want {
			t.Fatalf("connected=%v, want %v", body.Connected, want)
		}
	}
	if _, err := p.Exec(ctx, "INSERT INTO user_accounts(user_id,auth_provider,auth_provider_id) VALUES($1,'github',$2)", other, other.String()); err != nil {
		t.Fatal(err)
	}
	check(false)
	if _, err := p.Exec(ctx, "INSERT INTO user_accounts(user_id,auth_provider,auth_provider_id) VALUES($1,'github',$2)", user, user.String()); err != nil {
		t.Fatal(err)
	}
	check(true)
	if _, err := p.Exec(ctx, "UPDATE user_accounts SET revoked_at=now() WHERE user_id=$1", user); err != nil {
		t.Fatal(err)
	}
	check(false)
}
