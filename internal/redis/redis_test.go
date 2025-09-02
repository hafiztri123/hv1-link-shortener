package redis

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var redisClient *redis.Client

func TestMain(m *testing.M) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Printf("WARN: unable to load environment variable: %v", err)
	}

	redisURL := os.Getenv("REDIS_URL_TEST")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/1" // DB 1 will be used as test database
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("FATAL: unable to parse Redis URL: %v", err)
	}

	redisClient, err = NewClient(context.Background(), redisURL, opt)

	if err != nil {
		log.Fatalf("FATAL: unable to connect to Redis: %v", err)
	}

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("FATAL: unable to ping Redis: %v", err)
	}

	exitCode := m.Run()

	redisClient.Close()
	os.Exit(exitCode)
}

func setupTest(t *testing.T) func() {
	teardown := func() {
		err := redisClient.FlushDB(context.Background()).Err()
		if err != nil {
			t.Fatalf("FATAL: unable to flush Redis database: %v", err)
		}
	}

	teardown()
	return teardown
}

func TestSetAndGet(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	err := redisClient.Set(context.Background(), "testKey", "testValue", 0).Err()
	if err != nil {
		t.Errorf("FAIL: unable to set key in Redis: %v", err)
	}

	val, err := redisClient.Get(context.Background(), "testKey").Result()

	if err != nil {
		t.Errorf("FAIL: unable to get key from Redis: %v", err)
	}

	if val != "testValue" {
		t.Errorf("FAIL: expected 'testValue', got: %v", val)
	}
}

func TestGetNonExistentKey(t *testing.T) {
	teardown := setupTest(t)
	defer teardown()

	_, err := redisClient.Get(context.Background(), "nonExistentKey").Result()
	if err != redis.Nil {
		t.Errorf("FAIL: expected redis.Nil error, got: %v", err)
	}
}
