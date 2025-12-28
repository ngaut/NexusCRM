package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nexuscrm/backend/internal/domain/models"
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
		status := "pending"
		if currentStepOrder > 0 {
			if step.StepOrder < currentStepOrder {
				status = "completed"
			} else if step.StepOrder == currentStepOrder {
				status = "current"
			}
		}
		// If flow is completed, all steps are completed
		if instance.Status == models.FlowInstanceStatusCompleted {
			status = "completed"
		} else if instance.Status == models.FlowInstanceStatusFailed {
			if step.StepOrder == currentStepOrder {
				status = "skipped"
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
		ID:            data.GetString("id"),
		FlowID:        data.GetString("flow_id"),
		ObjectAPIName: data.GetString("object_api_name"),
		RecordID:      data.GetString("record_id"),
		Status:        data.GetString("status"),
	}

	if stepID, ok := data["current_step_id"].(string); ok && stepID != "" {
		instance.CurrentStepID = &stepID
	}

	if createdByID, ok := data["created_by_id"].(string); ok && createdByID != "" {
		instance.CreatedByID = &createdByID
	}

	// Parse context_data JSON
	if contextStr, ok := data["context_data"].(string); ok && contextStr != "" {
		var contextData map[string]interface{}
		if err := json.Unmarshal([]byte(contextStr), &contextData); err == nil {
			instance.ContextData = contextData
		}
	}

	// Parse dates
	if started, ok := data["started_date"].(time.Time); ok {
		instance.StartedDate = started
	}
	if paused, ok := data["paused_date"].(time.Time); ok {
		instance.PausedDate = &paused
	}
	if completed, ok := data["completed_date"].(time.Time); ok {
		instance.CompletedDate = &completed
	}

	return instance, nil
}

// Helper: Convert SObject to FlowStep
func (s *FlowInstanceService) sobjectToFlowStep(data models.SObject) (*models.FlowStep, error) {
	step := &models.FlowStep{
		ID:       data.GetString("id"),
		FlowID:   data.GetString("flow_id"),
		StepName: data.GetString("step_name"),
		StepType: data.GetString("step_type"),
	}

	// Handle step_order - database may return int, int64, or float64
	if order, ok := data["step_order"].(float64); ok {
		step.StepOrder = int(order)
	} else if order, ok := data["step_order"].(int64); ok {
		step.StepOrder = int(order)
	} else if order, ok := data["step_order"].(int); ok {
		step.StepOrder = order
	}

	// Optional fields
	if actionType, ok := data["action_type"].(string); ok && actionType != "" {
		step.ActionType = &actionType
	}
	if entryCondition, ok := data["entry_condition"].(string); ok && entryCondition != "" {
		step.EntryCondition = &entryCondition
	}
	if onSuccess, ok := data["on_success_step"].(string); ok && onSuccess != "" {
		step.OnSuccessStep = &onSuccess
	}
	if onFailure, ok := data["on_failure_step"].(string); ok && onFailure != "" {
		step.OnFailureStep = &onFailure
	}

	// Parse action_config JSON
	if configStr, ok := data["action_config"].(string); ok && configStr != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(configStr), &config); err == nil {
			step.ActionConfig = config
		}
	} else if config, ok := data["action_config"].(map[string]interface{}); ok {
		step.ActionConfig = config
	}

	return step, nil
}
