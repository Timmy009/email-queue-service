package service

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"email-queue-service/internal/core/domain"
	"email-queue-service/internal/core/ports"
	"email-queue-service/internal/pkg/logger"
)

// emailService implements the ports.EmailService interface.
type emailService struct {
	queue                   ports.Queue
	dlq                     ports.DeadLetterQueue
	logger                  *logger.Logger
	enqueuedCounter         prometheus.Counter
	processedCounter        prometheus.Counter
	failedCounter           prometheus.Counter
	retriedCounter          prometheus.Counter
	dlqCounter              prometheus.Counter
	processingDurationGauge prometheus.Histogram
	maxRetries              int
	retryDelaySeconds       int
}

// NewEmailService creates a new EmailService instance.
func NewEmailService(
	q ports.Queue,
	dlq ports.DeadLetterQueue,
	l *logger.Logger,
	enqueued prometheus.Counter,
	processed prometheus.Counter,
	failed prometheus.Counter,
	retried prometheus.Counter,
	dlqCount prometheus.Counter,
	processingDuration prometheus.Histogram,
	maxRetries int,
	retryDelaySeconds int,
) ports.EmailService {
	return &emailService{
		queue:                   q,
		dlq:                     dlq,
		logger:                  l,
		enqueuedCounter:         enqueued,
		processedCounter:        processed,
		failedCounter:           failed,
		retriedCounter:          retried,
		dlqCounter:              dlqCount,
		processingDurationGauge: processingDuration,
		maxRetries:              maxRetries,
		retryDelaySeconds:       retryDelaySeconds,
	}
}

// EnqueueEmail adds an email job to the queue.
func (s *emailService) EnqueueEmail(job domain.EmailJob) error {
	err := s.queue.Enqueue(job)
	if err != nil {
		s.logger.Errorf("Failed to enqueue email job: %v", err)
		s.failedCounter.Inc() // Increment failed counter if enqueue fails
		return fmt.Errorf("failed to enqueue email: %w", err)
	}
	s.logger.Printf("Enqueued email job for %s (retries: %d)", job.To, job.Retries)
	s.enqueuedCounter.Inc()
	return nil
}

// ProcessEmailJob simulates sending an email and handles retry/DLQ logic.
func (s *emailService) ProcessEmailJob(job domain.EmailJob) {
	s.logger.Printf("Processing email to: %s, Subject: %s (Attempt: %d)", job.To, job.Subject, job.Retries+1)
	start := time.Now()

	// Simulate external email sending service call with a chance of failure
	time.Sleep(1 * time.Second) // Simulate work
	success := rand.Intn(100) < 80 // 80% success rate for demonstration

	if success {
		s.logger.Printf("Successfully sent email to: %s", job.To)
		s.processedCounter.Inc()
		s.processingDurationGauge.Observe(time.Since(start).Seconds())
	} else {
		s.logger.Warnf("Failed to send email to: %s (Attempt: %d)", job.To, job.Retries+1)
		s.failedCounter.Inc()
		s.processingDurationGauge.Observe(time.Since(start).Seconds())

		if job.Retries < s.maxRetries {
			job.Retries++
			s.retriedCounter.Inc()
			s.logger.Printf("Retrying email to: %s in %d seconds (Attempt: %d/%d)", job.To, s.retryDelaySeconds, job.Retries+1, s.maxRetries+1)
			// Simulate delay before re-enqueuing for retry
			time.AfterFunc(time.Duration(s.retryDelaySeconds)*time.Second, func() {
				if err := s.queue.Enqueue(job); err != nil {
					s.logger.Errorf("Failed to re-enqueue email for retry to %s: %v", job.To, err)
					s.dlq.Store(job, fmt.Sprintf("Failed to re-enqueue after %d retries: %v", job.Retries, err))
					s.dlqCounter.Inc()
				}
			})
		} else {
			s.logger.Errorf("Email to %s permanently failed after %d retries. Moving to DLQ.", job.To, job.Retries)
			s.dlq.Store(job, fmt.Sprintf("Permanently failed after %d retries", job.Retries))
			s.dlqCounter.Inc()
		}
	}
}
