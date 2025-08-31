package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	IDOffset    uint64
}

func Load() (*Config, error) {
	errors := []string{}
	dbURL, ok := os.LookupEnv("DATABASE_URL")

	if !ok {
		errors = append(errors, "DATABASE_URL Environment variable not set")
	}

	redisURL, ok := os.LookupEnv("REDIS_URL")
	if !ok {
		errors = append(errors, "REDIS_URL Environment variable not set")
	}

	idOffset, ok := os.LookupEnv("ID_OFFSET")
	if !ok {
		errors = append(errors, "ID_OFFSET Environment variable not set")
	}

	convertedIdOffset, err := strconv.ParseUint(idOffset, 10, 64)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to parse ID_OFFSET: %v", err))
	}

	if (len(errors) > 0 ) {
		errorString := strings.Join(errors, "\n")
		return nil, fmt.Errorf("FATAL: %s", errorString)
	}

	return &Config{
		DatabaseURL: dbURL,
		RedisURL:    redisURL,
		IDOffset:    convertedIdOffset,
	}, nil

}
