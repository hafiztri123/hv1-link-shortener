package database

import (
	"context"
	"hafiztri123/app-link-shortener/internal/url"
	"hpj/hv1-link-shortener/shared/migrations"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestInsertIntegration(t *testing.T) {
	db, ctx := migrations.SetupTestDB(t)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	repo := url.NewRepository(db)

	idOffset := 1000
	shortCode, err := repo.FindOrCreateShortCode(ctx, "https://example.com", uint64(idOffset), nil)
	require.NoError(t, err)

	id := url.FromBase62(shortCode) - uint64(idOffset)

	retrievedURL, err := repo.GetByID(ctx, int64(id))
	require.NoError(t, err)
	require.Equal(t, "https://example.com", retrievedURL.LongURL)
}

func TestConnect_Success(t *testing.T) {
	ctx := context.Background()

	// Start a test postgres container
	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections")),
	)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, pgContainer.Terminate(ctx))
	}()

	// Get connection string and test our Connect function
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Wait for the database to be fully ready
	time.Sleep(1 * time.Second)

	// Test our Connect function
	db := Connect(connStr)
	require.NotNil(t, db)

	// Verify the connection works
	err = db.Ping()
	require.NoError(t, err)

	// Clean up
	db.Close()
}
