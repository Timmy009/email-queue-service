# Email Queue Microservice

This project implements a simple microservice in Go that accepts email jobs over HTTP, queues them, and processes them asynchronously using a worker system.

## Features

- **HTTP API**: Exposes a `POST /send-email` endpoint for enqueuing email jobs.
- **Pluggable Job Queue**: Supports both in-memory (Go channels) and Redis-backed queues.
- **Concurrent Workers**: Processes jobs asynchronously using multiple goroutine workers.
- **Simulated Email Sending**: Logs the email content and simulates a delay with a chance of failure.
- **Retry Logic**: Failed jobs are retried up to a configurable number of times with a delay.
- **Dead Letter Queue (DLQ)**: Permanently failed jobs (after exhausting retries) are moved to an in-memory DLQ for inspection.
- **Prometheus Metrics**: Exposes a `/metrics` endpoint with key operational metrics (queue length, jobs processed, failed, retried, DLQ).
- **Graceful Shutdown**: Handles `SIGINT` and `SIGTERM` signals to stop accepting new requests, drain the queue, and wait for active workers to finish.
- **Configurable**: Number of workers, queue capacity, HTTP port, retry settings, and queue type (in-memory/Redis) are configurable via environment variables.

## Project Structure

The project adheres to a clean, modular structure to promote maintainability and extensibility:

## Requirements

- Go 1.22+
- Docker (optional, for containerized deployment)
- Redis (required if `USE_REDIS_QUEUE=true`)

## How to Run

### 1. Build and Run Locally

