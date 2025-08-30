package database

import (
	"context"
	"hafiztri123/app-link-shortener/internal/url"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var testRepo *url.Repository

func TestMain(m *testing.M) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Printf("WARN: unable to load environment variable: %v", err)
	}

	dbURL, ok := os.LookupEnv("DATABASE_URL_TEST")
	if !ok {
		log.Println("WARN: unable to get 'DATABASE_URL_TEST' environment variable")
	}

	db := Connect(dbURL)
	defer db.Close()

	testRepo = url.NewRepository(db)

	os.Exit(m.Run())
}

func TestInsert(t *testing.T) {
	_, err := testRepo.DB.Exec("TRUNCATE TABLE urls RESTART IDENTITY")
	if err != nil {
		t.Fatalf("FAIL: unable to reset table for test: %v", err)
	}

	_, err = testRepo.Insert(context.Background(), "https://example.com")
	if err != nil {
		t.Errorf("FAIL: unable to insert data to test database: %v", err)
	}

}
