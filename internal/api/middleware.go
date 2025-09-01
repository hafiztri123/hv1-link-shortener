package api

import (
	"hafiztri123/app-link-shortener/internal/response"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
)

func RateLimiter(r rate.Limit, b int) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(r, b)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				response.Error(w, http.StatusTooManyRequests, "The API is at capacity, please try again later")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RedisRateLimiter(redisClient *redis.Client, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			now := time.Now().UnixNano()
			windowStart := now - window.Nanoseconds()

			pipe := redisClient.TxPipeline()
			pipe.ZRemRangeByScore(r.Context(), ip, "0", strconv.FormatInt(windowStart, 10))
			countCmd := pipe.ZCard(r.Context(), ip)
			pipe.ZAdd(r.Context(), ip, &redis.Z{Score: float64(now), Member: now})
			pipe.Expire(r.Context(), ip, window)
			_, err := pipe.Exec(r.Context())

			if err != nil {
				slog.Warn("Rate limiter failed", "error", err)
				next.ServeHTTP(w, r)
				return
			}

			if countCmd.Val() > int64(limit) {
				response.Error(w, http.StatusTooManyRequests, "Too many requests")
				return
			}

			next.ServeHTTP(w, r)

		})
	}

}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		slog.Info("http request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}
