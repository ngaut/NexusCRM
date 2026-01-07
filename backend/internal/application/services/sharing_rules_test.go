package services_test

import (
	"fmt"
	"testing"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// TestSharingRules tests that sharing rules grant access to matching records
func TestSharingRules_Integration(t *testing.T) {
	// Setup Dependencies
	dbConn, err := database.GetInstance()
	if err != nil {
		t.Fatalf("Failed to get DB instance: %v", err)
	}

	db := dbConn.DB()
	schemaManager := services.NewSchemaManager(db)
	metadataService := services.NewMetadataService(dbConn, schemaManager)
	permService := services.NewPermissionService(dbConn, metadataService)

	// Create test role for sharing
	salesRoleID := "test-sharing-sales-role"
	marketingRoleID := "test-sharing-marketing-role"

	// Clean up
	if _, err := db.Exec("DELETE FROM _System_SharingRule WHERE id LIKE 'test-sharing-%'"); err != nil {
		t.Logf("Failed to cleanup sharing rules: %v", err)
	}
	if _, err := db.Exec("DELETE FROM _System_User WHERE id LIKE 'test-sharing-%'"); err != nil {
		t.Logf("Failed to cleanup users: %v", err)
	}
	if _, err := db.Exec("DELETE FROM _System_Role WHERE id LIKE 'test-sharing-%'"); err != nil {
		t.Logf("Failed to cleanup roles: %v", err)
	}

	// Create roles
	_, err = db.Exec("INSERT INTO _System_Role (id, name, description, parent_role_id, created_date, last_modified_date) VALUES (?, ?, ?, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		salesRoleID, "Test Sales Role", "Sales role for testing")
	if err != nil {
		t.Fatalf("Failed to create sales role: %v", err)
	}
	_, err = db.Exec("INSERT INTO _System_Role (id, name, description, parent_role_id, created_date, last_modified_date) VALUES (?, ?, ?, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		marketingRoleID, "Test Marketing Role", "Marketing role for testing")
	if err != nil {
		t.Fatalf("Failed to create marketing role: %v", err)
	}

	// Refresh role hierarchy
	permService.RefreshRoleHierarchy()

	// Create test users
	salesUserID := "test-sharing-sales-user"
	marketingUserID := "test-sharing-marketing-user"
	ownerUserID := "test-sharing-owner"

	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, username, email, password, first_name, last_name, profile_id, role_id, is_active, created_date, last_modified_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", constants.TableUser),
		salesUserID, "sales_test", "sales_test@test.com", "hash", "Sales", "User", constants.ProfileStandardUser, salesRoleID, 1)
	if err != nil {
		t.Fatalf("Failed to create sales user: %v", err)
	}
	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, username, email, password, first_name, last_name, profile_id, role_id, is_active, created_date, last_modified_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", constants.TableUser),
		marketingUserID, "marketing_test", "marketing_test@test.com", "hash", "Marketing", "User", constants.ProfileStandardUser, marketingRoleID, 1)
	if err != nil {
		t.Fatalf("Failed to create marketing user: %v", err)
	}
	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, username, email, password, first_name, last_name, profile_id, role_id, is_active, created_date, last_modified_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", constants.TableUser),
		ownerUserID, "owner_test", "owner_test@test.com", "hash", "Owner", "User", constants.ProfileStandardUser, nil, 1)
	if err != nil {
		t.Fatalf("Failed to create owner user: %v", err)
	}

	// Create a sharing rule: Share Account records where industry=Technology with Sales role
	sharingRuleID := "test-sharing-rule-1"
	criteria := `industry == "Technology"`
	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, object_api_name, name, criteria, access_level, share_with_role_id, created_date, last_modified_date) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", constants.TableSharingRule),
		sharingRuleID, "Account", "Share Tech Accounts with Sales", criteria, "Read", salesRoleID)
	if err != nil {
		t.Fatalf("Failed to create sharing rule: %v", err)
	}

	// Mock schema
	schema := &models.ObjectMetadata{
		APIName: "Account",
	}

	// Test records
	techRecord := models.SObject{
		constants.FieldID:      "test-record-tech",
		constants.FieldOwnerID: ownerUserID,
		"industry":             "Technology",
	}

	nonTechRecord := models.SObject{
		constants.FieldID:      "test-record-finance",
		constants.FieldOwnerID: ownerUserID,
		"industry":             "Finance",
	}

	// Test cases
	t.Run("Sales user can read Tech record via sharing rule", func(t *testing.T) {
		salesSession := &models.UserSession{
			ID:        salesUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &salesRoleID,
		}
		if !permService.CheckRecordAccess(schema, techRecord, constants.PermRead, salesSession) {
			t.Error("Sales user should be able to read Tech record via sharing rule")
		}
	})

	t.Run("Sales user cannot read Finance record (criteria mismatch)", func(t *testing.T) {
		salesSession := &models.UserSession{
			ID:        salesUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &salesRoleID,
		}
		if permService.CheckRecordAccess(schema, nonTechRecord, constants.PermRead, salesSession) {
			t.Error("Sales user should NOT be able to read Finance record (criteria doesn't match)")
		}
	})

	t.Run("Marketing user cannot read Tech record (not in shared role)", func(t *testing.T) {
		marketingSession := &models.UserSession{
			ID:        marketingUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &marketingRoleID,
		}
		if permService.CheckRecordAccess(schema, techRecord, constants.PermRead, marketingSession) {
			t.Error("Marketing user should NOT be able to read Tech record (not in Sales role)")
		}
	})

	t.Run("Sales user cannot edit Tech record (rule is Read only)", func(t *testing.T) {
		salesSession := &models.UserSession{
			ID:        salesUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &salesRoleID,
		}
		if permService.CheckRecordAccess(schema, techRecord, constants.PermEdit, salesSession) {
			t.Error("Sales user should NOT be able to edit Tech record (Read access only)")
		}
	})

	// Test Edit access level
	editRuleID := "test-sharing-rule-edit"
	_, err = db.Exec("INSERT INTO _System_SharingRule (id, object_api_name, name, criteria, access_level, share_with_role_id, created_date, last_modified_date) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		editRuleID, "Account", "Edit Shared to Marketing", "[]", "Edit", marketingRoleID)
	if err != nil {
		t.Fatalf("Failed to create edit sharing rule: %v", err)
	}

	t.Run("Marketing user can edit any Account record (Edit access, no criteria)", func(t *testing.T) {
		marketingSession := &models.UserSession{
			ID:        marketingUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &marketingRoleID,
		}
		// Both tech and non-tech records should be editable
		if !permService.CheckRecordAccess(schema, techRecord, constants.PermEdit, marketingSession) {
			t.Error("Marketing user should be able to edit Tech record (Edit access)")
		}
		if !permService.CheckRecordAccess(schema, nonTechRecord, constants.PermRead, marketingSession) {
			t.Error("Marketing user should be able to read Finance record (Edit grants read too)")
		}
	})

	// Cleanup
	// Cleanup
	if _, err := db.Exec("DELETE FROM _System_SharingRule WHERE id LIKE 'test-sharing-%'"); err != nil {
		t.Logf("Cleanup failed: %v", err)
	}
	if _, err := db.Exec("DELETE FROM _System_User WHERE id LIKE 'test-sharing-%'"); err != nil {
		t.Logf("Cleanup failed: %v", err)
	}
	if _, err := db.Exec("DELETE FROM _System_Role WHERE id LIKE 'test-sharing-%'"); err != nil {
		t.Logf("Cleanup failed: %v", err)
	}

	t.Log("✅ Sharing Rules tests completed successfully")
}
