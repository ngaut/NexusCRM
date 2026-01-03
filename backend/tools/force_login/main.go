package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/backend/pkg/auth"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: force_login <user_id>")
	}
	userID := os.Args[1]

	// Initialize DB
	dbConn, err := database.GetInstance()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Fetch User
	var user struct {
		ID        string
		Username  string
		Email     string
		ProfileId string
		RoleId    sql.NullString
	}
	query := fmt.Sprintf("SELECT %s, %s, %s, %s, %s FROM %s WHERE %s = ? LIMIT 1",
		constants.FieldID, constants.FieldUsername, constants.FieldEmail, constants.FieldProfileID, constants.FieldRoleID,
		constants.TableUser, constants.FieldID)

	err = dbConn.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.Email, &user.ProfileId, &user.RoleId)
	if err != nil {
		log.Fatalf("Failed to find user %s: %v", userID, err)
	}

	// Create Session Object
	var roleIdPtr *string
	if user.RoleId.Valid {
		roleIdPtr = &user.RoleId.String
	}

	userSession := auth.UserSession{
		ID:        user.ID,
		Name:      user.Username,
		Email:     user.Email,
		ProfileId: user.ProfileId,
		RoleId:    roleIdPtr,
	}

	// Generate Token
	token, err := auth.GenerateToken(userSession)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	// Decode to get Claims (JTI/Expiry)
	claims, _ := auth.DecodeToken(token)
	expiresAt := time.Unix(claims.ExpiresAt.Unix(), 0)

	// Insert Session into DB
	sessionStruct := models.SystemSession{
		ID:           claims.RegisteredClaims.ID,
		UserID:       user.ID,
		Token:        token,
		ExpiresAt:    expiresAt,
		IPAddress:    "127.0.0.1",
		UserAgent:    "E2E Test Force Login",
		IsRevoked:    false,
		LastActivity: time.Now(),
	}

	// Manual SQL Insert to avoid dependency on PersistenceService complexity in this script
	// Assuming _System_Session table structure
	insertQuery := fmt.Sprintf(`
		INSERT INTO %s 
		(id, user_id, token, expires_at, ip_address, user_agent, is_revoked, last_activity, created_date, last_modified_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, constants.TableSession)

	_, err = dbConn.Exec(insertQuery,
		sessionStruct.ID,
		sessionStruct.UserID,
		sessionStruct.Token,
		sessionStruct.ExpiresAt,
		sessionStruct.IPAddress,
		sessionStruct.UserAgent,
		sessionStruct.IsRevoked,
		sessionStruct.LastActivity,
	)

	if err != nil {
		log.Fatalf("Failed to persist session: %v", err)
	}

	// Output Token to stdout
	fmt.Print(token)
}
