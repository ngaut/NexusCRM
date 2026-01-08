package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// PermissionService handles permission checks for objects and fields.
//
// This service queries the database for every permission check (per-request model).
// This ensures permission changes are immediately effective without cache invalidation.
//
// Permission evaluation order:
//  1. SuperUser check (system_admin profile bypasses all checks)
//  2. Object-level permissions from _System_ObjectPerms
//  3. Field-level permissions from _System_FieldPerms
//  4. Record-level access (ownership, sharing rules) - future enhancement
type PermissionService struct {
	metadata *MetadataService
	repo     *persistence.PermissionRepository
	userRepo *persistence.UserRepository
	formula  *formula.Engine

	// Role hierarchy cache: maps role_id -> parent_role_id
	roleHierarchyCache map[string]*string
	roleHierarchyMu    sync.RWMutex
}

// NewPermissionService creates a new PermissionService
func NewPermissionService(
	repo *persistence.PermissionRepository,
	metadata *MetadataService,
	userRepo *persistence.UserRepository,
) *PermissionService {
	ps := &PermissionService{
		metadata:           metadata,
		repo:               repo,
		userRepo:           userRepo,
		formula:            formula.NewEngine(),
		roleHierarchyCache: make(map[string]*string),
	}
	// Pre-load role hierarchy
	ps.refreshRoleHierarchy()
	return ps
}

// ==================== Object Permission Queries ====================

// loadEffectiveObjectPermission loads permissions considering Profile AND Permission Sets
func (ps *PermissionService) loadEffectiveObjectPermission(ctx context.Context, user *models.UserSession, objectAPIName string) (*models.SystemObjectPerms, error) {
	return ps.repo.LoadEffectiveObjectPermission(ctx, user, objectAPIName)
}

// loadEffectiveFieldPermission loads field permissions considering Profile AND Permission Sets
func (ps *PermissionService) loadEffectiveFieldPermission(ctx context.Context, user *models.UserSession, objectAPIName, fieldAPIName string) (*models.SystemFieldPerms, error) {
	return ps.repo.LoadEffectiveFieldPermission(ctx, user, objectAPIName, fieldAPIName)
}

// ==================== Core Permission Checks ====================

// CheckObjectPermissionWithUser checks if a user has permission for an operation on an object.
// Operations: "create", "read", "edit", "delete"
// Returns true if the operation is permitted.
func (ps *PermissionService) CheckObjectPermissionWithUser(ctx context.Context, objectAPIName string, operation string, user *models.UserSession) bool {
	// No user = no access
	if user == nil {
		return false
	}

	// SuperUser bypass - system_admin or IsSystemAdmin flag has full access
	if user.IsSystemAdmin || constants.IsSuperUser(user.ProfileID) {
		return true
	}

	// Query database for permission
	// Query database for effective permission (Profile OR Permission Sets)
	perm, err := ps.loadEffectiveObjectPermission(ctx, user, objectAPIName)
	if err != nil {
		return false
	}

	// No permission record = no access
	if perm == nil {
		return false
	}

	// Check specific operation
	switch strings.ToLower(operation) {
	case constants.PermRead:
		return perm.AllowRead
	case constants.PermCreate:
		return perm.AllowCreate
	case constants.PermEdit:
		return perm.AllowEdit
	case constants.PermDelete:
		return perm.AllowDelete
	default:
		return false
	}
}

// CheckPermissionOrErrorWithUser checks permission and returns a specific PermissionError if false
func (ps *PermissionService) CheckPermissionOrErrorWithUser(ctx context.Context, objectAPIName string, operation string, user *models.UserSession) error {
	if !ps.CheckObjectPermissionWithUser(ctx, objectAPIName, operation, user) {
		return errors.NewPermissionError(operation, objectAPIName)
	}
	return nil
}

// Record-level access functions are in permission_record_access.go:
// - CheckRecordAccess, checkManualShareAccess
// - checkTeamMemberAccess, accessLevelAllowsOperation

