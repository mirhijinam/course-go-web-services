package models

import "time"

var SessionLifetime time.Duration = time.Hour * 24 * 7

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}
