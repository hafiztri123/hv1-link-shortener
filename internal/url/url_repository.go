package url

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)


const PG_UNIQUE_CONSRAINT_VIOLATION_CODE = "23505"

type URLRepository interface {
	Insert(ctx context.Context, longURL string) (int64, error)
	FindOrCreateShortCode(ctx context.Context, longURL string, idOffset uint64) (string, error)
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

func (r *Repository) FindOrCreateShortCode(ctx context.Context, longURL string, idOffset uint64) (string, error) {
	tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return "", err
	}

	defer tx.Rollback()

	var shortCode sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT short_code FROM urls WHERE long_url = $1`, longURL).Scan(&shortCode)

	if err == nil && shortCode.Valid {
		return shortCode.String, nil
	}

	if err != sql.ErrNoRows {
		return "", err
	}

	var id int64
	err = tx.QueryRowContext(ctx, `INSERT INTO urls (long_url) VALUES ($1) RETURNING id`, longURL).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == PG_UNIQUE_CONSRAINT_VIOLATION_CODE {
			tx.Rollback()
			var existingShortCode string
			err = r.DB.QueryRowContext(ctx, `SELECT short_code FROM urls WHERE long_url = $1`, longURL).Scan(&existingShortCode)
			return existingShortCode, err
		}

		return "", err
	}

	newShortcode := toBase62(uint64(id) + idOffset)
	_, err = tx.ExecContext(ctx, `UPDATE urls SET short_code = $1 WHERE id = $2`, newShortcode, id)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return newShortcode, nil
}
