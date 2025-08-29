package url

import "time"

type URL struct {
	ID        int64
	ShortCode string
	LongURL   string
	CreatedAt time.Time
}

type CreateURLRequest struct {
	LongURL string `json:"long_url"`
}

type CreateURLResponse struct {
	ShortCode string `json:"short_url"`
}
