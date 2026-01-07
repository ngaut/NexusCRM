package services

import (
	"fmt"
	"strings"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Record-Level Access Checks ====================

// CheckRecordAccess checks if a user can access a specific record
// This checks record-level sharing rules and ownership
func (ps *PermissionService) CheckRecordAccess(schema *models.ObjectMetadata, record models.SObject, operation string, user *models.UserSession) bool {
	if user == nil {
		return false
	}

	// SuperRecord access/SuperUser check
	if constants.IsSuperUser(user.ProfileID) {
		return true
	}

	// Extract record ID for sharing checks
	recordID := ""
	if id, ok := record[constants.FieldID]; ok {
		if idStr, ok := id.(string); ok {
			recordID = idStr
		}
	}

	// Extract owner ID from record for subsequent checks
	var ownerIDStr string
	ownerID, hasOwner := record[constants.FieldOwnerID]
	if hasOwner {
		// handle both string and pointer types just in case
		switch v := ownerID.(type) {
		case string:
			ownerIDStr = v
		case *string:
			if v != nil {
				ownerIDStr = *v
			}
		default:
			ownerIDStr = fmt.Sprintf("%v", v)
		}
	}

	// 1. Ownership Check (User or Group)
	if hasOwner {
		// A. User Ownership
		if ownerIDStr == user.ID {
			return true
		}

		// B. Group Ownership (Queue)
		// Check if the ownerID is a Group that the user is a member of
		if ps.db != nil {
			query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE group_id = ? AND user_id = ?", constants.TableGroupMember)
			var count int
			if err := ps.db.QueryRow(query, ownerIDStr, user.ID).Scan(&count); err == nil && count > 0 {
				return true
			}
		}

	}

	// 2. Role Hierarchy Check (Read-Only access for managers)
	// If the user's role is above the record owner's role in the hierarchy,
	// grant READ access (not edit/delete)
	if operation == constants.PermRead && hasOwner && ownerIDStr != "" {
		ownerRoleID := ps.getRecordOwnerRoleID(ownerIDStr)
		if ps.isUserAboveInHierarchy(user.RoleID, ownerRoleID) {
			return true
		}
	}

	// 3. Sharing Rules Check
	// Evaluate all sharing rules for this object
	if schema != nil {
		rules := ps.metadata.GetSharingRules(schema.APIName)
		for _, rule := range rules {
			if ps.checkSharingRuleAccess(record, rule, user, operation) {
				return true
			}
		}
	}

	// 4. Manual Record Share Check (_System_RecordShare)
	if schema != nil && recordID != "" && ps.db != nil {
		if ps.checkManualShareAccess(schema.APIName, recordID, user, operation) {
			return true
		}
	}

	// 5. Team Member Check (_System_TeamMember)
	if schema != nil && recordID != "" && ps.db != nil {
		if ps.checkTeamMemberAccess(schema.APIName, recordID, user, operation) {
			return true
		}
	}

	// Default to DENY if not owner, not above in hierarchy, and no sharing rule matches
	return false
}

// checkManualShareAccess checks if user has access via manual record share
func (ps *PermissionService) checkManualShareAccess(objectAPIName, recordID string, user *models.UserSession, operation string) bool {
	// Check direct user share
	query := fmt.Sprintf(`
		SELECT access_level FROM %s 
		WHERE object_api_name = ? AND record_id = ? AND is_deleted = 0
		AND (share_with_user_id = ? OR share_with_group_id IN (
			SELECT group_id FROM %s WHERE user_id = ?
		))
	`, constants.TableRecordShare, constants.TableGroupMember)

	rows, err := ps.db.DB().Query(query, objectAPIName, recordID, user.ID, user.ID)
	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var accessLevel string
		if err := rows.Scan(&accessLevel); err != nil {
			continue
		}
		if ps.accessLevelAllowsOperation(accessLevel, operation) {
			return true
		}
	}
	return false
}

// checkTeamMemberAccess checks if user is a team member with access
func (ps *PermissionService) checkTeamMemberAccess(objectAPIName, recordID string, user *models.UserSession, operation string) bool {
	query := fmt.Sprintf(`
		SELECT access_level FROM %s 
		WHERE object_api_name = ? AND record_id = ? AND user_id = ? AND is_deleted = 0
	`, constants.TableTeamMember)

	var accessLevel string
	err := ps.db.DB().QueryRow(query, objectAPIName, recordID, user.ID).Scan(&accessLevel)
	if err != nil {
		return false
	}
	return ps.accessLevelAllowsOperation(accessLevel, operation)
}

// accessLevelAllowsOperation checks if access level permits the operation
func (ps *PermissionService) accessLevelAllowsOperation(accessLevel, operation string) bool {
	accessLevel = strings.ToLower(accessLevel)
	operation = strings.ToLower(operation)

	switch accessLevel {
	case constants.PermEdit:
		return operation == constants.PermRead || operation == constants.PermEdit
	case constants.PermRead:
		return operation == constants.PermRead
	default:
		return false
	}
}
