package main

import (
	"context"
	"hafiztri123/app-link-shortener/internal/api"
	"hafiztri123/app-link-shortener/internal/database"
	"hafiztri123/app-link-shortener/internal/redis"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found, using environment variable from OS")
	}

	db := database.Connect()
	redis, err := redis.NewClient(context.Background())
	if err != nil {
		log.Fatalf("FATAL: Could not connect to Redis: %v", err)
	}

	defer db.Close()

	server := api.NewServer(db, redis)
	router := server.RegisterRoutes()

	log.Println("INFO: Listening on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
