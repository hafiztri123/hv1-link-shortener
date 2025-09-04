package user

import (
	"context"
	"database/sql"
	"errors"

	"hafiztri123/app-link-shortener/internal/utils"


	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository interface {
	Insert(ctx context.Context, email string, password string) error
	GetByEmail(ctx context.Context, email string) (*User, error)


}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}


func (r *Repository) Insert(ctx context.Context, email string, password string) error {
	insertQuery := `INSERT INTO users (email, password) VALUES ($1, $2)`

	_, err := r.db.ExecContext(ctx, insertQuery, email, password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == utils.PG_UNIQUE_CONSRAINT_VIOLATION_CODE {
			return &EmailAlreadyExistsErr{email: email}
		}

		return err
	}
	return nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	getQuery := `SELECT id, email, password, created_at FROM users WHERE email = $1`

	var user User

	err := r.db.QueryRowContext(ctx, getQuery, email).Scan(
		&user.Id,
		&user.Email,
		&user.Password,
		&user.Created_at,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &InvalidCredentialErr{}

		}

		return nil, err
	}

	return &user, nil
}
