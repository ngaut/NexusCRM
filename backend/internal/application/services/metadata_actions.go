package services

import (
	"database/sql"
	"fmt"

	"github.com/nexuscrm/shared/pkg/models"
	"github.com/nexuscrm/shared/pkg/constants"
)

// ==================== Action CRUD ====================

// CreateAction creates a new action
func (ms *MetadataService) CreateAction(action *models.ActionMetadata) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Validate required fields
	if action.ID == "" {
		action.ID = GenerateID()
	}
	if action.Name == "" || action.Label == "" {
		return fmt.Errorf("action name and label are required")
	}
	if action.Type == "" {
		return fmt.Errorf("action type is required")
	}

	// Check for duplicate ID
	existing, _ := ms.queryAction(action.ID)
	if existing != nil {
		return fmt.Errorf("action with ID '%s' already exists", action.ID)
	}

	// Check for duplicate (object_api_name, name) - this is the unique constraint
	var existingCount int
	checkQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE object_api_name = ? AND name = ?", constants.TableAction)
	if err := ms.db.QueryRow(checkQuery, action.ObjectAPIName, action.Name).Scan(&existingCount); err != nil {
		return fmt.Errorf("failed to check for existing action: %w", err)
	}
	if existingCount > 0 {
		return fmt.Errorf("action '%s' already exists for object '%s'", action.Name, action.ObjectAPIName)
	}

	// Serialize config to JSON
	configJSON, err := MarshalJSONOrDefault(action.Config, "{}")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// Insert into database
	var targetObject sql.NullString
	if action.TargetObject != nil {
		targetObject = ToNullString(action.TargetObject)
	}

	query := fmt.Sprintf("INSERT INTO %s (id, object_api_name, name, label, type, icon, target_object, config) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", constants.TableAction)
	_, err = ms.db.Exec(query, action.ID, action.ObjectAPIName, action.Name, action.Label,
		action.Type, action.Icon, targetObject, configJSON)
	if err != nil {
		return fmt.Errorf("failed to insert action: %w", err)
	}

	return nil
}

// UpdateAction updates an existing action
func (ms *MetadataService) UpdateAction(actionID string, updates *models.ActionMetadata) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Find existing action
	existing, err := ms.queryAction(actionID)
	if err != nil || existing == nil {
		return fmt.Errorf("action with ID '%s' not found", actionID)
	}

	// Preserve ID
	updates.ID = actionID

	// Merge updates. We treat `updates` as the source of truth for modifyable fields.
	// ID is preserved from argument. Config is re-serialized.

	// Serialize config to JSON
	configJSON, err := MarshalJSONOrDefault(updates.Config, "{}")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// Update in database
	var targetObject sql.NullString
	if updates.TargetObject != nil {
		targetObject = ToNullString(updates.TargetObject)
	}

	query := fmt.Sprintf(`UPDATE %s SET object_api_name=?, name=?, label=?, type=?, icon=?, target_object=?, config=? WHERE id=?`, constants.TableAction)
	_, err = ms.db.Exec(query, updates.ObjectAPIName, updates.Name, updates.Label,
		updates.Type, updates.Icon, targetObject, configJSON, actionID)
	if err != nil {
		return fmt.Errorf("failed to update action: %w", err)
	}

	return nil
}

// DeleteAction deletes an action
func (ms *MetadataService) DeleteAction(actionID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.deleteMetadataRecord(constants.TableAction, actionID, "action")
}
