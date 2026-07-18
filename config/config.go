package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// VerifyEnvironment checks that Docker or the system has the required variables.
func VerifyEnvironment() error {
	
	// Silently fails on live bot
	_ = godotenv.Load()

	// The official Alpaca SDK explicitly looks for these variable names
	requiredVars := []string{"APCA_API_KEY_ID", "APCA_API_SECRET_KEY"}

	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			return fmt.Errorf("missing required environment variable: %s", v)
		}
	}

	return nil
}