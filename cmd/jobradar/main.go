// Package main...
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/samnodier/jobradar/internal/auth"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/queue"
)

type apiConfig struct {
	db           *database.Queries
	rdb          *redis.Client
	pool         *pgxpool.Pool
	queue        *queue.RedisQueue
	rootCtx      context.Context
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
	q := queue.NewRedisQueue(rdb)
	cfg := &apiConfig{
		db:           dbClient.Queries,
		rdb:          rdb,
		pool:         dbClient.Pool,
		queue:        q,
		rootCtx:      context.Background(),
		IsProduction: os.Getenv("APP_ENV") == "production",
	}

	wp := queue.NewWorkerPool(q, 2)
	wp.RegisterHandler(queue.JobScrapeRemoteOK, func(ctx context.Context, job *queue.Job) error {
		return cfg.scrapeRemoteOK(ctx)
	})
	wp.RegisterHandler(queue.JobMatchJob, func(ctx context.Context, job *queue.Job) error {
		return cfg.handleMatchJob(ctx, job)
	})

	ctx, cancel := context.WithCancel(cfg.rootCtx)
	wp.Start(ctx)
	defer cancel()

	// Match threshold for AI enqueueing
	aiMatchThreshold := 60
	if val := os.Getenv("AI_MATCH_THRESHOLD"); val != "" {
		aiMatchThreshold, err = strconv.Atoi(val)
		if err != nil {
			log.Fatalf("Invalid AI_MATCH_THRESHOLD format: %v", err)
		}
	}
	aiMatchThreshold = max(0, min(100, aiMatchThreshold))

	// Start a Cron Scheduler
	scrapeInterval := 6 * time.Hour
	if val := os.Getenv("SCRAPE_INTERVAL"); val != "" {
		scrapeInterval, err = time.ParseDuration(val)
		if err != nil {
			log.Fatalf("Invalid SCRAPE_INTERVAL format: %v", err)
		}
	}
	sched := queue.NewScheduler(q)
	sched.StartCron(ctx, queue.JobScrapeRemoteOK, scrapeInterval)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	originsStr := os.Getenv("ALLOWED_ORIGINS")
	if originsStr == "" {
		originsStr = "http://localhost:5173,http://127.0.0.1:5173"
	}
	allowedOrigins := strings.Split(originsStr, ",")
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
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
		r.Use(authHandler.TryAuth)

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
			r.Patch("/applications/{id}", cfg.handlerApplicationUpdate)
			r.Delete("/applications/{id}", cfg.handlerApplicationDelete)

			r.Get("/profile/experiences", cfg.handlerGetExperiences)
			r.Post("/profile/experiences", cfg.handlerCreateExperience)
			r.Patch("/profile/experiences/{id}", cfg.handlerUpdateExperience)
			r.Delete("/profile/experiences/{id}", cfg.handlerDeleteExperience)

			r.Get("/profile/educations", cfg.handlerGetEducations)
			r.Post("/profile/educations", cfg.handlerCreateEducation)
			r.Patch("/profile/educations/{id}", cfg.handlerUpdateEducation)
			r.Delete("/profile/educations/{id}", cfg.handlerDeleteEducation)

			r.Get("/profile/projects", cfg.handlerGetProjects)
			r.Post("/profile/projects", cfg.handlerCreateProject)
			r.Patch("/profile/projects/{id}", cfg.handlerUpdateProject)
			r.Delete("/profile/projects/{id}", cfg.handlerDeleteProject)

			r.Get("/profile/certifications", cfg.handlerGetCertifications)
			r.Post("/profile/certifications", cfg.handlerCreateCertification)
			r.Patch("/profile/certifications/{id}", cfg.handlerUpdateCertification)
			r.Delete("/profile/certifications/{id}", cfg.handlerDeleteCertification)

			r.Get("/profile/languages", cfg.handlerGetLanguages)
			r.Post("/profile/languages", cfg.handlerCreateLanguage)
			r.Patch("/profile/languages/{id}", cfg.handlerUpdateLanguage)
			r.Delete("/profile/languages/{id}", cfg.handlerDeleteLanguage)

			r.Get("/profile/preferences", cfg.handlerGetPreferences)
			r.Patch("/profile/preferences", cfg.handlerUpdatePreferences)

			r.Get("/saved_jobs", cfg.handlerSavedJobsGet)
			r.Post("/saved_jobs", cfg.handlerJobSave)
			r.Delete("/saved_jobs", cfg.handlerJobUnsave)

			r.Get("/users/me", cfg.handlerUserGet)
			r.Patch("/users/me", cfg.handlerUserUpdate)
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
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start HTTP server in a separate background goroutine
	go func() {
		fmt.Printf("Server starting on port %s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Graceful shutdown wiring
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")
	// Cancel the context to stop scheduler and workers
	cancel()
	// Stop workers (blocks until all workers have finished processing their current job)
	wp.Stop()

	// Shutdown HTTP server with a 5-second timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
