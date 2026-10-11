package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samnodier/jobradar/internal/database"
)

// Job retention keeps the feed, matching work and digest focused on postings that can
// plausibly still be applied to. Old public postings are deleted; anything the user saved,
// applied to or imported is kept (see DeleteExpiredPublicJobs for why that matters).
const (
	defaultJobRetentionDays = 90
	retentionBatchSize      = 500
	retentionInterval       = time.Hour
)

// jobRetentionFromEnv reads JOB_RETENTION_DAYS. Unset means the 90-day default; 0 turns
// retention off. Invalid values fail startup rather than silently keeping or deleting data.
func jobRetentionFromEnv() (time.Duration, error) {
	raw := os.Getenv("JOB_RETENTION_DAYS")
	if raw == "" {
		return defaultJobRetentionDays * 24 * time.Hour, nil
	}
	days, err := strconv.Atoi(raw)
	if err != nil || days < 0 {
		return 0, fmt.Errorf("JOB_RETENTION_DAYS must be a whole number of days (0 disables): %q", raw)
	}
	return time.Duration(days) * 24 * time.Hour, nil
}

// retentionCutoff returns the oldest posting time still kept, and false when retention is off.
func (cfg *apiConfig) retentionCutoff(now time.Time) (time.Time, bool) {
	if cfg.jobRetention <= 0 {
		return time.Time{}, false
	}
	return now.Add(-cfg.jobRetention), true
}

// pruneOldJobs deletes expired public postings in batches until none are left.
func (cfg *apiConfig) pruneOldJobs(ctx context.Context, now time.Time) (int64, error) {
	cutoff, ok := cfg.retentionCutoff(now)
	if !ok {
		return 0, nil
	}
	return deleteExpiredJobs(ctx, cfg.db, cutoff, retentionBatchSize)
}

func deleteExpiredJobs(ctx context.Context, q *database.Queries, cutoff time.Time, batch int32) (int64, error) {
	var total int64
	for {
		n, err := q.DeleteExpiredPublicJobs(ctx, database.DeleteExpiredPublicJobsParams{
			Cutoff:    pgtype.Timestamptz{Time: cutoff, Valid: true},
			BatchSize: batch,
		})
		total += n
		if err != nil {
			return total, err
		}
		if n < int64(batch) || ctx.Err() != nil {
			return total, ctx.Err()
		}
	}
}

// maybePruneOldJobs runs retention at most once per retentionInterval from the maintenance loop.
func (cfg *apiConfig) maybePruneOldJobs(ctx context.Context, last *time.Time) {
	now := time.Now()
	if !last.IsZero() && now.Sub(*last) < retentionInterval {
		return
	}
	*last = now
	callCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	removed, err := cfg.pruneOldJobs(callCtx, now)
	if err != nil && ctx.Err() == nil {
		log.Printf("job retention: removed=%d error=%v", removed, err)
		return
	}
	if removed > 0 {
		log.Printf("job retention: removed=%d postings older than %d days", removed, int(cfg.jobRetention.Hours()/24))
	}
}
