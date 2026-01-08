package services

import (
	"context"
	"log"
)

// ==================== Role Hierarchy ====================

// refreshRoleHierarchy loads the role hierarchy from the database into cache
func (ps *PermissionService) refreshRoleHierarchy() {
	ps.roleHierarchyMu.Lock()
	defer ps.roleHierarchyMu.Unlock()

	roles, err := ps.repo.GetAllRoles(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to load role hierarchy: %v", err)
		return
	}

	ps.roleHierarchyCache = make(map[string]*string)
	for _, role := range roles {
		ps.roleHierarchyCache[role.ID] = role.ParentRoleID
	}
}

// getRoleAncestors returns all ancestor role IDs for a given role (parent, grandparent, etc.)
func (ps *PermissionService) getRoleAncestors(roleID string) []string {
	ps.roleHierarchyMu.RLock()
	defer ps.roleHierarchyMu.RUnlock()

	ancestors := make([]string, 0)
	visited := make(map[string]bool) // Prevent infinite loops

	currentID := roleID
	for {
		parentID, exists := ps.roleHierarchyCache[currentID]
		if !exists || parentID == nil {
			break
		}
		if visited[*parentID] {
			// Circular reference detected - should not happen, but handle gracefully
			log.Printf("Warning: Circular role hierarchy detected at %s", *parentID)
			break
		}
		visited[*parentID] = true
		ancestors = append(ancestors, *parentID)
		currentID = *parentID
	}

	return ancestors
}

// isUserAboveInHierarchy checks if the user's role is an ancestor of the target role
// Returns true if userRoleID is above targetRoleID in the hierarchy
func (ps *PermissionService) isUserAboveInHierarchy(userRoleID, targetRoleID *string) bool {
	// Both must have roles for hierarchy check
	if userRoleID == nil || targetRoleID == nil {
		return false
	}

	// Same role - not above
	if *userRoleID == *targetRoleID {
		return false
	}

	// Get all ancestors of the target role
	ancestors := ps.getRoleAncestors(*targetRoleID)

	// Check if user's role is in the ancestor list
	for _, ancestorID := range ancestors {
		if ancestorID == *userRoleID {
			return true
		}
	}

	return false
}

// getRecordOwnerRoleID retrieves the role_id of the record owner
func (ps *PermissionService) getRecordOwnerRoleID(ctx context.Context, ownerID string) *string {
	roleID, err := ps.userRepo.GetUserRoleID(ctx, ownerID)
	if err != nil {
		return nil
	}
	return roleID
}

// RefreshRoleHierarchy reloads the role hierarchy cache
// Call this when roles are modified
func (ps *PermissionService) RefreshRoleHierarchy() {
	ps.refreshRoleHierarchy()
}
