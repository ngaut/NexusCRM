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
	"github.com/stretchr/testify/require"
)

// TestMasterDetail_CascadeDelete verifies that deleting a master record cascades to child records
func TestMasterDetail_CascadeDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 1. Setup
	conn, err := database.GetInstance()
	if err != nil {
		t.Logf("Skipping integration test: failed to connect to DB: %v", err)
		t.SkipNow()
	}
	db := conn.DB()

	// Initialize Services
	eventBus := NewEventBus()
	txManager := NewTransactionManager(conn)
	// NewSchemaManager takes *sql.DB
	schemaMgr := NewSchemaManager(conn.DB())
	metadataSvc := NewMetadataService(conn, schemaMgr)
	permSvc := NewPermissionService(conn, metadataSvc)

	// We need full persistence service
	ps := NewPersistenceService(conn, metadataSvc, permSvc, eventBus, txManager)

	// Context with Admin User
	ctx := context.Background()
	adminUser := &models.UserSession{
		ID:        "test-admin",
		Name:      "Test Admin",
		ProfileID: constants.ProfileSystemAdmin,
	}

	// Define Master Object (Project)
	masterObjName := fmt.Sprintf("project_%d", time.Now().UnixNano())
	masterCols := schemaMgr.GetStandardSystemColumns()

	masterDef := schema.TableDefinition{
		TableName:   masterObjName,
		Description: "Master Project",
		TableType:   string(constants.TableTypeCustomObject),
		Columns:     masterCols,
	}
	masterMeta := &models.ObjectMetadata{
		APIName:      masterObjName,
		Label:        "Master Project",
		PluralLabel:  "Master Projects",
		Description:  &masterDef.Description,
		IsCustom:     true,
		SharingModel: models.SharingModel(constants.SharingModelPublicReadWrite),
	}
	require.NoError(t, schemaMgr.CreateTableWithStrictMetadata(context.Background(), masterDef, masterMeta))
	// Clean up schema after test
	defer schemaMgr.DropTable(masterObjName)

	// Define Detail Object (Task)
	detailObjName := fmt.Sprintf("task_%d", time.Now().UnixNano())
	detailCols := schemaMgr.GetStandardSystemColumns()
	detailCols = append(detailCols,
		schema.ColumnDefinition{Name: "project_id", Type: "VARCHAR(36)", Nullable: true},
	)

	detailDef := schema.TableDefinition{
		TableName:   detailObjName,
		Description: "Detail Task",
		TableType:   "custom_object",
		Columns:     detailCols,
	}
	detailMeta := &models.ObjectMetadata{
		APIName:      detailObjName,
		Label:        "Detail Task",
		PluralLabel:  "Detail Tasks",
		Description:  &detailDef.Description,
		IsCustom:     true,
		SharingModel: models.SharingModel(constants.SharingModelPublicReadWrite),
	}
	require.NoError(t, schemaMgr.CreateTableWithStrictMetadata(context.Background(), detailDef, detailMeta))
	defer schemaMgr.DropTable(detailObjName)

	// 4. Update Metadata to enforce Master-Detail
	// CreateTableFromDefinition registers basic metadata. We need to update the relationship field.
	// We'll update _System_Field directly for simplicity to avoid constructing complex structs for SchemaManager.

	// Find Field ID dynamically to ensure proper targeting
	var fieldID string
	fieldIDQuery := fmt.Sprintf("SELECT id FROM %s WHERE object_id = (SELECT id FROM %s WHERE api_name = ?) AND api_name = ?",
		constants.TableField, constants.TableObject)
	err = db.QueryRow(fieldIDQuery, detailObjName, "project_id").Scan(&fieldID)
	require.NoError(t, err, "Failed to find field ID for project_id")

	updateQuery := fmt.Sprintf(`
		UPDATE %s 
		SET %s = ?, %s = ?, %s = ?, %s = ?
		WHERE id = ?
	`, constants.TableField,
		"type", "reference_to", "delete_rule", "is_master_detail")

	res, err := db.Exec(updateQuery,
		constants.FieldTypeLookup,              // type
		fmt.Sprintf("[\"%s\"]", masterObjName), // reference_to (JSON Array)
		constants.DeleteRuleCascade,            // delete_rule
		true,                                   // is_master_detail
		fieldID,
	)
	require.NoError(t, err)
	rows, _ := res.RowsAffected()
	require.Equal(t, int64(1), rows, "Metadata update should affect 1 row")

	require.NoError(t, metadataSvc.RefreshCache())

	// Create Data
	// Create Parent
	projectData := models.SObject{"name": "Alpha Project"}
	createdProject, err := ps.Insert(ctx, masterObjName, projectData, adminUser)
	require.NoError(t, err)
	projectID := createdProject[constants.FieldID].(string)

	// Create Child
	taskData := models.SObject{"name": "Sub Task 1", "project_id": projectID}
	createdTask, err := ps.Insert(ctx, detailObjName, taskData, adminUser)
	require.NoError(t, err)
	taskID := createdTask[constants.FieldID].(string)

	// Verify Child Exists
	var count int
	err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = ?", detailObjName), taskID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	// 6. Delete Parent (Should Cascade)
	t.Log("Deleting Parent Project...")
	err = ps.Delete(ctx, masterObjName, projectID, adminUser)
	require.NoError(t, err)

	// 7. Verify Parent Deleted (Soft Delete)
	var pDeleted bool
	pQuery := fmt.Sprintf("SELECT is_deleted FROM %s WHERE id = ?", masterObjName)
	err = db.QueryRow(pQuery, projectID).Scan(&pDeleted)
	assert.NoError(t, err)
	assert.True(t, pDeleted, "Parent should be soft deleted")

	// 8. Verify Child Deleted (Cascade Soft Delete)
	var tDeleted bool
	tQuery := fmt.Sprintf("SELECT is_deleted FROM %s WHERE id = ?", detailObjName)
	err = db.QueryRow(tQuery, taskID).Scan(&tDeleted)
	assert.NoError(t, err)
	assert.True(t, tDeleted, "Child should be soft deleted (Cascade)")
}
