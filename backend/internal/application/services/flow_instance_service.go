package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/nexuscrm/backend/internal/domain"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// FlowInstanceService manages flow instance lifecycle (create, pause, resume, complete)
type FlowInstanceService struct {
	persistence  *PersistenceService
	query        *QueryService
	metadata     *MetadataService
	stateMachine *domain.FlowStateMachine
}

// NewFlowInstanceService creates a new FlowInstanceService
func NewFlowInstanceService(persistence *PersistenceService, query *QueryService, metadata *MetadataService) *FlowInstanceService {
	return &FlowInstanceService{
		persistence:  persistence,
		query:        query,
		metadata:     metadata,
		stateMachine: domain.NewFlowStateMachine(),
	}
}

// mapStatusToState maps database status string to FlowState
func mapStatusToState(status string) domain.FlowState {
	switch status {
	case models.FlowInstanceStatusRunning:
		return domain.FlowStateRunning
	case models.FlowInstanceStatusPaused:
		return domain.FlowStatePaused
	case models.FlowInstanceStatusCompleted:
		return domain.FlowStateCompleted
	case models.FlowInstanceStatusFailed:
		return domain.FlowStateFailed
	default:
		return domain.FlowStateRunning // Default for unknown
	}
}

// CreateInstance creates a new flow instance when a multi-step flow starts
func (s *FlowInstanceService) CreateInstance(ctx context.Context, flow *models.Flow, objectAPIName, recordID string, user *models.UserSession) (*models.FlowInstance, error) {
	instance := &models.FlowInstance{
		FlowID:        flow.ID,
		ObjectAPIName: objectAPIName,
		RecordID:      recordID,
		Status:        models.FlowInstanceStatusRunning,
		StartedDate:   time.Now().UTC(),
		ContextData:   make(map[string]interface{}),
	}
	if user != nil {
		instance.CreatedByID = &user.ID
	}

	// Convert to SObject for persistence
	data := models.SObject{
		constants.FieldSysFlowInstance_FlowID:        instance.FlowID,
		constants.FieldSysFlowInstance_ObjectAPIName: instance.ObjectAPIName,
		constants.FieldSysFlowInstance_RecordID:      instance.RecordID,
		constants.FieldSysFlowInstance_Status:        instance.Status,
		constants.FieldSysFlowInstance_StartedDate:   instance.StartedDate,
		constants.FieldSysFlowInstance_ContextData:   "{}",
	}
	if instance.CreatedByID != nil {
		data[constants.FieldCreatedByID] = *instance.CreatedByID
	}

	created, err := s.persistence.Insert(ctx, constants.TableFlowInstance, data, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create flow instance: %w", err)
	}

	instance.ID = created.GetString(constants.FieldID)
	log.Printf("✅ FlowInstance created: %s for flow %s on %s/%s", instance.ID, flow.ID, objectAPIName, recordID)
	return instance, nil
}

// PauseInstance pauses a flow instance when it hits an approval step
func (s *FlowInstanceService) PauseInstance(ctx context.Context, instanceID, currentStepID string, user *models.UserSession) error {
	// Get current instance to validate state transition
	instance, err := s.GetInstance(ctx, instanceID, user)
	if err != nil {
		return fmt.Errorf("failed to get instance for pause: %w", err)
	}

	// Validate transition using state machine
	currentState := mapStatusToState(instance.Status)
	_, err = s.stateMachine.Transition(currentState, domain.TransitionPause)
	if err != nil {
		return fmt.Errorf("invalid state transition: %w", err)
	}

	now := time.Now().UTC()
	updates := models.SObject{
		constants.FieldSysFlowInstance_Status:        models.FlowInstanceStatusPaused,
		constants.FieldSysFlowInstance_CurrentStepID: currentStepID,
		constants.FieldSysFlowInstance_PausedDate:    now,
	}

	if err := s.persistence.Update(ctx, constants.TableFlowInstance, instanceID, updates, user); err != nil {
		return fmt.Errorf("failed to pause flow instance: %w", err)
	}

	log.Printf("⏸️ FlowInstance paused: %s at step %s", instanceID, currentStepID)
	return nil
}

// ResumeInstance resumes a paused flow instance after approval action
func (s *FlowInstanceService) ResumeInstance(ctx context.Context, instanceID, nextStepID string, user *models.UserSession) error {
	// Get current instance to validate state transition
	instance, err := s.GetInstance(ctx, instanceID, user)
	if err != nil {
		return fmt.Errorf("failed to get instance for resume: %w", err)
	}

	// Validate transition using state machine
	currentState := mapStatusToState(instance.Status)
	_, err = s.stateMachine.Transition(currentState, domain.TransitionResume)
	if err != nil {
		return fmt.Errorf("invalid state transition: %w", err)
	}

	updates := models.SObject{
		constants.FieldSysFlowInstance_Status:        models.FlowInstanceStatusRunning,
		constants.FieldSysFlowInstance_CurrentStepID: nextStepID,
		constants.FieldSysFlowInstance_PausedDate:    nil, // Clear paused date
	}

	if err := s.persistence.Update(ctx, constants.TableFlowInstance, instanceID, updates, user); err != nil {
		return fmt.Errorf("failed to resume flow instance: %w", err)
	}

	log.Printf("▶️ FlowInstance resumed: %s continuing to step %s", instanceID, nextStepID)
	return nil
}

