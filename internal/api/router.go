package api

import (
	"hafiztri123/app-link-shortener/internal/metrics"
	"hafiztri123/app-link-shortener/internal/url"
	"hafiztri123/app-link-shortener/internal/user"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	db         DB
	redis      *redis.Client
	urlService url.URLService
	userService user.UserService
}

func NewServer(db DB, redis *redis.Client, urlService url.URLService, userService user.UserService) *Server {
	return &Server{db: db, redis: redis, urlService: urlService, userService: userService}
}

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	// r.Use(RateLimiter(10, 20))
	r.Use(RedisRateLimiter(s.redis, 20, 1*time.Minute))

	r.Use(metrics.PrometheusMiddleware)
	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Get("/health", s.healthCheckHandler)
		v1.Post("/url/shorten", s.handleCreateURL)
		v1.Get("/url/{shortCode}", s.handleFetchURL)
		v1.Handle("/metrics", promhttp.Handler())
	})

	return r

}
