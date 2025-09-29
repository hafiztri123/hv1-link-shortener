package models

import "time"

type Click struct {
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	IPAddress string    `json:"ip_address"`
	Referer   string    `json:"referrer"`
	UserAgent string    `json:"user_agent"`
	Device    string    `json:"device"`
	OS        string    `json:"os"`
	Browser   string    `json:"browser"`
	Country   string    `json:"country"`
	City      string    `json:"city"`
}