// CheckFieldEditabilityWithUser checks if a field can be edited by the current user
func (ps *PermissionService) CheckFieldEditabilityWithUser(ctx context.Context, objectAPIName, fieldAPIName string, user *models.UserSession) bool {
	// System fields are never editable
	if isFieldSystemReadOnlyByName(fieldAPIName) {
		return false
	}

	// No user = no access
	if user == nil {
		return false
	}

	// SuperUser bypass
	if user.IsSystemAdmin || constants.IsSuperUser(user.ProfileID) {
		return true
	}

	// Check field-level permission (Effective)
	perm, err := ps.loadEffectiveFieldPermission(ctx, user, objectAPIName, fieldAPIName)
	if err != nil {
		return false
	}

	if perm != nil {
		return perm.Editable
	}

	// Fallback to object permission
	return ps.CheckObjectPermissionWithUser(ctx, objectAPIName, constants.PermEdit, user)
}

// CheckFieldVisibilityWithUser checks if a field is visible to the current user
func (ps *PermissionService) CheckFieldVisibilityWithUser(ctx context.Context, objectAPIName, fieldAPIName string, user *models.UserSession) bool {
	if user == nil {
		return false
	}

	// SuperUser bypass
	if user.IsSystemAdmin || constants.IsSuperUser(user.ProfileID) {
		return true
	}

	// Check field-level permission (Effective)
	perm, err := ps.loadEffectiveFieldPermission(ctx, user, objectAPIName, fieldAPIName)
	if err != nil {
		return false
	}

	// If explicit field permission exists, use it
	if perm != nil {
		return perm.Readable
	}

	// Fallback to object permission
	return ps.CheckObjectPermissionWithUser(ctx, objectAPIName, constants.PermRead, user)
}

// RefreshPermissions reloads permissions from the database
// This specifically refreshes the Role Hierarchy cache.
// Object/Field permissions are not cached (checked per-request), so they don't need refreshing.
func (ps *PermissionService) RefreshPermissions() error {
	ps.refreshRoleHierarchy()
	return nil
}

// GetEffectiveSchema returns the schema with field-level visibility applied
func (ps *PermissionService) GetEffectiveSchema(ctx context.Context, schema *models.ObjectMetadata, user *models.UserSession) *models.ObjectMetadata {
	if schema == nil {
		return nil
	}

	// Create a copy of the schema to avoid mutating the original (which might be cached)
	effectiveSchema := *schema
	effectiveSchema.Fields = make([]models.FieldMetadata, 0, len(schema.Fields))

	// Super users see all fields
	if user != nil && (user.IsSystemAdmin || constants.IsSuperUser(user.ProfileID)) {
		return schema
	}

	for _, field := range schema.Fields {
		if ps.CheckFieldVisibilityWithUser(ctx, schema.APIName, field.APIName, user) {
			effectiveSchema.Fields = append(effectiveSchema.Fields, field)
		}
	}

	return &effectiveSchema
}

// GetObjectPermissions retrieves all object permissions for a profile
func (ps *PermissionService) GetObjectPermissions(profileID string) ([]models.SystemObjectPerms, error) {
	return ps.repo.ListObjectPermissions(context.Background(), profileID)
}

// UpdateObjectPermission creates or updates an object permission
func (ps *PermissionService) UpdateObjectPermission(perm models.SystemObjectPerms) error {
	return ps.repo.UpsertObjectPermission(context.Background(), perm)
}

// UpdateObjectPermissionTx creates or updates an object permission within a transaction
func (ps *PermissionService) UpdateObjectPermissionTx(tx *sql.Tx, perm models.SystemObjectPerms) error {
	return ps.repo.UpsertObjectPermissionTx(context.Background(), tx, perm)
}

// updateObjectPermission creates or updates an object permission using the provided executor

// GetFieldPermissions retrieves all field permissions for a profile
func (ps *PermissionService) GetFieldPermissions(profileID string) ([]models.SystemFieldPerms, error) {
	return ps.repo.ListFieldPermissions(context.Background(), profileID)
}

// GrantFieldPermissions grants permissions for a field (Public wrapper for update)
func (ps *PermissionService) GrantFieldPermissions(ctx context.Context, perms models.SystemFieldPerms) error {
	return ps.UpdateFieldPermission(perms)
}

// UpdateFieldPermission creates or updates a field permission
func (ps *PermissionService) UpdateFieldPermission(perms models.SystemFieldPerms) error {
	return ps.repo.UpsertFieldPermission(context.Background(), perms)
}

