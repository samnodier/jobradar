package queue

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
)

type Scheduler struct {
	queue *RedisQueue
}

func NewScheduler(queue *RedisQueue) *Scheduler {
	return &Scheduler{
		queue: queue,
	}
}

// StartCron enqueues a jobType periodically at the specified internal
// Runs in a separate background goroutine so it does not block startup
func (s *Scheduler) StartCron(ctx context.Context, jobType JobType, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		log.Printf("Scheduler: Cron started for %s", jobType)
		s.enqueueCronJob(ctx, jobType)
		for {
			select {
			case <-ctx.Done():
				log.Printf("Scheduler: Cron stopped for %s", jobType)
				return
			case <-ticker.C:
				// Trigger the job on every ticker tick
				s.enqueueCronJob(ctx, jobType)
			}
		}
	}()
}

func (s *Scheduler) enqueueCronJob(ctx context.Context, jobType JobType) {
	job := Job{
		ID:        uuid.New().String(),
		Type:      jobType,
		Attempt:   0,
		MaxRetry:  3,
		CreatedAt: time.Now(),
	}
	if err := s.queue.Enqueue(ctx, &job); err != nil {
		log.Printf("Scheduler: failed to enqueue cron job %s: %v", jobType, err)
		return
	}
	log.Printf("Scheduler: Enqueued cron job %s (ID: %s)", jobType, job.ID)
}
