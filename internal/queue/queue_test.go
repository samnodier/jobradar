package queue

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Helper function to get a clean Redis connection pointing to DB 1 (for testing)
func setupTestRedis(t *testing.T) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",

		DB: 1,
	})
	if err := rdb.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("failed to flush test redis: %v",
			err)
	}
	return rdb
}

func TestEnqueueDequeue(t *testing.T) {
	rdb := setupTestRedis(t)
	defer func() { _ = rdb.Close() }() // Satisfies linter check

	q := NewRedisQueue(rdb)
	ctx := context.Background()
	job := &Job{
		ID: uuid.New().String(),

		Type: "test:job",

		Payload: []byte(`{"food":"bar"}`),

		Attempt: 0,

		MaxRetry: 3,

		CreatedAt: time.Now(),
	}

	if err := q.Enqueue(ctx, job); err != nil {
		t.Fatalf("failed to enqueue job: %v", err)
	}

	dequeued, err := q.Dequeue(context.Background(),
		1*time.Second)
	if err != nil {
		t.Fatalf("Dequeue failed: %v", err)
	}

	if dequeued == nil {
		t.Fatalf("Expected job, got nil (timeout)")
	}

	if dequeued.ID != job.ID {
		t.Errorf("Expected job ID %s, got %s",
			job.ID,
			dequeued.ID)
	}

	if string(dequeued.Payload) != string(job.
		Payload) {
		t.Errorf("Expected payload %s, got %s",

			string(job.Payload), string(dequeued.Payload))
	}
}

func TestWorkerPool_Success(t *testing.T) {
	rdb := setupTestRedis(t)
	defer func() { _ = rdb.Close() }() // Satisfies linter check

	q := NewRedisQueue(rdb)
	wp := NewWorkerPool(q, 2)

	ctx, cancel := context.WithCancel(context.Background())
	// Defer order matters! cancel will run before wp.Stop
	defer wp.Stop()
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	var processedJob *Job
	wp.RegisterHandler("test:success", func(ctx context.Context, job *Job) error {
		processedJob = job

		wg.Done()

		return nil
	})

	wp.Start(ctx)

	job := &Job{
		ID: uuid.New().String(),

		Type: "test:success",

		Payload: []byte("hello"),

		Attempt: 0,

		MaxRetry: 3,

		CreatedAt: time.Now(),
	}
	if err := q.Enqueue(ctx, job); err != nil {
		t.Fatalf("failed to enqueue job: %v", err)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()

		close(done)
	}()

	select {
	case <-done:
		if processedJob.ID != job.ID {
			t.Errorf("Expected processed job ID %s, got %s", job.ID, processedJob.ID)
		}

	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for worker to process job")
	}
}

func TestWorkerPool_RetryAndDLQ(t *testing.T) {
	rdb := setupTestRedis(t)
	defer func() { _ = rdb.Close() }() // Satisfies linter check

	q := NewRedisQueue(rdb)
	wp := NewWorkerPool(q, 1)

	ctx, cancel := context.WithCancel(context.Background())
	// Defer order matters! cancel will run before wp.Stop
	defer wp.Stop()
	defer cancel()

	wp.RegisterHandler("test:fail", func(ctx context.Context, job *Job) error {
		return errors.New("something went wrong")
	})

	wp.Start(ctx)

	job := &Job{
		ID: uuid.New().String(),

		Type: "test:fail",

		Payload: []byte("will-fail"),

		Attempt: 0,

		MaxRetry: 1,

		CreatedAt: time.Now(),
	}

	if err := q.Enqueue(ctx, job); err != nil {
		t.Fatalf("failed to enqueue: %v", err)
	}

	var dlqJob *Job
	limit := time.Now().Add(5 * time.Second)

	for time.Now().Before(limit) {
		res, err := rdb.BRPop(ctx,
			500*time.Millisecond,
			DLQName).Result()
		if err == nil && len(res) >= 2 {
			var j Job
			importErr := json.Unmarshal([]byte(res[1]), &j)
			if importErr == nil {
				dlqJob = &j
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	if dlqJob == nil {
		t.Fatal("Expected job to be pushed to DLQ, but it never arrived")
	}
	if dlqJob.ID != job.ID {
		t.Errorf("Expected DLQ job ID %s, got %s", job.ID, dlqJob.ID)
	}
	if dlqJob.Attempt <= job.MaxRetry {
		t.Errorf("Expected job attempts (%d) to be greater than MaxRetry (%d) before landing in DLQ", dlqJob.Attempt, job.MaxRetry)
	}
}
