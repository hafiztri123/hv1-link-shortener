package api

import (
	"bytes"
	"context"
	"fmt"
	"hafiztri123/app-link-shortener/internal/auth"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/time/rate"
)

func TestRateLimiter(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	rateLimiterHandler := RateLimiter(rate.Limit(2), 2)(testHandler)

	rrl := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/", nil)
	rateLimiterHandler.ServeHTTP(rrl, req1)
	assert.Equal(t, http.StatusOK, rrl.Code)

	rr2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/", nil)
	rateLimiterHandler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)

	rr3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/", nil)
	rateLimiterHandler.ServeHTTP(rr3, req3)
	assert.Equal(t, http.StatusTooManyRequests, rr3.Code)

	time.Sleep(500 * time.Millisecond)

	rr4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodGet, "/", nil)
	rateLimiterHandler.ServeHTTP(rr4, req4)
	assert.Equal(t, http.StatusOK, rr4.Code)
}

func TestLoggingMiddleware(t *testing.T) {
	var logBuffer bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))
	slog.SetDefault(logger)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	loggedHandler := LoggingMiddleware(testHandler)

	req, _ := http.NewRequest(http.MethodGet, "/test/path", nil)
	rr := httptest.NewRecorder()

	loggedHandler.ServeHTTP(rr, req)
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, `"level":"INFO`)
	assert.Contains(t, logOutput, `"msg":"http request"`)
	assert.Contains(t, logOutput, `"method":"GET"`)
	assert.Contains(t, logOutput, `"path":"/test/path"`)
	assert.Contains(t, logOutput, `"duration"`)

}

func TestRedisRateLimiter_Integration(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, redisContainer.Terminate(ctx))
	})

	host, err := redisContainer.Host(ctx)
	require.NoError(t, err)
	port, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(t, err)

	redisAddr := fmt.Sprintf("%s:%s", host, port.Port())
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	require.NoError(t, redisClient.Ping(ctx).Err())

	t.Run("allow request below the limit and blocks request above it", func(t *testing.T) {
		limit := 5
		window := 2 * time.Second

		limiterMiddleware := RedisRateLimiter(redisClient, limit, window)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		handler := limiterMiddleware(testHandler)

		for i := 0; i < limit; i++ {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, http.StatusOK, rr.Code, "request %d should be allowed", i+1)
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusTooManyRequests, rr.Code, "request %d should be blocked", limit+1)

		time.Sleep(window)
		req = httptest.NewRequest(http.MethodGet, "/", nil)
		rr = httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code, "request %d should be allowed", limit+2)
	})
}

func TestAuthMiddleware(t *testing.T) {
	tokenService := auth.NewTokenService("test123")
	reqBody := "test message"
	token, err := tokenService.GenerateToken(1, "example@mail.com")

	assert.NoError(t, err)

	testCases := []struct {
		name           string
		token          string
		permissive     bool
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "success",
			token:          "Bearer " + token,
			permissive:     false,
			wantStatusCode: http.StatusOK,
			wantBody:       reqBody,
		},

		{
			name:           "missing auth header",
			token:          "",
			permissive:     false,
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "authorization header required",
		},

		{
			name:           "missing 'bearer' in auth header",
			token:          token,
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "invalid authorization format",
		},

		{
			name:           "tampering the token",
			token:          "Bearer invalid",
			permissive:     false,
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "invalid token",
		},

		{
			name:           "permissive",
			token:          "Bearer " + token,
			permissive:     true,
			wantStatusCode: http.StatusOK,
			wantBody:       reqBody,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(reqBody))
			})

			middleware := AuthMiddleware(tokenService, tc.permissive)
			handler := middleware(testHandler)

			rrl, _ := http.NewRequest(http.MethodGet, "/", nil)

			if !tc.permissive {
				rrl.Header.Add("Authorization", tc.token)
			}

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, rrl)

			assert.Equal(t, tc.wantStatusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.wantBody)

		})
	}
}
