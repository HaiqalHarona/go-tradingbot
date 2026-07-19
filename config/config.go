package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds trading bot configuration parameters and risk rules.
type Config struct {
	AlpacaKeyID      string
	AlpacaSecretKey  string
	AlpacaBaseURL    string
	StockPool        []string
	EquityAllocation float64
	StopLossPct      float64
	TakeProfitPct    float64
	MaxOpenPositions int
}

// LoadConfig validates required API credentials and parses bot risk settings.
func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	requiredVars := []string{"APCA_API_KEY_ID", "APCA_API_SECRET_KEY"}
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			return nil, fmt.Errorf("missing required environment variable: %s", v)
		}
	}

	baseURL := os.Getenv("APCA_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://paper-api.alpaca.markets"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/v2")
	baseURL = strings.TrimSuffix(baseURL, "/")

	cfg := &Config{
		AlpacaKeyID:      os.Getenv("APCA_API_KEY_ID"),
		AlpacaSecretKey:  os.Getenv("APCA_API_SECRET_KEY"),
		AlpacaBaseURL:    baseURL,
		StockPool:        getEnvSlice("STOCK_POOL", []string{"AAPL", "HLAL", "NVDA", "SPUS", "TSLA"}),
		EquityAllocation: getEnvFloat("EQUITY_ALLOCATION_PCT", 0.02),
		StopLossPct:      getEnvFloat("STOP_LOSS_PCT", 0.015),
		TakeProfitPct:    getEnvFloat("TAKE_PROFIT_PCT", 0.030),
		MaxOpenPositions: getEnvInt("MAX_OPEN_POSITIONS", 4),
	}

	return cfg, nil
}

// VerifyEnvironment checks that required variables are present.
func VerifyEnvironment() error {
	_, err := LoadConfig()
	return err
}

func getEnvSlice(key string, defaultVal []string) []string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	rawItems := strings.Split(val, ",")
	var result []string
	for _, item := range rawItems {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return defaultVal
	}
	return result
}

func getEnvFloat(key string, defaultVal float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return defaultVal
	}
	return f
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}