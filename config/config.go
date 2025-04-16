package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBURL string
	JWTSecret string
	BcryptCost int
}

func Load() *Config {
	// Load from environment variables or .env file
	return &Config{
		DBURL: os.Getenv("DB_URL"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		BcryptCost: getInt("BCRYPT_COST", 10),
	}
}

func getInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}

	return defaultValue
}