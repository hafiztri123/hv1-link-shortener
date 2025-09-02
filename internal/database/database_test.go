package database

import (
	"context"
	"hafiztri123/app-link-shortener/internal/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestInsertIntegration(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections")),
	)

	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, pgContainer.Terminate(ctx))
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	time.Sleep(1 * time.Second) // wait for the database to be ready

	db := Connect(connStr)
	defer db.Close()

	createTableSQL := `
		CREATE TABLE urls (
			id SERIAL PRIMARY KEY,
			short_code VARCHAR(20) UNIQUE,
			long_url TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`
	_, err = db.Exec(createTableSQL)
	require.NoError(t, err)

	repo := url.NewRepository(db)

	id, err := repo.Insert(ctx, "https://example.com")
	require.NoError(t, err)
	require.Greater(t, id, int64(0))

	retrievedURL, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "https://example.com", retrievedURL.LongURL)
}
