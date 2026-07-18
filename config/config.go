package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config
type Config struct {
	AlpacaKeyID     string
	AlpacaSecretKey string
	AlpacaBaseURL   string
}

// Load Config
func LoadConfig() (*Config, error) {
	// load .env file
    _ = godotenv.Load()

	cfg := &Config{
		AlpacaKeyID:     os.Getenv("ALPACA_KEY_ID"),
		AlpacaSecretKey: os.Getenv("ALPACA_SECRET_KEY"),
		AlpacaBaseURL:   os.Getenv("ALPACA_BASE_URL"),
	}

	// Validate Creds
	if cfg.AlpacaKeyID == "" || cfg.AlpacaSecretKey == "" || cfg.AlpacaBaseURL == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	return cfg, nil
}
