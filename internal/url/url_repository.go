package url

import (
	"context"
	"database/sql"
)

type URLRepository interface {
	Insert(ctx context.Context, longURL string) (int64, error)
	UpdateShortCode(ctx context.Context, id int64, shortCode string) error
	GetByID(ctx context.Context, id int64) (*URL, error)
}

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Insert(ctx context.Context, longURL string) (int64, error) {
	var id int64
	query := "INSERT INTO urls (long_url) VALUES ($1) RETURNING id"
	err := r.DB.QueryRowContext(ctx, query, longURL).Scan(&id)
	return id, err
}

func (r *Repository) UpdateShortCode(ctx context.Context, id int64, shortCode string) error {
	query := "UPDATE urls set short_code = $1 WHERE id = $2"
	_, err := r.DB.ExecContext(ctx, query, shortCode, id)
	return err
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*URL, error) {
	query := "SELECT id, short_code, long_url, created_at FROM urls where id = $1"

	var url URL

	err := r.DB.QueryRowContext(ctx, query, id).Scan(&url.ID, &url.ShortCode, &url.LongURL, &url.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &url, nil
}
