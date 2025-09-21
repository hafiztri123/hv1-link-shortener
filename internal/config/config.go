package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseAddr string
	RedisAddr    string
	IDOffset     uint64
	SecretKey    string
}

func Load() (*Config, error) {
	errors := []string{}

	appUrl, ok := os.LookupEnv("APP_URL")
	if !ok {
		errors = append(errors, "APP_URL Environment variable not set")
	}

	dbPort, ok := os.LookupEnv("DB_PORT")
	if !ok {
		errors = append(errors, "DB_PORT Environment variable not set")
	}

	dbTransactionName, ok := os.LookupEnv("DB_TRANSACTION_NAME")
	if !ok {
		errors = append(errors, "DB_TRANSACTION_NAME Environment variable not set")
	}

	dbUser, ok := os.LookupEnv("DB_USER")
	if !ok {
		errors = append(errors, "DB_USER Environment variable not set")
	}

	dbPassword, ok := os.LookupEnv("DB_PASSWORD")
	if !ok {
		errors = append(errors, "DB_PASSWORD Environment variable not set")
	}

	dbSsl, ok := os.LookupEnv("DB_SSL")
	if !ok {
		errors = append(errors, "DB_SSL Environment variable not set")
	}

	redisPort, ok := os.LookupEnv("REDIS_PORT")
	if !ok {
		errors = append(errors, "REDIS_PORT Environment variable not set")
	}

	dbAddr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPassword, appUrl, dbPort, dbTransactionName, dbSsl)

	redisAddr := fmt.Sprintf("%s:%s", appUrl, redisPort)

	idOffset, ok := os.LookupEnv("ID_OFFSET")
	if !ok {
		errors = append(errors, "ID_OFFSET Environment variable not set")
	}

	jwt, ok := os.LookupEnv("JWT_SECRET")
	if !ok {
		errors = append(errors, "JWT_SECRET Environment variable not set")
	}

	convertedIdOffset, err := strconv.ParseUint(idOffset, 10, 64)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to parse ID_OFFSET: %v", err))
	}

	if len(errors) > 0 {
		errorString := strings.Join(errors, "\n")
		return nil, fmt.Errorf("FATAL: %s", errorString)
	}

	return &Config{
		DatabaseAddr: dbAddr,
		RedisAddr:    redisAddr,
		IDOffset:     convertedIdOffset,
		SecretKey:    jwt,
	}, nil

}
