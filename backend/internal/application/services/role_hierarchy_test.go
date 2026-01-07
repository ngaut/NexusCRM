package services_test

import (
	"fmt"
	"testing"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/stretchr/testify/require"
)

// TestRoleHierarchy_ManagerVisibility tests that managers can see records owned by subordinates
func TestRoleHierarchy_ManagerVisibility(t *testing.T) {
	// Setup Dependencies
	dbConn, err := database.GetInstance()
	if err != nil {
		t.Fatalf("Failed to get DB instance: %v", err)
	}

	db := dbConn.DB()
	schemaManager := services.NewSchemaManager(db)
	metadataService := services.NewMetadataService(dbConn, schemaManager)
	permService := services.NewPermissionService(dbConn, metadataService)

	// Create test roles
	// Hierarchy: CEO -> VP -> Manager -> Rep
	ceoRoleID := "test-role-ceo"
	vpRoleID := "test-role-vp"
	managerRoleID := "test-role-manager"
	repRoleID := "test-role-rep"

	// Clean up any existing test roles
	cleanupRoles := []string{repRoleID, managerRoleID, vpRoleID, ceoRoleID}
	for _, roleID := range cleanupRoles {
		if _, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", constants.TableRole), roleID); err != nil {
			t.Logf("Failed to cleanup role %s: %v", roleID, err)
		}
	}

	// Create roles in reverse order (children first, then parents)
	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, name, description, parent_role_id, created_date, last_modified_date) VALUES (?, 'CEO', 'Chief Executive Officer', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", constants.TableRole), ceoRoleID)
	require.NoError(t, err)

	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, name, description, parent_role_id, created_date, last_modified_date) VALUES (?, 'VP', 'Vice President', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", constants.TableRole), vpRoleID, ceoRoleID)
	require.NoError(t, err)

	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, name, description, parent_role_id, created_date, last_modified_date) VALUES (?, 'Manager', 'Sales Manager', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", constants.TableRole), managerRoleID, vpRoleID)
	require.NoError(t, err)

	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, name, description, parent_role_id, created_date, last_modified_date) VALUES (?, 'Rep', 'Sales Representative', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", constants.TableRole), repRoleID, managerRoleID)
	require.NoError(t, err)

	// Refresh role hierarchy cache
	permService.RefreshRoleHierarchy()

	// Create test users
	ceoUserID := "user_ceo"
	vpUserID := "user_vp"
	managerUserID := "user_mgr"
	repUserID := "user_rep1"
	siblingUserID := "user_rep2" // Sibling user

	// Clean up test users
	testUsers := []string{ceoUserID, vpUserID, managerUserID, repUserID, siblingUserID}
	for _, userID := range testUsers {
		if _, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", constants.TableUser), userID); err != nil {
			t.Logf("Failed to cleanup user %s: %v", userID, err)
		}
	}

	profileID := constants.ProfileStandardUser

	// Create users with roles - correct column names
	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, username, email, password, first_name, last_name, profile_id, role_id, is_active, created_date, last_modified_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", constants.TableUser),
		ceoUserID, "ceo_user", "ceo@example.com", "pass", "CEO", "User", profileID, ceoRoleID, 1)
	if err != nil {
		t.Fatalf("Failed to create ceo user: %v", err)
	}

	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, username, email, password, first_name, last_name, profile_id, role_id, is_active, created_date, last_modified_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", constants.TableUser),
		vpUserID, "vp_user", "vp@example.com", "pass", "VP", "User", profileID, vpRoleID, 1)
	if err != nil {
		t.Fatalf("Failed to create vp user: %v", err)
	}

	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, username, email, password, first_name, last_name, profile_id, role_id, is_active, created_date, last_modified_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", constants.TableUser),
		managerUserID, "manager_user", "manager@test.com", "pass", "Manager", "User", profileID, managerRoleID, 1)
	if err != nil {
		t.Fatalf("Failed to create manager user: %v", err)
	}

	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, username, email, password, first_name, last_name, profile_id, role_id, is_active, created_date, last_modified_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", constants.TableUser),
		repUserID, "rep_user", "rep@test.com", "pass", "Rep", "User", profileID, repRoleID, 1)
	if err != nil {
		t.Fatalf("Failed to create rep user: %v", err)
	}

	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, username, email, password, first_name, last_name, profile_id, role_id, is_active, created_date, last_modified_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", constants.TableUser),
		siblingUserID, "sibling_user", "sibling@test.com", "pass", "Sibling", "User", profileID, repRoleID, 1)
	if err != nil {
		t.Fatalf("Failed to create sibling user: %v", err)
	}

	// Test record owned by Rep
	repRecord := models.SObject{
		constants.FieldID:      "test-record-1",
		constants.FieldOwnerID: repUserID,
	}

	// Test cases
	t.Run("Rep can access own record", func(t *testing.T) {
		repSession := &models.UserSession{
			ID:        repUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &repRoleID,
		}
		if !permService.CheckRecordAccess(nil, repRecord, constants.PermRead, repSession) {
			t.Error("Rep should be able to access their own record")
		}
	})

	t.Run("Manager can read subordinate record", func(t *testing.T) {
		managerSession := &models.UserSession{
			ID:        managerUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &managerRoleID,
		}
		if !permService.CheckRecordAccess(nil, repRecord, constants.PermRead, managerSession) {
			t.Error("Manager should be able to read Rep's record via hierarchy")
		}
	})

	t.Run("VP can read subordinate record (2 levels down)", func(t *testing.T) {
		vpSession := &models.UserSession{
			ID:        vpUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &vpRoleID,
		}
		if !permService.CheckRecordAccess(nil, repRecord, constants.PermRead, vpSession) {
			t.Error("VP should be able to read Rep's record via hierarchy (2 levels)")
		}
	})

	t.Run("CEO can read any subordinate record", func(t *testing.T) {
		ceoSession := &models.UserSession{
			ID:        ceoUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &ceoRoleID,
		}
		if !permService.CheckRecordAccess(nil, repRecord, constants.PermRead, ceoSession) {
			t.Error("CEO should be able to read Rep's record via hierarchy (top of hierarchy)")
		}
	})

	t.Run("Rep cannot read Manager record", func(t *testing.T) {
		managerRecord := models.SObject{
			constants.FieldID:      "test-record-2",
			constants.FieldOwnerID: managerUserID,
		}
		repSession := &models.UserSession{
			ID:        repUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &repRoleID,
		}
		if permService.CheckRecordAccess(nil, managerRecord, constants.PermRead, repSession) {
			t.Error("Rep should NOT be able to read Manager's record (no upward visibility)")
		}
	})

	t.Run("Sibling cannot read peer record", func(t *testing.T) {
		siblingSession := &models.UserSession{
			ID:        siblingUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &repRoleID,
		}
		if permService.CheckRecordAccess(nil, repRecord, constants.PermRead, siblingSession) {
			t.Error("Sibling should NOT be able to read peer's record (same level)")
		}
	})

	t.Run("Manager cannot edit subordinate record via hierarchy", func(t *testing.T) {
		managerSession := &models.UserSession{
			ID:        managerUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &managerRoleID,
		}
		// Hierarchy grants READ only, not EDIT
		if permService.CheckRecordAccess(nil, repRecord, constants.PermEdit, managerSession) {
			t.Error("Manager should NOT be able to edit Rep's record via hierarchy (read-only)")
		}
	})

	// Cleanup
	for _, userID := range testUsers {
		if _, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", constants.TableUser), userID); err != nil {
			t.Logf("Failed to cleanup user %s: %v", userID, err)
		}
	}
	for _, roleID := range cleanupRoles {
		if _, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", constants.TableRole), roleID); err != nil {
			t.Logf("Failed to cleanup role %s: %v", roleID, err)
		}
	}

	t.Log("✅ Role Hierarchy tests completed successfully")
}
