package redis

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

func NewClient(ctx context.Context, connStr string, opt *redis.Options) (*redis.Client, error) {
	if opt == nil {
		opt = &redis.Options{
			Addr: connStr,
		}
	}

	rdb := redis.NewClient(opt)

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Println("SUCCESS: Connect to Redis")

	return rdb, nil
}