// CompleteInstance marks a flow instance as completed
func (s *FlowInstanceService) CompleteInstance(ctx context.Context, instanceID string, user *models.UserSession) error {
	// Get current instance to validate state transition
	instance, err := s.GetInstance(ctx, instanceID, user)
	if err != nil {
		return fmt.Errorf("failed to get instance for complete: %w", err)
	}

	// Validate transition using state machine
	currentState := mapStatusToState(instance.Status)
	_, err = s.stateMachine.Transition(currentState, domain.TransitionComplete)
	if err != nil {
		return fmt.Errorf("invalid state transition: %w", err)
	}

	now := time.Now().UTC()
	updates := models.SObject{
		constants.FieldSysFlowInstance_Status:        models.FlowInstanceStatusCompleted,
		constants.FieldSysFlowInstance_CompletedDate: now,
	}

	if err := s.persistence.Update(ctx, constants.TableFlowInstance, instanceID, updates, user); err != nil {
		return fmt.Errorf("failed to complete flow instance: %w", err)
	}

	log.Printf("✅ FlowInstance completed: %s", instanceID)
	return nil
}

// FailInstance marks a flow instance as failed
func (s *FlowInstanceService) FailInstance(ctx context.Context, instanceID, reason string, user *models.UserSession) error {
	// Get current instance to validate state transition
	instance, err := s.GetInstance(ctx, instanceID, user)
	if err != nil {
		return fmt.Errorf("failed to get instance for fail: %w", err)
	}

	// Validate transition using state machine
	currentState := mapStatusToState(instance.Status)
	_, err = s.stateMachine.Transition(currentState, domain.TransitionFail)
	if err != nil {
		return fmt.Errorf("invalid state transition: %w", err)
	}

	updates := models.SObject{
		constants.FieldSysFlowInstance_Status:      models.FlowInstanceStatusFailed,
		constants.FieldSysFlowInstance_ContextData: fmt.Sprintf(`{"error": "%s"}`, reason),
	}

	if err := s.persistence.Update(ctx, constants.TableFlowInstance, instanceID, updates, user); err != nil {
		return fmt.Errorf("failed to mark flow instance as failed: %w", err)
	}

	log.Printf("❌ FlowInstance failed: %s - %s", instanceID, reason)
	return nil
}

// GetInstance retrieves a flow instance by ID
func (s *FlowInstanceService) GetInstance(ctx context.Context, instanceID string, user *models.UserSession) (*models.FlowInstance, error) {
	filterExpr := fmt.Sprintf("%s == '%s'", constants.FieldID, instanceID)
	records, err := s.query.QueryWithFilter(ctx, constants.TableFlowInstance, filterExpr, user, "", "", 1)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}

	return s.sobjectToFlowInstance(records[0])
}

// GetInstanceByWorkItem finds the flow instance linked to an approval work item
func (s *FlowInstanceService) GetInstanceByWorkItem(ctx context.Context, workItemID string, user *models.UserSession) (*models.FlowInstance, error) {
	// First get the work item to find the flow_instance_id
	filterExpr := fmt.Sprintf("%s == '%s'", constants.FieldID, workItemID)
	workItems, err := s.query.QueryWithFilter(ctx, constants.TableApprovalWorkItem, filterExpr, user, "", "", 1)
	if err != nil {
		return nil, err
	}
	if len(workItems) == 0 {
		return nil, nil
	}

	instanceID, ok := workItems[0][constants.FieldSysApprovalWorkItem_FlowInstanceID].(string)
	if !ok || instanceID == "" {
		return nil, nil // No linked flow instance
	}

	return s.GetInstance(ctx, instanceID, user)
}

// GetFlowSteps retrieves all steps for a flow, ordered by step_order
func (s *FlowInstanceService) GetFlowSteps(ctx context.Context, flowID string, user *models.UserSession) ([]*models.FlowStep, error) {
	filterExpr := fmt.Sprintf("%s == '%s'", constants.FieldSysFlowStep_FlowID, flowID)
	records, err := s.query.QueryWithFilter(ctx, constants.TableFlowStep, filterExpr, user, "step_order", "ASC", 100)
	if err != nil {
		return nil, err
	}

	steps := make([]*models.FlowStep, 0, len(records))
	for _, record := range records {
		step, err := s.sobjectToFlowStep(record)
		if err != nil {
			log.Printf("⚠️ Failed to parse flow step: %v", err)
			continue
		}
		steps = append(steps, step)
	}

	return steps, nil
}

