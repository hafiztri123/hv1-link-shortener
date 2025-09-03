package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/require"
)

func TestPrometheusMiddleware(t *testing.T) {
	defer prometheus.Unregister(httpRequestsTotal)
	defer prometheus.Unregister(httpRequestDuration)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := PrometheusMiddleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handlerToTest.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Expected status code 200 but got %d", rr.Code)

	expectedCounterMetric := `
		# HELP http_requests_total Total number of HTTP requests.
		# TYPE http_requests_total counter
		http_requests_total{method="GET",path="/"} 1
	`

	err := testutil.CollectAndCompare(httpRequestsTotal, strings.NewReader(expectedCounterMetric))
	require.NoError(t, err, "Unexpected collecting result: %v", err)

	expectedHistogramMetric := `
		# HELP http_request_duration_seconds Histogram of request latencies.
		# TYPE http_request_duration_seconds histogram
		http_request_duration_seconds_count{method="GET",path="/"} 1
	`

	err = testutil.CollectAndCompare(httpRequestDuration, strings.NewReader(expectedHistogramMetric), "http_request_duration_seconds_count")
	require.NoError(t, err, "Unexpected collecting result: %v", err)

	count := testutil.CollectAndCount(httpRequestDuration)

	require.Equal(t, 1, count, "Expected 1 request but got %d", count)




}
