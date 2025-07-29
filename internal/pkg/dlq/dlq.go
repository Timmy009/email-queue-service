package dlq

import (
	"sync"
	"time"

	"email-queue-service/internal/core/domain"
	"email-queue-service/internal/core/ports"
	"email-queue-service/internal/pkg/logger"
)

// InMemoryDLQ implements a simple in-memory Dead Letter Queue.
type InMemoryDLQ struct {
	failedJobs []dlqEntry
	mu         sync.Mutex
	logger     *logger.Logger
}

type dlqEntry struct {
	Job       domain.EmailJob
	Reason    string
	Timestamp time.Time
}

// NewInMemoryDLQ creates a new in-memory DLQ.
func NewInMemoryDLQ(l *logger.Logger) ports.DeadLetterQueue {
	return &InMemoryDLQ{
		failedJobs: make([]dlqEntry, 0),
		logger:     l,
	}
}

// Store adds a failed job to the DLQ.
func (d *InMemoryDLQ) Store(job domain.EmailJob, reason string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	entry := dlqEntry{
		Job:       job,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	d.failedJobs = append(d.failedJobs, entry)
	d.logger.Errorf("DLQ: Stored failed job for %s (Reason: %s). Total DLQ jobs: %d", job.To, reason, len(d.failedJobs))

	// For demonstration, you might want to periodically log or inspect the DLQ
	// In a real system, this would persist to disk, a database, or another queue.
	d.logDLQContents()
}

// logDLQContents logs the current contents of the DLQ (for debugging/demonstration)
func (d *InMemoryDLQ) logDLQContents() {
	if len(d.failedJobs) > 0 {
		d.logger.Println("--- Current DLQ Contents ---")
		for i, entry := range d.failedJobs {
			d.logger.Printf("  %d. To: %s, Subject: %s, Retries: %d, Reason: %s, Time: %s",
				i+1, entry.Job.To, entry.Job.Subject, entry.Job.Retries, entry.Reason, entry.Timestamp.Format(time.RFC3339))
		}
		d.logger.Println("--------------------------")
	}
}

// Ensure InMemoryDLQ implements the ports.DeadLetterQueue interface
var _ ports.DeadLetterQueue = (*InMemoryDLQ)(nil)
