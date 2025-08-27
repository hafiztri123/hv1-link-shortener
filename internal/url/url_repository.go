package url

import (
	"context"
	"database/sql"
)

type URLRepository interface {
	Insert(ctx context.Context, longURL string) (int64, error)
	UpdateShortCode(ctx context.Context, id int64, shortURL string) error
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Insert(ctx context.Context, longURL string) (int64, error) {
	var id int64
	query := "INSERT INTO urls (long_url) VALUES ($1) RETURNING id"
	err := r.db.QueryRowContext(ctx, query, longURL).Scan(&id)
	return id, err
}

func (r *Repository) UpdateShortCode(ctx context.Context, id int64, shortURL string) error {
	query := "UPDATE urls set short_url = $1 WHERE id = $2"
	_, err := r.db.ExecContext(ctx, query, shortURL, id)
	return err
}
