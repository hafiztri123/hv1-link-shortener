package url

import (
	"hpj/hv1-link-shortener/shared/migrations"
	"sync"
	"testing"

	"hafiztri123/app-link-shortener/internal/auth"
	"hafiztri123/app-link-shortener/internal/user"
	_ "hafiztri123/app-link-shortener/internal/utils"

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

	urls, err := repo.GetByUserID_Bulk(ctx, userId)
	require.NoError(t, err)
	assert.Equal(t, 2, len(urls))
}

func TestRepository_Shortening_Bulk(t *testing.T) {
	db, ctx := migrations.SetupTestDB(t)
	repo := NewRepository(db)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	longURL := "https://www.google.com/search?q=golang-testing"
	longURL2 := "https://www.google.com/search?q=golang-testing-2"

	result, err := repo.FindOrCreateShortCode_Bulk(ctx, []string{longURL, longURL2}, 1000, nil)
	require.NoError(t, err)
	assert.Equal(t, 2, len(result))

	longURL3 := "https://www.google.com/search?q=golang-testing-"
	result, err = repo.FindOrCreateShortCode_Bulk(ctx, []string{longURL3}, 1000, nil)

	require.NoError(t, err)
	assert.Equal(t, 1, len(result))

	t.Log(result)
}

func TestRepository_Shortening_Race(t *testing.T) {
	db, ctx := migrations.SetupTestDB(t)
	repo := NewRepository(db)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	var wg sync.WaitGroup
	worker := 5
	for i := 0; i < worker; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, err := repo.FindOrCreateShortCode(ctx, "https://www.google.com/search?q=golang-testing", 1000, nil)
				require.NoError(t, err)
			}
		}()
	}

	wg.Wait()
}

func TestRepository_GetByUserID_Bulk_ScanError(t *testing.T) {
	db, ctx := migrations.SetupTestDB(t)
	repo := NewRepository(db)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	// Test with non-existent user ID - should return empty slice
	urls, err := repo.GetByUserID_Bulk(ctx, 999)
	require.NoError(t, err)
	assert.Equal(t, 0, len(urls))
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	db, ctx := migrations.SetupTestDB(t)
	repo := NewRepository(db)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	// Test with non-existent ID
	_, err := repo.GetByID(ctx, 999)
	require.Error(t, err)
}
