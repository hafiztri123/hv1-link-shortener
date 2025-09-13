package test

import (
	"context"
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

	ctxWithValue := context.WithValue(ctx, auth.UserContextKey, claims)

	_, err = urlService.CreateShortCode(ctxWithValue, "https://example.com")
	assert.NoError(t, err)

	_, err = urlService.CreateShortCode(ctxWithValue, "https://example2.com")
	assert.NoError(t, err)

	urls, err := urlService.FetchUserURLHistory(ctxWithValue, claims.UserID)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(urls))

}
