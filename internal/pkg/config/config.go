package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds the application's configuration.
type Config struct {
	HTTPPort          int
	WorkerCount       int
	QueueCapacity     int
	MaxRetries        int
	RetryDelaySeconds int
	UseRedisQueue     bool
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
}

// LoadConfig loads configuration from environment variables or uses default values.
func LoadConfig() *Config {
	httpPortStr := os.Getenv("HTTP_PORT")
	httpPort, err := strconv.Atoi(httpPortStr)
	if err != nil || httpPort == 0 {
		httpPort = 8080 // Default HTTP port
		log.Printf("HTTP_PORT not set or invalid, using default: %d", httpPort)
	}

	workerCountStr := os.Getenv("WORKER_COUNT")
	workerCount, err := strconv.Atoi(workerCountStr)
	if err != nil || workerCount == 0 {
		workerCount = 3 // Default number of workers
		log.Printf("WORKER_COUNT not set or invalid, using default: %d", workerCount)
	}

	queueCapacityStr := os.Getenv("QUEUE_CAPACITY")
	queueCapacity, err := strconv.Atoi(queueCapacityStr)
	if err != nil || queueCapacity == 0 {
		queueCapacity = 100 // Default queue capacity
		log.Printf("QUEUE_CAPACITY not set or invalid, using default: %d", queueCapacity)
	}

	maxRetriesStr := os.Getenv("MAX_RETRIES")
	maxRetries, err := strconv.Atoi(maxRetriesStr)
	if err != nil || maxRetries < 0 {
		maxRetries = 3 // Default max retries
		log.Printf("MAX_RETRIES not set or invalid, using default: %d", maxRetries)
	}

	retryDelaySecondsStr := os.Getenv("RETRY_DELAY_SECONDS")
	retryDelaySeconds, err := strconv.Atoi(retryDelaySecondsStr)
	if err != nil || retryDelaySeconds <= 0 {
		retryDelaySeconds = 5 // Default retry delay in seconds
		log.Printf("RETRY_DELAY_SECONDS not set or invalid, using default: %d", retryDelaySeconds)
	}

	useRedisQueue := os.Getenv("USE_REDIS_QUEUE") == "true"
	redisAddr := os.Getenv("REDIS_ADDR")
	if useRedisQueue && redisAddr == "" {
		redisAddr = "localhost:6379" // Default Redis address
		log.Printf("REDIS_ADDR not set, using default: %s", redisAddr)
	}

	redisPassword := os.Getenv("REDIS_PASSWORD") // Can be empty
	redisDBStr := os.Getenv("REDIS_DB")
	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		redisDB = 0 // Default Redis DB
		log.Printf("REDIS_DB not set or invalid, using default: %d", redisDB)
	}

	return &Config{
		HTTPPort:          httpPort,
		WorkerCount:       workerCount,
		QueueCapacity:     queueCapacity,
		MaxRetries:        maxRetries,
		RetryDelaySeconds: retryDelaySeconds,
		UseRedisQueue:     useRedisQueue,
		RedisAddr:         redisAddr,
		RedisPassword:     redisPassword,
		RedisDB:           redisDB,
	}
}
