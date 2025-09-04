package user

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "hafiztri123/app-link-shortener/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)


func setupTestDB(t *testing.T) *sql.DB {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := sql.Open("sqlite3_proxy", dsn)
	require.NoError(t, err)

	createTableSQL := `
		CREATE TABLE users (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err = db.ExecContext(context.Background(), createTableSQL)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestRepository_IntegrationFlow(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	password := "admin"
	email := "example@mail.com"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	require.NoError(t, err)

	err = repo.Insert(ctx, email, string(hashedPassword))
	require.NoError(t, err)

	user, err := repo.GetByEmail(ctx, email)
	assert.NoError(t, err)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, string(hashedPassword), user.Password)

	err = repo.Insert(ctx, email, string(hashedPassword))
	assert.Error(t, err)

	_, err = repo.GetByEmail(ctx, "invalid@mail.com")
	assert.Error(t, err)

}
