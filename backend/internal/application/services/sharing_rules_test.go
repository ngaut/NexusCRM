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
)

// TestSharingRules tests that sharing rules grant access to matching records
func TestSharingRules_Integration(t *testing.T) {
	// Setup Dependencies
	dbConn, err := database.GetInstance()
	if err != nil {
		t.Fatalf("Failed to get DB instance: %v", err)
	}

	db := dbConn.DB()
	schemaRepo := persistence.NewSchemaRepository(db)
	schemaManager := services.NewSchemaManager(schemaRepo)
	metadataRepo := persistence.NewMetadataRepository(dbConn.DB())
	metadataService := services.NewMetadataService(metadataRepo, schemaManager)
	userRepo := persistence.NewUserRepository(db)
	permRepo := persistence.NewPermissionRepository(dbConn.DB())
	permService := services.NewPermissionService(permRepo, metadataService, userRepo)

	// Create test role for sharing
	salesRoleID := "test-sharing-sales-role"
	marketingRoleID := "test-sharing-marketing-role"

	// Clean up
	if _, err := db.Exec(fmt.Sprintf("%s %s %s %s %s 'test-sharing-%%'", persistence.KeywordDeleteFrom, constants.TableSharingRule, persistence.KeywordWhere, constants.FieldID, persistence.KeywordLike)); err != nil {
		t.Logf("Failed to cleanup sharing rules: %v", err)
	}
	if _, err := db.Exec(fmt.Sprintf("%s %s %s %s %s 'test-sharing-%%'", persistence.KeywordDeleteFrom, constants.TableUser, persistence.KeywordWhere, constants.FieldID, persistence.KeywordLike)); err != nil {
		t.Logf("Failed to cleanup users: %v", err)
	}
	if _, err := db.Exec(fmt.Sprintf("%s %s %s %s %s 'test-sharing-%%'", persistence.KeywordDeleteFrom, constants.TableRole, persistence.KeywordWhere, constants.FieldID, persistence.KeywordLike)); err != nil {
		t.Logf("Failed to cleanup roles: %v", err)
	}

	// Create roles
	// Create roles
	roleCols := fmt.Sprintf("%s, %s, %s, %s, %s, %s", constants.FieldID, constants.FieldSysRole_Name, constants.FieldSysRole_Description, constants.FieldSysRole_ParentRoleID, constants.FieldCreatedDate, constants.FieldLastModifiedDate)
	_, err = db.Exec(fmt.Sprintf("%s %s (%s) %s (?, ?, ?, %s, %s, %s)", persistence.KeywordInsertInto, constants.TableRole, roleCols, persistence.KeywordValues, persistence.KeywordNull, persistence.FuncCurrentTimestamp, persistence.FuncCurrentTimestamp),
		salesRoleID, "Test Sales Role", "Sales role for testing")
	if err != nil {
		t.Fatalf("Failed to create sales role: %v", err)
	}
	_, err = db.Exec(fmt.Sprintf("%s %s (%s) %s (?, ?, ?, %s, %s, %s)", persistence.KeywordInsertInto, constants.TableRole, roleCols, persistence.KeywordValues, persistence.KeywordNull, persistence.FuncCurrentTimestamp, persistence.FuncCurrentTimestamp),
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

	userCols := fmt.Sprintf("%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s", constants.FieldID, constants.FieldUsername, constants.FieldEmail, constants.FieldPassword, constants.FieldFirstName, constants.FieldLastName, constants.FieldProfileID, constants.FieldRoleID, constants.FieldIsActive, constants.FieldCreatedDate, constants.FieldLastModifiedDate)
	_, err = db.Exec(fmt.Sprintf("%s %s (%s) %s (?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)", persistence.KeywordInsertInto, constants.TableUser, userCols, persistence.KeywordValues, persistence.FuncCurrentTimestamp, persistence.FuncCurrentTimestamp),
		salesUserID, "sales_test", "sales_test@test.com", "hash", "Sales", "User", constants.ProfileStandardUser, salesRoleID, 1)
	if err != nil {
		t.Fatalf("Failed to create sales user: %v", err)
	}
	_, err = db.Exec(fmt.Sprintf("%s %s (%s) %s (?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)", persistence.KeywordInsertInto, constants.TableUser, userCols, persistence.KeywordValues, persistence.FuncCurrentTimestamp, persistence.FuncCurrentTimestamp),
		marketingUserID, "marketing_test", "marketing_test@test.com", "hash", "Marketing", "User", constants.ProfileStandardUser, marketingRoleID, 1)
	if err != nil {
		t.Fatalf("Failed to create marketing user: %v", err)
	}
	_, err = db.Exec(fmt.Sprintf("%s %s (%s) %s (?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)", persistence.KeywordInsertInto, constants.TableUser, userCols, persistence.KeywordValues, persistence.FuncCurrentTimestamp, persistence.FuncCurrentTimestamp),
		ownerUserID, "owner_test", "owner_test@test.com", "hash", "Owner", "User", constants.ProfileStandardUser, nil, 1)
	if err != nil {
		t.Fatalf("Failed to create owner user: %v", err)
	}

	// Create a sharing rule: Share Account records where industry=Technology with Sales role
	sharingRuleID := "test-sharing-rule-1"
	criteria := `industry == "Technology"`
	sharingCols := fmt.Sprintf("%s, %s, %s, %s, %s, %s, %s, %s", constants.FieldID, constants.FieldSysSharingRule_ObjectAPIName, constants.FieldSysSharingRule_Name, constants.FieldSysSharingRule_Criteria, constants.FieldSysSharingRule_AccessLevel, constants.FieldSysSharingRule_ShareWithRoleID, constants.FieldCreatedDate, constants.FieldLastModifiedDate)
	_, err = db.Exec(fmt.Sprintf("%s %s (%s) %s (?, ?, ?, ?, ?, ?, %s, %s)", persistence.KeywordInsertInto, constants.TableSharingRule, sharingCols, persistence.KeywordValues, persistence.FuncCurrentTimestamp, persistence.FuncCurrentTimestamp),
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

	salesSession := &models.UserSession{
		ID:        salesUserID,
		ProfileID: constants.ProfileStandardUser,
		RoleID:    &salesRoleID,
	}
	marketingSession := &models.UserSession{
		ID:        marketingUserID,
		ProfileID: constants.ProfileStandardUser,
		RoleID:    &marketingRoleID,
	}
	adminSession := &models.UserSession{ // Assuming an admin session for owner access
		ID:        ownerUserID,
		ProfileID: constants.ProfileSystemAdmin, // Or a profile that grants full access
		RoleID:    nil,                          // Admins might not have a specific role for hierarchy
	}

	// Test Case 1: Tech support record (matches criteria)
	// Sales User -> Should have READ access (via sharing rule)
	t.Run("Sales User Access to Tech Record", func(t *testing.T) {
		if !permService.CheckRecordAccess(context.Background(), schema, techRecord, constants.PermRead, salesSession) {
			t.Errorf("Expected Sales user to have READ access to Tech record via sharing rule")
		}
	})

	// Test Case 2: Non-Tech support record (does NOT match criteria)
	// Sales User -> Should NOT have access (Role Hierarchy doesn't apply as owner is admin, sharing rule criteria fails)
	t.Run("Sales User Access to Non-Tech Record", func(t *testing.T) {
		if permService.CheckRecordAccess(context.Background(), schema, nonTechRecord, constants.PermRead, salesSession) {
			t.Errorf("Expected Sales user to NOT have access to Non-Tech record")
		}
	})

	// Test Case 3: Tech support record
	// Marketing User -> Should NOT have READ access (No matching sharing rule)
	t.Run("Marketing User Access to Tech Record", func(t *testing.T) {
		if permService.CheckRecordAccess(context.Background(), schema, techRecord, constants.PermRead, marketingSession) {
			t.Errorf("Expected Marketing user to NOT have READ access to Tech record (No sharing rule applies)")
		}
	})

	// Test Case 4: Edit Access
	// Sales User -> Should NOT have EDIT access (Sharing rule is Read Only)
	t.Run("Sales User Edit Access", func(t *testing.T) {
		if permService.CheckRecordAccess(context.Background(), schema, techRecord, constants.PermEdit, salesSession) {
			t.Errorf("Expected Sales user to NOT have EDIT access (Sharing rule is Read Only)")
		}
	})

	// Test Case 5: Owner Access (Admin)
	// Admin -> Should have EDIT access
	t.Run("Owner Access", func(t *testing.T) {
		if !permService.CheckRecordAccess(context.Background(), schema, techRecord, constants.PermEdit, adminSession) {
			t.Errorf("Expected Owner (Admin) to have EDIT access")
		}
	})

	// Test Case 6: Marketing User Access (Verify no access to confirm isolation)
	t.Run("Marketing User No Access", func(t *testing.T) {
		if permService.CheckRecordAccess(context.Background(), schema, techRecord, constants.PermEdit, marketingSession) {
			t.Errorf("Expected Marketing user to NOT have EDIT access")
		}
		if permService.CheckRecordAccess(context.Background(), schema, nonTechRecord, constants.PermRead, marketingSession) {
			t.Errorf("Expected Marketing user to NOT have READ access to Finance record")
		}
	})

	// Test Edit access level
	editRuleID := "test-sharing-rule-edit"
	_, err = db.Exec(fmt.Sprintf("%s %s (%s) %s (?, ?, ?, ?, ?, ?, %s, %s)", persistence.KeywordInsertInto, constants.TableSharingRule, sharingCols, persistence.KeywordValues, persistence.FuncCurrentTimestamp, persistence.FuncCurrentTimestamp),
		editRuleID, "Account", "Edit Shared to Marketing", "[]", "Edit", marketingRoleID)
	if err != nil {
		t.Fatalf("Failed to create edit sharing rule: %v", err)
	}

	t.Run("Marketing user can edit any Account record (Edit access, no criteria)", func(t *testing.T) {
		// Both tech and non-tech records should be editable
		if !permService.CheckRecordAccess(context.Background(), schema, techRecord, constants.PermEdit, marketingSession) {
			t.Error("Marketing user should be able to edit Tech record (Edit access)")
		}
		if !permService.CheckRecordAccess(context.Background(), schema, nonTechRecord, constants.PermRead, marketingSession) {
			t.Error("Marketing user should be able to read Finance record (Edit grants read too)")
		}
	})

	// Cleanup
	if _, err := db.Exec(fmt.Sprintf("%s %s %s %s %s 'test-sharing-%%'", persistence.KeywordDeleteFrom, constants.TableSharingRule, persistence.KeywordWhere, constants.FieldID, persistence.KeywordLike)); err != nil {
		t.Logf("Cleanup failed: %v", err)
	}
	if _, err := db.Exec(fmt.Sprintf("%s %s %s %s %s 'test-sharing-%%'", persistence.KeywordDeleteFrom, constants.TableUser, persistence.KeywordWhere, constants.FieldID, persistence.KeywordLike)); err != nil {
		t.Logf("Cleanup failed: %v", err)
	}
	if _, err := db.Exec(fmt.Sprintf("%s %s %s %s %s 'test-sharing-%%'", persistence.KeywordDeleteFrom, constants.TableRole, persistence.KeywordWhere, constants.FieldID, persistence.KeywordLike)); err != nil {
		t.Logf("Cleanup failed: %v", err)
	}

	t.Log("✅ Sharing Rules tests completed successfully")
}
