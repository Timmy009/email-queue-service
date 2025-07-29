package ports

import "email-queue-service/internal/core/domain"

// EmailService defines the interface for managing email jobs.
type EmailService interface {
	// EnqueueEmail adds an email job to the queue.
	EnqueueEmail(job domain.EmailJob) error
	// ProcessEmailJob simulates sending an email and handles retry/DLQ logic.
	ProcessEmailJob(job domain.EmailJob)
}

// DeadLetterQueue defines the interface for storing failed jobs.
type DeadLetterQueue interface {
	Store(job domain.EmailJob, reason string)
}
