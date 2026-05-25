// Package queue
package queue

import (
	"context"
	"time"
)

type JobType string

const (
	JobScrapeRemoteOK JobType = "scrape:remoteok"
)

type Job struct {
	ID        string    `json:"id"`
	Type      JobType   `json:"type"`
	Payload   []byte    `json:"payload"`
	Attempt   int       `json:"attempt"`
	MaxRetry  int       `json:"max_retry"`
	CreatedAt time.Time `json:"created_at"`
}

type JobHandler func(ctx context.Context, job *Job) error
