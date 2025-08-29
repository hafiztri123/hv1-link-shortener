package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	IDOffset    uint64
}

func Load() *Config {
	dbURL, ok := os.LookupEnv("DATABASE_URL")

	if !ok {
		log.Fatal("DATABASE_URL Environment variable not set")
	}

	redisURL, ok := os.LookupEnv("REDIS_URL")
	if !ok {
		log.Fatal("REDIS_URL Environment variable not set")
	}

	idOffset, ok := os.LookupEnv("ID_OFFSET")
	if !ok {
		log.Fatal("ID_OFFSET Environment variable not set")
	}

	convertedIdOffset, err := strconv.ParseUint(idOffset, 10, 64)
	if err != nil {
		log.Fatalf("Failed to convert offset to int: %v", err)
	}

	return &Config{
		DatabaseURL: dbURL,
		RedisURL:    redisURL,
		IDOffset:    convertedIdOffset,
	}

}