// GetStep retrieves a single step by ID
func (s *FlowInstanceService) GetStep(ctx context.Context, stepID string, user *models.UserSession) (*models.FlowStep, error) {
	filterExpr := fmt.Sprintf("%s == '%s'", constants.FieldID, stepID)
	records, err := s.query.QueryWithFilter(ctx, constants.TableFlowStep, filterExpr, user, "", "", 1)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}

	return s.sobjectToFlowStep(records[0])
}

// GetNextStep determines the next step based on current step and outcome
func (s *FlowInstanceService) GetNextStep(currentStep *models.FlowStep, success bool, allSteps []*models.FlowStep) *models.FlowStep {
	var nextStepID *string
	if success {
		nextStepID = currentStep.OnSuccessStep
	} else {
		nextStepID = currentStep.OnFailureStep
	}

	// If explicit next step specified, find it
	if nextStepID != nil && *nextStepID != "" {
		for _, step := range allSteps {
			if step.ID == *nextStepID {
				return step
			}
		}
	}

	// Otherwise, find next step by order
	if success {
		for _, step := range allSteps {
			if step.StepOrder > currentStep.StepOrder {
				return step
			}
		}
	}

	// No more steps
	return nil
}

// ResumeAfterApproval handles post-approval flow continuation.
// This method encapsulates all flow resumption logic that was previously in the REST handler.
// It determines the next step, executes it, and completes the flow if no more steps remain.
func (s *FlowInstanceService) ResumeAfterApproval(
	ctx context.Context,
	instanceID, stepID string,
	approved bool,
	stepExecutor func(ctx context.Context, instance *models.FlowInstance, step *models.FlowStep, record models.SObject, user *models.UserSession) error,
	user *models.UserSession,
) error {
	// Execute whole flow resumption logic in a transaction
	return s.persistence.RunInTransaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		// Get the current step to determine next step
		currentStep, err := s.GetStep(txCtx, stepID, user)
		if err != nil || currentStep == nil {
			return fmt.Errorf("failed to get flow step %s: %w", stepID, err)
		}

		// Get all steps for this flow to find the next one
		allSteps, err := s.GetFlowSteps(txCtx, currentStep.FlowID, user)
		if err != nil {
			return fmt.Errorf("failed to get flow steps: %w", err)
		}

		// Determine next step based on approval outcome
		nextStep := s.GetNextStep(currentStep, approved, allSteps)

		if nextStep == nil {
			// No more steps - complete the flow
			return s.CompleteInstance(txCtx, instanceID, user)
		}

		// Resume flow to next step (Update status)
		if err := s.ResumeInstance(txCtx, instanceID, nextStep.ID, user); err != nil {
			return fmt.Errorf("failed to resume flow instance: %w", err)
		}

		log.Printf("▶️ Flow instance %s resumed to step: %s", instanceID, nextStep.StepName)

		// Get instance to know Object and Record ID
		instance, err := s.GetInstance(txCtx, instanceID, user)
		if err != nil || instance == nil {
			return fmt.Errorf("failed to load resumed instance %s: %w", instanceID, err)
		}

		// Get record data
		filterExprRecord := fmt.Sprintf("%s == '%s'", constants.FieldID, instance.RecordID)
		records, err := s.query.QueryWithFilter(txCtx, instance.ObjectAPIName, filterExprRecord, user, "", "", 1)
		if err != nil || len(records) == 0 {
			return fmt.Errorf("failed to load record %s/%s: %w", instance.ObjectAPIName, instance.RecordID, err)
		}
		record := records[0]

		// Execute the next step (recursive/complex side effects)
		// We pass txCtx so if it does DB ops via persistence, it joins the tx.
		if err := stepExecutor(txCtx, instance, nextStep, record, user); err != nil {
			_ = s.FailInstance(txCtx, instanceID, err.Error(), user)
			return fmt.Errorf("failed to execute step %s: %w", nextStep.StepName, err)
		}

		// After successful non-approval step execution, check if there's a next step
		// If not, complete the flow instance
		if nextStep.StepType != constants.FlowStepTypeApproval {
			subsequentNextStep := s.GetNextStep(nextStep, true, allSteps)
			if subsequentNextStep == nil {
				return s.CompleteInstance(txCtx, instanceID, user)
			}
		}

		return nil
	})
}

// Progress tracking types and helper functions are in flow_instance_progress.go:
// - FlowInstanceProgress, FlowStepProgress types
// - GetInstanceProgress, sobjectToFlowInstance, sobjectToFlowStep
