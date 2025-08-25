package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect() *sql.DB {
	connStr, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Unable to open database connection: %v\n", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	log.Println("SUCCESS: Connect to database")
	return db
}
