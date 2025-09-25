package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	RedisAddr string
	AnalyticsDBAddr string
	StreamName string
	ConsumerGroup string
	ConsumerName string
	BatchSize int64
	ProcessingDelay int
}

func Load() (*Config, error) {
	requiredVars := map[string]string{
		"APP_URL": os.Getenv("APP_URL"),
		"DB_USER": os.Getenv("DB_USER"),
		"DB_PASSWORD" : os.Getenv("DB_PASSWORD"),
		"DB_ANALYTICS_NAME": os.Getenv("DB_ANALYTICS_NAME"),
		"DB_PORT": os.Getenv("DB_PORT"),
		"REDIS_PORT": os.Getenv("REDIS_PORT"),
	}

	for key, value := range requiredVars {
		if value == "" {
			return nil, fmt.Errorf("missing environment variable: %s", key)
		}
	}

	dbSsl := GetEnvOrDefault("DB_SSL", "disable")
	analyticsDBAddr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		requiredVars["DB_USER"],
		requiredVars["DB_PASSWORD"],
		requiredVars["APP_URL"],
		requiredVars["DB_PORT"],
		requiredVars["DB_ANALYTICS_NAME"],
		dbSsl,
	)

	redisAddr := fmt.Sprintf("%s:%s", requiredVars["APP_URL"], requiredVars["REDIS_PORT"])

	batchSize, _ := strconv.ParseInt(GetEnvOrDefault("WORKER_BATCH_SIZE", "100"), 10, 64)
	processingDelay, _ := strconv.Atoi(GetEnvOrDefault("WORKER_PROCESSING_DELAY", "1"))

	return &Config{
		RedisAddr: redisAddr,
		AnalyticsDBAddr: analyticsDBAddr,
		StreamName: GetEnvOrDefault("REDIS_STREAM_NAME", "clicks_stream"),
		ConsumerGroup: GetEnvOrDefault("CONSUME_GROUP", "analytics_workers"),
		ConsumerName: GetEnvOrDefault("CONSUMER_NAME", "worker_1"),
		BatchSize: batchSize,
		ProcessingDelay: processingDelay,
	}, nil





}

func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue

}