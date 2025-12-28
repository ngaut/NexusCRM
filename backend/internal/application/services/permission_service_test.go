package services

import (
	"testing"

	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/pkg/constants"
)

// MockPermissionService provides test helpers for PermissionService
// Note: These tests don't require a database connection since they test
// the logic flow (SuperUser bypass, nil user handling, etc.)

func TestCheckObjectPermissionWithUser_NilUser(t *testing.T) {
	// Create a minimal permission service (db and metadata will be nil)
	ps := &PermissionService{}

	// Nil user should return false
	result := ps.CheckObjectPermissionWithUser(constants.ObjectAccount, constants.PermRead, nil)
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
		result := ps.CheckObjectPermissionWithUser(constants.ObjectAccount, op, user)
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
		result := ps.CheckFieldEditabilityWithUser(constants.ObjectAccount, field, user)
		if result {
			t.Errorf("Expected system field %s to be non-editable, got editable", field)
		}
	}
}

func TestCheckFieldVisibilityWithUser_NilUser(t *testing.T) {
	ps := &PermissionService{}

	// Nil user should return false
	result := ps.CheckFieldVisibilityWithUser(constants.ObjectAccount, constants.FieldName, nil)
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
	result := ps.CheckFieldVisibilityWithUser(constants.ObjectAccount, constants.FieldName, user)
	if !result {
		t.Error("Expected true for SuperUser field visibility, got false")
	}
}

func TestCheckPermissionOrErrorWithUser_NilUser(t *testing.T) {
	ps := &PermissionService{}

	err := ps.CheckPermissionOrErrorWithUser(constants.ObjectAccount, constants.PermRead, nil)
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

	err := ps.CheckPermissionOrErrorWithUser(constants.ObjectAccount, constants.PermRead, user)
	if err != nil {
		t.Errorf("Expected no error for SuperUser, got: %v", err)
	}
}
func TestGetEffectiveSchema_NilSchema(t *testing.T) {
	ps := &PermissionService{}
	result := ps.GetEffectiveSchema(nil, nil)
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
		APIName: constants.ObjectAccount,
		Fields: []models.FieldMetadata{
			{APIName: constants.FieldName},
			{APIName: "Secret"},
		},
	}

	result := ps.GetEffectiveSchema(schema, user)
	if len(result.Fields) != 2 {
		t.Errorf("Expected SuperUser to see all 2 fields, got %d", len(result.Fields))
	}
}

func TestGetEffectiveSchema_NilUser(t *testing.T) {
	ps := &PermissionService{}
	schema := &models.ObjectMetadata{
		APIName: constants.ObjectAccount,
		Fields: []models.FieldMetadata{
			{APIName: constants.FieldName},
			{APIName: "Secret"},
		},
	}

	// Nil user -> CheckFieldVisibility returns false -> Filtering happens
	result := ps.GetEffectiveSchema(schema, nil)
	if len(result.Fields) != 0 {
		t.Errorf("Expected nil user to see 0 fields, got %d", len(result.Fields))
	}
}
func TestCheckRecordAccess_Owner(t *testing.T) {
	ps := &PermissionService{}
	user := &models.UserSession{
		ID:        "user-123",
		ProfileID: constants.ProfileStandardUser,
	}
	record := models.SObject{
		constants.FieldOwnerID: "user-123",
		constants.FieldName:    "My Account",
	}

	if !ps.CheckRecordAccess(nil, record, constants.PermEdit, user) {
		t.Error("Expected owner to have access")
	}
}

func TestCheckRecordAccess_NonOwner(t *testing.T) {
	ps := &PermissionService{}
	user := &models.UserSession{
		ID:        "user-456",
		ProfileID: constants.ProfileStandardUser,
	}
	record := models.SObject{
		constants.FieldOwnerID: "user-123", // Different owner
		constants.FieldName:    "Someone Else's Account",
	}

	if ps.CheckRecordAccess(nil, record, constants.PermEdit, user) {
		t.Error("Expected non-owner to be denied access")
	}
}

func TestCheckRecordAccess_SuperUser(t *testing.T) {
	ps := &PermissionService{}
	user := &models.UserSession{
		ID:        "admin-user",
		ProfileID: constants.ProfileSystemAdmin,
	}
	record := models.SObject{
		constants.FieldOwnerID: "user-123",
		constants.FieldName:    "User's Account",
	}

	if !ps.CheckRecordAccess(nil, record, constants.PermEdit, user) {
		t.Error("Expected SuperUser to have access regardless of owner")
	}
}

func TestCheckRecordAccess_NoOwnerField(t *testing.T) {
	ps := &PermissionService{}
	user := &models.UserSession{
		ID:        "user-123",
		ProfileID: constants.ProfileStandardUser,
	}
	record := models.SObject{
		constants.FieldName: "Orphan Record",
		// No owner_id
	}

	if ps.CheckRecordAccess(nil, record, constants.PermEdit, user) {
		t.Error("Expected denial for record without owner_id")
	}
}

func TestCheckRecordAccess_PointerOwner(t *testing.T) {
	ps := &PermissionService{}
	user := &models.UserSession{
		ID:        "user-123",
		ProfileID: constants.ProfileStandardUser,
	}
	ownerID := "user-123"
	record := models.SObject{
		constants.FieldOwnerID: &ownerID,
		constants.FieldName:    "Pointer Owner Rec",
	}

	if !ps.CheckRecordAccess(nil, record, constants.PermEdit, user) {
		t.Error("Expected owner (via pointer) to have access")
	}
}
