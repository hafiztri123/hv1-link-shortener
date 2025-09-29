package config

import (
	"os"
	"strconv"
)

type Config struct {
	RabbitMQAddr    string
	AnalyticsDBAddr string
	QueueName       string
	WorkerCount     int
	BatchSize       int
	ProcessingDelay int
}

func Load() (*Config, error) {
	workerCount, err := strconv.Atoi(getEnvOrDefault("WORKER_COUNT", "3"))
	if err != nil {
		return nil, err
	}

	batchSize, err := strconv.Atoi(getEnvOrDefault("BATCH_SIZE", "50"))
	if err != nil {
		return nil, err
	}

	processingDelay, err := strconv.Atoi(getEnvOrDefault("PROCESSING_DELAY", "1"))
	if err != nil {
		return nil, err
	}

	return &Config{
		RabbitMQAddr:    getEnvOrDefault("RABBITMQ_ADDR", "amqp://guest:guest@localhost:5672/"),
		AnalyticsDBAddr: getEnvOrDefault("ANALYTICS_DB", "analytics_db"),
		QueueName:       getEnvOrDefault("CLICKS_QUEUE", "click_events"),
		WorkerCount:     workerCount,
		BatchSize:       batchSize,
		ProcessingDelay: processingDelay,
	}, nil

}

func getEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue

}