// isFieldSystemReadOnlyByName checks if a field is a system read-only field based on its name
func isFieldSystemReadOnlyByName(fieldAPIName string) bool {
	return constants.IsSystemField(fieldAPIName)
}

// isFieldSystemReadOnly checks if a field is system read-only by looking up metadata
// This version takes metadata service and object name for richer checking
func isFieldSystemReadOnly(metadata *MetadataService, objectAPIName string, fieldAPIName string) bool {
	// First check if it's a common system field
	if isFieldSystemReadOnlyByName(fieldAPIName) {
		return true
	}

	// If we have metadata, check the field's is_system flag
	if metadata != nil {
		schema := metadata.GetSchema(context.Background(), objectAPIName)
		if schema != nil {
			for _, field := range schema.Fields {
				if field.APIName == fieldAPIName {
					return field.IsSystem
				}
			}
		}
	}

	return false
}

// GrantInitialPermissions grants default permissions for a new object to all profiles
func (ps *PermissionService) GrantInitialPermissions(ctx context.Context, objectAPIName string) error {
	return ps.repo.GrantInitialPermissions(ctx, objectAPIName)
}

// Role hierarchy functions are in permission_role_hierarchy.go:
// - refreshRoleHierarchy, getRoleAncestors, isUserAboveInHierarchy, getRecordOwnerRoleID, RefreshRoleHierarchy

// CreateRole creates a new role
func (ps *PermissionService) CreateRole(ctx context.Context, name, description string, parentRoleID *string) (*models.SystemRole, error) {
	// Validate parent role if provided
	if parentRoleID != nil {
		parent, err := ps.repo.GetRole(ctx, *parentRoleID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate parent role: %w", err)
		}
		if parent == nil {
			return nil, errors.NewValidationError("parent_role_id", "parent role does not exist")
		}
	}

	id, err := ps.repo.CreateRole(ctx, name, description, parentRoleID)
	if err != nil {
		return nil, err
	}

	// Refresh cache
	ps.refreshRoleHierarchy()

	return &models.SystemRole{
		ID:           id,
		Name:         name,
		Description:  description,
		ParentRoleID: parentRoleID,
	}, nil
}

// GetRole retrieves a role by ID
func (ps *PermissionService) GetRole(ctx context.Context, id string) (*models.SystemRole, error) {
	return ps.repo.GetRole(ctx, id)
}

// GetAllRoles retrieves all roles
func (ps *PermissionService) GetAllRoles(ctx context.Context) ([]*models.SystemRole, error) {
	return ps.repo.GetAllRoles(ctx)
}

// UpdateRole updates an existing role
func (ps *PermissionService) UpdateRole(ctx context.Context, id string, name, description string, parentRoleID *string) (*models.SystemRole, error) {
	// Validate parent role if provided
	if parentRoleID != nil {
		// Prevent self-reference
		if *parentRoleID == id {
			return nil, errors.NewValidationError("parent_role_id", "cannot be self")
		}

		parent, err := ps.repo.GetRole(ctx, *parentRoleID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate parent role: %w", err)
		}
		if parent == nil {
			return nil, errors.NewValidationError("parent_role_id", "parent role does not exist")
		}
	}

	if err := ps.repo.UpdateRole(ctx, id, name, description, parentRoleID); err != nil {
		return nil, err
	}

	// Refresh cache
	ps.refreshRoleHierarchy()

	return ps.GetRole(ctx, id)
}

// DeleteRole deletes a role
func (ps *PermissionService) DeleteRole(ctx context.Context, id string) error {
	if err := ps.repo.DeleteRole(ctx, id); err != nil {
		return err
	}
	// Refresh cache
	ps.refreshRoleHierarchy()
	return nil
}

// Sharing rule functions are in permission_sharing_rules.go:

// - isUserInRoleOrBelow, checkSharingRuleAccess, evaluateSharingCriteria

// Permission Set functions are in permission_perm_sets.go:
// - GetPermissionSetObjectPermissions, GetPermissionSetFieldPermissions
// - UpdatePermissionSetObjectPermission, UpdatePermissionSetFieldPermission
// - GetEffectiveObjectPermissions, GetEffectiveFieldPermissions
// - CreatePermissionSet, UpdatePermissionSet, DeletePermissionSet
