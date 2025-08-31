package api

import (
	"hafiztri123/app-link-shortener/internal/response"
	"net/http"

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