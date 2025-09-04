package api

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type DB interface {
	Ping() error
}

type Redis interface {
	Ping(ctx context.Context) *redis.StatusCmd
}

