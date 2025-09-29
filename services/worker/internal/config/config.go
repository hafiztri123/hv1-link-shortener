package config

import (
	"fmt"
	"hpj/hv1-link-shortener/shared/utils"
)

type Config struct {
	RabbitMQAddr    string
	AnalyticsDBAddr string
	ClickQueueLabel string
}

func Load() *Config {
	rabbitmqAddr := fmt.Sprintf("amqp://%s:%s@%s:%s",
		utils.GetEnvOrDefault("RABBITMQ_USER", "guest"),
		utils.GetEnvOrDefault("RABBITMQ_PASSWORD", "guest"),
		utils.GetEnvOrDefault("RABBITMQ_HOST", "localhost"),
		utils.GetEnvOrDefault("RABBITMQ_PORT", "5672"),
	)

	analyticsDbAddr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		utils.GetEnvOrDefault("DB_USER", "admin"),
		utils.GetEnvOrDefault("DB_PASSWORD", "admin"),
		utils.GetEnvOrDefault("DB_HOST", "localhost"),
		utils.GetEnvOrDefault("DB_PORT", "5432"),
		utils.GetEnvOrDefault("ANALYTICS_DB", "analytics_db"),
		utils.GetEnvOrDefault("DB_SSL", "disable"),
	)
	return &Config{
		RabbitMQAddr:    rabbitmqAddr,
		ClickQueueLabel: utils.GetEnvOrDefault("CLICK_QUEUE_LABEL", "click_event"),
		AnalyticsDBAddr: analyticsDbAddr,
	}
}
