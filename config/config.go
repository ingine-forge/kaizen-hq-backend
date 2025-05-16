package config

import (
	"os"
	"strconv"
)

type TornAPIConfig struct {
	BaseURL string
}

type CorsConfig struct {
	ClientDomain string
	ClientPort   string
}

type Config struct {
	DBURL           string
	JWTSecret       string
	BcryptCost      int
	DiscordBotToken string
	TornAPI         TornAPIConfig
}

func Load() *Config {
	// Load from environment variables or .env file
	return &Config{
		DBURL:           os.Getenv("DB_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		BcryptCost:      getInt("BCRYPT_COST", 10),
		DiscordBotToken: os.Getenv("DISCORD_BOT_TOKEN"),
		TornAPI: TornAPIConfig{
			BaseURL: "https://api.torn.com/",
		},
		CORS: CorsConfig{
			ClientDomain: os.Getenv("CLIENT_DOMAIN"),
			ClientPort:   os.Getenv("CLIENT_PORT"),
		},
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
