package url

import (
	"context"
	"database/sql"
	"errors"
	"hafiztri123/app-link-shortener/internal/utils"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"
)

type URLRepository interface {
	FindOrCreateShortCode(context.Context, string, uint64, *int64) (string, error)
	GetByID(context.Context, int64) (*URL, error)
	GetByUserIDBulk(context.Context, int64) ([]*URL, error)
}

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
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

func (r *Repository) GetByUserIDBulk(ctx context.Context, userId int64) ([]*URL, error) {
	fetchQuery := `SELECT id, short_code, long_url, created_at FROM urls WHERE user_id = $1`

	rows, err := r.DB.QueryContext(ctx, fetchQuery, userId)
	if err != nil {
		slog.Error("database operation error occured", "error", err)
		return nil, err
	}
	defer rows.Close()

	var urls []*URL

	for rows.Next() {
		var url URL

		err := rows.Scan(&url.ID, &url.ShortCode, &url.LongURL, &url.CreatedAt)
		if err != nil {
			return nil, err
		}

		urls = append(urls, &url)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil

}

func (r *Repository) FindOrCreateShortCode(ctx context.Context, longURL string, idOffset uint64, userId *int64) (string, error) {
	tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		slog.Error("Failed to begin transaction", "error", err)
		return "", err
	}

	defer tx.Rollback()

	var shortCode sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT short_code FROM urls WHERE long_url = $1 AND user_id IS NOT DISTINCT FROM $2`, longURL, userId).Scan(&shortCode)

	if err == nil && shortCode.Valid {
		return shortCode.String, nil
	}

	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	var id int64
	err = tx.QueryRowContext(ctx, `INSERT INTO urls (long_url, user_id) VALUES ($1, $2) RETURNING id`, longURL, userId).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == utils.PG_UNIQUE_CONSRAINT_VIOLATION_CODE {
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
