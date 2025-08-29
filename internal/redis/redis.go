package redis

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

func NewClient(ctx context.Context, connStr string) (*redis.Client, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr: connStr,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Println("SUCCESS: Connect to Redis")

	return rdb, nil
}
