package url

import (
	"testing"

	"hafiztri123/app-link-shortener/internal/auth"
	"hafiztri123/app-link-shortener/internal/user"
	_ "hafiztri123/app-link-shortener/internal/utils"
	"hafiztri123/app-link-shortener/migrations"

	_ "github.com/mattn/go-sqlite3" // Driver for in-memory SQLite
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_IntegrationFlow(t *testing.T) {
	db, ctx := migrations.SetupTestDB(t)
	repo := NewRepository(db)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	longURL := "https://www.google.com/search?q=golang-testing"

	shortCode1, err := repo.FindOrCreateShortCode(ctx, longURL, 1000, nil)
	require.NoError(t, err)
	require.NotEmpty(t, shortCode1)

	shortCode2, err := repo.FindOrCreateShortCode(ctx, longURL, 1000, nil)
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

func TestRepository_FetchHistoryURL(t *testing.T) {
	db, ctx := migrations.SetupTestDB(t)
	repo := NewRepository(db)
	userId := int64(1)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	tokenService := auth.NewTokenService("secret")

	userService := user.NewService(db, user.NewRepository(db), tokenService)

	err := userService.Register(ctx, user.RegisterRequest{
		Email:    "test",
		Password: "password",
	})
	token, err := userService.Login(ctx, user.LoginRequest{
		Email:    "test",
		Password: "password",
	})

	assert.NoError(t, err)

	claims, err := tokenService.ValidateToken(token)
	assert.NoError(t, err)
	userId = claims.UserID

	longURL := "https://www.google.com/search?q=golang-testing"
	longURL2 := "https://www.google.com/search?q=golang-testing-2"

	shortCode1, err := repo.FindOrCreateShortCode(ctx, longURL, 1000, &userId)
	require.NoError(t, err)
	require.NotEmpty(t, shortCode1)

	shortCode2, err := repo.FindOrCreateShortCode(ctx, longURL2, 1000, &userId)
	require.NoError(t, err)
	require.NotEmpty(t, shortCode2)

	urls, err := repo.GetByUserIDBulk(ctx, userId)
	require.NoError(t, err)
	assert.Equal(t, 2, len(urls))
}
