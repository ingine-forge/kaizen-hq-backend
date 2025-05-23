package account

import (
	"kaizen-hq/config"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	TornID    int       `json:"torn_id"`
	Password  string    `json:"-"` // Skip in JSON responses
	APIKey    string    `json:"api_key,omitempty"`
	DiscordID string    `json:"discord_id"`
	CreatedAt time.Time `json:"created_at"`
	LastLogin time.Time `json:"last_login"`
}

// HashPassword takes a plain text password and creates a hashed version
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), config.Load().BcryptCost)
	return string(bytes), err
}

// CheckPasswordHash compares a password with a hash to see if they match
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
