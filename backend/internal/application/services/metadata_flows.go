package services

import (
	"fmt"

	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/pkg/constants"
)

// ==================== Flow CRUD Methods ====================

// GetFlow returns a flow by its ID
func (ms *MetadataService) GetFlow(flowID string) *models.Flow {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	flow, err := ms.queryFlow(flowID)
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
		existingFlows, err := ms.queryFlowsByObject(flow.TriggerObject)
		if err != nil {
			return fmt.Errorf("failed to query duplicate flows: %w", err)
		}
		if err := ms.validationSvc.ValidateFlow(flow, existingFlows); err != nil {
			return err
		}
	}

	// Serialize actionConfig
	actionConfigJSON, err := MarshalJSONOrDefault(flow.ActionConfig, "{}")
	if err != nil {
		return fmt.Errorf("failed to set default action config: %w", err)
	}

	// Insert into database
	query := fmt.Sprintf(`INSERT INTO %s (id, name, trigger_object, trigger_type, trigger_condition, action_type, action_config, status, flow_type, last_modified_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, constants.TableFlow)
	_, err = ms.db.Exec(query, flow.ID, flow.Name, flow.TriggerObject, flow.TriggerType, flow.TriggerCondition,
		flow.ActionType, actionConfigJSON, flow.Status, flow.FlowType, flow.LastModified)
	if err != nil {
		return fmt.Errorf("failed to create flow: %w", err)
	}

	// Create steps
	if len(flow.Steps) > 0 {
		return ms.saveFlowSteps(flow.ID, flow.Steps)
	}

	return nil
}

// UpdateFlow updates an existing flow
func (ms *MetadataService) UpdateFlow(flowID string, updates *models.Flow) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Find existing flow
	existingFlow, err := ms.queryFlow(flowID)
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
		existingFlows, err := ms.queryFlowsByObject(existingFlow.TriggerObject)
		if err != nil {
			return fmt.Errorf("failed to query duplicate flows: %w", err)
		}
		if err := ms.validationSvc.ValidateFlow(existingFlow, existingFlows); err != nil {
			return err
		}
	}

	// Serialize actionConfig
	actionConfigJSON, err := MarshalJSONOrDefault(existingFlow.ActionConfig, "{}")
	if err != nil {
		return fmt.Errorf("failed to serialize action config: %w", err)
	}

	// Update database
	query := fmt.Sprintf(`UPDATE %s SET name = ?, trigger_object = ?, trigger_type = ?, trigger_condition = ?, 
		action_type = ?, action_config = ?, status = ?, flow_type = ?, last_modified_date = ? WHERE id = ?`, constants.TableFlow)
	_, err = ms.db.Exec(query, existingFlow.Name, existingFlow.TriggerObject, existingFlow.TriggerType, existingFlow.TriggerCondition,
		existingFlow.ActionType, actionConfigJSON, existingFlow.Status, existingFlow.FlowType, existingFlow.LastModified, flowID)
	if err != nil {
		return fmt.Errorf("failed to update flow: %w", err)
	}

	// Update steps (Delete and Re-create)
	if existingFlow.FlowType == constants.FlowTypeMultistep || updates.FlowType == constants.FlowTypeMultistep || len(updates.Steps) > 0 {
		if err := ms.deleteFlowSteps(flowID); err != nil {
			return err
		}
		if len(updates.Steps) > 0 {
			if err := ms.saveFlowSteps(flowID, updates.Steps); err != nil {
				return err
			}
		}
	} else if len(updates.Steps) > 0 {
		// If steps provided in update, save them
		if err := ms.deleteFlowSteps(flowID); err != nil {
			return err
		}
		if err := ms.saveFlowSteps(flowID, updates.Steps); err != nil {
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
	existing, err := ms.queryFlow(flowID)
	if err != nil || existing == nil {
		return fmt.Errorf("flow with ID '%s' not found", flowID)
	}

	// Delete from database
	_, err = ms.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", constants.TableFlow), flowID)
	if err != nil {
		return fmt.Errorf("failed to delete flow: %w", err)
	}

	return nil
}

// saveFlowSteps saves flow steps to database
func (ms *MetadataService) saveFlowSteps(flowID string, steps []models.FlowStep) error {
	query := fmt.Sprintf(`INSERT INTO %s (id, flow_id, step_name, step_type, step_order, 
		action_type, action_config, entry_condition, on_success_step, on_failure_step)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, constants.TableFlowStep)

	for _, step := range steps {
		if step.ID == "" {
			step.ID = GenerateID()
		}
		step.FlowID = flowID

		actionConfigJSON, err := MarshalJSONOrDefault(step.ActionConfig, "{}")
		if err != nil {
			return fmt.Errorf("failed to serialize step action config: %w", err)
		}

		_, err = ms.db.Exec(query, step.ID, step.FlowID, step.StepName, step.StepType, step.StepOrder,
			step.ActionType, actionConfigJSON, step.EntryCondition, step.OnSuccessStep, step.OnFailureStep)
		if err != nil {
			return fmt.Errorf("failed to create flow step: %w", err)
		}
	}
	return nil
}

// deleteFlowSteps deletes all steps for a flow
func (ms *MetadataService) deleteFlowSteps(flowID string) error {
	_, err := ms.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE flow_id = ?", constants.TableFlowStep), flowID)
	if err != nil {
		return fmt.Errorf("failed to delete existing flow steps: %w", err)
	}
	return nil
}

// queryFlowsByObject returns all flows for a specific object (lightweight)
func (ms *MetadataService) queryFlowsByObject(objectName string) ([]*models.Flow, error) {
	query := fmt.Sprintf("SELECT id, trigger_object, trigger_type, status FROM %s WHERE trigger_object = ?", constants.TableFlow)
	rows, err := ms.db.Query(query, objectName)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var flows []*models.Flow
	for rows.Next() {
		var f models.Flow
		if err := rows.Scan(&f.ID, &f.TriggerObject, &f.TriggerType, &f.Status); err != nil {
			return nil, err
		}
		flows = append(flows, &f)
	}
	return flows, nil
}
