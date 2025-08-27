package api

import (
	"context"
	"database/sql"
	"fmt"
	"hafiztri123/app-link-shortener/internal/url"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

type Server struct {
	db    *sql.DB
	redis *redis.Client
	urlService *url.Service
}

func NewServer(db *sql.DB, redis *redis.Client, urlService *url.Service) *Server {
	return &Server{db: db, redis: redis, urlService: urlService }
}

func (s *Server) RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.healthCheckHandler)
	mux.HandleFunc("/api/v1/shorten", s.handleCreateURL)

	return mux

}

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	err := s.db.Ping()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err != nil {
		http.Error(w, "Database not connected", http.StatusInternalServerError)
		log.Printf("Database health check failed: %v", err)
		return
	}

	err = s.redis.Ping(ctx).Err()
	if err != nil {
		http.Error(w, "Redis not connected", http.StatusInternalServerError)
		log.Printf("Redis health check failed: %v", err)
		return

	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "DB and Redis is connected")
}
