package ports

import (
	"context"

	"github.com/nexuscrm/backend/internal/domain/models"
)

// FlowInstanceManager manages flow instance lifecycle.
// This interface enables FlowExecutor to use FlowInstanceService without direct dependency.
type FlowInstanceManager interface {
	// CreateInstance creates a new flow instance when a multi-step flow starts
	CreateInstance(ctx context.Context, flow *models.Flow, objectAPIName, recordID string, user *models.UserSession) (*models.FlowInstance, error)

	// PauseInstance pauses a flow instance when it hits an approval step
	PauseInstance(ctx context.Context, instanceID, currentStepID string, user *models.UserSession) error
	ResumeInstance(ctx context.Context, instanceID, nextStepID string, user *models.UserSession) error
	CompleteInstance(ctx context.Context, instanceID string, user *models.UserSession) error
	FailInstance(ctx context.Context, instanceID, reason string, user *models.UserSession) error

	// GetFlowSteps retrieves all steps for a flow, ordered by step_order
	GetFlowSteps(ctx context.Context, flowID string, user *models.UserSession) ([]*models.FlowStep, error)
	GetStep(ctx context.Context, stepID string, user *models.UserSession) (*models.FlowStep, error)
}

// ApprovalPersistence provides persistence operations for approval work items.
// This interface enables FlowExecutor to create work items without direct PersistenceService dependency.
type ApprovalPersistence interface {
	// Insert creates a new record in the specified table
	Insert(ctx context.Context, tableName string, data models.SObject, user *models.UserSession) (models.SObject, error)
}
