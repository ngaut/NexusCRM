package services

import (
	"testing"

	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// SetupIntegrationTest initializes a database connection and basic services for integration testing.
// It skips the test if the database is not available.
// It takes `t` to handle skipping.
func SetupIntegrationTest(t *testing.T) (*database.TiDBConnection, *MetadataService) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	conn, err := database.GetInstance()
	if err != nil {
		t.Skip("Database not available: " + err.Error())
	}

	db := conn.DB()
	sm := NewSchemaManager(db)
	ms := NewMetadataService(conn, sm)

	return conn, ms
}

// GetTestUser returns a UserSession suitable for testing permissions.
func GetTestUser(id string, profileID string) *models.UserSession {
	if profileID == "" {
		profileID = constants.ProfileSystemAdmin
	}
	return &models.UserSession{
		ID:        id,
		Name:      "Test User " + id,
		ProfileID: profileID,
	}
}
