package worker

import (
	"sync"

	"email-queue-service/internal/core/domain"
	"email-queue-service/internal/core/ports"
	"email-queue-service/internal/pkg/logger"
)

// WorkerPool manages a pool of concurrent workers.
type WorkerPool struct {
	numWorkers int
	jobQueue   ports.Queue // Changed to interface
	logger     *logger.Logger
	wg         sync.WaitGroup // To wait for all workers to finish
	stopChan   chan struct{}  // To signal workers to stop
}

// NewWorkerPool creates a new WorkerPool.
func NewWorkerPool(numWorkers int, jobQueue ports.Queue, l *logger.Logger) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		jobQueue:   jobQueue,
		logger:     l,
		stopChan:   make(chan struct{}),
	}
}

// Start begins the worker goroutines.
func (wp *WorkerPool) Start(processor func(domain.EmailJob)) {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i+1, processor)
	}
}

// worker is the goroutine function for each worker.
func (wp *WorkerPool) worker(id int, processor func(domain.EmailJob)) {
	defer wp.wg.Done()
	wp.logger.Printf("Worker %d started.", id)

	for {
		select {
		case <-wp.stopChan:
			// Received stop signal.
			// If the queue is closed and empty, the Dequeue will return false.
			// If the queue is still open but no more jobs are coming,
			// we should still try to drain existing jobs.
			wp.logger.Printf("Worker %d received stop signal. Attempting to drain remaining jobs...", id)
			// Fall through to continue processing any jobs left in the queue buffer (for in-memory)
			// or until Dequeue returns false (for Redis or closed in-memory).
		default:
			job, ok := wp.jobQueue.Dequeue()
			if !ok {
				// Queue is closed and empty, or an error occurred during dequeue (e.g., Redis connection lost)
				wp.logger.Printf("Worker %d detected queue closed/empty or dequeue error. Exiting.", id)
				return
			}
			processor(job)
		}
	}
}

// Stop signals all workers to stop and waits for them to finish.
func (wp *WorkerPool) Stop() {
	// Signal all workers to stop.
	close(wp.stopChan)
	wp.logger.Println("Signaled workers to stop.")

	// Wait for all workers to complete their current tasks and exit.
	wp.wg.Wait()
	wp.logger.Println("All workers have finished.")
}
