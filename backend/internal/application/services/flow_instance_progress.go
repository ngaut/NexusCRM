package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// FlowInstanceProgress contains instance data with computed step progress
type FlowInstanceProgress struct {
	ID               string             `json:"id"`
	FlowID           string             `json:"flow_id"`
	Status           string             `json:"status"`
	CurrentStepID    *string            `json:"current_step_id,omitempty"`
	CurrentStepOrder int                `json:"current_step_order"`
	TotalSteps       int                `json:"total_steps"`
	Steps            []FlowStepProgress `json:"steps"`
}

// FlowStepProgress contains step data with status for UI display
type FlowStepProgress struct {
	ID        string `json:"id"`
	StepOrder int    `json:"step_order"`
	StepName  string `json:"step_name"`
	StepType  string `json:"step_type"`
	Status    string `json:"status"` // pending, completed, current, skipped
}

// GetInstanceProgress returns flow instance with computed step progress for UI
func (s *FlowInstanceService) GetInstanceProgress(ctx context.Context, instanceID string, user *models.UserSession) (*FlowInstanceProgress, error) {
	// Get the instance
	instance, err := s.GetInstance(ctx, instanceID, user)
	if err != nil || instance == nil {
		return nil, err
	}

	// Get all steps for this flow
	steps, err := s.GetFlowSteps(ctx, instance.FlowID, user)
	if err != nil {
		return nil, err
	}

	// Build progress response
	progress := &FlowInstanceProgress{
		ID:            instance.ID,
		FlowID:        instance.FlowID,
		Status:        instance.Status,
		CurrentStepID: instance.CurrentStepID,
		TotalSteps:    len(steps),
		Steps:         make([]FlowStepProgress, 0, len(steps)),
	}

	// Determine current step order
	currentStepOrder := 0
	if instance.CurrentStepID != nil {
		for _, step := range steps {
			if step.ID == *instance.CurrentStepID {
				currentStepOrder = step.StepOrder
				break
			}
		}
	}
	progress.CurrentStepOrder = currentStepOrder

	// Build step progress list with status
	for _, step := range steps {
		status := constants.ProgressStatusPending
		if currentStepOrder > 0 {
			if step.StepOrder < currentStepOrder {
				status = constants.ProgressStatusCompleted
			} else if step.StepOrder == currentStepOrder {
				status = constants.ProgressStatusCurrent
			}
		}
		// If flow is completed, all steps are completed
		if instance.Status == models.FlowInstanceStatusCompleted {
			status = constants.ProgressStatusCompleted
		} else if instance.Status == models.FlowInstanceStatusFailed {
			if step.StepOrder == currentStepOrder {
				status = constants.ProgressStatusSkipped
			}
		}

		progress.Steps = append(progress.Steps, FlowStepProgress{
			ID:        step.ID,
			StepOrder: step.StepOrder,
			StepName:  step.StepName,
			StepType:  step.StepType,
			Status:    status,
		})
	}

	return progress, nil
}

// Helper: Convert SObject to FlowInstance
func (s *FlowInstanceService) sobjectToFlowInstance(data models.SObject) (*models.FlowInstance, error) {
	instance := &models.FlowInstance{
		ID:            data.GetString(constants.FieldID),
		FlowID:        data.GetString(constants.FieldSysFlowInstance_FlowID),
		ObjectAPIName: data.GetString(constants.FieldSysFlowInstance_ObjectAPIName),
		RecordID:      data.GetString(constants.FieldSysFlowInstance_RecordID),
		Status:        data.GetString(constants.FieldSysFlowInstance_Status),
	}

	if stepID, ok := data[constants.FieldSysFlowInstance_CurrentStepID].(string); ok && stepID != "" {
		instance.CurrentStepID = &stepID
	}

	if createdByID, ok := data[constants.FieldCreatedByID].(string); ok && createdByID != "" {
		instance.CreatedByID = &createdByID
	}

	// Parse context_data JSON
	if contextStr, ok := data[constants.FieldSysFlowInstance_ContextData].(string); ok && contextStr != "" {
		var contextData map[string]interface{}
		if err := json.Unmarshal([]byte(contextStr), &contextData); err == nil {
			instance.ContextData = contextData
		}
	}

	// Parse dates
	if started, ok := data[constants.FieldSysFlowInstance_StartedDate].(time.Time); ok {
		instance.StartedDate = started
	}
	if paused, ok := data[constants.FieldSysFlowInstance_PausedDate].(time.Time); ok {
		instance.PausedDate = &paused
	}
	if completed, ok := data[constants.FieldSysFlowInstance_CompletedDate].(time.Time); ok {
		instance.CompletedDate = &completed
	}

	return instance, nil
}

// Helper: Convert SObject to FlowStep
func (s *FlowInstanceService) sobjectToFlowStep(data models.SObject) (*models.FlowStep, error) {
	step := &models.FlowStep{
		ID:       data.GetString(constants.FieldID),
		FlowID:   data.GetString(constants.FieldSysFlowStep_FlowID),
		StepName: data.GetString(constants.FieldSysFlowStep_StepName),
		StepType: data.GetString(constants.FieldSysFlowStep_StepType),
	}

	// Handle step_order - database may return int, int64, or float64
	if order, ok := data[constants.FieldSysFlowStep_StepOrder].(float64); ok {
		step.StepOrder = int(order)
	} else if order, ok := data[constants.FieldSysFlowStep_StepOrder].(int64); ok {
		step.StepOrder = int(order)
	} else if order, ok := data[constants.FieldSysFlowStep_StepOrder].(int); ok {
		step.StepOrder = order
	}

	// Optional fields
	if actionType, ok := data[constants.FieldSysFlowStep_ActionType].(string); ok && actionType != "" {
		step.ActionType = &actionType
	}
	if entryCondition, ok := data[constants.FieldSysFlowStep_EntryCondition].(string); ok && entryCondition != "" {
		step.EntryCondition = &entryCondition
	}
	if onSuccess, ok := data[constants.FieldSysFlowStep_OnSuccessStep].(string); ok && onSuccess != "" {
		step.OnSuccessStep = &onSuccess
	}
	if onFailure, ok := data[constants.FieldSysFlowStep_OnFailureStep].(string); ok && onFailure != "" {
		step.OnFailureStep = &onFailure
	}

	// Parse action_config JSON
	if configStr, ok := data[constants.FieldSysFlowStep_ActionConfig].(string); ok && configStr != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(configStr), &config); err == nil {
			step.ActionConfig = config
		}
	} else if config, ok := data[constants.FieldSysFlowStep_ActionConfig].(map[string]interface{}); ok {
		step.ActionConfig = config
	}

	return step, nil
}
