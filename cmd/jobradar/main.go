// Package main...
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/samnodier/jobradar/internal/auth"
	"github.com/samnodier/jobradar/internal/crypto"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/digest"
	"github.com/samnodier/jobradar/internal/llm"
	"github.com/samnodier/jobradar/internal/queue"
)

type apiConfig struct {
	db                  *database.Queries
	rdb                 *redis.Client
	pool                *pgxpool.Pool
	queue               *queue.RedisQueue
	enrichQueue         *queue.RedisQueue
	rootCtx             context.Context
	crypto              *crypto.Service
	aiMatchThreshold    int32
	aiEnrichmentEnabled bool
	newEnricher         func(ctx context.Context, provider, apiKey string) (llm.Enricher, error)
	newExtractor        func(ctx context.Context, provider, apiKey string) (llm.Extractor, error)
	IsProduction        bool
	digest              *digest.Service
}

func main() {
	preview := flag.Bool("digest-preview", false, "render a digest preview without sending")
	sendOnce := flag.Bool("digest-once", false, "send today's personal digest once")
	flag.Parse()
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("failed to load env variables: %v", err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	dbURL := os.Getenv("DB_URL")
	dbClient, err := database.NewClient(dbURL)
	if err != nil {
		log.Fatalf("failed to create a db client: %v", err)
	}

	defer dbClient.Pool.Close()
	digestConfig, emailSender, err := digest.FromEnv()
	if err != nil {
		log.Fatal(err)
	}
	digestService := &digest.Service{Pool: dbClient.Pool, Config: digestConfig, Sender: emailSender}
	if *preview {
		html, err := digestService.Preview(ctx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(html)
		return
	}
	if *sendOnce {
		if !digestConfig.Enabled {
			log.Fatal("DIGEST_ENABLED must be true to send")
		}
		if err := digestService.Tick(ctx, time.Now(), true); err != nil {
			log.Fatal(err)
		}
		log.Println("Digest cycle complete; check email_digests or the dashboard for delivery status")
		return
	}

	// Get the encryption key
	encKey := os.Getenv("ENCRYPTION_KEY")
	// Initialize the crypto service - Fail Loud
	cryptoService, err := crypto.New(encKey)
	if err != nil {
		log.Fatalf("CRITICAL: Failed to initialize crypto service: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
	defer rdb.Close()
	q := queue.NewRedisQueue(rdb)
	enrichQueue := queue.NewNamedRedisQueue(rdb, "jobradar:queue:enrich")

	// Match threshold for AI enqueueing
	var aiMatchThreshold int32 = 60
	if val := os.Getenv("AI_MATCH_THRESHOLD"); val != "" {
		var parsed int
		parsed, err = strconv.Atoi(val)
		if err != nil {
			log.Fatalf("Invalid AI_MATCH_THRESHOLD format: %v", err)
		}
		aiMatchThreshold = int32(parsed)
	}
	aiMatchThreshold = max(int32(0), min(int32(100), aiMatchThreshold))

	cfg := &apiConfig{
		db:                  dbClient.Queries,
		digest:              digestService,
		rdb:                 rdb,
		pool:                dbClient.Pool,
		queue:               q,
		enrichQueue:         enrichQueue,
		rootCtx:             ctx,
		crypto:              cryptoService,
		newEnricher:         llm.NewEnricher,
		newExtractor:        llm.NewExtractor,
		aiMatchThreshold:    aiMatchThreshold,
		aiEnrichmentEnabled: os.Getenv("AI_ENRICHMENT_ENABLED") == "true",
		IsProduction:        os.Getenv("APP_ENV") == "production",
	}

	wp := queue.NewWorkerPool(q, 2)
	wp.RegisterHandler(queue.JobMatchJob, cfg.handleDurableMatch)
	enrichPool := queue.NewWorkerPool(enrichQueue, 1)
	enrichPool.RegisterHandler(queue.JobEnrichMatch, cfg.handleEnrichJob)
	// Drain legacy enrichment tasks off the main queue without blocking matching.
	wp.RegisterHandler(queue.JobEnrichMatch, func(ctx context.Context, job *queue.Job) error { return enrichQueue.Enqueue(ctx, job) })
	sourceList, err := cfg.configuredSources()
	if err != nil {
		log.Fatal(err)
	}
	sched := queue.NewScheduler(q)
	for _, source := range sourceList {
		source := source
		jobType := queue.JobType("scrape:" + source.source.Name())
		wp.RegisterHandler(jobType, func(ctx context.Context, job *queue.Job) error { return cfg.scrapeSource(ctx, source.source) })
	}
	wp.Start(ctx)
	enrichPool.Start(ctx)
	for _, source := range sourceList {
		sched.StartCron(ctx, queue.JobType("scrape:"+source.source.Name()), source.interval)
	}
	var background sync.WaitGroup
	background.Add(1)
	go func() { defer background.Done(); cfg.runMaintenance(ctx) }()
	background.Add(1)
	go func() { defer background.Done(); digestService.Run(ctx) }()

	router := chi.NewRouter()
	router.Use(requestLog)
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
		PersonalEmail: os.Getenv("PERSONAL_EMAIL"),
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
		r.Use(privateNoCache)
		r.Use(authHandler.TryAuth)

		r.Get("/health", cfg.handlerHealth)

		r.Get("/jobs", cfg.handlerJobsGet)
		r.Get("/jobs/stats", cfg.handlerJobStatsGet)
		r.Get("/jobs/{jobID}", cfg.handlerJobGetByID)

		r.Get("/auth/onboarding", authHandler.HandleOnboardingGet)
		r.Post("/auth/onboarding", authHandler.HandleOnboardingComplete)
		r.Group(func(r chi.Router) {
			r.Use(authHandler.RequireAuth)

			r.Post("/jobs/import", cfg.handlerExtractJob)
			r.Post("/jobs/import/confirm", cfg.handlerImportConfirm)

			r.Get("/notifications", cfg.handlerNotifications)
			r.Patch("/notifications/{notificationID}/read", cfg.handlerNotificationRead)
			r.Get("/digest", cfg.handlerDigestStatus)
			r.Get("/digest/preview", cfg.handlerDigestPreview)
			r.Get("/system/status", cfg.handlerSystemStatus)
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

			r.Get("/profile/matching", cfg.handlerMatchSettingsGet)
			r.Put("/profile/matching", cfg.handlerMatchSettingsPut)
			r.Get("/profile/preferences", cfg.handlerGetPreferences)
			r.Patch("/profile/preferences", cfg.handlerUpdatePreferences)

			r.Get("/saved_jobs", cfg.handlerSavedJobsGet)
			r.Post("/saved_jobs", cfg.handlerJobSave)
			r.Delete("/saved_jobs", cfg.handlerJobUnsave)

			r.Get("/users/me", cfg.handlerUserGet)
			r.Patch("/users/me", cfg.handlerUserUpdate)
			r.Delete("/users/me", cfg.HandleDeleteAccount)
			r.Put("/users/me/api-keys/{provider}", cfg.handlerUserSetAPIKey)
			r.Delete("/users/me/api-keys/{provider}", cfg.handlerUserDeleteAPIKey)
		})
	})

	router.Mount("/auth", authHandler.Routes())
	router.NotFound(spaHandler(os.Getenv("FRONTEND_DIST")))

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      90 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start HTTP server in a separate background goroutine
	go func() {
		fmt.Printf("Server starting on port %s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Graceful shutdown wiring
	<-ctx.Done()
	log.Println("Shutting down server gracefully...")
	wp.Stop()
	enrichPool.Stop()
	background.Wait()

	// Shutdown HTTP server with a 5-second timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
