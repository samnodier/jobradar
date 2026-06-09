package queue

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type WorkerPool struct {
	queue    *RedisQueue
	handlers map[JobType]JobHandler
	ctx      context.Context
	wg       sync.WaitGroup
	workers  int
}

func NewWorkerPool(queue *RedisQueue, workers int) *WorkerPool {
	return &WorkerPool{
		queue:    queue,
		handlers: make(map[JobType]JobHandler),
		workers:  workers,
	}
}

// RegisterHandler maps a JobType to its handler function
func (wp *WorkerPool) RegisterHandler(jobType JobType, handler JobHandler) {
	wp.handlers[jobType] = handler
}

// Start spawns the requested number of workers goroutines
// Each running workerLoop
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.ctx = ctx
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.workerLoop(ctx, i)
	}
}

// Stop waits for all worker goroutines to clean up and exit
func (wp *WorkerPool) Stop() {
	wp.wg.Wait()
	log.Println("All workers shut down cleanly.")
}

// workerLoop is an infinite loop with a select
// - either shutdown signal or dequeue
func (wp *WorkerPool) workerLoop(ctx context.Context, workerID int) {
	defer wp.wg.Done()
	log.Printf("Worker %d started", workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d stopping due to context cancellation...", workerID)
			return
		default:
			job, err := wp.queue.Dequeue(ctx, 2*time.Second)
			// handle context cancelled errors explicitly
			if err != nil {
				if errors.Is(err, context.Canceled) {
					continue
				}
				log.Printf("Worker %d: unexpected dequeue error: %v", workerID, err)
				continue
			}
			// If job is nil, loop back to check ctx.Done()
			if job == nil {
				continue
			}
			// Execute the job
			if err := wp.executeJob(ctx, job); err != nil {
				wp.handleJobFailure(job, err)
			}
		}
	}
}

// executeJob looks up the handler by job.Type in the handlers map and calls it
func (wp *WorkerPool) executeJob(ctx context.Context, job *Job) error {
	handler, ok := wp.handlers[job.Type]
	if !ok {
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
	return handler(ctx, job)
}

// handleJobFailure increments job.Attempt and
// either sends to DLQ (max retries hit)
// or schedules a retry with exponential backoff
func (wp *WorkerPool) handleJobFailure(job *Job, err error) {
	job.Attempt++
	if job.Attempt > job.MaxRetry {
		log.Printf("Job %s exceed max retries (%d). Sending to DLQ. Error: %v", job.ID, job.MaxRetry, err)
		// Use a timeout context so this doesn't hang forever if Redis is down
		dlqCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := wp.queue.EnqueueDLQ(dlqCtx, job); err != nil {
			log.Printf("CRITICAL: Failed to push job %s to DLQ. Error: %v", job.ID, err)
		}
		return
	}

	maxBackoff := float64(int(1) << job.Attempt)
	backoffSec := rand.Float64() * maxBackoff
	backoffDur := min(time.Duration(backoffSec*float64(time.Second)), 5*time.Minute)
	log.Printf("Scheduling retry %d/%d for job %s in %s", job.Attempt, job.MaxRetry, job.ID, backoffDur)

	// Tell the WorkerPool WaitGroup to track this retry goroutine
	wp.wg.Add(1)
	go func(j *Job) {
		// Decrement the WaitGroup when this goroutine finishes
		defer wp.wg.Done()

		// Create an interruptible timer instead of a blocking sleep
		timer := time.NewTimer(backoffDur)
		defer timer.Stop() // Prevents memory leak if the timer is abandoned

		select {
		case <-wp.ctx.Done():
			// Server shutdown detected
			log.Printf("Shutdown detected. Re-enqueuing job %s immediately to avoid data loss.", j.ID)
			saveCtx, saveCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer saveCancel()
			if err := wp.queue.Enqueue(saveCtx, j); err != nil {
				log.Printf("CRITICAL: Failed to save job %s during shutdown. Error: %v", j.ID, err)
			}
		case <-timer.C: // Restored: Listen to the backoff timer!
			enqCtx, enqCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer enqCancel()

			if err := wp.queue.Enqueue(enqCtx, j); err != nil {
				log.Printf("Failed to re-enqueue job %s: %v", j.ID, err)
			}
		}
	}(job)
}
