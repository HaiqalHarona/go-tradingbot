package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config is an exported type (starts with capital letter) that holds
// configuration values. Exported identifiers are visible from other packages.
type Config struct {
	AlpacaKeyID     string
	AlpacaSecretKey string
	AlpacaBaseURL   string
}

// LoadConfig loads configuration from environment variables and returns
// a pointer to Config and an error if something goes wrong.
// Function signatures with (type, error) are common in Go for returning
// a result and an error value.
func LoadConfig() (*Config, error) {
	// Load .env file into environment variables. This returns (error),
	// and the blank identifier `_` is used to explicitly ignore that value.
	_ = godotenv.Load()

	// &Config{...} is a composite literal; the & returns a pointer to the struct.
	// os.Getenv reads an environment variable and returns a string (empty string if not set).
	cfg := &Config{
		AlpacaKeyID:     os.Getenv("ALPACA_KEY_ID"),
		AlpacaSecretKey: os.Getenv("ALPACA_SECRET_KEY"),
		AlpacaBaseURL:   os.Getenv("ALPACA_BASE_URL"),
	}

	// Validate credentials. In Go, the empty string "" is the zero value for string.
	if cfg.AlpacaKeyID == "" || cfg.AlpacaSecretKey == "" || cfg.AlpacaBaseURL == "" {
		// fmt.Errorf formats and returns an error value.
		return nil, fmt.Errorf("missing required environment variables")
	}

	// Return the pointer to the config and a nil error when successful.
	return cfg, nil
}
