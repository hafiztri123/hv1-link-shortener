package api

import (
	"hafiztri123/app-link-shortener/internal/auth"
	"hafiztri123/app-link-shortener/internal/metrics"
	"hafiztri123/app-link-shortener/internal/url"
	"hafiztri123/app-link-shortener/internal/user"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type DB interface {
	Ping() error
}

type Server struct {
	db           DB
	redis        *redis.Client
	urlService   url.URLService
	userService  user.UserService
	tokenService *auth.TokenService
}

func NewServer(db DB, redis *redis.Client, urlService url.URLService, userService user.UserService, ts *auth.TokenService) *Server {
	return &Server{db: db, redis: redis, urlService: urlService, userService: userService, tokenService: ts}
}

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	// r.Use(RateLimiter(10, 20))
	r.Use(RedisRateLimiter(s.redis, 20, 1*time.Minute))

	r.Use(metrics.PrometheusMiddleware)
	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Get("/health", s.healthCheckHandler)
		v1.Get("/url/{shortCode}", s.handleFetchURL)
		v1.Get("/url/{shortCode}/qr", s.handleGenerateQR)
		v1.Post("/user/register", s.handleRegister)
		v1.Post("/user/login", s.handleLogin)
		v1.Handle("/metrics", promhttp.Handler())

		v1.Route("/url", func(protected chi.Router) {
			protected.Use(auth.AuthMiddleware(s.tokenService, true))
			protected.Post("/shorten", s.handleCreateURL)
			protected.Post("/shorten/bulk", s.handleCreateURL_Bulk)
		})

		v1.Route("/user", func(protected chi.Router) {
			protected.Use(auth.AuthMiddleware(s.tokenService, false))
			protected.Get("/history", s.handleFetchUserURLHistory)
		})
	})

	return r

}
