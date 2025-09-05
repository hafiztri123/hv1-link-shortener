package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	t.Run("success case - valid config", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "test_db_url")
		t.Setenv("REDIS_URL", "test_redis_url")
		t.Setenv("ID_OFFSET", "123")
		t.Setenv("JWT_SECRET", "jwt_secret")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.Equal(t, "test_db_url", cfg.DatabaseURL)
		assert.Equal(t, "test_redis_url", cfg.RedisURL)
		assert.Equal(t, uint64(123), cfg.IDOffset)
	})

	t.Run("failure case - missing DATABASE_URL", func(t *testing.T) {
		t.Setenv("REDIS_URL", "test_redis_url")
		t.Setenv("ID_OFFSET", "123")
		os.Unsetenv("DATABASE_URL")

		_, err := Load()

		assert.Error(t, err)
	})
}
