package user

import (
	"context"
	"database/sql"
	"errors"
	"hafiztri123/app-link-shortener/internal/utils"

	"github.com/jackc/pgx/v5/pgconn"
)

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
			return err
		}

		return err
	}
	return nil
}
