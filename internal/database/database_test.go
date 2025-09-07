package database

import (
	"hafiztri123/app-link-shortener/internal/url"
	"hafiztri123/app-link-shortener/migrations"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInsertIntegration(t *testing.T) {
	db, ctx := migrations.SetupTestDB(t)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	repo := url.NewRepository(db)

	idOffset := 1000
	shortCode, err := repo.FindOrCreateShortCode(ctx, "https://example.com", uint64(idOffset))
	require.NoError(t, err)

	id := url.FromBase62(shortCode) - uint64(idOffset)

	retrievedURL, err := repo.GetByID(ctx, int64(id))
	require.NoError(t, err)
	require.Equal(t, "https://example.com", retrievedURL.LongURL)
}
