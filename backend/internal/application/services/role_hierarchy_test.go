package services_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/stretchr/testify/require"
)

// TestRoleHierarchy_ManagerVisibility tests that managers can see records owned by subordinates
func TestRoleHierarchy_ManagerVisibility(t *testing.T) {
	// Setup Dependencies
	dbConn, err := database.GetInstance()
	require.NoError(t, err)

	db := dbConn.DB()
	schemaRepo := persistence.NewSchemaRepository(db)
	schemaManager := services.NewSchemaManager(schemaRepo)
	metadataRepo := persistence.NewMetadataRepository(dbConn.DB())
	metadataService := services.NewMetadataService(metadataRepo, schemaManager)
	userRepo := persistence.NewUserRepository(db)
	permRepo := persistence.NewPermissionRepository(dbConn.DB())
	permService := services.NewPermissionService(permRepo, metadataService, userRepo)

	ctx := context.Background()

	// Helper to create role
	createRole := func(name, label, parentID string) string {
		id := services.GenerateID()
		// Note: 'label' is not in the schema based on previous error, using 'description' for label text
		var query string
		var err error

		roleCols := fmt.Sprintf("%s, %s, %s, %s, %s, %s", constants.FieldID, constants.FieldSysRole_Name, constants.FieldSysRole_Description, constants.FieldSysRole_ParentRoleID, constants.FieldCreatedDate, constants.FieldLastModifiedDate)
		if parentID == "" {
			query = fmt.Sprintf("INSERT INTO %s (%s) VALUES (?, ?, ?, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", constants.TableRole, roleCols)
			_, err = db.Exec(query, id, name, label)
		} else {
			query = fmt.Sprintf("INSERT INTO %s (%s) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", constants.TableRole, roleCols)
			_, err = db.Exec(query, id, name, label, parentID)
		}
		require.NoError(t, err)

		// Register cleanup
		t.Cleanup(func() {
			_, _ = db.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s = ?", constants.TableRole, constants.FieldID), id)
		})
		return id
	}

	// Create test roles hierarchy: CEO -> VP -> Manager -> Rep
	ceoRoleID := createRole("CEO_"+services.GenerateID(), "CEO", "")
	vpRoleID := createRole("VP_"+services.GenerateID(), "VP", ceoRoleID)
	managerRoleID := createRole("Manager_"+services.GenerateID(), "Manager", vpRoleID)
	repRoleID := createRole("Rep_"+services.GenerateID(), "Rep", managerRoleID)

	// Refresh role hierarchy cache
	permService.RefreshRoleHierarchy()

	// Helper to create user
	createUser := func(roleID string) (*models.UserSession, string) {
		userID := services.GenerateID()
		username := "user_" + userID + "@test.com"

		userCols := fmt.Sprintf("%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s", constants.FieldID, constants.FieldUsername, constants.FieldEmail, constants.FieldPassword, constants.FieldFirstName, constants.FieldLastName, constants.FieldProfileID, constants.FieldRoleID, constants.FieldIsActive, constants.FieldCreatedDate, constants.FieldLastModifiedDate)
		_, err := db.Exec(fmt.Sprintf("INSERT INTO %s (%s) VALUES (?, ?, ?, 'pass', 'Test', 'User', ?, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", constants.TableUser, userCols),
			userID, username, username, constants.ProfileStandardUser, roleID)
		require.NoError(t, err)

		t.Cleanup(func() {
			_, _ = db.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s = ?", constants.TableUser, constants.FieldID), userID)
		})

		return &models.UserSession{
			ID:        userID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &roleID,
		}, userID
	}

	// Create test users
	ceoSession, _ := createUser(ceoRoleID)
	vpSession, _ := createUser(vpRoleID)
	managerSession, managerUserID := createUser(managerRoleID)
	repSession, repUserID := createUser(repRoleID)
	siblingSession, _ := createUser(repRoleID)

	// Test record owned by Rep
	repRecord := models.SObject{
		constants.FieldID:      "test-record-" + services.GenerateID(),
		constants.FieldOwnerID: repUserID,
	}

	// Test cases
	t.Run("Rep can access own record", func(t *testing.T) {
		if !permService.CheckRecordAccess(ctx, nil, repRecord, constants.PermRead, repSession) {
			t.Error("Rep should be able to access their own record")
		}
	})

	t.Run("Manager can read subordinate record", func(t *testing.T) {
		if !permService.CheckRecordAccess(ctx, nil, repRecord, constants.PermRead, managerSession) {
			t.Error("Manager should be able to read Rep's record via hierarchy")
		}
	})

	t.Run("VP can read subordinate record (2 levels down)", func(t *testing.T) {
		if !permService.CheckRecordAccess(ctx, nil, repRecord, constants.PermRead, vpSession) {
			t.Error("VP should be able to read Rep's record via hierarchy (2 levels)")
		}
	})

	t.Run("CEO can read any subordinate record", func(t *testing.T) {
		if !permService.CheckRecordAccess(ctx, nil, repRecord, constants.PermRead, ceoSession) {
			t.Error("CEO should be able to read Rep's record via hierarchy (top of hierarchy)")
		}
	})

	t.Run("Rep cannot read Manager record", func(t *testing.T) {
		managerRecord := models.SObject{
			constants.FieldID:      "test-record-" + services.GenerateID(),
			constants.FieldOwnerID: managerUserID,
		}
		if permService.CheckRecordAccess(ctx, nil, managerRecord, constants.PermRead, repSession) {
			t.Error("Rep should NOT be able to read Manager's record (no upward visibility)")
		}
	})

	t.Run("Sibling cannot read peer record", func(t *testing.T) {
		if permService.CheckRecordAccess(ctx, nil, repRecord, constants.PermRead, siblingSession) {
			t.Error("Sibling should NOT be able to read peer's record (same level)")
		}
	})

	t.Run("Manager cannot edit subordinate record via hierarchy", func(t *testing.T) {
		// Hierarchy grants READ only, not EDIT
		if permService.CheckRecordAccess(ctx, nil, repRecord, constants.PermEdit, managerSession) {
			t.Error("Manager should NOT be able to edit Rep's record via hierarchy (read-only)")
		}
	})

	t.Log("✅ Role Hierarchy tests completed successfully")
}
