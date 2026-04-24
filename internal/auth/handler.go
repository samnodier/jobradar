package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/samnodier/jobradar/internal/database"
	"golang.org/x/oauth2"
)

type Handler struct {
	cfg        AuthConfig
	oauthCfg   *oauth2.Config
	rdb        *redis.Client
	httpClient *http.Client
	db         *database.Queries
	pool       *pgxpool.Pool
}

func NewHandler(cfg AuthConfig, rdb *redis.Client, db *database.Queries, pool *pgxpool.Pool) (*Handler, error) {
	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if rdb == nil {
		return nil, fmt.Errorf("redis client is nil")
	}

	if db == nil {
		return nil, fmt.Errorf("database queries is nil")
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
		db:   db,
		pool: pool,
	}, nil
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/users/me", h.HandlerUserGet)
	r.Get("/github/login", h.handleGitHubLogin)
	r.Get("/github/callback", h.handleGitHubCallback)
	r.Get("/onboarding", h.handleOnboardingGet)
	r.Post("/onboarding", h.handleOnboardingComplete)
	r.Post("/logout", h.handleLogout)

	return r
}
