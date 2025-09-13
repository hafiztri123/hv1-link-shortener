package url

import (
	"database/sql"
	"time"
)

type URL struct {
	ID        int64
	ShortCode sql.NullString
	LongURL   string
	CreatedAt time.Time
}

type CreateURLRequest struct {
	LongURL string `json:"long_url"`
}

type CreateURLResponse struct {
	ShortCode string `json:"short_code"`
}

type CreateURLRequest_Bulk struct {
	LongURLs []string `json:"long_urls"`
}
