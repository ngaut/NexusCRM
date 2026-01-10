package services

import (
	"context"
	"fmt"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Flow CRUD Methods ====================

// GetFlow returns a flow by its ID
func (ms *MetadataService) GetFlow(ctx context.Context, flowID string) *models.Flow {
	if err := ms.ensureCacheInitialized(); err != nil {
		return nil
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if ms.flowMap == nil {
		return nil
	}
	return ms.flowMap[flowID]
}

// CreateFlow creates a new flow
func (ms *MetadataService) CreateFlow(ctx context.Context, flow *models.Flow) error {
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
		existingFlows, err := ms.repo.GetFlowsByObject(ctx, flow.TriggerObject)
		if err != nil {
			return fmt.Errorf("failed to query duplicate flows: %w", err)
		}
		if err := ms.validationSvc.ValidateFlow(flow, existingFlows); err != nil {
			return err
		}
	}

	// Insert into database via Repo
	if err := ms.repo.CreateFlow(ctx, flow); err != nil {
		return fmt.Errorf("failed to create flow: %w", err)
	}

	// Create steps
	if len(flow.Steps) > 0 {
		if err := ms.repo.SaveFlowSteps(ctx, flow.ID, flow.Steps); err != nil {
			return err
		}
	}

	// Invalidate cache to include new flow
	ms.invalidateCacheLocked()
	return nil
}

// UpdateFlow updates an existing flow
func (ms *MetadataService) UpdateFlow(ctx context.Context, flowID string, updates *models.Flow) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Find existing flow
	existingFlow, err := ms.repo.GetFlow(ctx, flowID)
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
		existingFlows, err := ms.repo.GetFlowsByObject(ctx, existingFlow.TriggerObject)
		if err != nil {
			return fmt.Errorf("failed to query duplicate flows: %w", err)
		}
		if err := ms.validationSvc.ValidateFlow(existingFlow, existingFlows); err != nil {
			return err
		}
	}

	// Update database via Repo
	if err := ms.repo.UpdateFlow(ctx, flowID, existingFlow); err != nil {
		return fmt.Errorf("failed to update flow: %w", err)
	}

	// Update steps (Delete and Re-create)
	if existingFlow.FlowType == constants.FlowTypeMultistep || updates.FlowType == constants.FlowTypeMultistep || len(updates.Steps) > 0 {
		if err := ms.repo.DeleteFlowSteps(ctx, flowID); err != nil {
			return err
		}
		if len(updates.Steps) > 0 {
			if err := ms.repo.SaveFlowSteps(ctx, flowID, updates.Steps); err != nil {
				return err
			}
		}
	} else if len(updates.Steps) > 0 {
		// If steps provided in update, save them
		if err := ms.repo.DeleteFlowSteps(ctx, flowID); err != nil {
			return err
		}
		if err := ms.repo.SaveFlowSteps(ctx, flowID, updates.Steps); err != nil {
			return err
		}
	}

	// Invalidate cache to reflect updated flow
	ms.invalidateCacheLocked()
	return nil
}

// DeleteFlow deletes a flow
func (ms *MetadataService) DeleteFlow(ctx context.Context, flowID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Check existence
	existing, err := ms.repo.GetFlow(ctx, flowID)
	if err != nil || existing == nil {
		return fmt.Errorf("flow with ID '%s' not found", flowID)
	}

	// Delete from database via Repo
	if err := ms.repo.DeleteFlow(ctx, flowID); err != nil {
		return err
	}

	// Invalidate cache
	ms.invalidateCacheLocked()
	return nil
}
