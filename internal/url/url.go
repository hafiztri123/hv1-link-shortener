package url

import "time"

type URL struct {
	ID        int64
	ShortURL  string
	LongURL   string
	CreatedAt time.Time
}

type CreateURLRequest struct {
	LongURL string `json:"long_url"`
}

type CreateURLResponse struct {
	ShortURL string `json:"short_url"`
}
