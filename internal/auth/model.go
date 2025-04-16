package auth

import (
	"kaizen-hq/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	TornID    int       `json:"torn_id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Skip in JSON responses
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	TornID   int    `json:"torn_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
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
