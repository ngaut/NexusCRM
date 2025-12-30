package models

import (
	"time"
)

// Flow represents a workflow/automation
type Flow struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Status           string                 `json:"status"`    // Active, Draft
	FlowType         string                 `json:"flow_type"` // simple, multistep
	Description      *string                `json:"description,omitempty"`
	TriggerObject    string                 `json:"trigger_object"`
	TriggerType      string                 `json:"trigger_type"` // beforeCreate, afterCreate, etc.
	TriggerCondition string                 `json:"trigger_condition"`
	ActionType       string                 `json:"action_type"`
	ActionConfig     map[string]interface{} `json:"action_config,omitempty"`
	LastModified     string                 `json:"last_modified"`
	Steps            []FlowStep             `json:"steps,omitempty"` // For multi-step flows
}

// FlowStep represents a step within a multi-step flow
type FlowStep struct {
	ID             string                 `json:"id"`
	FlowID         string                 `json:"flow_id"`
	StepOrder      int                    `json:"step_order"`
	StepName       string                 `json:"step_name"`
	StepType       string                 `json:"step_type"` // action, approval, decision
	ActionType     *string                `json:"action_type,omitempty"`
	ActionConfig   map[string]interface{} `json:"action_config,omitempty"`
	EntryCondition *string                `json:"entry_condition,omitempty"` // Formula - skip if false
	OnSuccessStep  *string                `json:"on_success_step,omitempty"` // Next step ID on success
	OnFailureStep  *string                `json:"on_failure_step,omitempty"` // Step ID on failure/rejection
}

// FlowInstance represents a running or paused flow execution
type FlowInstance struct {
	ID            string                 `json:"id"`
	FlowID        string                 `json:"flow_id"`
	ObjectAPIName string                 `json:"object_api_name"`
	RecordID      string                 `json:"record_id"`
	Status        string                 `json:"status"` // Running, Paused, Completed, Failed
	CurrentStepID *string                `json:"current_step_id,omitempty"`
	ContextData   map[string]interface{} `json:"context_data,omitempty"` // Variables passed between steps
	StartedDate   time.Time              `json:"started_date"`
	PausedDate    *time.Time             `json:"paused_date,omitempty"`
	CompletedDate *time.Time             `json:"completed_date,omitempty"`
	CreatedByID   *string                `json:"created_by_id,omitempty"`
}

// FlowInstance status constants
const (
	FlowInstanceStatusRunning   = "Running"
	FlowInstanceStatusPaused    = "Paused"
	FlowInstanceStatusCompleted = "Completed"
	FlowInstanceStatusFailed    = "Failed"
)

// TransformationTarget represents a transformation target
type TransformationTarget struct {
	TargetObject string            `json:"target_object"`
	Required     bool              `json:"required"`
	FieldMapping map[string]string `json:"field_mapping"`
}

// TransformationConfig represents a transformation configuration
type TransformationConfig struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	SourceObject  string                 `json:"source_object"`
	StatusField   *string                `json:"status_field,omitempty"`
	TriggerStatus *string                `json:"trigger_status,omitempty"`
	TargetStatus  *string                `json:"target_status,omitempty"`
	Targets       []TransformationTarget `json:"targets"`
}
