package url

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3" // Driver for in-memory SQLite
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3_proxy", ":memory:")
	require.NoError(t, err)

	createTableSQL := `
		CREATE TABLE urls (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			short_code TEXT,
			long_url TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err = db.ExecContext(context.Background(), createTableSQL)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestRepository_IntegrationFlow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	longURL := "https://www.google.com/search?q=golang-testing"

	id, err := repo.Insert(ctx, longURL)
	require.NoError(t, err)
	require.Equal(t, int64(1), id, "Expected: 1, Actual: %d", id)

	retrievedURL, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, retrievedURL, "Expected: URL")
	require.Equal(t, id, retrievedURL.ID, "Expected: %d, Actual: %d", id, retrievedURL.ID)
	require.Equal(t, longURL, retrievedURL.LongURL, "Expected: %s, Actual: %s", longURL, retrievedURL.LongURL)
	require.False(t, retrievedURL.ShortCode.Valid, "Expected: empty string, Actual: %s", retrievedURL.ShortCode)

	shortCode := "GoTest"
	err = repo.UpdateShortCode(ctx, id, shortCode)
	require.NoError(t, err)

	updatedURL, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, updatedURL, "Expected: URL")
	require.True(t, updatedURL.ShortCode.Valid, "Expected: %s, Actual: %s", shortCode, updatedURL.ShortCode)

	_, err = repo.GetByID(ctx, 999)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows, "Expected: ErrNoRows")
}
