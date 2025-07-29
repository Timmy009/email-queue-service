package handlers

import (
	"encoding/json"
	"net/http"

	"email-queue-service/internal/core/domain"
	"email-queue-service/internal/core/ports"
	"email-queue-service/internal/pkg/logger"
)

// EmailHandler handles HTTP requests related to emails.
type EmailHandler struct {
	emailService ports.EmailService
	logger       *logger.Logger
}

// NewEmailHandler creates a new EmailHandler.
func NewEmailHandler(es ports.EmailService, l *logger.Logger) *EmailHandler {
	return &EmailHandler{
		emailService: es,
		logger:       l,
	}
}

// SendEmail handles the POST /send-email endpoint.
func (h *EmailHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var job domain.EmailJob
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		h.logger.Errorf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := job.Validate(); err != nil {
		h.logger.Warnf("Invalid email job received: %v", err)
		http.Error(w, err.Error(), http.StatusUnprocessableEntity) // 422 Bad Request for invalid input
		return
	}

	// Initialize retries to 0 for new jobs
	job.Retries = 0

	err := h.emailService.EnqueueEmail(job)
	if err != nil {
		h.logger.Errorf("Error enqueuing email: %v", err)
		// Check if the error indicates a full queue
		if err.Error() == "failed to enqueue email: queue is full, cannot enqueue job" { // For in-memory queue
			http.Error(w, "Service Unavailable: Email queue is full", http.StatusServiceUnavailable) // 503 Service Unavailable
		} else if err.Error() == "failed to enqueue email: failed to enqueue job to Redis: redis: client is closed" { // Example for Redis
			http.Error(w, "Service Unavailable: Redis queue is unavailable", http.StatusServiceUnavailable)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted) // 202 Accepted
	w.Write([]byte("Email job enqueued successfully"))
}
