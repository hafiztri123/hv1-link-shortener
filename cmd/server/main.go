package main

import (
	"context"
	"hafiztri123/app-link-shortener/internal/api"
	"hafiztri123/app-link-shortener/internal/auth"
	"hafiztri123/app-link-shortener/internal/config"
	"hafiztri123/app-link-shortener/internal/database"
	"hafiztri123/app-link-shortener/internal/redis"
	"hafiztri123/app-link-shortener/internal/url"
	"hafiztri123/app-link-shortener/internal/user"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Warn(".env file not found, using environment variable from OS", "error", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()

	if err != nil {
		slog.Error("Could not load config", "error", err)
		os.Exit(1)
	}

	db := database.Connect(cfg.DatabaseURL)
	redis, err := redis.NewClient(context.Background(), cfg.RedisURL, nil)

	if err != nil {
		slog.Error("Could not connect to redis", "error", err)
		os.Exit(1)
	}

	tokenService := auth.NewTokenService(cfg.SecretKey)

	urlRepo := url.NewRepository(db)
	urlService := url.NewService(urlRepo, redis, cfg.IDOffset)

	userRepo := user.NewRepository(db)
	userService := user.NewService(db, userRepo, tokenService)

	server := api.NewServer(db, redis, urlService, userService, tokenService)
	router := server.RegisterRoutes()

	defer db.Close()
	defer redis.Close()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		slog.Info("Listening on port :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	slog.Info("Shutdown signal received, starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Server shutdown successfully")

}
