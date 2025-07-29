package shutdown

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// WaitForShutdown blocks until a SIGINT or SIGTERM signal is received,
// then executes the provided cleanup functions.
func WaitForShutdown(stopFuncs ...func()) {
	// Create a channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	sig := <-sigChan
	log.Printf("Received signal: %v. Initiating graceful shutdown...", sig)

	// Execute cleanup functions in reverse order
	for i := len(stopFuncs) - 1; i >= 0; i-- {
		stopFuncs[i]()
	}
}
