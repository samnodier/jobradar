package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
)

type Handler struct {
	cfg        AuthConfig
	oauthCfg   *oauth2.Config
	rdb        *redis.Client
	httpClient *http.Client
}

func NewHandler(cfg AuthConfig, rdb *redis.Client) (*Handler, error) {
	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if rdb == nil {
		return nil, fmt.Errorf("redis client is nil")
	}

	// Build the oauth config
	oauthCfg := cfg.GitHubOAuth()

	return &Handler{
		cfg:      cfg,
		oauthCfg: oauthCfg,
		rdb:      rdb,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/github/login", h.handleGitHubLogin)
	r.Get("/github/callback", h.handleGitHubCallback)

	return r
}
