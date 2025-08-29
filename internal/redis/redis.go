package redis

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

func NewClient(ctx context.Context) (*redis.Client, error) {
	rdbAddr, ok := os.LookupEnv("REDIS_URL")
	if !ok {
		log.Fatal("REDIS_URL Environment variable  not set")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: rdbAddr,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Println("SUCCESS: Connect to Redis")

	return rdb, nil
}
