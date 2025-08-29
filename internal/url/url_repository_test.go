package url

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var testRepo *Repository

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		log.Fatal("TEST_DATABASE_URL environment variable not set")
	}

	testDB, err := sql.Open("pgx", testDBURL)
	if err != nil {
		log.Fatalf("Unable to open test database connection: %v\n", err)
	}

	defer testDB.Close()

	testRepo = NewRepository(testDB)


}
