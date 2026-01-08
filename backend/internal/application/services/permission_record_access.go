package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Record-Level Access Checks ====================

// CheckRecordAccess checks if a user can access a specific record
// This checks record-level sharing rules and ownership
// CheckRecordAccess checks if a user can access a specific record
// This checks record-level sharing rules and ownership
func (ps *PermissionService) CheckRecordAccess(ctx context.Context, schema *models.ObjectMetadata, record models.SObject, operation string, user *models.UserSession) bool {
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
		isMember, err := ps.repo.IsUserInGroup(ctx, ownerIDStr, user.ID)
		if err == nil && isMember {
			return true
		}

	}

	// 2. Role Hierarchy Check (Read-Only access for managers)
	// If the user's role is above the record owner's role in the hierarchy,
	// grant READ access (not edit/delete)
	if operation == constants.PermRead && hasOwner && ownerIDStr != "" {
		ownerRoleID := ps.getRecordOwnerRoleID(ctx, ownerIDStr)
		if ps.isUserAboveInHierarchy(user.RoleID, ownerRoleID) {
			return true
		}
	}

	// 3. Sharing Rules Check
	// Evaluate all sharing rules for this object
	if schema != nil {
		rules := ps.metadata.GetSharingRules(ctx, schema.APIName)
		for _, rule := range rules {
			if ps.checkSharingRuleAccess(ctx, record, rule, user, operation) {
				return true
			}
		}
	}

	// 4. Manual Record Share Check (_System_RecordShare)
	if schema != nil && recordID != "" {
		if ps.checkManualShareAccess(ctx, schema.APIName, recordID, user, operation) {
			return true
		}
	}

	// 5. Team Member Check (_System_TeamMember)
	if schema != nil && recordID != "" {
		if ps.checkTeamMemberAccess(ctx, schema.APIName, recordID, user, operation) {
			return true
		}
	}

	// Default to DENY if not owner, not above in hierarchy, and no sharing rule matches
	return false
}

// checkManualShareAccess checks if user has access via manual record share
func (ps *PermissionService) checkManualShareAccess(ctx context.Context, objectAPIName, recordID string, user *models.UserSession, operation string) bool {
	// Check direct user share and group share via repository
	levels, err := ps.repo.GetManualShareAccessLevels(ctx, objectAPIName, recordID, user.ID)
	if err != nil {
		return false
	}

	for _, level := range levels {
		if ps.accessLevelAllowsOperation(level, operation) {
			return true
		}
	}
	return false
}

// checkTeamMemberAccess checks if user is a team member with access
func (ps *PermissionService) checkTeamMemberAccess(ctx context.Context, objectAPIName, recordID string, user *models.UserSession, operation string) bool {
	accessLevel, err := ps.repo.GetTeamMemberAccessLevel(ctx, objectAPIName, recordID, user.ID)
	if err != nil || accessLevel == nil {
		return false
	}
	return ps.accessLevelAllowsOperation(*accessLevel, operation)
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
