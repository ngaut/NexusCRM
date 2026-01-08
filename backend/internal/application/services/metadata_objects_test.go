package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestMetadataService_Objects_ThemeColor_Integration(t *testing.T) {
	// 1. Setup
	conn, _ := SetupIntegrationTest(t)
	db := conn.DB()

	// Services
	repo := persistence.NewMetadataRepository(db)
	schemaRepo := persistence.NewSchemaRepository(db)
	schemaMgr := NewSchemaManager(schemaRepo)
	ms := NewMetadataService(repo, schemaMgr)

	apiName := fmt.Sprintf("color_test_%d", time.Now().UnixNano())
	cleanup := func() {
		_ = schemaMgr.DropTable(apiName) // Drops metadata too
	}
	defer cleanup()

	// 3. Define Object with ThemeColor
	color := "#123456"
	objDef := models.ObjectMetadata{
		APIName:      apiName,
		Label:        "Color Test",
		PluralLabel:  "Color Tests",
		ThemeColor:   &color,
		SharingModel: constants.SharingModelPrivate,
		Fields: []models.FieldMetadata{
			{APIName: "name", Label: "Name", Type: constants.FieldTypeText, IsNameField: true},
		},
	}

	// 4. Create Object
	err := ms.CreateSchemaOptimized(context.Background(), &objDef)
	assert.NoError(t, err)

	// 5. Verify Metadata Persistence
	// Direct DB Query to _System_Object
	ctx := context.Background()
	var storedColor string
	query := fmt.Sprintf("SELECT theme_color FROM %s WHERE api_name = ?", constants.TableObject)
	err = db.QueryRowContext(ctx, query, apiName).Scan(&storedColor)

	assert.NoError(t, err)
	assert.Equal(t, color, storedColor, "Theme color should be persisted correctly")
}
