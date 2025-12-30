package services_test

import (
	"fmt"
	"testing"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/backend/pkg/constants"
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
	db.Exec("DELETE FROM _System_SharingRule WHERE id LIKE 'test-sharing-%'")
	db.Exec("DELETE FROM _System_User WHERE id LIKE 'test-sharing-%'")
	db.Exec("DELETE FROM _System_Role WHERE id LIKE 'test-sharing-%'")

	// Create roles
	_, err = db.Exec("INSERT INTO _System_Role (id, name, description, parent_role_id) VALUES (?, ?, ?, NULL)",
		salesRoleID, "Test Sales Role", "Sales role for testing")
	if err != nil {
		t.Fatalf("Failed to create sales role: %v", err)
	}
	_, err = db.Exec("INSERT INTO _System_Role (id, name, description, parent_role_id) VALUES (?, ?, ?, NULL)",
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

	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, username, email, password, first_name, last_name, profile_id, role_id, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", constants.TableUser),
		salesUserID, "sales_test", "sales_test@test.com", "hash", "Sales", "User", constants.ProfileStandardUser, salesRoleID, 1)
	if err != nil {
		t.Fatalf("Failed to create sales user: %v", err)
	}
	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, username, email, password, first_name, last_name, profile_id, role_id, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", constants.TableUser),
		marketingUserID, "marketing_test", "marketing_test@test.com", "hash", "Marketing", "User", constants.ProfileStandardUser, marketingRoleID, 1)
	if err != nil {
		t.Fatalf("Failed to create marketing user: %v", err)
	}
	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, username, email, password, first_name, last_name, profile_id, role_id, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", constants.TableUser),
		ownerUserID, "owner_test", "owner_test@test.com", "hash", "Owner", "User", constants.ProfileStandardUser, nil, 1)
	if err != nil {
		t.Fatalf("Failed to create owner user: %v", err)
	}

	// Create a sharing rule: Share Account records where industry=Technology with Sales role
	sharingRuleID := "test-sharing-rule-1"
	criteria := `industry == "Technology"`
	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (id, object_api_name, name, criteria, access_level, share_with_role_id) VALUES (?, ?, ?, ?, ?, ?)", constants.TableSharingRule),
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
		"id":       "test-record-tech",
		"owner_id": ownerUserID,
		"industry": "Technology",
	}

	nonTechRecord := models.SObject{
		"id":       "test-record-finance",
		"owner_id": ownerUserID,
		"industry": "Finance",
	}

	// Test cases
	t.Run("Sales user can read Tech record via sharing rule", func(t *testing.T) {
		salesSession := &models.UserSession{
			ID:        salesUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &salesRoleID,
		}
		if !permService.CheckRecordAccess(schema, techRecord, "read", salesSession) {
			t.Error("Sales user should be able to read Tech record via sharing rule")
		}
	})

	t.Run("Sales user cannot read Finance record (criteria mismatch)", func(t *testing.T) {
		salesSession := &models.UserSession{
			ID:        salesUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &salesRoleID,
		}
		if permService.CheckRecordAccess(schema, nonTechRecord, "read", salesSession) {
			t.Error("Sales user should NOT be able to read Finance record (criteria doesn't match)")
		}
	})

	t.Run("Marketing user cannot read Tech record (not in shared role)", func(t *testing.T) {
		marketingSession := &models.UserSession{
			ID:        marketingUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &marketingRoleID,
		}
		if permService.CheckRecordAccess(schema, techRecord, "read", marketingSession) {
			t.Error("Marketing user should NOT be able to read Tech record (not in Sales role)")
		}
	})

	t.Run("Sales user cannot edit Tech record (rule is Read only)", func(t *testing.T) {
		salesSession := &models.UserSession{
			ID:        salesUserID,
			ProfileID: constants.ProfileStandardUser,
			RoleID:    &salesRoleID,
		}
		if permService.CheckRecordAccess(schema, techRecord, "edit", salesSession) {
			t.Error("Sales user should NOT be able to edit Tech record (Read access only)")
		}
	})

	// Test Edit access level
	editRuleID := "test-sharing-rule-edit"
	_, err = db.Exec("INSERT INTO _System_SharingRule (id, object_api_name, name, criteria, access_level, share_with_role_id) VALUES (?, ?, ?, ?, ?, ?)",
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
		if !permService.CheckRecordAccess(schema, techRecord, "edit", marketingSession) {
			t.Error("Marketing user should be able to edit Tech record (Edit access)")
		}
		if !permService.CheckRecordAccess(schema, nonTechRecord, "read", marketingSession) {
			t.Error("Marketing user should be able to read Finance record (Edit grants read too)")
		}
	})

	// Cleanup
	db.Exec("DELETE FROM _System_SharingRule WHERE id LIKE 'test-sharing-%'")
	db.Exec("DELETE FROM _System_User WHERE id LIKE 'test-sharing-%'")
	db.Exec("DELETE FROM _System_Role WHERE id LIKE 'test-sharing-%'")

	t.Log("✅ Sharing Rules tests completed successfully")
}