1.  **Navigate to the project directory:**
    \`\`\`bash
    cd email-queue-service
    \`\`\`
    (If your directory is still named `email-queue-service (6)`, you might need to rename it to `email-queue-service` first: `mv "email-queue-service (6)" email-queue-service`)
2.  **Initialize Go module:**
    \`\`\`bash
    go mod init email-queue-service
    \`\`\`
    _(If `go.mod` already exists, this command might say it already exists. That's fine, just proceed to the next step.)_
3.  **Download dependencies and tidy module:**
    \`\`\`bash
    go mod tidy
    \`\`\`
    **This step is absolutely crucial.** It will now correctly resolve all external dependencies (Prometheus, Redis) and update your `go.sum` file based on the corrected import paths.
4.  **Build the application:**
    \`\`\`bash
    go build -o bin/email-service ./cmd/email-service
    \`\`\`
5.  **Run the application:**

    **Using In-Memory Queue (Default):**
    \`\`\`bash

    # Using default configuration

    go run ./cmd/email-service
    ./bin/email-service

    # Or, with custom configuration via environment variables

    HTTP_PORT=8080 WORKER_COUNT=5 QUEUE_CAPACITY=200 MAX_RETRIES=5 RETRY_DELAY_SECONDS=10 ./bin/email-service
    \`\`\`

    **Using Redis Queue:**
    First, ensure you have a Redis instance running. You can start one quickly with Docker:
    \`\`\`bash
    docker run --name my-redis -p 6379:6379 -d redis/redis-stack-server:latest
    \`\`\`
    Then, run the application with Redis enabled:
    \`\`\`bash
    USE_REDIS_QUEUE=true REDIS_ADDR=localhost:6379 HTTP_PORT=8080 WORKER_COUNT=5 MAX_RETRIES=3 RETRY_DELAY_SECONDS=5 ./bin/email-service
    \`\`\`
    If your Redis requires a password or different DB:
    \`\`\`bash
    USE_REDIS_QUEUE=true REDIS_ADDR=localhost:6379 REDIS_PASSWORD=your_password REDIS_DB=1 HTTP_PORT=8080 WORKER_COUNT=5 ./bin/email-service
    \`\`\`

### 2. Run with Docker

1.  **Build the Docker image:**
    \`\`\`bash
    docker build -t email-queue-service .
    \`\`\`
2.  **Run the Docker container (In-Memory Queue):**
    \`\`\`bash
    docker run -p 8080:8080 -e WORKER_COUNT=5 -e QUEUE_CAPACITY=200 -e MAX_RETRIES=3 -e RETRY_DELAY_SECONDS=5 email-queue-service
    \`\`\`
3.  **Run the Docker container (Redis Queue):**
    First, ensure your Redis container is running and accessible from the network where your `email-queue-service` container will run. If running both on the same Docker network (e.g., via `docker-compose`), you can use the service name.
    \`\`\`bash
    # Example if Redis is running as 'my-redis' on the same network
    docker run -p 8080:8080 \
     -e USE_REDIS_QUEUE=true \
     -e REDIS_ADDR=host.docker.internal:6379 \ # Use host.docker.internal to reach host's Redis from container
    -e WORKER_COUNT=5 \
     -e MAX_RETRIES=3 \
     -e RETRY_DELAY_SECONDS=5 \
     email-queue-service
    \`\`\`

## API Endpoints

### `POST /send-email`

Enqueues an email job for asynchronous processing.

**Request Body:**

\`\`\`json
{
"to": "recipient@example.com",
"subject": "Your Subject Here",
"body": "This is the body of your email."
}
\`\`\`

**Headers:**

`Content-Type: application/json`

**Responses:**

- **`202 Accepted`**: Email job successfully enqueued.
  \`\`\`
  Email job enqueued successfully
  \`\`\`
- **`422 Unprocessable Entity`**: Invalid input (e.g., missing fields, invalid email format).
  \`\`\`
  recipient 'to' field is required
  \`\`\`
  or
  \`\`\`
  invalid email format for 'to' field: mail: missing '@' or angle-addr
  \`\`\`
- **`503 Service Unavailable`**: The email queue is full (for in-memory) or Redis is unavailable.
  \`\`\`
  Service Unavailable: Email queue is full
  \`\`\`
  or
  \`\`\`
  Service Unavailable: Redis queue is unavailable
  \`\`\`

### `GET /metrics`

Exposes Prometheus metrics for scraping.

**Example Usage (using `curl`)**

**Successful Request:**

\`\`\`bash
curl -X POST \
 http://localhost:8080/send-email \
 -H 'Content-Type: application/json' \
 -d '{
"to": "test@example.com",
"subject": "Hello from Email Service",
"body": "This is a test email sent via the queue."
}'
\`\`\`

**Check Metrics:**

\`\`\`bash
curl http://localhost:8080/metrics
\`\`\`
You will see output similar to:
\`\`\`

# HELP email_jobs_enqueued_total Total number of email jobs enqueued.

# TYPE email_jobs_enqueued_total counter

email_jobs_enqueued_total 1.0

# HELP email_jobs_processed_total Total number of email jobs successfully processed.

# TYPE email_jobs_processed_total counter

email_jobs_processed_total 1.0

# HELP email_queue_length Current number of jobs in the email queue.

# TYPE email_queue_length gauge

email_queue_length 0.0
...
\`\`\`

## Graceful Shutdown

The service is designed to shut down gracefully upon receiving `SIGINT` (Ctrl+C) or `SIGTERM` signals.

1.  The HTTP server stops accepting new requests.
2.  The job queue (in-memory channel or Redis client) is closed, preventing new jobs from being enqueued.
3.  Active workers are allowed to finish processing any jobs currently in the queue or being processed.
4.  The application exits cleanly.

## Configuration

The following environment variables can be used to configure the service:

- `HTTP_PORT`: The port on which the HTTP server will listen (default: `8080`).
- `WORKER_COUNT`: The number of concurrent workers to process email jobs (default: `3`).
- `QUEUE_CAPACITY`: The maximum number of email jobs the **in-memory** queue can hold (default: `100`). _Only applicable if `USE_REDIS_QUEUE` is `false`._
- `MAX_RETRIES`: The maximum number of times a failed email job will be retried (default: `3`).
- `RETRY_DELAY_SECONDS`: The delay in seconds before a failed job is re-enqueued for retry (default: `5`).
- `USE_REDIS_QUEUE`: Set to `true` to use Redis as the job queue. Otherwise, the in-memory queue is used (default: `false`).
- `REDIS_ADDR`: The address of the Redis server (e.g., `localhost:6379`). Required if `USE_REDIS_QUEUE` is `true`.
- `REDIS_PASSWORD`: The password for the Redis server (optional).
- `REDIS_DB`: The Redis database number to use (default: `0`).

---
