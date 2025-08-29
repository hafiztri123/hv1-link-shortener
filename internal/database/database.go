package database

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(connStr string) *sql.DB {


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
