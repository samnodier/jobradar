package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
	"github.com/samnodier/jobradar/internal/password"
)

func NormalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > 254 {
		return "", errors.New("enter a valid email address")
	}
	return email, nil
}

// JSON and origin validation protect login and credential changes from browser CSRF.
func (h *Handler) decodePasswordRequest(w http.ResponseWriter, r *http.Request, dst any) bool {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		httpx.RespondError(w, 415, "application/json required")
		return false
	}
	if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		httpx.RespondError(w, 403, "cross-site request rejected")
		return false
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		allowed := false
		for _, candidate := range append([]string{h.cfg.AppBaseURL}, h.cfg.AllowedOrigins...) {
			u, err := url.Parse(candidate)
			if err == nil && origin == u.Scheme+"://"+u.Host {
				allowed = true
			}
		}
		if !allowed {
			httpx.RespondError(w, 403, "untrusted origin")
			return false
		}
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if decoder.Decode(dst) != nil || decoder.Decode(new(any)) != io.EOF {
		httpx.RespondError(w, 400, "invalid request body")
		return false
	}
	return true
}

const passwordRateScript = `
local a = redis.call('INCR', KEYS[1])
if a == 1 then redis.call('EXPIRE', KEYS[1], 900) end
local b = redis.call('INCR', KEYS[2])
if b == 1 then redis.call('EXPIRE', KEYS[2], 900) end
if a > 10 or b > 50 then return 0 end
return 1`

func (h *Handler) allowPasswordAttempt(w http.ResponseWriter, r *http.Request, identity string) bool {
	// RemoteAddr only: arbitrary X-Forwarded-For must not bypass throttling.
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	allowed, err := h.rdb.Eval(r.Context(), passwordRateScript, []string{"password:account:" + HashReadToken(identity), "password:ip:" + HashReadToken(ip)}).Int()
	if err != nil {
		httpx.RespondError(w, 503, "authentication temporarily unavailable")
		return false
	}
	if allowed == 0 {
		w.Header().Set("Retry-After", "900")
		httpx.RespondError(w, 429, "too many attempts; try again in 15 minutes")
		return false
	}
	return true
}

func (h *Handler) handlePasswordLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !h.decodePasswordRequest(w, r, &req) {
		return
	}
	email, err := NormalizeEmail(req.Email)
	if err != nil {
		httpx.RespondError(w, 400, err.Error())
		return
	}
	if !h.allowPasswordAttempt(w, r, email) {
		return
	}
	user, err := h.db.GetPasswordLogin(r.Context(), email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		httpx.RespondError(w, 503, "authentication temporarily unavailable")
		return
	}
	hash := user.PasswordHash
	if err != nil {
		hash = password.DummyHash
	}
	valid, verifyErr := password.Verify(req.Password, hash)
	if verifyErr != nil {
		httpx.RespondError(w, 503, "authentication temporarily unavailable")
		return
	}
	if err != nil || !valid || (h.cfg.PersonalEmail != "" && !strings.EqualFold(email, h.cfg.PersonalEmail)) {
		httpx.RespondError(w, 401, "invalid email or password")
		return
	}
	h.finishPasswordLogin(w, r, user.ID, user.Email, user.AuthVersion)
}
func (h *Handler) finishPasswordLogin(w http.ResponseWriter, r *http.Request, id uuid.UUID, email string, version int64) {
	token, err := h.createSessionVersion(r.Context(), id, email, version)
	if err != nil {
		httpx.RespondError(w, 503, "password accepted; sign in again when sessions are available")
		return
	}
	if cookie, err := r.Cookie("session_id"); err == nil {
		_ = h.deleteSession(r.Context(), cookie.Value)
	}
	h.setSessionCookie(w, token)
	httpx.RespondJSON(w, 200, map[string]string{"redirect_to": "/jobs"})
}
func (h *Handler) handlePasswordStatus(w http.ResponseWriter, r *http.Request) {
	session, _ := SessionFromContext(r.Context())
	_, err := h.db.GetPasswordHash(r.Context(), session.UserID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		httpx.RespondError(w, 500, "could not load login settings")
		return
	}
	passwordConfigured := err == nil
	githubConnected, err := h.db.HasGitHubAccount(r.Context(), session.UserID)
	if err != nil {
		httpx.RespondError(w, 500, "could not load login settings")
		return
	}
	httpx.RespondJSON(w, 200, map[string]any{"email": session.Email, "password_configured": passwordConfigured, "github_enabled": h.cfg.GitHubEnabled(), "github_connected": githubConnected})
}

