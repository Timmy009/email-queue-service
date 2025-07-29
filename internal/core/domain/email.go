package domain

import (
	"fmt"
	"net/mail"
)

// EmailJob represents an email sending task.
type EmailJob struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Retries int    `json:"retries"` // Added for retry logic
}

// Validate checks if the EmailJob fields are valid.
func (j *EmailJob) Validate() error {
	if j.To == "" {
		return fmt.Errorf("recipient 'to' field is required")
	}
	if j.Subject == "" {
		return fmt.Errorf("subject field is required")
	}
	if j.Body == "" {
		return fmt.Errorf("body field is required")
	}

	// Simple email format validation
	if _, err := mail.ParseAddress(j.To); err != nil {
		return fmt.Errorf("invalid email format for 'to' field: %w", err)
	}

	return nil
}
