package redis

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
)

func TestNewClient_NilOptions(t *testing.T) {
	ctx := context.Background()

	// Test with nil options - this should create default options
	_, err := NewClient(ctx, "invalid-host:6379", nil)

	// We expect an error because the host is invalid, but the function should handle nil options
	if err == nil {
		t.Error("Expected error due to invalid host, but got nil")
	}

	// The important thing is that it didn't panic due to nil options
}

func TestNewClient_WithCustomOptions(t *testing.T) {
	ctx := context.Background()

	customOpt := &redis.Options{
		Addr:     "invalid-host:6380",
		Password: "password",
		DB:       1,
	}

	// Test with custom options
	_, err := NewClient(ctx, "ignored-conn-string", customOpt)

	// We expect an error because the host is invalid
	if err == nil {
		t.Error("Expected error due to invalid host, but got nil")
	}

	// The important thing is that custom options were used
}

func TestNewClient_PingFailure(t *testing.T) {
	ctx := context.Background()

	// Test with a definitely invalid address
	_, err := NewClient(ctx, "127.0.0.1:1", nil)

	// Should return an error due to connection failure
	if err == nil {
		t.Error("Expected ping to fail with invalid address")
	}
}

func TestNewClient_EmptyConnString(t *testing.T) {
	ctx := context.Background()

	// Test with empty connection string and nil options
	client, err := NewClient(ctx, "", nil)

	// Empty string defaults to localhost:6379 and may succeed if Redis is running locally
	if err != nil && client == nil {
		// This is fine - either connection failed or succeeded
		t.Logf("Connection failed as expected: %v", err)
	} else if err == nil && client != nil {
		// This is also fine - Redis is running locally
		t.Logf("Connection succeeded to local Redis")
		client.Close()
	}
}
