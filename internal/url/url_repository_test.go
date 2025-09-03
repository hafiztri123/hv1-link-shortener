package url

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3" // Driver for in-memory SQLite
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := sql.Open("sqlite3_proxy", dsn)
	require.NoError(t, err)

	createTableSQL := `
		CREATE TABLE urls (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			short_code TEXT,
			long_url TEXT UNIQUE NOT NULL,
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
	repo := NewRepository(db)
	ctx := context.Background()

	longURL := "https://www.google.com/search?q=golang-testing"

	shortCode1, err := repo.FindOrCreateShortCode(ctx, longURL, 1000)
	require.NoError(t, err)
	require.NotEmpty(t, shortCode1)

	shortCode2, err := repo.FindOrCreateShortCode(ctx, longURL, 1000)
	assert.NoError(t, err)
	assert.Equal(t, shortCode1, shortCode2, "Expected: %s, Actual: %s", shortCode1, shortCode2)

	retrievedURL, err := repo.GetByID(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, retrievedURL, "Expected: URL")
	require.Equal(t, int64(1), retrievedURL.ID)
	require.Equal(t, longURL, retrievedURL.LongURL)
	require.True(t, retrievedURL.ShortCode.Valid)
	require.Equal(t, shortCode1, retrievedURL.ShortCode.String)
}
