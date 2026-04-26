package auth

import (
	"context"
	"encoding/json"
)

type PendingSignup struct {
	GitHubID    int64  `json:"github_id"`
	Login       string `json:"login"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
	AccessToken string `json:"access_token"`
}

func (h *Handler) savePendingSignup(ctx context.Context, token string, signup PendingSignup) error {
	data, err := json.Marshal(signup)
	if err != nil {
		return err
	}

	// Store in Redis with TTL 15 min
	return h.rdb.Set(ctx, "auth:pending_signup:"+token, data, h.cfg.OnboardingTTL).Err()
}

func (h *Handler) getPendingSignup(ctx context.Context, token string) (*PendingSignup, error) {
	val, err := h.rdb.Get(ctx, "auth:pending_signup:"+token).Result()
	if err != nil {
		return nil, err
	}

	var signup PendingSignup
	if err := json.Unmarshal([]byte(val), &signup); err != nil {
		return nil, err
	}
	return &signup, nil
}

func (h *Handler) deletePendingSignup(ctx context.Context, token string) error {
	return h.rdb.Del(ctx, "auth:pending_signup:"+token).Err()
}
