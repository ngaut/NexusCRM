package services

import (
	"context"
	"fmt"

	"github.com/nexuscrm/shared/pkg/models"
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
	if action.Name == "" {
		return fmt.Errorf("Action Name is required")
	}
	if action.Label == "" {
		return fmt.Errorf("Action Label is required")
	}
	if action.Type == "" {
		return fmt.Errorf("Action Type is required")
	}

	// Check for duplicate ID
	existing, _ := ms.repo.GetAction(context.Background(), action.ID)
	if existing != nil {
		return fmt.Errorf("action with ID '%s' already exists", action.ID)
	}

	// Check for duplicate (object_api_name, name) - this is the unique constraint
	exists, err := ms.repo.CheckActionExists(context.Background(), action.ObjectAPIName, action.Name)
	if err != nil {
		return fmt.Errorf("failed to check for existing action: %w", err)
	}
	if exists {
		return fmt.Errorf("action '%s' already exists for object '%s'", action.Name, action.ObjectAPIName)
	}

	// Insert into database via Repo
	if err := ms.repo.CreateAction(context.Background(), action); err != nil {
		return fmt.Errorf("failed to insert action: %w", err)
	}

	return nil
}

// UpdateAction updates an existing action
func (ms *MetadataService) UpdateAction(actionID string, updates *models.ActionMetadata) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Find existing action
	existing, err := ms.repo.GetAction(context.Background(), actionID)
	if err != nil || existing == nil {
		return fmt.Errorf("action with ID '%s' not found", actionID)
	}

	// Preserve ID
	updates.ID = actionID

	// Update in database via Repo
	if err := ms.repo.UpdateAction(context.Background(), actionID, updates); err != nil {
		return fmt.Errorf("failed to update action: %w", err)
	}

	return nil
}

// DeleteAction deletes an action
func (ms *MetadataService) DeleteAction(actionID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.repo.DeleteAction(context.Background(), actionID)
}
