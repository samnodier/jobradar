package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"
	"github.com/samnodier/jobradar/internal/httpx"
)

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

type GitHubEmail struct {
	Email      string `json:"email"`
	Verified   bool   `json:"verified"`
	Primary    bool   `json:"primary"`
	Visibility string `json:"visibility"`
}

func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (h *Handler) handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Generate the CSRF state
	state := generateRandomState()

	// Store state in Redis
	err := h.rdb.Set(
		ctx,
		"oauth:github:state:"+state,
		"valid",
		h.cfg.OnboardingTTL,
	).Err()
	if err != nil {
		http.Error(w, "failed to store oauth state", http.StatusInternalServerError)
		return
	}

	// Store state in a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.IsProduction,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.cfg.OnboardingTTL.Seconds()),
	})

	authURL := h.oauthCfg.AuthCodeURL(state)

	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *Handler) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// read the cookie
	cookie, err := r.Cookie("oauth_state")
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "missing state cookie")
		return
	}

	// read the query state
	state := r.URL.Query().Get("state")
	if state == "" {
		httpx.RespondError(w, http.StatusBadRequest, "missing state")
		return
	}

	if state != cookie.Value {
		httpx.RespondError(w, http.StatusBadRequest, "invalid state")
		return
	}

	// check redis
	key := "oauth:github:state:" + state
	_, err = h.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		httpx.RespondError(w, http.StatusBadRequest, "invalid or expired state")
		return
	} else if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "invalid state")
		return
	}

	// Delete redis one (one-time use)
	err = h.rdb.Del(ctx, key).Err()
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to delete state")
		return
	}

	// Delete cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		httpx.RespondError(w, http.StatusBadRequest, "missing code")
		return
	}

	token, err := h.oauthCfg.Exchange(ctx, code)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to exchange code")
		return
	}

	user, err := h.fetchGitHubUser(ctx, token.AccessToken)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to fetch github user")
		return
	}

	_ = user // TODO: Store pending signup
}

func (h *Handler) fetchGitHubUser(ctx context.Context, accessToken string) (*GitHubUser, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.github.com/user",
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "jobradar")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api error: %d", resp.StatusCode)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	if user.Email == "" {
		userEmail, err := h.fetchGitHubPrimaryEmail(ctx, accessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user email: %w", err)
		}
		user.Email = userEmail
	}

	return &user, nil
}

// If fetching user fails, try fetching from http://api.github.com/user/email
func (h *Handler) fetchGitHubPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.github.com/user/emails",
		nil,
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
	req.Header.Set("User-Agent", "jobradar")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github email api error: %d", resp.StatusCode)
	}

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Verified && email.Primary {
			return email.Email, nil
		}
	}

	return "", errors.New("failed to get user email")
}
