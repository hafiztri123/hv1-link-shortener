package metadata

import (
	"context"
	"database/sql"
	"fmt"
	"hpj/hv1-link-shortener/shared/models"
	"strings"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) InsertMetadata(ctx context.Context, data *models.Click) error {
	stmt :=
		`INSERT INTO clicks 
	(
	url_path,
	ip_address, 
	referer, 
	user_agent, 
	device, 
	os, 
	browser, 
	country, 
	city, 
	timestamp
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, stmt,
		data.Path,
		data.IPAddress,
		data.Referer,
		data.UserAgent,
		data.Device,
		data.OS,
		data.Browser,
		data.Country,
		data.City,
		data.Timestamp,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) InsertMetadataBatch(ctx context.Context, datas []*models.Click) error {
	value := make([]string, 0, len(datas))
	args := make([]any, 0, len(datas)*10)

	for i, data := range datas {
		value = append(value, fmt.Sprintf(
			`($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)`,
			i*10+1,
			i*10+2,
			i*10+3,
			i*10+4,
			i*10+5,
			i*10+6,
			i*10+7,
			i*10+8,
			i*10+9,
			i*10+10,
		))

		args = append(args,
			data.Path,
			data.IPAddress,
			data.Referer,
			data.UserAgent,
			data.Device,
			data.OS,
			data.Browser,
			data.Country,
			data.City,
			data.Timestamp,
		)
	}

	query := fmt.Sprintf(
		`INSERT INTO clicks 
	(
	url_path,
	ip_address, 
	referer, 
	user_agent, 
	device, 
	os, 
	browser, 
	country, 
	city, 
	timestamp
	) VALUES %s`, strings.Join(value, ","))

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
