package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestMetadataService_Themes_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 1. Setup Connection
	conn, err := database.GetInstance()
	if err != nil {
		t.SkipNow()
	}
	db := conn.DB()

	// 2. Initialize ServiceManager (or just MetadataService)
	// We need SchemaManager via NewSchemaManager(db *sql.DB) but NewMetadataService takes connection struct
	// This implies we need to match signatures.
	// We can't easily construct a fake TiDBConnection because it's in infrastructure package.
	// But conn from GetInstance IS valid.

	// Helper to clean up
	ctx := context.Background()
	themeName := fmt.Sprintf("Test Theme %d", time.Now().UnixNano())

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM "+constants.TableTheme+" WHERE name = ?", themeName)
	}()

	sm := NewSchemaManager(db)
	ms := NewMetadataService(conn, sm)

	// 3. Test CreateTheme (UpsertTheme)
	colors := map[string]interface{}{
		"brand": "#ff0000",
		"text":  "#000000",
	}
	newTheme := models.Theme{
		Name:     themeName,
		IsActive: false,
		Colors:   colors,
		Density:  "compact",
	}

	err = ms.UpsertTheme(context.Background(), &newTheme)
	assert.NoError(t, err)
	// We need to fetch it to get ID because UpsertTheme by Name likely generated one
	// But we passed empty ID, so it generated one. We need to query it back by Name
	var fetchedID string
	err = db.QueryRowContext(ctx, "SELECT id FROM "+constants.TableTheme+" WHERE name = ?", themeName).Scan(&fetchedID)
	assert.NoError(t, err)
	newTheme.ID = fetchedID

	// 4. Test GetActiveTheme
	_, err = ms.GetActiveTheme(context.Background())
	assert.NoError(t, err)

	// 5. Test ActivateTheme
	err = ms.ActivateTheme(context.Background(), newTheme.ID)
	assert.NoError(t, err)

	// Verify it is active
	active2, err := ms.GetActiveTheme(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, active2)
	assert.Equal(t, newTheme.ID, active2.ID)
	assert.True(t, active2.IsActive)
}
