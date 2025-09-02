package api

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
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

func TestRedisRateLimiter(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("request is allowed", func(t *testing.T) {
		db, mock := redismock.NewClientMock()

		now := time.Now().UnixNano()
		windowStart := now - (1 * time.Minute).Nanoseconds()
		ip := "127.0.0.1:1234"

		mock.ExpectTxPipeline()
		mock.ExpectZRemRangeByScore(ip, "0", strconv.FormatInt(windowStart, 10))
		mock.ExpectZAdd(ip, &redis.Z{Score: float64(now), Member: now})
		mock.ExpectZCard(ip).SetVal(5)
		mock.ExpectExpire(ip, 1*time.Minute)
		mock.ExpectTxPipelineExec()

		limiter := RedisRateLimiter(db, 10, 1*time.Minute)
		handler := limiter(testHandler)


		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, rr.Code, http.StatusOK)
	})
}
