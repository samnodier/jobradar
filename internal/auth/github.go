package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
	"golang.org/x/oauth2"
)

type GitHubUser struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type GitHubEmail struct {
	Email      string `json:"email"`
	Verified   bool   `json:"verified"`
	Primary    bool   `json:"primary"`
	Visibility string `json:"visibility"`
}

func (h *Handler) handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stateValue := "login"
	if r.URL.Query().Get("link") == "1" {
		session, ok := h.recentGitHubLinkSession(r)
		if !ok {
			httpx.RespondError(w, 401, "sign in again before connecting GitHub")
			return
		}
		stateValue = "link:" + session.UserID.String()
	}
	// Generate the CSRF state
	state := generateRandomToken()

	// Store state in Redis
	err := h.rdb.Set(
		ctx,
		"oauth:github:state:"+state,
		stateValue,
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
	stateValue, err := h.rdb.GetDel(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		httpx.RespondError(w, http.StatusBadRequest, "invalid or expired state")
		return
	} else if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "invalid state")
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

	token, err := h.oauthCfg.Exchange(context.WithValue(ctx, oauth2.HTTPClient, h.httpClient), code)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to exchange code")
		return
	}

	user, err := h.fetchGitHubUser(ctx, token.AccessToken)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to fetch github user")
		return
	}

	if h.cfg.PersonalEmail != "" && !strings.EqualFold(user.Email, h.cfg.PersonalEmail) {
		httpx.RespondError(w, http.StatusForbidden, "this is a personal Jobradar instance")
		return
	}
	providerID := strconv.FormatInt(user.ID, 10)
	if strings.HasPrefix(stateValue, "link:") {
		session, ok := h.recentGitHubLinkSession(r)
		target, e := uuid.Parse(strings.TrimPrefix(stateValue, "link:"))
		if !ok || e != nil || session.UserID != target {
			httpx.RespondError(w, 401, "linking session expired; sign in again")
			return
		}
		if !strings.EqualFold(session.Email, user.Email) {
			httpx.RespondError(w, 409, "GitHub's verified primary email must match your Jobradar email")
			return
		}
		linked, e := h.db.GetUserByProviderIdentity(ctx, database.GetUserByProviderIdentityParams{AuthProvider: "github", AuthProviderID: providerID})
		if e == nil && linked.ID != target {
			httpx.RespondError(w, 409, "GitHub is already connected to another account")
			return
		}
		if e != nil && !errors.Is(e, pgx.ErrNoRows) {
			httpx.RespondError(w, 500, "could not connect GitHub")
			return
		}
		if errors.Is(e, pgx.ErrNoRows) {
			_, e = h.db.CreateUserAccount(ctx, database.CreateUserAccountParams{UserID: target, AuthProvider: "github", AuthProviderID: providerID})
			if e != nil {
				httpx.RespondError(w, 409, "this account already has a GitHub connection")
				return
			}
		}
		http.Redirect(w, r, h.cfg.AppBaseURL+"/profile", http.StatusFound)
		return
	}

	dbUser, err := h.db.GetUserByProviderIdentity(ctx, database.GetUserByProviderIdentityParams{
		AuthProvider:   "github",
		AuthProviderID: providerID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		// real DB error
		httpx.RespondError(w, http.StatusInternalServerError, "database error")
		return
	}

	if err == nil {
		// Returning user - redirect to app

		sessionID, err := h.createSession(ctx, dbUser.ID, user.Email)
		if err != nil {
			redirectWithError(w, r, "/login", "session_create_failed")
			return
		}

		h.setSessionCookie(w, sessionID)

		http.Redirect(w, r, h.cfg.AppBaseURL, http.StatusFound)
		return
	}

	// Never attach GitHub to an existing local account based on email alone.
	if _, e := h.db.GetAuthIdentityByEmail(ctx, user.Email); e == nil {
		http.Redirect(w, r, h.cfg.AppBaseURL+"/login?error=account_exists", http.StatusFound)
		return
	} else if !errors.Is(e, pgx.ErrNoRows) {
		httpx.RespondError(w, 500, "database error")
		return
	}
	// New user - store pending sign up
	pendingToken := generateRandomToken()
	signup := PendingSignup{
		GitHubID:  user.ID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
	}
	if err := h.savePendingSignup(ctx, pendingToken, signup); err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to store pending signup")
		return
	}

	http.Redirect(w, r, h.cfg.AppBaseURL+"/auth/onboarding?token="+pendingToken, http.StatusFound)
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
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api error: %d", resp.StatusCode)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	// The public profile email is not proof of verification. Always use the
	// verified primary email from GitHub's authenticated email endpoint.
	userEmail, err := h.fetchGitHubPrimaryEmail(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch verified email: %w", err)
	}
	user.Email = userEmail

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
	defer func() {
		_ = resp.Body.Close()
	}()

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

func (h *Handler) recentGitHubLinkSession(r *http.Request) (*Session, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil, false
	}
	session, err := h.getSession(r.Context(), cookie.Value)
	if err != nil || !recentSession(session) {
		return nil, false
	}
	if h.cfg.PersonalEmail != "" && !strings.EqualFold(session.Email, h.cfg.PersonalEmail) {
		return nil, false
	}
	return session, true
}
