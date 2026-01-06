package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nexuscrm/backend/internal/domain/schema"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/stretchr/testify/assert"
)

// TestSchemaManager_Integration_ACID creates a real table and verifies metadata registration
// This test requires a running TiDB instance configured in environment variables
func TestSchemaManager_Integration_ACID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 1. Setup Connection
	// We use the singleton getter which uses env vars (TIDB_HOST etc.)
	conn, err := database.GetInstance()
	if err != nil {
		t.Logf("Skipping integration test: failed to connect to DB: %v", err)
		t.SkipNow()
	}
	db := conn.DB()

	sm := NewSchemaManager(db)
	tableName := fmt.Sprintf("test_acid_obj_%d", time.Now().UnixNano())

	// Helper to cleanup
	cleanup := func() {
		// We use DropTable which also cleans metadata
		_ = sm.DropTable(tableName)
	}
	defer cleanup()

	// 2. Define Table
	def := schema.TableDefinition{
		TableName:   tableName,
		TableType:   string(constants.TableTypeCustomObject),
		Category:    "test",
		Description: "Test Object for ACID",
		Columns: []schema.ColumnDefinition{
			{Name: constants.FieldID, Type: "VARCHAR(36)", Nullable: false, PrimaryKey: true, LogicalType: "ID"}, // Added ID to satisfy check
			{Name: constants.FieldName, Type: "VARCHAR(255)", Nullable: false, LogicalType: string(constants.FieldTypeText)},
			{Name: "amount", Type: "DECIMAL(10,2)", Nullable: true, LogicalType: string(constants.FieldTypeCurrency)},
		},
	}

	// 3. Execute Create (Transactional)
	// 3. Execute Create (Transactional)
	t.Logf("Creating table %s...", tableName)

	// Create Object Metadata explicitly for strict mode
	desc := "Test Object for ACID"
	objMeta := &models.ObjectMetadata{
		APIName:      tableName,
		Label:        "Test Object",
		PluralLabel:  "Test Objects",
		Description:  &desc,
		IsCustom:     true,
		SharingModel: models.SharingModel(constants.SharingModelPublicReadWrite),
	}

	err = sm.CreateTableWithStrictMetadata(context.Background(), def, objMeta)
	assert.NoError(t, err, "CreateTableWithStrictMetadata should succeed")

	// 4. Verify Physical Table Existence
	// Using a direct query to information_schema or checking via SQL
	ctx := context.Background()
	var exists int
	checkTableQuery := "SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ? AND table_schema = DATABASE()"
	err = db.QueryRowContext(ctx, checkTableQuery, tableName).Scan(&exists)
	assert.NoError(t, err)
	assert.Equal(t, 1, exists, "Physical table must exist in database")

	// 5. Verify Metadata (_System_Object)
	var objID string
	var isCustom bool
	checkObjQuery := fmt.Sprintf("SELECT id, is_custom FROM %s WHERE api_name = ?", constants.TableObject)
	err = db.QueryRowContext(ctx, checkObjQuery, tableName).Scan(&objID, &isCustom)
	assert.NoError(t, err, "Object metadata must exist")
	assert.NotEmpty(t, objID)
	assert.True(t, isCustom, "Should be custom object")

	// 6. Verify Field Metadata (_System_Field)
	// Check 'name' field
	var fieldID string
	var fieldType string
	checkFieldQuery := fmt.Sprintf("SELECT id, type FROM %s WHERE object_id = ? AND api_name = ?", constants.TableField)

	err = db.QueryRowContext(ctx, checkFieldQuery, objID, "name").Scan(&fieldID, &fieldType)
	assert.NoError(t, err, "Field 'name' metadata must exist")
	assert.Equal(t, string(constants.FieldTypeText), fieldType)

	// Check 'amount' field
	err = db.QueryRowContext(ctx, checkFieldQuery, objID, "amount").Scan(&fieldID, &fieldType)
	assert.NoError(t, err, "Field 'amount' metadata must exist")
	assert.Equal(t, string(constants.FieldTypeCurrency), fieldType)

	// 7. Verify System Fields (Auto-registered)
	// ID, CreatedDate, etc.
	err = db.QueryRowContext(ctx, checkFieldQuery, objID, constants.FieldID).Scan(&fieldID, &fieldType)
	assert.NoError(t, err, "System field 'id' metadata must exist")

	t.Log("✅ Integration Test Passed: Table and Metadata created correctly.")
}
