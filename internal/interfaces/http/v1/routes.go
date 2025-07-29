package v1

import (
	"net/http"

	"email-queue-service/internal/interfaces/http/v1/handlers"
)

// SetupRoutes registers the API routes with the given ServeMux.
func SetupRoutes(mux *http.ServeMux, emailHandler *handlers.EmailHandler) {
	mux.HandleFunc("/send-email", emailHandler.SendEmail)
}
