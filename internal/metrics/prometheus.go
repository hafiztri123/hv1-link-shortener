package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path"}, // Labels
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of request latencies.",
			Buckets: prometheus.DefBuckets, // A default set of buckets for latency.
		},
		[]string{"method", "path"}, // Labels
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)

	prometheus.MustRegister(httpRequestDuration)
}

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues(r.Method, r.URL.Path))
		defer timer.ObserveDuration()

		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path).Inc()

		next.ServeHTTP(w,r)
	})
}
