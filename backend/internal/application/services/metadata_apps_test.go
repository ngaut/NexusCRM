package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/nexuscrm/shared/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataService_Apps_IsDefault(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 1. Setup
	conn, ms := SetupIntegrationTest(t)
	db := conn.DB()

	// Ensure tables exist (SchemaManager usually does this)
	// For integration test, we assume core tables exist or we create them if missing.
	// But usually InitializeSchema runs at startup.
	// Let's create temp tables or rely on existing ones.
	// Ideally we use unique IDs to avoid collision.

	t.Run("Create App With Default", func(t *testing.T) {
		appID := fmt.Sprintf("app_TestDefault_%d", time.Now().UnixNano())
		app := &models.AppConfig{
			ID:          appID,
			Label:       "Test Default App",
			Description: "Unit Test App",
			Icon:        "Star",
			IsDefault:   true,
		}

		err := ms.CreateApp(app)
		require.NoError(t, err)
		defer ms.DeleteApp(appID)

		// Verify DB
		var isDefault bool
		err = db.QueryRow("SELECT is_default FROM _System_App WHERE id = ?", appID).Scan(&isDefault)
		require.NoError(t, err)
		assert.True(t, isDefault, "IsDefault should be true")
	})

	t.Run("Update App To Default", func(t *testing.T) {
		// Create non-default app
		appID := fmt.Sprintf("app_TestUpdate_%d", time.Now().UnixNano())
		initialApp := &models.AppConfig{
			ID:        appID,
			Label:     "Test Update",
			IsDefault: false,
		}
		err := ms.CreateApp(initialApp)
		require.NoError(t, err)
		defer ms.DeleteApp(appID)

		updates := &models.AppConfig{
			ID:        appID,
			IsDefault: true,
		}

		err = ms.UpdateApp(appID, updates)
		require.NoError(t, err)

		// Verify DB
		var isDefault bool
		err = db.QueryRow("SELECT is_default FROM _System_App WHERE id = ?", appID).Scan(&isDefault)
		require.NoError(t, err)
		assert.True(t, isDefault, "IsDefault should update to true")
	})
}
