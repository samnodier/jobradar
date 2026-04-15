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

	// Define routes
	router.Get("/api/health", cfg.handlerHealth)
	router.Get("/api/jobs", cfg.handlerJobsGet)
	router.Get("/api/jobs/{jobID}", cfg.handlerJobGetByID)

	// Start the server
	const port = ":8080"
	fmt.Printf("Server starting on port %s\n", port)

	go startScraping(cfg, 6*time.Hour)

	server := &http.Server{
		Addr:    port,
		Handler: router,
	}

	log.Fatal(server.ListenAndServe())
}
