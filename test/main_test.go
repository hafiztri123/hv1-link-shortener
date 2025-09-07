package test

import (
	"hafiztri123/app-link-shortener/internal/auth"
	"hafiztri123/app-link-shortener/internal/url"
	"hafiztri123/app-link-shortener/internal/user"
	"hafiztri123/app-link-shortener/migrations"
	"testing"

	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLHistory(t *testing.T) {
	db, ctx := migrations.SetupTestDB(t)
	redis, _ := redismock.NewClientMock()

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	jwtService := auth.NewTokenService("secret")
	userService := user.NewService(db, user.NewRepository(db), jwtService)
	urlService := url.NewService(url.NewRepository(db), redis, 0)

	err := userService.Register(ctx, user.RegisterRequest{
		Email:    "test",
		Password: "password",
	})

	assert.NoError(t, err)

	token, err := userService.Login(ctx, user.LoginRequest{
		Email:    "test",
		Password: "password",
	})

	claims, err := jwtService.ValidateToken(token)
	assert.NoError(t, err)

	_, err = urlService.CreateShortCode(ctx, "https://example.com", &claims.UserID)
	assert.NoError(t, err)

	_, err = urlService.CreateShortCode(ctx, "https://example2.com", &claims.UserID)
	assert.NoError(t, err)

	urls, err := urlService.FetchUserURLHistory(ctx, claims.UserID)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(urls))

}
