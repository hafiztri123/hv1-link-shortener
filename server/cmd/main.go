package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	err := godotenv.Load()
	if err != nil {
		slog.Error("error initializing slog for logging purposes", "err", err)
		os.Exit(1)
	}

	
}