// savePassword is called inside a transaction after locking the user row.
func savePassword(ctx context.Context, q *database.Queries, id uuid.UUID, hash string) (int64, error) {
	if err := q.SetPasswordHash(ctx, database.SetPasswordHashParams{UserID: id, PasswordHash: hash}); err != nil {
		return 0, err
	}
	if err := q.RevokeUserReadTokens(ctx, id); err != nil {
		return 0, err
	}
	if err := q.DeletePasswordSetupTokens(ctx, id); err != nil {
		return 0, err
	}
	return q.InvalidateUserAuthentication(ctx, id)
}
func (h *Handler) handlePasswordChange(w http.ResponseWriter, r *http.Request) {
	session, _ := SessionFromContext(r.Context())
	var req struct {
		Current  string `json:"current_password"`
		Password string `json:"password"`
	}
	if !h.decodePasswordRequest(w, r, &req) || !h.allowPasswordAttempt(w, r, session.Email) {
		return
	}
	if err := password.Validate(req.Password); err != nil {
		httpx.RespondError(w, 400, err.Error())
		return
	}
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		httpx.RespondError(w, 503, "authentication unavailable")
		return
	}
	defer tx.Rollback(r.Context())
	q := h.db.WithTx(tx)
	user, err := q.LockAuthIdentity(r.Context(), session.UserID)
	if err != nil || user.AuthVersion != session.AuthVersion {
		httpx.RespondError(w, 401, "sign in again")
		return
	}
	hash, err := q.GetPasswordHash(r.Context(), session.UserID)
	if err == nil {
		valid, e := password.Verify(req.Current, hash)
		if e != nil {
			httpx.RespondError(w, 503, "authentication unavailable")
			return
		}
		if !valid {
			httpx.RespondError(w, 401, "current password is incorrect")
			return
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		if !recentSession(session) {
			httpx.RespondError(w, 401, "sign in with GitHub again before adding a password")
			return
		}
	} else {
		httpx.RespondError(w, 503, "authentication unavailable")
		return
	}
	encoded, err := password.Hash(req.Password)
	if err != nil {
		httpx.RespondError(w, 503, "password service busy; try again")
		return
	}
	version, err := savePassword(r.Context(), q, session.UserID, encoded)
	if err != nil {
		httpx.RespondError(w, 500, "could not save password")
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		httpx.RespondError(w, 500, "could not save password")
		return
	}
	h.finishPasswordLogin(w, r, session.UserID, user.Email, version)
}
func recentSession(session *Session) bool {
	age := time.Now().Unix() - session.CreatedAt
	return session.CreatedAt > 0 && age >= 0 && age <= 600
}

func (h *Handler) handlePasswordSetup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if !h.decodePasswordRequest(w, r, &req) || !h.allowPasswordAttempt(w, r, "setup:"+HashReadToken(req.Token)) {
		return
	}
	if len(req.Token) != 67 {
		httpx.RespondError(w, 400, "invalid or expired setup link")
		return
	}
	if err := password.Validate(req.Password); err != nil {
		httpx.RespondError(w, 400, err.Error())
		return
	}
	tokenHash := HashReadToken(req.Token)
	id, err := h.db.GetPasswordSetupToken(r.Context(), tokenHash)
	if err != nil {
		httpx.RespondError(w, 400, "invalid or expired setup link")
		return
	}
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		httpx.RespondError(w, 503, "authentication unavailable")
		return
	}
	defer tx.Rollback(r.Context())
	q := h.db.WithTx(tx)
	user, err := q.LockAuthIdentity(r.Context(), id)
	if err != nil {
		httpx.RespondError(w, 400, "invalid or expired setup link")
		return
	}
	if h.cfg.PersonalEmail != "" && !strings.EqualFold(user.Email, h.cfg.PersonalEmail) {
		httpx.RespondError(w, 403, "this is a personal Jobradar instance")
		return
	}
	consumed, err := q.ConsumePasswordSetupToken(r.Context(), tokenHash)
	if err != nil || consumed != id {
		httpx.RespondError(w, 400, "invalid or expired setup link")
		return
	}
	encoded, err := password.Hash(req.Password)
	if err != nil {
		httpx.RespondError(w, 503, "password service busy; try again")
		return
	}
	version, err := savePassword(r.Context(), q, id, encoded)
	if err != nil {
		httpx.RespondError(w, 500, "could not save password")
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		httpx.RespondError(w, 500, "could not save password")
		return
	}
	h.finishPasswordLogin(w, r, id, user.Email, version)
}
