package energy

import "time"

type LogResponse struct {
	Log map[string]LogEntry `json:"log"`
}

type LogEntry struct {
	Title     string `json:"title"`
	Timestamp int64  `json:"timestamp"`
	Data      struct {
		EnergyUsed float64 `json:"energy_used"`
	} `json:"data"`
}

type UserEnergyRecord struct {
	TornID     int64     `json:"torn_id" db:"torn_id"`
	Username   string    `json:"username" db:"username"`
	APIKey     string    `json:"-" db:"api_key"` // Never expose in JSON
	EnergyUsed float64   `json:"energy_used" db:"energy_used"`
	Date       time.Time `json:"date" db:"date"`
}
