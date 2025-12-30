package services

import (
	"fmt"
	"log"
	"strings"

	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Sharing Rules ====================

// isUserInRoleOrBelow checks if the user's role matches the target role
// or is a child (subordinate) of the target role.
// This enables hierarchical sharing: if you share with "Sales Manager",
// all users in "Sales Rep" role (below Sales Manager) also get access.
func (ps *PermissionService) isUserInRoleOrBelow(userRoleID, targetRoleID *string) bool {
	if userRoleID == nil || targetRoleID == nil {
		return false
	}

	// Exact match
	if *userRoleID == *targetRoleID {
		return true
	}

	// Check if target role is an ancestor of user's role (user is below target)
	ancestors := ps.getRoleAncestors(*userRoleID)
	for _, ancestorID := range ancestors {
		if ancestorID == *targetRoleID {
			return true
		}
	}

	return false
}

// checkSharingRuleAccess evaluates if a sharing rule grants access to a record
func (ps *PermissionService) checkSharingRuleAccess(
	record models.SObject,
	rule *models.SharingRule,
	user *models.UserSession,
	operation string,
) bool {
	matchesIdentity := false

	// 1. Check if user belongs to the target role/group
	// A. Role-based sharing
	if rule.ShareWithRoleID != nil {
		if ps.isUserInRoleOrBelow(user.RoleID, rule.ShareWithRoleID) {
			matchesIdentity = true
		}
	}

	// B. Group-based sharing
	if !matchesIdentity && rule.ShareWithGroupID != nil {
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE group_id = ? AND user_id = ?", constants.TableGroupMember)
		var count int
		if err := ps.db.DB().QueryRow(query, *rule.ShareWithGroupID, user.ID).Scan(&count); err == nil && count > 0 {
			matchesIdentity = true
		}
	}

	if !matchesIdentity {
		return false
	}

	// 2. Check if operation is allowed by access level
	// AccessLevel: "Read" or "Edit"
	// "Edit" includes read, so Edit grants both read and edit
	switch strings.ToLower(rule.AccessLevel) {
	case constants.PermRead:
		if operation != constants.PermRead {
			return false
		}
	case constants.PermEdit:
		if operation != constants.PermRead && operation != constants.PermEdit {
			return false
		}
	default:
		return false
	}

	// 3. Evaluate criteria (if any)
	if rule.Criteria == "" || rule.Criteria == "[]" {
		// No criteria = share all records with this role
		return true
	}

	// Parse and evaluate criteria
	return ps.evaluateSharingCriteria(record, rule.Criteria)
}

// evaluateSharingCriteria evaluates a formula expression against a record.
// Examples: state == "CA", amount > 1000, priority == "High" || region == "West"
func (ps *PermissionService) evaluateSharingCriteria(record models.SObject, criteria string) bool {
	// Empty criteria = match all records
	if criteria == "" || criteria == "[]" {
		return true
	}

	if ps.formula == nil {
		log.Printf("Warning: Cannot evaluate formula criteria - formula engine not initialized")
		return false
	}

	// Build formula context from record
	ctx := &formula.Context{
		Record: record,
	}

	result, err := ps.formula.Evaluate(criteria, ctx)
	if err != nil {
		log.Printf("Warning: Failed to evaluate sharing rule criteria: %v", err)
		return false
	}

	// Result should be a boolean
	if boolResult, ok := result.(bool); ok {
		return boolResult
	}

	log.Printf("Warning: Sharing rule criteria did not evaluate to boolean: %v", result)
	return false
}
