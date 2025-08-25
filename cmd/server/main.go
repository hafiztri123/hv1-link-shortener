package main

import (
	"hafiztri123/app-link-shortener/internal/api"
	"hafiztri123/app-link-shortener/internal/database"
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
	defer db.Close()

	server := api.NewServer(db)
	router := server.RegisterRoutes()

	log.Println("INFO: Listening on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
