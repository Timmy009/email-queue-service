package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"

	"email-queue-service/internal/core/domain"
	"email-queue-service/internal/core/ports"
	"email-queue-service/internal/pkg/logger"
)

const (
	redisQueueKey = "email_jobs_queue"
	redisTimeout  = 5 * time.Second
)

// RedisQueue implements the ports.Queue interface using Redis LIST commands.
type RedisQueue struct {
	client           *redis.Client
	logger           *logger.Logger
	queueLengthGauge prometheus.Gauge
	closed           bool
}

// NewRedisClient creates and pings a new Redis client.
func NewRedisClient(addr, password string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	return client
}

// NewRedisQueue creates a new RedisQueue instance.
func NewRedisQueue(client *redis.Client, l *logger.Logger, queueLengthGauge prometheus.Gauge) *RedisQueue {
	q := &RedisQueue{
		client:           client,
		logger:           l,
		queueLengthGauge: queueLengthGauge,
		closed:           false,
	}
	// Initialize gauge with current queue length
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()
	length, err := client.LLen(ctx, redisQueueKey).Result()
	if err != nil {
		l.Errorf("Failed to get initial Redis queue length: %v", err)
	} else {
		q.queueLengthGauge.Set(float64(length))
	}
	return q
}

// Enqueue adds a job to the Redis queue.
func (q *RedisQueue) Enqueue(job domain.EmailJob) error {
	if q.closed {
		return fmt.Errorf("queue is closed, cannot enqueue new jobs")
	}

	jobBytes, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	// RPUSH adds the job to the tail of the list
	err = q.client.RPush(ctx, redisQueueKey, jobBytes).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue job to Redis: %w", err)
	}
	q.queueLengthGauge.Inc()
	return nil
}

// Dequeue retrieves a job from the Redis queue.
func (q *RedisQueue) Dequeue() (domain.EmailJob, bool) {
	ctx := context.Background() // Use a long-lived context for blocking pop

	// BLPOP blocks until an element is available, or timeout occurs
	// Timeout of 0 means block indefinitely
	result, err := q.client.BLPop(ctx, 0, redisQueueKey).Result()
	if err == redis.Nil {
		// No elements in the list, and BLPOP timed out (shouldn't happen with 0 timeout)
		return domain.EmailJob{}, false
	}
	if err != nil {
		q.logger.Errorf("Failed to dequeue job from Redis: %v", err)
		return domain.EmailJob{}, false
	}

	// result[0] is the key, result[1] is the value
	jobBytes := []byte(result[1])
	var job domain.EmailJob
	if err := json.Unmarshal(jobBytes, &job); err != nil {
		q.logger.Errorf("Failed to unmarshal job from Redis: %v", err)
		return domain.EmailJob{}, false
	}
	q.queueLengthGauge.Dec()
	return job, true
}

// Close closes the Redis client connection.
func (q *RedisQueue) Close() {
	if !q.closed {
		q.client.Close()
		q.closed = true
		q.logger.Println("Redis client closed.")
	}
}

// IsClosed returns true if the queue is closed.
func (q *RedisQueue) IsClosed() bool {
	return q.closed
}

// Ensure RedisQueue implements the ports.Queue interface
var _ ports.Queue = (*RedisQueue)(nil)
