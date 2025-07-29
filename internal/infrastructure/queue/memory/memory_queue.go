package memory

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"email-queue-service/internal/core/domain"
	"email-queue-service/internal/core/ports"
)

// MemoryQueue implements an in-memory job queue using a Go channel.
type MemoryQueue struct {
	Jobs     chan domain.EmailJob
	capacity int
	mu       sync.Mutex // Protects access to the channel state (e.g., closed status)
	closed   bool
	queueLengthGauge prometheus.Gauge
}

// NewMemoryQueue creates a new MemoryQueue with the given capacity.
func NewMemoryQueue(capacity int, queueLengthGauge prometheus.Gauge) *MemoryQueue {
	q := &MemoryQueue{
		Jobs:     make(chan domain.EmailJob, capacity),
		capacity: capacity,
		closed:   false,
		queueLengthGauge: queueLengthGauge,
	}
	q.queueLengthGauge.Set(0) // Initialize gauge
	return q
}

// Enqueue adds a job to the queue. Returns an error if the queue is full or closed.
func (q *MemoryQueue) Enqueue(job domain.EmailJob) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return fmt.Errorf("queue is closed, cannot enqueue new jobs")
	}

	select {
	case q.Jobs <- job:
		q.queueLengthGauge.Inc() // Increment gauge on enqueue
		return nil
	default:
		return fmt.Errorf("queue is full, cannot enqueue job")
	}
}

// Dequeue retrieves a job from the queue. Returns the job and true if successful,
// or an empty job and false if the queue is closed and empty.
func (q *MemoryQueue) Dequeue() (domain.EmailJob, bool) {
	job, ok := <-q.Jobs
	if ok {
		q.queueLengthGauge.Dec() // Decrement gauge on dequeue
	}
	return job, ok
}

// Close closes the job channel, signaling that no more jobs will be enqueued.
func (q *MemoryQueue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.closed {
		close(q.Jobs)
		q.closed = true
	}
}

// IsClosed returns true if the queue is closed.
func (q *MemoryQueue) IsClosed() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.closed
}

// Ensure MemoryQueue implements the ports.Queue interface
var _ ports.Queue = (*MemoryQueue)(nil)
