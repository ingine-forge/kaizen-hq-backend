package energy

import "time"

// Log response represents the log received from the api
type LogResponse struct {
	Log map[string]LogEntry `json:"log"`
}

// LogEntry represents each log entry inside the log response
type LogEntry struct {
	Title     string `json:"title"`
	Timestamp int64  `json:"timestamp"`
	Data      struct {
		EnergyUsed float64 `json:"energy_used"`
	} `json:"data"`
}

// UserEnergyRecord represents the table for energy for each user in the database
type UserEnergyRecord struct {
	TornID     int64     `json:"torn_id" db:"torn_id"`
	Username   string    `json:"username" db:"username"`
	APIKey     string    `json:"-" db:"api_key"` // Never expose in JSON
	EnergyUsed float64   `json:"energy_used" db:"energy_used"`
	Date       time.Time `json:"date" db:"date"`
}

// EnergyUsageRecord represents the table for energy_usage
type EnergyUsage struct {
	EnergyUsed float64   `json:"energy" db:"energy"`
	Date       time.Time `json:"date" db:"date"`
}

// UserProfile represents the user profile received from the api
// ? Maybe this should have it's own repository?
// TODO: look more into it
type UserProfile struct {
	PlayerID int    `json:"player_id"`
	Name     string `json:"name"`
	Signup   string `json:"signup"` // "YYYY-MM-DD HH:MM:SS"
}

// EnergyQueryParams represents the from and to time to be used as parameters for the endpoint
type EnergyQueryParams struct {
	From int64 `json:"from"` // Unix timestamp
	To   int64 `json:"to"`   // Unix timestamp
}

type EnergyUsageRequest struct {
	TornID int `json:"torn_id"`
}
