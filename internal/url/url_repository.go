package url

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hafiztri123/app-link-shortener/internal/utils"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

type URLRepository interface {
	FindOrCreateShortCode(context.Context, string, uint64, *int64) (string, error)
	FindOrCreateShortCode_Bulk(context.Context, []string, uint64, *int64) ([]CreateShortCodeBulkResult, error)
	GetByID(context.Context, int64) (*URL, error)
	GetByUserID_Bulk(context.Context, int64) ([]*URL, error)
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

func (r *Repository) GetByUserID_Bulk(ctx context.Context, userId int64) ([]*URL, error) {
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
	tx, err := r.DB.BeginTx(ctx, nil)
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

func (r *Repository) FindOrCreateShortCode_Bulk(ctx context.Context, longURLs []string, idOffset uint64, userId *int64) ([]CreateShortCodeBulkResult, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if len(longURLs) > 0 {
		err = r.insertURLsIgnoreConflicts(ctx, tx, longURLs, userId)
		if err != nil {
			return nil, err
		}
	}

	placeholder := utils.SelectPlaceholderBuilder(len(longURLs), 2)
	selectQuery := fmt.Sprintf(`
		SELECT id, short_code, long_url
		FROM urls
		WHERE user_id IS NOT DISTINCT FROM $1 AND long_url IN (%s)
	`, placeholder)

	args := []any{userId}
	args = append(args, utils.StringSliceToAny(longURLs)...)

	rows, err := tx.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var urlsToUpdate []struct {
		id      int64
		longURL string
	}

	urlToShortCode := make(map[string]string)

	for rows.Next() {
		var id int64
		var shortCode sql.NullString
		var longURL string

		if err := rows.Scan(&id, &shortCode, &longURL); err != nil {
			return nil, err
		}

		if shortCode.Valid {
			urlToShortCode[longURL] = shortCode.String
		} else {
			urlsToUpdate = append(urlsToUpdate, struct {
				id      int64
				longURL string
			}{
				id:      id,
				longURL: longURL,
			})
		}
	}

	if len(urlsToUpdate) > 0 {
		err = r.bulkUpdateShortCodes(ctx, tx, urlsToUpdate, idOffset, urlToShortCode)
		if err != nil {
			return nil, err
		}
	}

	result := make([]CreateShortCodeBulkResult, len(longURLs))
	for i, url := range longURLs {
		result[i] = CreateShortCodeBulkResult{
			LongURL:   url,
			ShortCode: urlToShortCode[url],
		}
	}

	return result, tx.Commit()

}

func (r *Repository) insertURLsIgnoreConflicts(ctx context.Context, tx *sql.Tx, longURLs []string, userId *int64) error {
	fieldCount := 2
	placeholderGroups := make([]string, len(longURLs))
	args := make([]any, 0, len(longURLs)*fieldCount)

	for i, url := range longURLs {
		baseIndex := i * fieldCount
		placeholderGroups[i] = fmt.Sprintf("($%d, $%d)", baseIndex+1, baseIndex+2)
		args = append(args, url, userId)
	}

	query := fmt.Sprintf(`
		INSERT INTO urls (long_url, user_id)
		VALUES %s
		ON CONFLICT (long_url) DO NOTHING
	`, strings.Join(placeholderGroups, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) bulkUpdateShortCodes(ctx context.Context, tx *sql.Tx, urlsToUpdate []struct {
	id      int64
	longURL string
}, idOffset uint64, urlToShortCode map[string]string) error {
	if len(urlsToUpdate) == 0 {
		return nil
	}

	caseClauses := make([]string, len(urlsToUpdate))
	inClauses := make([]string, len(urlsToUpdate))
	args := make([]any, 0, len(urlsToUpdate)*2)

	for i, item := range urlsToUpdate {
		shortCode := toBase62(uint64(item.id) + idOffset)
		urlToShortCode[item.longURL] = shortCode

		caseClauses[i] = fmt.Sprintf("WHEN $%d THEN $%d", i*2+1, i*2+2)
		inClauses[i] = fmt.Sprintf("$%d", i*2+1)
		args = append(args, item.id, shortCode)
	}

	query := fmt.Sprintf(`
		UPDATE urls
		SET short_code = CASE id %s END
		WHERE id IN (%s)
	`, strings.Join(caseClauses, " "), strings.Join(inClauses, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}
