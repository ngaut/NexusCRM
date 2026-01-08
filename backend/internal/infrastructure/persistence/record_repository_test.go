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
		CREATE TABLE %s (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255),
			label VARCHAR(255),
			type VARCHAR(50),
			email VARCHAR(255),
			is_deleted TINYINT DEFAULT 0,
			created_date DATETIME,
			last_modified_date DATETIME
		)
	`, tableName))
	require.NoError(t, err, "Failed to create test table")

	recordID := fmt.Sprintf("test_rec_%d", time.Now().UnixNano())

	cleanup := func() {
		// Clean up table
		t.Logf("Dropping test table: %s", tableName)
		_, _ = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
	}
	defer cleanup()

	ctx := context.Background()

	// 1. Insert
	newRecord := models.SObject{
		constants.FieldID:    recordID,
		"name":               "Test Group",
		"label":              "Test Label",
		"type":               "Queue",
		"email":              "test@example.com",
		"created_date":       time.Now(),
		"last_modified_date": time.Now(),
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
	assert.Equal(t, "Test Group", rec["name"])

	// 4. Update
	updates := models.SObject{
		"name": "Updated Group",
	}
	err = repo.Update(ctx, nil, tableName, recordID, updates)
	assert.NoError(t, err)

	recUpdated, err := repo.FindOne(ctx, nil, tableName, recordID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Group", recUpdated["name"])

	// 5. Delete
	err = repo.Delete(ctx, nil, tableName, recordID)
	assert.NoError(t, err)

	existsAfter, err := repo.Exists(ctx, nil, tableName, recordID)
	assert.NoError(t, err)
	assert.False(t, existsAfter, "Record should be deleted")
}
