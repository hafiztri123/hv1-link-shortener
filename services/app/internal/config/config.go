package config

import (
	"fmt"
	"hpj/hv1-link-shortener/shared/utils"
	"strconv"
)

type Config struct {
	DatabaseAddr string
	RedisAddr    string
	IDOffset     uint64
	SecretKey    string
	RabbitMQAddr string
	ClickQueueLabel string
}

func Load() (*Config, error) {

	databaseAddr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		utils.GetEnvOrDefault("DB_USER", "admin"),
		utils.GetEnvOrDefault("DB_PASSWORD", "admin"),
		utils.GetEnvOrDefault("DB_HOST", "admin"),
		utils.GetEnvOrDefault("DB_PORT", "5432"),
		utils.GetEnvOrDefault("TRANSACTION_DB", "app_db"),
		utils.GetEnvOrDefault("DB_SSL", "disable"),
	)

	rabbitmqAddr:= fmt.Sprintf("amqp://%s:%s@%s:%s",
		utils.GetEnvOrDefault("RABBITMQ_USER", "guest"),
		utils.GetEnvOrDefault("RABBITMQ_PASSWORD", "guest"),
		utils.GetEnvOrDefault("RABBITMQ_HOST", "localhost"),
		utils.GetEnvOrDefault("RABBITMQ_PORT", "5672"),
	)


	redisAddr := fmt.Sprintf("%s:%s", utils.GetEnvOrDefault("REDIS_HOST", "localhost"), utils.GetEnvOrDefault("REDIS_PORT", "6379"))

	convertedIdOffset, err := strconv.ParseUint(utils.GetEnvOrDefault("ID_OFFSET", "100000000"), 10, 64)

	if err != nil {
		return nil, err
	}

	return &Config{
		DatabaseAddr: databaseAddr,
		RedisAddr:    redisAddr,
		IDOffset:     convertedIdOffset,
		SecretKey:    utils.GetEnvOrDefault("JWT", "secret"),
		RabbitMQAddr: rabbitmqAddr,
		ClickQueueLabel: utils.GetEnvOrDefault("CLICK_QUEUE_LABEL", "click_event"),
	}, nil

}
