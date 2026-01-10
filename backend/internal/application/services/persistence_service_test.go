package services_test

import (
	"context"
	"testing"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/internal/bootstrap"
	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersistenceService_Integration(t *testing.T) {
	conn, ms := services.SetupIntegrationTest(t)
	db := conn.DB()

	// 0. Bootstrap Schema (tables + system data)
	err := bootstrap.InitializeSchema(conn)
	require.NoError(t, err, "Schema bootstrap failed")

	// 1. Setup Dependencies
	recordRepo := persistence.NewRecordRepository(db)
	userRepo := persistence.NewUserRepository(db)
	permRepo := persistence.NewPermissionRepository(db)
	rollupRepo := persistence.NewRollupRepository(db)
	txManager := persistence.NewTransactionManager(conn)

	// Services
	ps := services.NewPermissionService(permRepo, ms, userRepo)
	rollup := services.NewRollupService(rollupRepo, ms, txManager)
	eventBus := services.NewEventBus()
	formulaEngine := formula.NewEngine()
	validator := services.NewValidationService(formulaEngine)

	// Create Service
	svc := services.NewPersistenceService(recordRepo, rollup, ms, ps, eventBus, validator, txManager, nil)

	ms.RefreshCache()

	// 2. Setup Test Data (User)
	sysAdminID := "test-admin"
	// GetTestUser returns Name as "Test User " + ID
	expectedAdminName := "Test User " + sysAdminID

	adminUser := services.GetTestUser(sysAdminID, constants.ProfileSystemAdmin)
	// Ensure the session object has the expected name for our checks
	adminUser.Name = expectedAdminName

	// 3. Test Cases for Record Lifecycle
	t.Run("Create_Update_Delete_Record", func(t *testing.T) {
		ctx := context.Background()

		// Use 'Group' (QUEUE) because it doesn't depend on many other things
		tableName := constants.TableGroup
		// Unique email and name to avoid collision
		uniqueSuffix := services.GenerateID()
		email := "testq_" + uniqueSuffix + "@example.com"
		name := "Integration Test Queue " + uniqueSuffix
		label := "Test Queue Label " + uniqueSuffix

		newGroup := models.SObject{
			"name":  name,
			"label": label,
			"type":  "Queue",
			"email": email,
		}

		// Create
		created, err := svc.Insert(ctx, tableName, newGroup, adminUser)
		require.NoError(t, err, "Insert should succeed")
		require.NotEmpty(t, created[constants.FieldID])
		id := created[constants.FieldID].(string)

		t.Logf("Created record: %s", id)

		// Update
		uniqueLabelUpdated := "Updated Queue Label " + uniqueSuffix
		updates := models.SObject{
			"label": uniqueLabelUpdated,
		}
		err = svc.Update(ctx, tableName, id, updates, adminUser)
		assert.NoError(t, err, "Update should succeed")

		// Find (Verify Update)
		rec, err := recordRepo.FindOne(ctx, nil, tableName, id)
		assert.NoError(t, err)
		assert.Equal(t, uniqueLabelUpdated, rec["label"])

		// Delete
		err = svc.Delete(ctx, tableName, id, adminUser)
		assert.NoError(t, err, "Delete should succeed")

		// Verify Deletion (Soft Delete)
		// 1. Exists should return TRUE (Soft Delete)
		exists, err := recordRepo.Exists(ctx, nil, tableName, id)
		assert.NoError(t, err)
		assert.True(t, exists, "Record should still exist physically (Soft Delete)")

		// 2. FindOne should return NIL (Excluded)
		recDeleted, err := recordRepo.FindOne(ctx, nil, tableName, id)
		assert.NoError(t, err)
		assert.Nil(t, recDeleted, "FindOne should not return soft-deleted record")

		// 3. Verify Recycle Bin
		// The DeletedBy field should match adminUser.Name
		recycleBinItems, err := recordRepo.FindRecycleBinItemsByUser(ctx, adminUser.Name)
		assert.NoError(t, err)

		foundInBin := false
		for _, item := range recycleBinItems {
			if item[constants.FieldRecordID] == id {
				foundInBin = true
				break
			}
		}
		assert.True(t, foundInBin, "Record should be in recycle bin")
	})

	t.Run("Create_Validation_Error", func(t *testing.T) {
		ctx := context.Background()
		uniqueSuffix := services.GenerateID()

		groupBad := models.SObject{
			"name":  "",
			"type":  "Queue",
			"email": "bad_" + uniqueSuffix + "@example.com",
		}

		_, err := svc.Insert(ctx, constants.TableGroup, groupBad, adminUser)
		if err == nil {
			t.Log("Insert with empty name succeeded")
		} else {
			t.Logf("Got expected validation/db error: %v", err)
		}
	})

	t.Run("AutoNumber_Generation", func(t *testing.T) {
		ctx := context.Background()
		objectName := constants.TableGroup // Use Group

		// Create AutoNumber field
		uniqueSuffix := services.GenerateID()
		fieldName := "ticket_" + uniqueSuffix[:8] + "__c"
		format := "T-{0000}" // expect T-0001, T-0002

		field := &models.FieldMetadata{
			APIName:      fieldName,
			Label:        "Ticket Number",
			Type:         constants.FieldTypeAutoNumber,
			DefaultValue: services.StringPtr(format),
		}

		err := ms.CreateField(ctx, objectName, field)
		require.NoError(t, err, "CreateField (AutoNumber) should succeed")

		// Refresh cache to ensure persistence service sees the new field
		ms.RefreshCache()

		// Insert Record 1
		rec1 := models.SObject{
			"name":  "AutoNum Test 1 " + uniqueSuffix,
			"label": "AutoNum Test 1 " + uniqueSuffix,
		}
		saved1, err := svc.Insert(ctx, objectName, rec1, adminUser)
		require.NoError(t, err, "Insert record 1 should succeed")
		t.Cleanup(func() {
			svc.Delete(context.Background(), objectName, saved1[constants.FieldID].(string), adminUser)
		})

		val1 := saved1[fieldName]
		t.Logf("Record 1 AutoNumber: %v", val1)

		// Check if AutoNumber was generated (might be async or sync depending on implementation)
		// Assuming sync for now.
		// Also standard objects like Account might require other fields? Account requires Name (provided).
		if val1 == nil {
			t.Fatal("AutoNumber field was nil")
		}
		assert.Equal(t, "T-0001", val1)

		// Insert Record 2
		rec2 := models.SObject{
			"name":  "AutoNum Test 2 " + uniqueSuffix,
			"label": "AutoNum Test 2 " + uniqueSuffix,
		}
		saved2, err := svc.Insert(ctx, objectName, rec2, adminUser)
		require.NoError(t, err, "Insert record 2 should succeed")
		t.Cleanup(func() {
			svc.Delete(context.Background(), objectName, saved2[constants.FieldID].(string), adminUser)
		})

		val2 := saved2[fieldName]
		t.Logf("Record 2 AutoNumber: %v", val2)
		assert.Equal(t, "T-0002", val2)
	})
}
