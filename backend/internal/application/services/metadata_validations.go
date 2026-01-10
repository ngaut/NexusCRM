package services

import (
	"context"
	"fmt"
	"strings"

	appErrors "github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Validation Rule CRUD ====================

// CreateValidationRule creates a new validation rule
func (ms *MetadataService) CreateValidationRule(ctx context.Context, rule *models.ValidationRule) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Validate
	rule.ObjectAPIName = strings.ToLower(rule.ObjectAPIName)
	if rule.ObjectAPIName == "" {
		return fmt.Errorf("object API name is required")
	}
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	if rule.Condition == "" {
		return fmt.Errorf("error condition formula is required")
	}
	if rule.ErrorMessage == "" {
		return fmt.Errorf("error message is required")
	}

	if rule.ID == "" {
		rule.ID = GenerateID()
	}

	if err := ms.repo.CreateValidationRule(ctx, rule); err != nil {
		return fmt.Errorf("failed to insert validation rule: %w", err)
	}

	// Invalidate cache
	ms.invalidateCacheLocked()
	return nil
}

// UpdateValidationRule updates an existing validation rule
func (ms *MetadataService) UpdateValidationRule(ctx context.Context, id string, updates *models.ValidationRule) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Find the rule in DB
	existingRule, err := ms.repo.GetValidationRule(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get validation rule: %w", err)
	}
	if existingRule == nil {
		return fmt.Errorf("validation rule with ID '%s' not found", id)
	}

	// Merge updates
	if updates.Name != "" {
		existingRule.Name = updates.Name
	}
	// Active is boolean
	existingRule.Active = updates.Active

	if updates.Condition != "" {
		existingRule.Condition = updates.Condition
	}
	if updates.ErrorMessage != "" {
		existingRule.ErrorMessage = updates.ErrorMessage
	}

	// Update DB
	if err := ms.repo.UpdateValidationRule(ctx, id, existingRule); err != nil {
		return fmt.Errorf("failed to update validation rule: %w", err)
	}

	// Invalidate cache
	ms.invalidateCacheLocked()
	return nil
}

// DeleteValidationRule deletes a validation rule
func (ms *MetadataService) DeleteValidationRule(ctx context.Context, id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if err := ms.repo.DeleteValidationRule(ctx, id); err != nil {
		return fmt.Errorf("failed to delete validation rule: %w", err)
	}

	// Invalidate cache
	ms.invalidateCacheLocked()
	return nil
}

// ValidateMaxMasterDetailFields checks if an object can accept another Master-Detail field
// Limit is 2 per object (Standard PaaS limit)
// Returns error if limit reached
func (ms *MetadataService) ValidateMaxMasterDetailFields(obj *models.ObjectMetadata, excludeField string) error {
	count := 0
	for _, f := range obj.Fields {
		// Build a count of EXISTING master-detail fields
		if f.IsMasterDetail {
			// If we are updating a field, don't count itself (if it was already MD)
			// But usually this logic is for "Add New" or "Change Type to MD"
			if excludeField != "" && strings.EqualFold(f.APIName, excludeField) {
				continue
			}
			count++
		}
	}

	if count >= 2 {
		return appErrors.NewValidationError("is_master_detail", fmt.Sprintf("Object '%s' already has the maximum of 2 Master-Detail relationships", obj.APIName))
	}

	return nil
}
