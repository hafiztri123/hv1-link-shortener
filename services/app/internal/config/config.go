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
}

func Load() (*Config, error) {
	appUrl := utils.GetEnvOrDefault("APP_URL", "localhost")

	databaseAddr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		utils.GetEnvOrDefault("DB_USER", "admin"),
		utils.GetEnvOrDefault("DB_PASSWORD", "admin"),
		appUrl,
		utils.GetEnvOrDefault("DB_PORT", "5432"),
		utils.GetEnvOrDefault("TRANSACTION_DB", "app_db"),
		utils.GetEnvOrDefault("DB_SSL", "disable"),
	)

	redisAddr := fmt.Sprintf("%s:%s", appUrl, utils.GetEnvOrDefault("REDIS_PORT", "6379"))

	convertedIdOffset, err := strconv.ParseUint(utils.GetEnvOrDefault("ID_OFFSET", "100000000"), 10, 64)

	if err != nil {
		return nil, err
	}

	return &Config{
		DatabaseAddr: databaseAddr,
		RedisAddr:    redisAddr,
		IDOffset:     convertedIdOffset,
		SecretKey:    utils.GetEnvOrDefault("JWT", "secret"),
	}, nil

}
