package persistence

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup DB
	conn, err := database.GetInstance()
	require.NoError(t, err)
	db := conn.DB()

	repo := NewRecordRepository(db)

	// Create a test table explicitly
	tableName := fmt.Sprintf("test_record_repo_%d", time.Now().UnixNano())

	// Use raw SQL to create table (simplest for this integration test)
	// We mimic a standard object layout
	t.Logf("Creating test table: %s", tableName)
	_, err = db.Exec(fmt.Sprintf(`
		%s %s (
			%s %s %s,
			%s %s,
			%s %s,
			%s %s,
			%s %s,
			%s %s %s 0,
			%s %s,
			%s %s
		)
	`, KeywordCreateTable, tableName,
		constants.FieldID, SQLTypeVarchar36, KeywordPrimaryKey,
		constants.FieldSysGroup_Name, SQLTypeVarchar255,
		constants.FieldSysGroup_Label, SQLTypeVarchar255,
		constants.FieldSysGroup_Type, SQLTypeVarchar50,
		constants.FieldSysGroup_Email, SQLTypeVarchar255,
		constants.FieldIsDeleted, SQLTypeTinyInt, KeywordDefault,
		constants.FieldCreatedDate, SQLTypeDateTime,
		constants.FieldLastModifiedDate, SQLTypeDateTime))
	require.NoError(t, err, "Failed to create test table")

	recordID := fmt.Sprintf("test_rec_%d", time.Now().UnixNano())

	cleanup := func() {
		// Clean up table
		t.Logf("Dropping test table: %s", tableName)
		_, _ = db.Exec(fmt.Sprintf("%s %s %s", KeywordDropTable, KeywordIfExists, tableName))
	}
	defer cleanup()

	ctx := context.Background()

	// 1. Insert
	newRecord := models.SObject{
		constants.FieldID:               recordID,
		constants.FieldSysGroup_Name:    "Test Group",
		constants.FieldSysGroup_Label:   "Test Label",
		constants.FieldSysGroup_Type:    "Queue",
		constants.FieldSysGroup_Email:   "test@example.com",
		constants.FieldCreatedDate:      time.Now(),
		constants.FieldLastModifiedDate: time.Now(),
	}

	err = repo.Insert(ctx, nil, tableName, newRecord)
	assert.NoError(t, err, "Insert should succeed")

	// 2. Exists
	exists, err := repo.Exists(ctx, nil, tableName, recordID)
	assert.NoError(t, err)
	assert.True(t, exists, "Record should exist")

	// 3. FindOne
	rec, err := repo.FindOne(ctx, nil, tableName, recordID)
	assert.NoError(t, err)
	assert.NotNil(t, rec)
	assert.Equal(t, "Test Group", rec[constants.FieldSysGroup_Name])

	// 4. Update
	updates := models.SObject{
		constants.FieldSysGroup_Name: "Updated Group",
	}
	err = repo.Update(ctx, nil, tableName, recordID, updates)
	assert.NoError(t, err)

	recUpdated, err := repo.FindOne(ctx, nil, tableName, recordID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Group", recUpdated[constants.FieldSysGroup_Name])

	// 5. Delete
	err = repo.Delete(ctx, nil, tableName, recordID)
	assert.NoError(t, err)

	existsAfter, err := repo.Exists(ctx, nil, tableName, recordID)
	assert.NoError(t, err)
	assert.False(t, existsAfter, "Record should be deleted")
}
