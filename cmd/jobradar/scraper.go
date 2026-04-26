package main

import (
	"context"
	"log"
	"time"
)

func startScraping(cfg *apiConfig, interval time.Duration) {
	log.Printf("Scraping on interval: %s", interval)

	ticker := time.NewTicker(interval)

	for ; ; <-ticker.C {
		cfg.scrapeRemoteOK(context.Background())
	}
}
