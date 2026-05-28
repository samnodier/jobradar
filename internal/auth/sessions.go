package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt int64     `json:"created_at"`
}

func (h *Handler) createSession(ctx context.Context, userID uuid.UUID, email string) (string, error) {
	sessionID := generateRandomToken()

	session := Session{
		UserID:    userID,
		Email:     email,
		CreatedAt: time.Now().Unix(),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return "", err
	}

	err = h.rdb.Set(
		ctx,
		"session:"+sessionID,
		data,
		h.cfg.SessionTTL,
	).Err()
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func (h *Handler) getSession(ctx context.Context, sessionID string) (*Session, error) {
	val, err := h.rdb.Get(ctx, "session:"+sessionID).Result()
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// refreshSessionIfNeeded bumbs the session TTL in Redis and the browser if the session is getting old.
func (h *Handler) refreshSessionIfNeeded(w http.ResponseWriter, ctx context.Context, sessionID string) {
	sessionKey := "session:" + sessionID

	// Check how much time is left on this session
	ttl, err := h.rdb.TTL(ctx, sessionKey).Result()
	if err != nil {
		return // Silently fail
	}

	halfTTL := h.cfg.SessionTTL / 2

	// If it is older than half its lifespan, bump it!
	if ttl < halfTTL {
		// Update Redis
		err = h.rdb.Expire(ctx, sessionKey, h.cfg.SessionTTL).Err()
		if err != nil {
			return // If Redis fails, don't update the browser cookie
		}
		// Update the Browser Cookie using helper
		h.setSessionCookie(w, sessionID)
	}
}

func (h *Handler) deleteSession(ctx context.Context, sessionID string) error {
	return h.rdb.Del(ctx, "session:"+sessionID).Err()
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.IsProduction,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.cfg.SessionTTL.Seconds()),
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.IsProduction,
		SameSite: http.SameSiteLaxMode,
	})
}
