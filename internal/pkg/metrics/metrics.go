package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// EmailJobsEnqueuedTotal counts the total number of email jobs enqueued.
	EmailJobsEnqueuedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "email_jobs_enqueued_total",
		Help: "Total number of email jobs enqueued.",
	})

	// EmailJobsProcessedTotal counts the total number of email jobs successfully processed.
	EmailJobsProcessedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "email_jobs_processed_total",
		Help: "Total number of email jobs successfully processed.",
	})

	// EmailJobsFailedTotal counts the total number of email jobs that failed (including retries).
	EmailJobsFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "email_jobs_failed_total",
		Help: "Total number of email jobs that failed (including retries).",
	})

	// EmailJobsRetriedTotal counts the total number of email jobs that were retried.
	EmailJobsRetriedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "email_jobs_retried_total",
		Help: "Total number of email jobs that were retried.",
	})

	// EmailJobsDLQTotal counts the total number of email jobs moved to the Dead Letter Queue.
	EmailJobsDLQTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "email_jobs_dlq_total",
		Help: "Total number of email jobs moved to the Dead Letter Queue.",
	})

	// EmailQueueLength gauges the current number of jobs in the queue.
	EmailQueueLength = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "email_queue_length",
		Help: "Current number of jobs in the email queue.",
	})

	// EmailProcessingDuration measures the duration of email processing.
	EmailProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "email_processing_duration_seconds",
		Help:    "Duration of email processing in seconds.",
		Buckets: prometheus.DefBuckets, // Default buckets: .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
	})
)

// InitMetrics registers the metrics. This function is called once at startup.
func InitMetrics() {
	// Metrics are automatically registered with promauto.New* functions.
	// This function serves as a placeholder if manual registration was needed,
	// or to simply ensure the package is imported and its init functions run.
}
