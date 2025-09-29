package api

import (
	"hafiztri123/app-link-shortener/internal/auth"
	"hafiztri123/app-link-shortener/internal/metrics"
	"hafiztri123/app-link-shortener/internal/rabbitmq"
	"hafiztri123/app-link-shortener/internal/url"
	"hafiztri123/app-link-shortener/internal/user"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/oschwald/maxminddb-golang"
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
	geoDb        *maxminddb.Reader
	rabbitMq     *rabbitmq.RabbitMQ
}

func NewServer(db DB, redis *redis.Client, urlService url.URLService, userService user.UserService, ts *auth.TokenService, geoDb *maxminddb.Reader, rabbitMq *rabbitmq.RabbitMQ) *Server {
	return &Server{
		db:           db,
		redis:        redis,
		urlService:   urlService,
		userService:  userService,
		tokenService: ts,
		geoDb:        geoDb,
		rabbitMq:     rabbitMq,
	}
}

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(RedisRateLimiter(s.redis, 20, 1*time.Minute))
	r.Use(metrics.PrometheusMiddleware)
	r.Use(MetadataMiddleware(s.geoDb))

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Get("/health", s.healthCheckHandler)
		v1.Post("/user/register", s.handleRegister)
		v1.Post("/user/login", s.handleLogin)
		v1.Handle("/metrics", promhttp.Handler())

		v1.Route("/url", func(url chi.Router) {

			url.Get("/{shortCode}", s.handleFetchURL)
			url.Get("/{shortCode}/qr", s.handleGenerateQR)

			url.Group(func(protected chi.Router) {
				protected.Use(AuthMiddleware(s.tokenService, true))
				protected.Post("/shorten", s.handleCreateURL)
				protected.Post("/shorten/bulk", s.handleCreateURL_Bulk)
			})
		})

		// User routes
		v1.Route("/user", func(user chi.Router) {
			user.Use(AuthMiddleware(s.tokenService, false))
			user.Get("/history", s.handleFetchUserURLHistory)
		})
	})

	return r
}
