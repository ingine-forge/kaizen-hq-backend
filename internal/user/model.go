package user

import "time"

type User struct {
	TornID    int       `json:"torn_id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Skip in JSON responses
	APIKey    string    `json:"api_key,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}