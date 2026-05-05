// Package main...
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/samnodier/jobradar/internal/auth"
	"github.com/samnodier/jobradar/internal/database"
)

type apiConfig struct {
	db           *database.Queries
	rdb          *redis.Client
	pool         *pgxpool.Pool
	IsProduction bool
}

func main() {
	fmt.Println("jobradar starting...")
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load env variables: %v", err)
	}
	dbURL := os.Getenv("DB_URL")
	dbClient, err := database.NewClient(dbURL)
	if err != nil {
		log.Fatalf("failed to create a db client: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	cfg := &apiConfig{
		db:           dbClient.Queries,
		rdb:          rdb,
		pool:         dbClient.Pool,
		IsProduction: os.Getenv("APP_ENV") == "production",
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	onboardingTTL := 10 * time.Minute
	if val := os.Getenv("ONBOARDING_TTL"); val != "" {
		onboardingTTL, err = time.ParseDuration(val)
		if err != nil {
			log.Fatalf("Invalid ONBOARDING_TTL format: %v", err)
		}
	}
	sessionTTL := 168 * time.Hour
	if val := os.Getenv("SESSION_TTL"); val != "" {
		sessionTTL, err = time.ParseDuration(val)
		if err != nil {
			log.Fatalf("Invalid SESSION_TTL format: %v", err)
		}
	}
	authCfg := auth.AuthConfig{
		GitHub: auth.GitHubOAuthConfig{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
		},
		AppBaseURL:    os.Getenv("APP_BASE_URL"),
		OnboardingTTL: onboardingTTL,
		SessionTTL:    sessionTTL,
		IsProduction:  os.Getenv("APP_ENV") == "production",
	}

	authHandler, err := auth.NewHandler(authCfg, rdb, dbClient.Queries, dbClient.Pool)
	if err != nil {
		log.Fatalf("failed to create auth handler: %v", err)
	}

	// Define routes
	router.Route("/api", func(r chi.Router) {
		r.Get("/health", cfg.handlerHealth)

		r.Get("/jobs", cfg.handlerJobsGet)
		r.Get("/jobs/stats", cfg.handlerJobStatsGet)
		r.Get("/jobs/{jobID}", cfg.handlerJobGetByID)

		r.Get("/auth/onboarding", authHandler.HandleOnboardingGet)
		r.Post("/auth/onboarding", authHandler.HandleOnboardingComplete)
		r.Group(func(r chi.Router) {
			r.Use(authHandler.RequireAuth)

			r.Get("/applications", cfg.handlerApplicationsGet)
			r.Post("/applications", cfg.handlerApplicationCreate)
			r.Get("/applications/{applicationID}", cfg.handlerApplicationGetByID)

			r.Post("/saved_jobs", cfg.handlerJobSave)

			r.Delete("/users/me", cfg.HandleDeleteAccount)
		})
	})

	router.Mount("/auth", authHandler.Routes())

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	fmt.Printf("Server starting on port %s\n", port)

	go startScraping(cfg, 6*time.Hour)

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	log.Fatal(server.ListenAndServe())
}
