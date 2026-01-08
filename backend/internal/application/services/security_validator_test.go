package services_test

import (
	"context"
	"testing"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPermissionChecker implements services.PermissionChecker
type MockPermissionChecker struct {
	mock.Mock
}

func (m *MockPermissionChecker) CheckObjectPermissionWithUser(ctx context.Context, objectName string, permission string, user *models.UserSession) bool {
	args := m.Called(ctx, objectName, permission, user)
	return args.Bool(0)
}

func (m *MockPermissionChecker) CheckFieldVisibilityWithUser(ctx context.Context, objectName string, fieldName string, user *models.UserSession) bool {
	args := m.Called(ctx, objectName, fieldName, user)
	return args.Bool(0)
}

// MockMetadataProvider implements services.MetadataProvider
type MockMetadataProvider struct {
	mock.Mock
}

func (m *MockMetadataProvider) GetSchema(ctx context.Context, objectName string) *models.ObjectMetadata {
	args := m.Called(ctx, objectName)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.ObjectMetadata)
}

func TestSecurityValidator_ValidateAndRewrite(t *testing.T) {
	// Setup user
	stdUser := &models.UserSession{
		ID:        "user-123",
		ProfileID: constants.ProfileStandardUser,
	}

	adminUser := &models.UserSession{
		ID:        "admin-123",
		ProfileID: constants.ProfileSystemAdmin,
	}

	// Schema with OwnerID
	accountSchema := &models.ObjectMetadata{
		APIName: "Account",
		Fields: []models.FieldMetadata{
			{APIName: "name", Type: constants.FieldTypeText},
			{APIName: "owner_id", Type: constants.FieldTypeLookup},
			{APIName: "industry", Type: constants.FieldTypePicklist},
		},
	}

	// Schema without OwnerID
	configSchema := &models.ObjectMetadata{
		APIName: "SystemConfig",
		Fields: []models.FieldMetadata{
			{APIName: "config_key", Type: constants.FieldTypeText},
			{APIName: "value", Type: constants.FieldTypeText},
		},
	}

	t.Run("Permissions: Block Table Access", func(t *testing.T) {
		mockPerms := new(MockPermissionChecker)
		mockMeta := new(MockMetadataProvider)

		// Expect permission check failure
		// Note: AST traversal visits Fields before From clause, so Field check happens first!
		mockPerms.On("CheckFieldVisibilityWithUser", mock.Anything, "Account", "name", stdUser).Return(true)
		mockPerms.On("CheckObjectPermissionWithUser", mock.Anything, "Account", constants.PermRead, stdUser).Return(false)

		validator := services.NewSecurityValidator(mockPerms, mockMeta)

		sql := "SELECT name FROM Account"
		_, _, err := validator.ValidateAndRewrite(context.Background(), sql, nil, stdUser)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied: cannot read table 'Account'")

		mockPerms.AssertExpectations(t)
	})

	t.Run("Permissions: Block Field Access", func(t *testing.T) {
		mockPerms := new(MockPermissionChecker)
		mockMeta := new(MockMetadataProvider)

		// Table OK, Field "salary" blocked
		mockPerms.On("CheckObjectPermissionWithUser", mock.Anything, "Employee", constants.PermRead, stdUser).Return(true)
		mockPerms.On("CheckFieldVisibilityWithUser", mock.Anything, "Employee", "salary", stdUser).Return(false)
		mockPerms.On("CheckFieldVisibilityWithUser", mock.Anything, "Employee", "name", stdUser).Return(true) // name OK

		validator := services.NewSecurityValidator(mockPerms, mockMeta)

		sql := "SELECT name, salary FROM Employee"
		_, _, err := validator.ValidateAndRewrite(context.Background(), sql, nil, stdUser)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied: cannot read field 'salary'")
	})

	t.Run("RLS: Standard User - Inject Owner Filter", func(t *testing.T) {
		mockPerms := new(MockPermissionChecker)
		mockMeta := new(MockMetadataProvider)

		mockPerms.On("CheckObjectPermissionWithUser", mock.Anything, "Account", constants.PermRead, stdUser).Return(true)
		mockPerms.On("CheckFieldVisibilityWithUser", mock.Anything, "Account", "name", stdUser).Return(true)
		mockPerms.On("CheckFieldVisibilityWithUser", mock.Anything, "Account", "industry", stdUser).Return(true)

		// Return schema with owner_id
		mockMeta.On("GetSchema", mock.Anything, "Account").Return(accountSchema)

		validator := services.NewSecurityValidator(mockPerms, mockMeta)

		sql := "SELECT name FROM Account WHERE industry = 'Tech'"
		rewritten, _, err := validator.ValidateAndRewrite(context.Background(), sql, nil, stdUser)

		assert.NoError(t, err)
		// rewritten SQL should contain owner_id check
		// Note: The AST restoration puts backticks usually
		assert.Contains(t, rewritten, "owner_id")
		assert.Contains(t, rewritten, "user-123")
		// TiDB parser output format check: `owner_id` or just owner_id?
		// We can check stricter logic if needed.
		// Expected: SELECT `name` FROM `Account` WHERE `industry` = 'Tech' AND `owner_id` = 'user-123'
	})

	t.Run("RLS: Admin User - No injection", func(t *testing.T) {
		mockPerms := new(MockPermissionChecker)
		mockMeta := new(MockMetadataProvider)

		mockPerms.On("CheckObjectPermissionWithUser", mock.Anything, "Account", constants.PermRead, adminUser).Return(true)
		mockPerms.On("CheckFieldVisibilityWithUser", mock.Anything, "Account", "name", adminUser).Return(true)

		// Metadata fetch might still happen or logic checks profile first.
		// implementation: if !constants.IsSuperUser(user.ProfileID) { ... }
		// So it should skip RLS injection completely without calling GetSchema or ApplyRLS

		validator := services.NewSecurityValidator(mockPerms, mockMeta)

		sql := "SELECT name FROM Account"
		rewritten, _, err := validator.ValidateAndRewrite(context.Background(), sql, nil, adminUser)

		assert.NoError(t, err)
		assert.NotContains(t, rewritten, "owner_id")
		assert.NotContains(t, rewritten, "user-123")
	})

	t.Run("RLS: Object without Owner - No injection", func(t *testing.T) {
		mockPerms := new(MockPermissionChecker)
		mockMeta := new(MockMetadataProvider)

		mockPerms.On("CheckObjectPermissionWithUser", mock.Anything, "SystemConfig", constants.PermRead, stdUser).Return(true)
		mockPerms.On("CheckFieldVisibilityWithUser", mock.Anything, "SystemConfig", "config_key", stdUser).Return(true)

		mockMeta.On("GetSchema", mock.Anything, "SystemConfig").Return(configSchema)

		validator := services.NewSecurityValidator(mockPerms, mockMeta)

		sql := "SELECT config_key FROM SystemConfig"
		rewritten, _, err := validator.ValidateAndRewrite(context.Background(), sql, nil, stdUser)

		assert.NoError(t, err)
		assert.NotContains(t, rewritten, "owner_id")
	})

	// Complex query (Join) - Currently skipped by logical limitation, but ensures no panic/error
	t.Run("Complex Join - Skip RLS (Current Limitation)", func(t *testing.T) {
		mockPerms := new(MockPermissionChecker)
		mockMeta := new(MockMetadataProvider)

		// Permission checks traverse all tables
		mockPerms.On("CheckObjectPermissionWithUser", mock.Anything, "Account", constants.PermRead, stdUser).Return(true)
		mockPerms.On("CheckObjectPermissionWithUser", mock.Anything, "Contact", constants.PermRead, stdUser).Return(true)
		// Join conditions involve fields: Account.id, Contact.account_id.
		// Select * might implicitly involve fields if expanded? No, parser sees *.
		// But ON clause has fields.
		mockPerms.On("CheckFieldVisibilityWithUser", mock.Anything, "Account", "id", stdUser).Return(true)
		mockPerms.On("CheckFieldVisibilityWithUser", mock.Anything, "Contact", "account_id", stdUser).Return(true)

		validator := services.NewSecurityValidator(mockPerms, mockMeta)

		sql := "SELECT * FROM Account JOIN Contact ON Account.id = Contact.account_id"
		rewritten, _, err := validator.ValidateAndRewrite(context.Background(), sql, nil, stdUser)

		assert.NoError(t, err)
		// Should NOT inject owner_id because Join structure doesn't match single table check in ApplyRLS
		assert.NotContains(t, rewritten, "owner_id")
	})
}
