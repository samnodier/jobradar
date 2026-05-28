package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	QueueName = "jobradar:queue:main"
	DLQName   = "jobradar:queue:dlq"
)

type RedisQueue struct {
	rdb *redis.Client
}

func NewRedisQueue(rdb *redis.Client) *RedisQueue {
	return &RedisQueue{
		rdb: rdb,
	}
}

// Enqueue JSON-encodes a Job and does LPUSH (push to the left of the list)
func (rq *RedisQueue) Enqueue(ctx context.Context, job *Job) error {
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshall the job: %w", err)
	}
	err = rq.rdb.LPush(ctx, QueueName, jobBytes).Err()
	if err != nil {
		return fmt.Errorf("failed to push to redis: %w", err)
	}
	return nil
}

// EnqueueDLQ same as Enqueue but pushes to the dead letter queue
func (rq *RedisQueue) EnqueueDLQ(ctx context.Context, job *Job) error {
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshall the job: %w", err)
	}
	err = rq.rdb.LPush(ctx, DLQName, jobBytes).Err()
	if err != nil {
		return fmt.Errorf("failed to push to redis: %w", err)
	}
	return nil
}

// Dequeue does the BRPOP (pop form the right, blocking right pop up to timeout seconds)
func (rq *RedisQueue) Dequeue(ctx context.Context, timeout time.Duration) (*Job, error) {
	result, err := rq.rdb.BRPop(ctx, timeout, QueueName).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to pop from redis: %w", err)
	}
	if len(result) < 2 {
		return nil, fmt.Errorf("invalid BRPOP result")
	}
	jobString := result[1]
	var job Job
	if err := json.Unmarshal([]byte(jobString), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshall the job: %w", err)
	}
	return &job, nil
}
