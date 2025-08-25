package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	db *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{db: db}
}

func (s *Server) RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.healthCheckHandler)

	return mux

}

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	err := s.db.Ping()
	if err != nil {
		http.Error(w, "Database not connected", http.StatusInternalServerError)
		log.Printf("Database health check failed: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Database is connected")
}
