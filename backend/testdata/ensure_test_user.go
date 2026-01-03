package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/backend/pkg/auth"
	"github.com/nexuscrm/shared/pkg/constants"
)

// EnsureTestUser creates or updates the E2E test user
func main() {
	// Get database connection
	db, err := database.GetInstance()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	testEmail := "admin@test.com"
	testPassword := "Admin123!"
	testName := "E2E Test Admin"

	// Hash password
	hashedPassword, err := auth.HashPassword(testPassword)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Get System Administrator profile ID
	// NOTE: This MUST exist (created by bootstrap.InitializeSystemData)
	// If it doesn't exist, something is seriously wrong - FAIL HARD
	var profileID string
	err = db.QueryRow(fmt.Sprintf("SELECT id FROM %s WHERE id = ? LIMIT 1", constants.TableProfile), constants.ProfileSystemAdmin).Scan(&profileID)
	if err != nil {
		log.Fatalf("CRITICAL: %s profile does not exist! Did bootstrap.InitializeSystemData() run? Error: %v", constants.ProfileSystemAdmin, err)
	}

	// Check if user exists
	var existingUserID string
	err = db.QueryRow(fmt.Sprintf("SELECT id FROM %s WHERE email = ?", constants.TableUser), testEmail).Scan(&existingUserID)

	if err == sql.ErrNoRows {
		// Create new user
		userID := services.GenerateID()
		_, err = db.Exec(fmt.Sprintf(`
			INSERT INTO %s (id, username, email, password, profile_id)
			VALUES (?, ?, ?, ?, ?)
		`, constants.TableUser), userID, testName, testEmail, hashedPassword, profileID)

		if err != nil {
			log.Fatalf("Failed to create test user: %v", err)
		}

		fmt.Printf("✅ Created test user: %s (ID: %s)\n", testEmail, userID)
	} else if err != nil {
		log.Fatalf("Database error: %v", err)
	} else {
		// Update existing user's password AND ProfileId (to ensure consistency)
		_, err = db.Exec(fmt.Sprintf(`
			UPDATE %s
			SET password = ?,
			    username = ?,
			    profile_id = ?
			WHERE email = ?
		`, constants.TableUser), hashedPassword, testName, profileID, testEmail)

		if err != nil {
			log.Fatalf("Failed to update test user: %v", err)
		}

		fmt.Printf("✅ Updated test user: %s (ID: %s, ProfileId: %s)\n", testEmail, existingUserID, profileID)
	}

	fmt.Println("\nTest user credentials:")
	fmt.Printf("  Email: %s\n", testEmail)
	fmt.Printf("  Password: %s\n", testPassword)
	fmt.Println("\nTest data setup complete! 🎉")
	os.Exit(0)
}
