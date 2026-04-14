package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	ctx := context.Background()
	cfg.scrapeRemoteOK(ctx)
	fmt.Println("Done!")
}
