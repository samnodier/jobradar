package auth

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/testutil"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type githubTransport struct {
	base   http.RoundTripper
	target string
}

func (t githubTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	clone := r.Clone(r.Context())
	u := *r.URL
	target, _ := url.Parse(t.target)
	u.Scheme = target.Scheme
	u.Host = target.Host
	clone.URL = &u
	return t.base.RoundTrip(clone)
}
func TestGitHubVerifiedEmailExplicitLinkAndReturningLogin(t *testing.T) {
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
	_, err := IssuePasswordSetup(ctx, p, email)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Exec(ctx, "DELETE FROM users WHERE email=$1", email)
	owner, err := q.GetAuthIdentityByEmail(ctx, email)
	if err != nil {
		t.Fatal(err)
	}
	providerID := time.Now().UnixNano()
	verified := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/token":
			io.WriteString(w, `{"access_token":"test","token_type":"bearer"}`)
		case "/user":
			io.WriteString(w, `{"id":`+strconv.FormatInt(providerID, 10)+`,"name":"Owner","email":"unverified@example.test"}`)
		case "/user/emails":
			if verified {
				io.WriteString(w, `[{"email":"`+email+`","primary":true,"verified":true}]`)
			} else {
				io.WriteString(w, `[{"email":"`+email+`","primary":true,"verified":false}]`)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	h, err := NewHandler(AuthConfig{AppBaseURL: "http://example.test", GitHub: GitHubOAuthConfig{ClientID: "test", ClientSecret: "test", RedirectURL: "http://example.test/auth/github/callback"}}, rdb, q, p)
	if err != nil {
		t.Fatal(err)
	}
	h.httpClient = &http.Client{Transport: githubTransport{base: http.DefaultTransport, target: server.URL}}
	h.oauthCfg.Endpoint = oauth2.Endpoint{AuthURL: server.URL + "/authorize", TokenURL: server.URL + "/token", AuthStyle: oauth2.AuthStyleInParams}
	sessionID, err := h.createSession(ctx, owner.ID, email)
	if err != nil {
		t.Fatal(err)
	}
	router := h.Routes()
	cycle := func(link, session bool) *httptest.ResponseRecorder {
		path := "/github/login"
		if link {
			path += "?link=1"
		}
		start := httptest.NewRequest("GET", path, nil)
		if session {
			start.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		}
		first := httptest.NewRecorder()
		router.ServeHTTP(first, start)
		if first.Code != 302 {
			return first
		}
		loc, _ := url.Parse(first.Header().Get("Location"))
		state := loc.Query().Get("state")
		req := httptest.NewRequest("GET", "/github/callback?code=ok&state="+url.QueryEscape(state), nil)
		for _, cookie := range first.Result().Cookies() {
			req.AddCookie(cookie)
		}
		if session {
			req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		replay := httptest.NewRecorder()
		router.ServeHTTP(replay, req)
		if replay.Code != 400 {
			t.Fatal("OAuth state replay accepted")
		}
		return response
	}
	if cycle(true, false).Code != 401 {
		t.Fatal("unauthenticated linking accepted")
	}
	w := cycle(false, false)
	if w.Code != 302 || !strings.Contains(w.Header().Get("Location"), "account_exists") {
		t.Fatal("silently linked account", w.Code, w.Body.String())
	}
	verified = false
	w = cycle(true, true)
	if w.Code != 500 {
		t.Fatal("unverified email accepted", w.Code)
	}
	verified = true
	w = cycle(true, true)
	if w.Code != 302 || w.Header().Get("Location") != "http://example.test/profile" {
		t.Fatal("link failed", w.Code, w.Body.String())
	}
	w = cycle(false, false)
	if w.Code != 302 {
		t.Fatal("returning login failed", w.Code, w.Body.String())
	}
	var returningCookie *http.Cookie
	found := false
	for _, c := range w.Result().Cookies() {
		if c.Name == "session_id" {
			s, e := h.getSession(ctx, c.Value)
			if e != nil || s.UserID != owner.ID {
				t.Fatal("wrong returning account", e)
			}
			returningCookie = c
			found = true
		}
	}
	if !found {
		t.Fatal("no returning session")
	}
	// The same account can add a password after GitHub sign-in, then use email.
	body, _ := json.Marshal(map[string]string{"password": "a second way to sign in"})
	req := httptest.NewRequest("PUT", "/password", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(returningCookie)
	req.RemoteAddr = uuid.NewString() + ":1"
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatal("adding password to GitHub account failed", response.Code, response.Body.String())
	}
	body, _ = json.Marshal(map[string]string{"email": email, "password": "a second way to sign in"})
	req = httptest.NewRequest("POST", "/password/login", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = uuid.NewString() + ":1"
	response = httptest.NewRecorder()
	router.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatal("email login after GitHub failed", response.Code)
	}
	session, e := h.getSession(ctx, response.Result().Cookies()[0].Value)
	if e != nil || session.UserID != owner.ID {
		t.Fatal("login methods split the account", e)
	}

}
