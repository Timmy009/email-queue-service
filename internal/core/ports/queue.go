package ports

import "email-queue-service/internal/core/domain"

// Queue defines the interface for a job queue.
type Queue interface {
	// Enqueue adds a job to the queue. Returns an error if the queue is full or closed.
	Enqueue(job domain.EmailJob) error
	// Dequeue retrieves a job from the queue. Returns the job and true if successful,
	// or an empty job and false if the queue is closed and empty.
	Dequeue() (domain.EmailJob, bool)
	// Close closes the job channel, signaling that no more jobs will be enqueued.
	Close()
	// IsClosed returns true if the queue is closed.
	IsClosed() bool
}
