package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	t.Run("success case - valid config", func(t *testing.T) {
		t.Setenv("APP_URL", "localhost")
		t.Setenv("DB_PORT", "5432")
		t.Setenv("TRANSACTION_DB", "testdb")
		t.Setenv("DB_USER", "testuser")
		t.Setenv("DB_PASSWORD", "testpass")
		t.Setenv("DB_SSL", "disable")
		t.Setenv("REDIS_PORT", "6379")
		t.Setenv("ID_OFFSET", "123")
		t.Setenv("JWT", "jwt_secret")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.Equal(t, "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable", cfg.DatabaseAddr)
		assert.Equal(t, "localhost:6379", cfg.RedisAddr)
		assert.Equal(t, uint64(123), cfg.IDOffset)
		assert.Equal(t, "jwt_secret", cfg.SecretKey)
	})

	t.Run("success case - missing APP_URL", func(t *testing.T) {
		t.Setenv("DB_PORT", "5432")
		t.Setenv("TRANSACTION_DB", "testdb")
		t.Setenv("DB_USER", "testuser")
		t.Setenv("DB_PASSWORD", "testpass")
		t.Setenv("DB_SSL", "disable")
		t.Setenv("REDIS_PORT", "6379")
		t.Setenv("ID_OFFSET", "123")
		t.Setenv("JWT", "jwt_secret")
		os.Unsetenv("APP_URL")

		_, err := Load()

		assert.NoError(t, err)
	})

	t.Run("success case - missing DB_PORT", func(t *testing.T) {
		t.Setenv("DB_PORT", "5432")
		t.Setenv("TRANSACTION_DB", "testdb")
		t.Setenv("DB_USER", "testuser")
		t.Setenv("DB_PASSWORD", "testpass")
		t.Setenv("DB_SSL", "disable")
		t.Setenv("REDIS_PORT", "6379")
		t.Setenv("ID_OFFSET", "123")
		t.Setenv("JWT", "jwt_secret")
		os.Unsetenv("DB_PORT")

		_, err := Load()

		assert.NoError(t, err)
	})

	t.Run("failure case - invalid ID_OFFSET", func(t *testing.T) {
		t.Setenv("APP_URL", "localhost")
		t.Setenv("DB_PORT", "5432")
		t.Setenv("TRANSACTION_DB", "testdb")
		t.Setenv("DB_USER", "testuser")
		t.Setenv("DB_PASSWORD", "testpass")
		t.Setenv("DB_SSL", "disable")
		t.Setenv("REDIS_PORT", "6379")
		t.Setenv("ID_OFFSET", "invalid")
		t.Setenv("JWT", "jwt_secret")

		_, err := Load()

		assert.Error(t, err)
	})
}
