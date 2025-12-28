package services_test

import (
	"testing"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/backend/pkg/constants"
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
		db.Exec("DELETE FROM _System_Role WHERE id = ?", roleID)
	}

	// Create roles in reverse order (children first, then parents)
	_, err = db.Exec("INSERT INTO _System_Role (id, name, description, parent_role_id) VALUES (?, 'CEO', 'Chief Executive Officer', NULL)", ceoRoleID)
	if err != nil {
		t.Fatalf("Failed to create CEO role: %v", err)
	}
	_, err = db.Exec("INSERT INTO _System_Role (id, name, description, parent_role_id) VALUES (?, 'VP', 'Vice President', ?)", vpRoleID, ceoRoleID)
	if err != nil {
		t.Fatalf("Failed to create VP role: %v", err)
	}
	_, err = db.Exec("INSERT INTO _System_Role (id, name, description, parent_role_id) VALUES (?, 'Manager', 'Sales Manager', ?)", managerRoleID, vpRoleID)
	if err != nil {
		t.Fatalf("Failed to create Manager role: %v", err)
	}
	_, err = db.Exec("INSERT INTO _System_Role (id, name, description, parent_role_id) VALUES (?, 'Rep', 'Sales Representative', ?)", repRoleID, managerRoleID)
	if err != nil {
		t.Fatalf("Failed to create Rep role: %v", err)
	}

	// Refresh role hierarchy cache
	permService.RefreshRoleHierarchy()

	// Create test users
	repUserID := "test-user-rep"
	managerUserID := "test-user-manager"
	vpUserID := "test-user-vp"
	ceoUserID := "test-user-ceo"
	siblingUserID := "test-user-sibling"

	// Clean up test users
	testUsers := []string{repUserID, managerUserID, vpUserID, ceoUserID, siblingUserID}
	for _, userID := range testUsers {
		db.Exec("DELETE FROM _System_User WHERE id = ?", userID)
	}

	// Create users with roles - correct column names
	_, err = db.Exec("INSERT INTO _System_User (id, username, email, password, first_name, last_name, profile_id, role_id, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		repUserID, "rep_user", "rep@test.com", "hash", "Rep", "User", constants.ProfileStandardUser, repRoleID, 1)
	if err != nil {
		t.Fatalf("Failed to create rep user: %v", err)
	}
	_, err = db.Exec("INSERT INTO _System_User (id, username, email, password, first_name, last_name, profile_id, role_id, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		managerUserID, "manager_user", "manager@test.com", "hash", "Manager", "User", constants.ProfileStandardUser, managerRoleID, 1)
	if err != nil {
		t.Fatalf("Failed to create manager user: %v", err)
	}
	_, err = db.Exec("INSERT INTO _System_User (id, username, email, password, first_name, last_name, profile_id, role_id, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		vpUserID, "vp_user", "vp@test.com", "hash", "VP", "User", constants.ProfileStandardUser, vpRoleID, 1)
	if err != nil {
		t.Fatalf("Failed to create vp user: %v", err)
	}
	_, err = db.Exec("INSERT INTO _System_User (id, username, email, password, first_name, last_name, profile_id, role_id, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		ceoUserID, "ceo_user", "ceo@test.com", "hash", "CEO", "User", constants.ProfileStandardUser, ceoRoleID, 1)
	if err != nil {
		t.Fatalf("Failed to create ceo user: %v", err)
	}
	// Sibling user (same level as Rep, different manager chain)
	_, err = db.Exec("INSERT INTO _System_User (id, username, email, password, first_name, last_name, profile_id, role_id, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		siblingUserID, "sibling_user", "sibling@test.com", "hash", "Sibling", "User", constants.ProfileStandardUser, repRoleID, 1)
	if err != nil {
		t.Fatalf("Failed to create sibling user: %v", err)
	}

	// Test record owned by Rep
	repRecord := models.SObject{
		"id":       "test-record-1",
		"owner_id": repUserID,
	}

	// Test cases
	t.Run("Rep can access own record", func(t *testing.T) {
		repSession := &models.UserSession{
			ID:        repUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &repRoleID,
		}
		if !permService.CheckRecordAccess(nil, repRecord, "read", repSession) {
			t.Error("Rep should be able to access their own record")
		}
	})

	t.Run("Manager can read subordinate record", func(t *testing.T) {
		managerSession := &models.UserSession{
			ID:        managerUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &managerRoleID,
		}
		if !permService.CheckRecordAccess(nil, repRecord, "read", managerSession) {
			t.Error("Manager should be able to read Rep's record via hierarchy")
		}
	})

	t.Run("VP can read subordinate record (2 levels down)", func(t *testing.T) {
		vpSession := &models.UserSession{
			ID:        vpUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &vpRoleID,
		}
		if !permService.CheckRecordAccess(nil, repRecord, "read", vpSession) {
			t.Error("VP should be able to read Rep's record via hierarchy (2 levels)")
		}
	})

	t.Run("CEO can read any subordinate record", func(t *testing.T) {
		ceoSession := &models.UserSession{
			ID:        ceoUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &ceoRoleID,
		}
		if !permService.CheckRecordAccess(nil, repRecord, "read", ceoSession) {
			t.Error("CEO should be able to read Rep's record via hierarchy (top of hierarchy)")
		}
	})

	t.Run("Rep cannot read Manager record", func(t *testing.T) {
		managerRecord := models.SObject{
			"id":       "test-record-2",
			"owner_id": managerUserID,
		}
		repSession := &models.UserSession{
			ID:        repUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &repRoleID,
		}
		if permService.CheckRecordAccess(nil, managerRecord, "read", repSession) {
			t.Error("Rep should NOT be able to read Manager's record (no upward visibility)")
		}
	})

	t.Run("Sibling cannot read peer record", func(t *testing.T) {
		siblingSession := &models.UserSession{
			ID:        siblingUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &repRoleID,
		}
		if permService.CheckRecordAccess(nil, repRecord, "read", siblingSession) {
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
		if permService.CheckRecordAccess(nil, repRecord, "edit", managerSession) {
			t.Error("Manager should NOT be able to edit Rep's record via hierarchy (read-only)")
		}
	})

	// Cleanup
	for _, userID := range testUsers {
		db.Exec("DELETE FROM _System_User WHERE id = ?", userID)
	}
	for _, roleID := range cleanupRoles {
		db.Exec("DELETE FROM _System_Role WHERE id = ?", roleID)
	}

	t.Log("✅ Role Hierarchy tests completed successfully")
}
