package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/samnodier/jobradar/internal/queue"
)

func (cfg *apiConfig) handleEnrichJob(ctx context.Context, qJob *queue.Job) error {
	var payload EnrichJobPayload
	if err := json.Unmarshal(qJob.Payload, &payload); err != nil {
		log.Printf("handleEnrichJob: failed to unmarshal payload: %v", err)
		return nil
	}
	log.Printf("would enrich job %s for user %s", payload.JobID, payload.UserID)
	return nil
}
