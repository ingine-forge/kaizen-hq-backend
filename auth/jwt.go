package auth

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// TokenExpiration defines how long tokens are valid
const TokenExpiration = 24 * time.Hour

// GetJWTSecret returns the secret key from environment variables
func GetJWTSecret() []byte {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		if gin.Mode() == gin.ReleaseMode {
			log.Fatal("JWT_SECRET environment variable is required in production")
		}

		// Fallback for development only
		return []byte("default-fallback-secret-for-dev-only")
	}

	return []byte(secretKey)
}

// Claims represents the JWT claims
type Claims struct {
	TornID   int    `json:"torn_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(user *User) (string, error) {
	// Set expiration time
	expirationTime := time.Now().Add(TokenExpiration)

	// Create claims
	claims := &Claims{
		TornID:   user.TornID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	tokenString, err := token.SignedString(GetJWTSecret())
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken checks if a token is valid and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate sigining method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return GetJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	// Extract claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
