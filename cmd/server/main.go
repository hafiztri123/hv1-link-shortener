package main

import (
	"context"
	"hafiztri123/app-link-shortener/internal/api"
	"hafiztri123/app-link-shortener/internal/config"
	"hafiztri123/app-link-shortener/internal/database"
	"hafiztri123/app-link-shortener/internal/redis"
	"hafiztri123/app-link-shortener/internal/url"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found, using environment variable from OS")
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()

	if err != nil {
		log.Fatalf("FATAL: Could not load config: %v", err)
	}

	db := database.Connect(cfg.DatabaseURL)
	redis, err := redis.NewClient(context.Background(), cfg.RedisURL)

	urlRepo := url.NewRepository(db)
	urlService := url.NewService(urlRepo, redis, cfg.IDOffset)

	if err != nil {
		log.Fatalf("FATAL: Could not connect to Redis: %v", err)
	}

	defer db.Close()
	defer redis.Close()

	server := api.NewServer(db, redis, urlService)
	router := server.RegisterRoutes()

	log.Println("INFO: Listening on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
