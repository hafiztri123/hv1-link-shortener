package analytics

import "time"

type Click struct {
	Timestamp time.Time `json:"timestamp"`
	URLPath   string    `json:"url_path"`
	IPAddress string    `json:"ip_address"`
	Referrer  string    `json:"referrer"`
	UserAgent string    `json:"user_agent"`
	Device    string    `json:"device"`
	OS        string    `json:"os"`
	Browser   string    `json:"browser"`
	Country   string    `json:"country"`
	City      string    `json:"city"`
}

type ClickBatch struct {
	Clicks []Click
	MessageIDs []string
}
