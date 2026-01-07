package services

import (
	"context"
	"fmt"

	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Permission Set Permissions ====================

// GetPermissionSetObjectPermissions retrieves all object permissions for a permission set
func (ps *PermissionService) GetPermissionSetObjectPermissions(permissionSetID string) ([]models.SystemObjectPerms, error) {
	return ps.repo.ListPermissionSetObjectPermissions(context.Background(), permissionSetID)
}

// GetPermissionSetFieldPermissions retrieves all field permissions for a permission set
func (ps *PermissionService) GetPermissionSetFieldPermissions(permissionSetID string) ([]models.SystemFieldPerms, error) {
	return ps.repo.ListPermissionSetFieldPermissions(context.Background(), permissionSetID)
}

// UpdatePermissionSetObjectPermission creates or updates an object permission for a permission set
func (ps *PermissionService) UpdatePermissionSetObjectPermission(perm models.SystemObjectPerms) error {
	if perm.PermissionSetID == nil || *perm.PermissionSetID == "" {
		return fmt.Errorf("permission_set_id is required")
	}
	return ps.repo.UpsertObjectPermission(context.Background(), perm)
}

// UpdatePermissionSetFieldPermission creates or updates a field permission for a permission set
func (ps *PermissionService) UpdatePermissionSetFieldPermission(perm models.SystemFieldPerms) error {
	if perm.PermissionSetID == nil || *perm.PermissionSetID == "" {
		return fmt.Errorf("permission_set_id is required")
	}
	return ps.repo.UpsertFieldPermission(context.Background(), perm)
}

// ==================== Effective Permissions (Admin View) ====================

// GetEffectiveObjectPermissions returns the merged object permissions for a specific user
func (ps *PermissionService) GetEffectiveObjectPermissions(userID string) ([]models.SystemObjectPerms, error) {
	ctx := context.Background()
	profileID, err := ps.repo.GetUserProfileID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return ps.repo.ListEffectiveObjectPermissionsForUser(ctx, userID, profileID)
}

// GetEffectiveFieldPermissions returns the merged field permissions for a specific user
func (ps *PermissionService) GetEffectiveFieldPermissions(userID string) ([]models.SystemFieldPerms, error) {
	ctx := context.Background()
	profileID, err := ps.repo.GetUserProfileID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return ps.repo.ListEffectiveFieldPermissionsForUser(ctx, userID, profileID)
}

// ==================== Permission Set Management ====================

// CreatePermissionSet creates a new permission set
func (ps *PermissionService) CreatePermissionSet(name, label, description string) (string, error) {
	return ps.repo.CreatePermissionSet(context.Background(), name, label, description)
}

// UpdatePermissionSet updates a permission set
func (ps *PermissionService) UpdatePermissionSet(id string, name, label, description string, isActive bool) error {
	return ps.repo.UpdatePermissionSet(context.Background(), id, name, label, description, isActive)
}

// DeletePermissionSet deletes a permission set
func (ps *PermissionService) DeletePermissionSet(id string) error {
	return ps.repo.DeletePermissionSet(context.Background(), id)
}
