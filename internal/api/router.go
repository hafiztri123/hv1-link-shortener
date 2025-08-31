package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	db         DB
	redis      Redis
	urlService URLService
}

func NewServer(db DB, redis Redis, urlService URLService) *Server {
	return &Server{db: db, redis: redis, urlService: urlService}
}

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Get("/health", s.healthCheckHandler)
		v1.Post("/url/shorten", s.handleCreateURL)
		v1.Get("/url/{shortCode}", s.handleFetchURL)
	})

	return r

}
