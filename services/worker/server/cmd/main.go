package main

import (
	"context"
	"hafiztri123/worker-link-shortener/internal/config"
	"hafiztri123/worker-link-shortener/internal/queue"
	"hafiztri123/worker-link-shortener/internal/queue/metadata"
	"hpj/hv1-link-shortener/shared/database"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	err := godotenv.Load()
	if err != nil {
		slog.Warn("error initializing env", "err", err)
	}

	cfg := config.Load()

	db := database.Connect(cfg.AnalyticsDBAddr)
	metadataRepository := metadata.NewRepository(db)

	consumer, err := queue.NewConsumer(cfg.RabbitMQAddr, cfg.ClickQueueLabel, metadataRepository)

	if err != nil {
		slog.Error("failed to create consumer", "error", err)
		os.Exit(1)
	}
	defer consumer.Close()

	dlqConsumer, err := queue.NewDLQConsumer(cfg.RabbitMQAddr, cfg.ClickQueueLabel)
	if err != nil {
		slog.Error("failed to create dlq consumer", "error", err)
		os.Exit(1)
	}
	defer dlqConsumer.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		errChan <- consumer.StartConsuming(ctx)
	}()

	go func() {
		errChan <- dlqConsumer.StartConsuming(ctx)
	}()

	select {
	case err := <-errChan:
		slog.Error("Consumer error", "error", err)
	case sig := <-sigChan:
		slog.Info("shutting down", "signal", sig)
		cancel()
	}
}
