package services

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/pkg/constants"
	appErrors "github.com/nexuscrm/backend/pkg/errors"
)

// ==================== Validation Rule CRUD ====================

// CreateValidationRule creates a new validation rule
func (ms *MetadataService) CreateValidationRule(rule *models.ValidationRule) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Validate
	if rule.ObjectAPIName == "" || rule.Name == "" || rule.Condition == "" || rule.ErrorMessage == "" {
		return fmt.Errorf("objectApiName, name, condition, and errorMessage are required")
	}

	if rule.ID == "" {
		rule.ID = GenerateID()
	}

	// Insert into DB. Note: `condition` is a reserved keyword in some SQL dialects, best to backtick or avoid.
	query := fmt.Sprintf("INSERT INTO %s (id, object_api_name, name, active, `condition`, error_message) VALUES (?, ?, ?, ?, ?, ?)", constants.TableValidation)
	_, err := ms.db.Exec(query, rule.ID, rule.ObjectAPIName, rule.Name, rule.Active, rule.Condition, rule.ErrorMessage)
	if err != nil {
		return fmt.Errorf("failed to insert validation rule: %w", err)
	}

	return nil
}

// UpdateValidationRule updates an existing validation rule
func (ms *MetadataService) UpdateValidationRule(id string, updates *models.ValidationRule) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Find the rule in DB
	queryFetch := fmt.Sprintf("SELECT id, object_api_name, name, active, condition, error_message FROM %s WHERE id = ?", constants.TableValidation)

	existingRulePtr, err := ms.scanValidationRule(ms.db.QueryRow(queryFetch, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("validation rule with ID '%s' not found: %w", id, err)
		}
		return fmt.Errorf("failed to scan validation rule: %w", err)
	}
	existingRule := *existingRulePtr

	// Merge updates
	if updates.Name != "" {
		existingRule.Name = updates.Name
	}
	// Active is boolean, so we always update it if provided? Use pointer or assume true/false is intent?
	// Existing implementation: existingRule.Active = updates.Active.
	// This means if I pass false, it becomes false. If I don't pass it (default false), it becomes false.
	// This might be unintended if I just want to update Name.
	// But sticking to previous logic:
	existingRule.Active = updates.Active

	if updates.Condition != "" {
		existingRule.Condition = updates.Condition
	}
	if updates.ErrorMessage != "" {
		existingRule.ErrorMessage = updates.ErrorMessage
	}

	// Update DB
	queryUpdate := fmt.Sprintf("UPDATE %s SET name = ?, active = ?, `condition` = ?, error_message = ? WHERE id = ?", constants.TableValidation)
	_, err = ms.db.Exec(queryUpdate, existingRule.Name, existingRule.Active, existingRule.Condition, existingRule.ErrorMessage, id)
	if err != nil {
		return fmt.Errorf("failed to update validation rule: %w", err)
	}

	return nil
}

// DeleteValidationRule deletes a validation rule
func (ms *MetadataService) DeleteValidationRule(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.deleteMetadataRecord(constants.TableValidation, id, "validation rule")
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
