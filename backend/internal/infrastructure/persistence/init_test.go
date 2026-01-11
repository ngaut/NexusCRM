package persistence_test

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env file for integration tests
	// The .env file is at project root (nexuscrm/.env), not in backend/
	// When tests run from internal/infrastructure/persistence/, we need to go up to find it
	paths := []string{
		"../../../../.env", // From internal/infrastructure/persistence/ to nexuscrm/
		"../../../.env",    // From internal/infrastructure/
		"../../.env",       // Fallback
		"../.env",          // Another fallback
		".env",             // Current directory
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			if err := godotenv.Load(p); err == nil {
				log.Printf("📁 Loaded .env from %s for tests", p)
				return
			}
		}
	}

	log.Println("⚠️  No .env file found for tests - database tests may fail")
}
