package services

import (
	"context"
	"testing"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// MockPermissionService provides test helpers for PermissionService
// Note: These tests don't require a database connection since they test
// the logic flow (SuperUser bypass, nil user handling, etc.)

func TestCheckObjectPermissionWithUser_NilUser(t *testing.T) {
	// Create a minimal permission service (db and metadata will be nil)
	ps := &PermissionService{}

	// Nil user should return false
	result := ps.CheckObjectPermissionWithUser(context.Background(), "Account", constants.PermRead, nil)
	if result {
		t.Error("Expected false for nil user, got true")
	}
}

func TestCheckObjectPermissionWithUser_SuperUser(t *testing.T) {
	// Create a minimal permission service
	ps := &PermissionService{}

	// Super user (system_admin) should bypass all checks
	user := &models.UserSession{
		ID:        "test-user-id",
		Name:      "Test Admin",
		ProfileID: constants.ProfileSystemAdmin,
	}

	operations := []string{constants.PermRead, constants.PermCreate, constants.PermEdit, constants.PermDelete}
	for _, op := range operations {
		result := ps.CheckObjectPermissionWithUser(context.Background(), "Account", op, user)
		if !result {
			t.Errorf("Expected true for SuperUser %s operation, got false", op)
		}
	}
}

func TestCheckFieldEditabilityWithUser_SystemField(t *testing.T) {
	ps := &PermissionService{}

	user := &models.UserSession{
		ID:        "test-user-id",
		Name:      "Test User",
		ProfileID: constants.ProfileSystemAdmin,
	}

	// System fields should never be editable, even for super users
	systemFields := []string{
		constants.FieldID,
		constants.FieldCreatedDate,
		constants.FieldCreatedByID,
		constants.FieldLastModifiedDate,
		constants.FieldLastModifiedByID,
		constants.FieldIsDeleted,
	}

	for _, field := range systemFields {
		result := ps.CheckFieldEditabilityWithUser(context.Background(), "Account", field, user)
		if result {
			t.Errorf("Expected system field %s to be non-editable, got editable", field)
		}
	}
}

func TestCheckFieldVisibilityWithUser_NilUser(t *testing.T) {
	ps := &PermissionService{}

	// Nil user should return false
	result := ps.CheckFieldVisibilityWithUser(context.Background(), "Account", constants.FieldName, nil)
	if result {
		t.Error("Expected false for nil user, got true")
	}
}

func TestCheckFieldVisibilityWithUser_SuperUser(t *testing.T) {
	ps := &PermissionService{}

	user := &models.UserSession{
		ID:        "test-user-id",
		Name:      "Test Admin",
		ProfileID: constants.ProfileSystemAdmin,
	}

	// Super user should see all fields
	result := ps.CheckFieldVisibilityWithUser(context.Background(), "Account", constants.FieldName, user)
	if !result {
		t.Error("Expected true for SuperUser field visibility, got false")
	}
}

func TestCheckPermissionOrErrorWithUser_NilUser(t *testing.T) {
	ps := &PermissionService{}

	err := ps.CheckPermissionOrErrorWithUser(context.Background(), "Account", constants.PermRead, nil)
	if err == nil {
		t.Error("Expected error for nil user, got nil")
	}
}

func TestCheckPermissionOrErrorWithUser_SuperUser(t *testing.T) {
	ps := &PermissionService{}

	user := &models.UserSession{
		ID:        "test-user-id",
		Name:      "Test Admin",
		ProfileID: constants.ProfileSystemAdmin,
	}

	err := ps.CheckPermissionOrErrorWithUser(context.Background(), "Account", constants.PermRead, user)
	if err != nil {
		t.Errorf("Expected no error for SuperUser, got: %v", err)
	}
}
func TestGetEffectiveSchema_NilSchema(t *testing.T) {
	ps := &PermissionService{}
	result := ps.GetEffectiveSchema(context.Background(), nil, nil)
	if result != nil {
		t.Error("Expected nil result for nil schema")
	}
}

func TestGetEffectiveSchema_SuperUser(t *testing.T) {
	ps := &PermissionService{}
	user := &models.UserSession{
		ID:        "admin",
		ProfileID: constants.ProfileSystemAdmin,
	}
	schema := &models.ObjectMetadata{
		APIName: "Account",
		Fields: []models.FieldMetadata{
			{APIName: constants.FieldName},
			{APIName: "Secret"},
		},
	}

	result := ps.GetEffectiveSchema(context.Background(), schema, user)
	if len(result.Fields) != 2 {
		t.Errorf("Expected SuperUser to see all 2 fields, got %d", len(result.Fields))
	}
}

func TestGetEffectiveSchema_NilUser(t *testing.T) {
	ps := &PermissionService{}
	schema := &models.ObjectMetadata{
		APIName: "Account",
		Fields: []models.FieldMetadata{
			{APIName: constants.FieldName},
			{APIName: "Secret"},
		},
	}

	// Nil user -> CheckFieldVisibility returns false -> Filtering happens
	result := ps.GetEffectiveSchema(context.Background(), schema, nil)
	if len(result.Fields) != 0 {
		t.Errorf("Expected nil user to see 0 fields, got %d", len(result.Fields))
	}
}

func TestCheckRecordAccess_Owner(t *testing.T) {
	// Setup
	ps := &PermissionService{}
	user := &models.UserSession{ID: "user1", ProfileID: "profile1"}

	// Record owned by user
	record := models.SObject{
		constants.FieldID:      "rec1",
		constants.FieldOwnerID: "user1",
	}

	// Expect IsUserInGroup to NOT be called for owner check (optimization)
	// The implementation checks owner first: if ownerIDStr == user.ID { return true }

	if !ps.CheckRecordAccess(context.Background(), nil, record, constants.PermEdit, user) {
		t.Errorf("Expected owner to have access")
	}
}

// TestCheckRecordAccess_NonOwner removed as it requires DB mocking (PermissionService uses concrete repos)
// Covered by sharing_rules_test.go integration tests.

func TestCheckRecordAccess_SuperUser(t *testing.T) {
	ps := &PermissionService{}
	// System Admin Profile
	user := &models.UserSession{ID: "user1", ProfileID: constants.ProfileSystemAdmin}

	record := models.SObject{
		constants.FieldID:      "rec1",
		constants.FieldOwnerID: "other_user",
	}

	// Should pass without DB lookup due to SuperUser check returning early
	if !ps.CheckRecordAccess(context.Background(), nil, record, constants.PermEdit, user) {
		t.Errorf("Expected SuperUser to have access")
	}
}

func TestCheckRecordAccess_NoOwnerField(t *testing.T) {
	ps := &PermissionService{}
	user := &models.UserSession{ID: "user1", ProfileID: "profile1"}

	// Record with no owner field
	record := models.SObject{
		constants.FieldID: "rec1",
	}

	// Should fail safely
	// No owner field -> hasOwner = false. Skips Group/Owner checks.
	// Skips Hierarchy.
	// Skips Sharing Rules (nil schema).
	// Skips Manual Share (nil schema).

	if ps.CheckRecordAccess(context.Background(), nil, record, constants.PermEdit, user) {
		t.Errorf("Expected no access when no owner field present")
	}
}

func TestCheckRecordAccess_PointerOwner(t *testing.T) {
	ps := &PermissionService{}
	user := &models.UserSession{ID: "user1", ProfileID: "profile1"}

	ownerID := "user1"
	record := models.SObject{
		constants.FieldID:      "rec1",
		constants.FieldOwnerID: &ownerID,
	}

	if !ps.CheckRecordAccess(context.Background(), nil, record, constants.PermEdit, user) {
		t.Errorf("Expected owner (pointer) to have access")
	}
}
