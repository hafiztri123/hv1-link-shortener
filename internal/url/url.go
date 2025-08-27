package url

import "time"

type URL struct {
	ID        int64
	ShortURL  string
	LongURL   string
	CreatedAt time.Time
}
