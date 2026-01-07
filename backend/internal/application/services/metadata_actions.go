package services

import (
	"context"
	"fmt"

	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Action CRUD ====================

// CreateAction creates a new action
func (ms *MetadataService) CreateAction(ctx context.Context, action *models.ActionMetadata) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Validate required fields
	if action.ID == "" {
		action.ID = GenerateID()
	}
	if action.Name == "" {
		return fmt.Errorf("action name is required")
	}
	if action.Label == "" {
		return fmt.Errorf("action label is required")
	}
	if action.Type == "" {
		return fmt.Errorf("action type is required")
	}

	// Check for duplicate ID
	existing, _ := ms.repo.GetAction(ctx, action.ID)
	if existing != nil {
		return fmt.Errorf("action with ID '%s' already exists", action.ID)
	}

	// Check for duplicate (object_api_name, name) - this is the unique constraint
	exists, err := ms.repo.CheckActionExists(ctx, action.ObjectAPIName, action.Name)
	if err != nil {
		return fmt.Errorf("failed to check for existing action: %w", err)
	}
	if exists {
		return fmt.Errorf("action '%s' already exists for object '%s'", action.Name, action.ObjectAPIName)
	}

	// Insert into database via Repo
	if err := ms.repo.CreateAction(ctx, action); err != nil {
		return fmt.Errorf("failed to insert action: %w", err)
	}

	return nil
}

// UpdateAction updates an existing action
func (ms *MetadataService) UpdateAction(ctx context.Context, actionID string, updates *models.ActionMetadata) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Find existing action
	existing, err := ms.repo.GetAction(ctx, actionID)
	if err != nil || existing == nil {
		return fmt.Errorf("action with ID '%s' not found", actionID)
	}

	// Preserve ID
	updates.ID = actionID

	// Update in database via Repo
	if err := ms.repo.UpdateAction(ctx, actionID, updates); err != nil {
		return fmt.Errorf("failed to update action: %w", err)
	}

	return nil
}

// DeleteAction deletes an action
func (ms *MetadataService) DeleteAction(ctx context.Context, actionID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.repo.DeleteAction(ctx, actionID)
}
