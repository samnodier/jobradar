package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
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
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/methods", func(w http.ResponseWriter, r *http.Request) {
		httpx.RespondJSON(w, 200, map[string]bool{"password": true, "github": h.cfg.GitHubEnabled()})
	})
	r.Post("/password/login", h.handlePasswordLogin)
	r.Post("/password/setup", h.handlePasswordSetup)

	r.Group(func(r chi.Router) {
		r.Use(h.RequireAuth)
		r.Get("/users/me", h.HandlerUserGet)
		r.Get("/password", h.handlePasswordStatus)
		r.Put("/password", h.handlePasswordChange)
	})

	if h.cfg.GitHubEnabled() {
		r.Get("/github/login", h.handleGitHubLogin)
		r.Get("/github/callback", h.handleGitHubCallback)
	}
	r.Post("/logout", h.handleLogout)

	return r
}
