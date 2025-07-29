package main

import (
	"context"
	"fmt"

	"net/http"

	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"email-queue-service/internal/core/ports"
	"email-queue-service/internal/core/service"
	"email-queue-service/internal/infrastructure/queue/memory"
	"email-queue-service/internal/infrastructure/queue/redis"
	"email-queue-service/internal/infrastructure/worker"
	"email-queue-service/internal/interfaces/http/v1"
	"email-queue-service/internal/interfaces/http/v1/handlers"
	"email-queue-service/internal/pkg/config"
	"email-queue-service/internal/pkg/dlq"
	"email-queue-service/internal/pkg/logger"
	"email-queue-service/internal/pkg/metrics"
	"email-queue-service/internal/pkg/shutdown"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	appLogger := logger.NewLogger()

	// Initialize metrics
	metrics.InitMetrics()
	appLogger.Println("Prometheus metrics initialized.")

	// Initialize Dead Letter Queue
	deadLetterQueue := dlq.NewInMemoryDLQ(appLogger)
	appLogger.Println("In-memory Dead Letter Queue initialized.")

	var emailQueue ports.Queue
	if cfg.UseRedisQueue {
		redisClient := redis.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
		emailQueue = redis.NewRedisQueue(redisClient, appLogger, metrics.EmailQueueLength)
		appLogger.Printf("Initialized Redis queue at %s", cfg.RedisAddr)
	} else {
		emailQueue = memory.NewMemoryQueue(cfg.QueueCapacity, metrics.EmailQueueLength)
		appLogger.Printf("Initialized in-memory queue with capacity: %d", cfg.QueueCapacity)
	}

	// Initialize email service
	emailService := service.NewEmailService(
		emailQueue,
		deadLetterQueue,
		appLogger,
		metrics.EmailJobsEnqueuedTotal,
		metrics.EmailJobsProcessedTotal,
		metrics.EmailJobsFailedTotal,
		metrics.EmailJobsRetriedTotal,
		metrics.EmailJobsDLQTotal,
		metrics.EmailProcessingDuration,
		cfg.MaxRetries,
		cfg.RetryDelaySeconds,
	)

	// Initialize worker pool
	workerPool := worker.NewWorkerPool(cfg.WorkerCount, emailQueue, appLogger)
	appLogger.Printf("Starting %d email workers...", cfg.WorkerCount)
	workerPool.Start(emailService.ProcessEmailJob) // Pass the processing function

	// Initialize HTTP handlers and routes
	emailHandler := handlers.NewEmailHandler(emailService, appLogger)
	mux := http.NewServeMux()
	v1.SetupRoutes(mux, emailHandler)

	// Add Prometheus metrics handler
	mux.Handle("/metrics", promhttp.Handler())
	appLogger.Println("Prometheus metrics endpoint exposed at /metrics")

	// Start HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: mux,
	}

	go func() {
		appLogger.Printf("HTTP server listening on port %d", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Graceful shutdown
	shutdown.WaitForShutdown(func() {
		appLogger.Println("Shutting down server...")

		// 1. Stop accepting new HTTP requests
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			appLogger.Errorf("HTTP server shutdown error: %v", err)
		} else {
			appLogger.Println("HTTP server gracefully stopped.")
		}

		// 2. Close the queue (no new jobs can be enqueued)
		emailQueue.Close()
		appLogger.Println("Email queue closed for new jobs.")

		// 3. Signal workers to stop and wait for them to finish current jobs
		workerPool.Stop()
		appLogger.Println("All workers stopped.")

		appLogger.Println("Application shutdown complete.")
	})
}
