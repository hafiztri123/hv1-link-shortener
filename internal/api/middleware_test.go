package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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