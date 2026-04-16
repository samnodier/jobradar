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
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/samnodier/jobradar/internal/auth"
	"github.com/samnodier/jobradar/internal/database"
)

type apiConfig struct {
	db *database.Queries
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
	cfg := &apiConfig{
		db: dbClient.Queries,
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

	authCfg := auth.AuthConfig{
		GitHub: auth.GitHubOAuthConfig{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
		},
		AppBaseURL:    os.Getenv("APP_BASE_URL"),
		OnboardingTTL: 10 * time.Minute,
		IsProduction:  os.Getenv("APP_ENV") == "production",
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	authHandler, err := auth.NewHandler(authCfg, rdb)
	if err != nil {
		log.Fatalf("failed to create auth handler: %v", err)
	}

	// Define routes
	router.Get("/api/health", cfg.handlerHealth)
	router.Get("/api/jobs", cfg.handlerJobsGet)
	router.Get("/api/jobs/{jobID}", cfg.handlerJobGetByID)
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
