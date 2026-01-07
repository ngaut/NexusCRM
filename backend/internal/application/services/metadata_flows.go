package services

import (
	"context"
	"fmt"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Flow CRUD Methods ====================

// GetFlow returns a flow by its ID
func (ms *MetadataService) GetFlow(flowID string) *models.Flow {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	flow, err := ms.repo.GetFlow(context.Background(), flowID)
	if err != nil {
		return nil
	}
	return flow
}

// CreateFlow creates a new flow
func (ms *MetadataService) CreateFlow(flow *models.Flow) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Generate ID if not provided
	if flow.ID == "" {
		flow.ID = GenerateID()
	}

	// Set defaults
	if flow.Status == "" {
		flow.Status = constants.FlowStatusDraft
	}
	flow.LastModified = NowTimestamp()

	// Validate Flow
	if ms.validationSvc != nil {
		existingFlows, err := ms.repo.GetFlowsByObject(context.Background(), flow.TriggerObject)
		if err != nil {
			return fmt.Errorf("failed to query duplicate flows: %w", err)
		}
		if err := ms.validationSvc.ValidateFlow(flow, existingFlows); err != nil {
			return err
		}
	}

	// Insert into database via Repo
	if err := ms.repo.CreateFlow(context.Background(), flow); err != nil {
		return fmt.Errorf("failed to create flow: %w", err)
	}

	// Create steps
	if len(flow.Steps) > 0 {
		return ms.repo.SaveFlowSteps(context.Background(), flow.ID, flow.Steps)
	}

	return nil
}

// UpdateFlow updates an existing flow
func (ms *MetadataService) UpdateFlow(flowID string, updates *models.Flow) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Find existing flow
	existingFlow, err := ms.repo.GetFlow(context.Background(), flowID)
	if err != nil || existingFlow == nil {
		return fmt.Errorf("flow with ID '%s' not found", flowID)
	}

	// Apply updates
	if updates.Name != "" {
		existingFlow.Name = updates.Name
	}
	if updates.TriggerObject != "" {
		existingFlow.TriggerObject = updates.TriggerObject
	}
	if updates.TriggerType != "" {
		existingFlow.TriggerType = updates.TriggerType
	}
	if updates.TriggerCondition != "" {
		existingFlow.TriggerCondition = updates.TriggerCondition
	}
	if updates.ActionType != "" {
		existingFlow.ActionType = updates.ActionType
	}
	if updates.ActionConfig != nil {
		existingFlow.ActionConfig = updates.ActionConfig
	}
	if updates.Status != "" {
		existingFlow.Status = updates.Status
	}
	if updates.FlowType != "" {
		existingFlow.FlowType = updates.FlowType
	}
	existingFlow.LastModified = NowTimestamp()

	// Validate Update
	if ms.validationSvc != nil {
		// For duplicates, we need to check against OTHER flows.
		// Since we are updating, we need to make sure the modified flow doesn't clash.
		// We pass the "updates applied" flow to validator?
		// existingFlow already has updates applied in memory above.
		existingFlows, err := ms.repo.GetFlowsByObject(context.Background(), existingFlow.TriggerObject)
		if err != nil {
			return fmt.Errorf("failed to query duplicate flows: %w", err)
		}
		if err := ms.validationSvc.ValidateFlow(existingFlow, existingFlows); err != nil {
			return err
		}
	}

	// Update database via Repo
	if err := ms.repo.UpdateFlow(context.Background(), flowID, existingFlow); err != nil {
		return fmt.Errorf("failed to update flow: %w", err)
	}

	// Update steps (Delete and Re-create)
	if existingFlow.FlowType == constants.FlowTypeMultistep || updates.FlowType == constants.FlowTypeMultistep || len(updates.Steps) > 0 {
		if err := ms.repo.DeleteFlowSteps(context.Background(), flowID); err != nil {
			return err
		}
		if len(updates.Steps) > 0 {
			if err := ms.repo.SaveFlowSteps(context.Background(), flowID, updates.Steps); err != nil {
				return err
			}
		}
	} else if len(updates.Steps) > 0 {
		// If steps provided in update, save them
		if err := ms.repo.DeleteFlowSteps(context.Background(), flowID); err != nil {
			return err
		}
		if err := ms.repo.SaveFlowSteps(context.Background(), flowID, updates.Steps); err != nil {
			return err
		}
	}

	return nil
}

// DeleteFlow deletes a flow
func (ms *MetadataService) DeleteFlow(flowID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Check existence
	existing, err := ms.repo.GetFlow(context.Background(), flowID)
	if err != nil || existing == nil {
		return fmt.Errorf("flow with ID '%s' not found", flowID)
	}

	// Delete from database via Repo
	return ms.repo.DeleteFlow(context.Background(), flowID)
}
