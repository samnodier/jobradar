package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/queue"
)

func (cfg *apiConfig) dispatchPendingMatches(ctx context.Context) error {
	rows, err := cfg.db.ListPendingMatches(ctx)
	if err != nil {
		return err
	}
	for _, row := range rows {
		body, _ := json.Marshal(MatchJobPayload{JobID: row.JobID.String(), UserID: row.UserID.String(), Revision: row.Revision})
		err = cfg.queue.Enqueue(ctx, &queue.Job{ID: fmt.Sprintf("%s-%s-%d", row.JobID, row.UserID, row.Revision), Type: queue.JobMatchJob, Payload: body, MaxRetry: 3})
		if err != nil {
			return err
		}
		// Mark only AFTER enqueue. A crash between these steps permits a duplicate,
		// which the revision check handles; reversing the order could delay recovery.
		if err := cfg.db.MarkMatchEnqueued(ctx, database.MarkMatchEnqueuedParams{UserID: row.UserID, JobID: row.JobID, Revision: row.Revision}); err != nil {
			return err
		}
	}
	return nil
}

func (cfg *apiConfig) runMaintenance(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	var lastPrune time.Time
	for {
		cfg.maybePruneOldJobs(ctx, &lastPrune)
		callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		if err := cfg.dispatchPendingMatches(callCtx); err != nil && ctx.Err() == nil {
			log.Printf("pending match dispatch: %v", err)
		}
		cancel()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (cfg *apiConfig) handleDurableMatch(ctx context.Context, job *queue.Job) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var payload MatchJobPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return nil
	}
	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return nil
	}
	jobID, err := uuid.Parse(payload.JobID)
	if err != nil {
		return nil
	}
	// Serialize duplicate deliveries for a pair without blocking profile edits.
	tx, err := cfg.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1, 0))", payload.UserID+payload.JobID); err != nil {
		return err
	}
	// Keep reads, score writes and acknowledgement on this connection. Besides
	// atomic completion, this avoids workers exhausting a small pool while each
	// holds one connection and waits for a second.
	worker := *cfg
	worker.db = cfg.db.WithTx(tx)
	revision, err := worker.db.GetPendingMatchRevision(ctx, database.GetPendingMatchRevisionParams{UserID: userID, JobID: jobID})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if payload.Revision != 0 && revision != payload.Revision {
		return nil
	}
	if err = worker.handleMatchJob(ctx, job); err != nil {
		return err
	}
	if err = worker.db.CompletePendingMatch(ctx, database.CompletePendingMatchParams{UserID: userID, JobID: jobID, Revision: revision}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
