package services

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/nexuscrm/backend/pkg/constants"
)

// ==================== Role Hierarchy ====================

// refreshRoleHierarchy loads the role hierarchy from the database into cache
func (ps *PermissionService) refreshRoleHierarchy() {
	ps.roleHierarchyMu.Lock()
	defer ps.roleHierarchyMu.Unlock()

	query := fmt.Sprintf("SELECT id, parent_role_id FROM %s WHERE is_deleted = 0", constants.TableRole)
	rows, err := ps.db.DB().Query(query)
	if err != nil {
		log.Printf("Warning: Failed to load role hierarchy: %v", err)
		return
	}
	defer rows.Close()

	ps.roleHierarchyCache = make(map[string]*string)
	for rows.Next() {
		var id string
		var parentID sql.NullString
		if err := rows.Scan(&id, &parentID); err != nil {
			log.Printf("Warning: Failed to scan role: %v", err)
			continue
		}
		if parentID.Valid {
			ps.roleHierarchyCache[id] = &parentID.String
		} else {
			ps.roleHierarchyCache[id] = nil
		}
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
func (ps *PermissionService) getRecordOwnerRoleID(ownerID string) *string {
	query := fmt.Sprintf("SELECT role_id FROM %s WHERE id = ?", constants.TableUser)
	var roleID sql.NullString
	err := ps.db.DB().QueryRow(query, ownerID).Scan(&roleID)
	if err != nil {
		return nil
	}
	if roleID.Valid {
		return &roleID.String
	}
	return nil
}

// RefreshRoleHierarchy reloads the role hierarchy cache
// Call this when roles are modified
func (ps *PermissionService) RefreshRoleHierarchy() {
	ps.refreshRoleHierarchy()
}
