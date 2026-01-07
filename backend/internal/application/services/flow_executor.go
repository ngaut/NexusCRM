package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	"time"

	"github.com/nexuscrm/backend/internal/domain/events"
	"github.com/nexuscrm/backend/internal/domain/ports"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// FlowExecutor connects Flows to EventBus events for metadata-driven automation.
// It uses interfaces for all dependencies to enable proper unit testing.
type FlowExecutor struct {
	metadata            ports.FlowMetadataProvider
	actionSvc           ports.ActionExecutor
	eventBus            ports.EventPublisher
	formula             ports.FormulaEvaluator
	flowInstanceManager ports.FlowInstanceManager
	approvalPersistence ports.ApprovalPersistence
}

// NewFlowExecutor creates a new FlowExecutor with interface dependencies.
// This allows for proper dependency injection and testing.
func NewFlowExecutor(
	metadata ports.FlowMetadataProvider,
	actionSvc ports.ActionExecutor,
	eventBus ports.EventPublisher,
	flowInstanceManager ports.FlowInstanceManager,
	approvalPersistence ports.ApprovalPersistence,
) *FlowExecutor {
	return &FlowExecutor{
		metadata:            metadata,
		actionSvc:           actionSvc,
		eventBus:            eventBus,
		formula:             formula.NewEngine(),
		flowInstanceManager: flowInstanceManager,
		approvalPersistence: approvalPersistence,
	}
}

// NewFlowExecutorWithFormula creates a FlowExecutor with a custom formula evaluator.
// This is useful for testing to inject a mock formula engine.
func NewFlowExecutorWithFormula(
	metadata ports.FlowMetadataProvider,
	actionSvc ports.ActionExecutor,
	eventBus ports.EventPublisher,
	formulaEngine ports.FormulaEvaluator,
	flowInstanceManager ports.FlowInstanceManager,
	approvalPersistence ports.ApprovalPersistence,
) *FlowExecutor {
	return &FlowExecutor{
		metadata:            metadata,
		actionSvc:           actionSvc,
		eventBus:            eventBus,
		formula:             formulaEngine,
		flowInstanceManager: flowInstanceManager,
		approvalPersistence: approvalPersistence,
	}
}

// RegisterFlowHandlers subscribes to EventBus events and executes matching Flows
func (fe *FlowExecutor) RegisterFlowHandlers() {
	// Dynamically subscribe to all events supported by metadata
	supportedEvents := fe.metadata.GetSupportedEvents()

	for _, eventType := range supportedEvents {
		// Capture variable for usage in closure
		evtType := eventType

		fe.eventBus.Subscribe(events.EventType(evtType), func(ctx context.Context, payload interface{}) error {
			recordPayload, ok := payload.(RecordEventPayload)
			if !ok {
				return nil
			}

			// Map event to flow trigger type
			triggerType := fe.mapEventToTrigger(evtType)
			if triggerType == "" {
				return nil
			}

			return fe.executeMatchingFlows(ctx, triggerType, recordPayload)
		})
	}

	log.Printf("✅ FlowExecutor: Registered handlers for %d supported events", len(supportedEvents))
}

// mapEventToTrigger converts platform event names to flow trigger types
func (fe *FlowExecutor) mapEventToTrigger(eventType string) string {
	switch eventType {
	case string(events.RecordAfterCreate):
		return constants.TriggerAfterCreate
	case string(events.RecordAfterUpdate):
		return constants.TriggerAfterUpdate
	case string(events.RecordAfterDelete):
		return constants.TriggerAfterDelete
	case string(events.RecordBeforeCreate):
		return constants.TriggerBeforeCreate
	case string(events.RecordBeforeUpdate):
		return constants.TriggerBeforeUpdate
	case string(events.RecordBeforeDelete):
		return constants.TriggerBeforeDelete
	default:
		return ""
	}
}

// executeMatchingFlows finds and executes Flows that match the trigger
func (fe *FlowExecutor) executeMatchingFlows(ctx context.Context, triggerType string, payload RecordEventPayload) error {
	flows := fe.metadata.GetFlows(ctx)
	log.Printf("🔍 FlowExecutor: checking %d flows for trigger '%s' on object '%s'", len(flows), triggerType, payload.ObjectAPIName)

	for _, flow := range flows {
		// Skip inactive flows
		if flow.Status != constants.FlowStatusActive {
			continue
		}

		// Check if flow triggers on this object (case-insensitive)
		if !strings.EqualFold(flow.TriggerObject, payload.ObjectAPIName) {
			continue
		}

		// Check if flow triggers on this event type (case-insensitive for safety)
		if !strings.EqualFold(flow.TriggerType, triggerType) {
			continue
		}

		// Check trigger condition
		if flow.TriggerCondition != "" {
			// Evaluate condition formula
			formulaCtx := fe.createFormulaContext(payload)

			result, err := fe.formula.Evaluate(flow.TriggerCondition, formulaCtx)
			if err != nil {
				log.Printf("⚠️ Flow %s: condition evaluation failed: %v", flow.Name, err)
				continue
			}

			// Condition must evaluate to true
			if conditionMet, ok := result.(bool); !ok || !conditionMet {
				continue
			}
		}

		// Execute the flow action
		log.Printf("🔄 Flow %s: executing %s action on %s", flow.Name, flow.ActionType, payload.ObjectAPIName)

		if err := fe.executeFlowAction(ctx, flow, payload); err != nil {
			log.Printf("❌ Flow %s: execution failed: %v", flow.Name, err)
			// Continue with other flows even if one fails
			continue
		}

		log.Printf("✅ Flow %s: executed successfully", flow.Name)
	}

	return nil
}

// createFormulaContext helper to create context for formula evaluation
func (fe *FlowExecutor) createFormulaContext(payload RecordEventPayload) *formula.Context {
	ctx := &formula.Context{
		Record: payload.Record,
	}
	if payload.CurrentUser != nil {
		ctx.User = payload.CurrentUser.ToMap()
	}
	if payload.OldRecord != nil {
		ctx.Prior = *payload.OldRecord
	}
	return ctx
}

// executeFlowAction executes a single flow's action
func (fe *FlowExecutor) executeFlowAction(ctx context.Context, flow *models.Flow, payload RecordEventPayload) error {
	// For multi-step flows with empty ActionType, invoke multi-step execution
	if flow.FlowType == constants.FlowTypeMultistep && flow.ActionType == "" {
		return fe.executeMultiStepFlow(ctx, flow, payload)
	}
	return fe.executeActionLogic(ctx, flow.ActionType, flow.ActionConfig, flow.ID, flow.TriggerType, payload)
}

// executeActionLogic executes a generic action based on type and config
func (fe *FlowExecutor) executeActionLogic(ctx context.Context, actionType string, config map[string]interface{}, flowID, triggerType string, payload RecordEventPayload) error {
	isBeforeTrigger := strings.EqualFold(triggerType, constants.TriggerBeforeCreate) ||
		strings.EqualFold(triggerType, constants.TriggerBeforeUpdate) ||
		strings.EqualFold(triggerType, constants.TriggerBeforeDelete)

	if strings.EqualFold(actionType, constants.ActionTypeExecuteAction) {
		// Execute an action by ID from config
		if actionID, ok := config[constants.ConfigActionID].(string); ok {
			return fe.actionSvc.ExecuteAction(ctx, actionID, payload.Record, payload.CurrentUser)
		}
		return fmt.Errorf("flow action missing action_id in config")
	}

	if strings.EqualFold(actionType, constants.ActionTypeUpdateRecord) {
		// Update the current record with specified field values
		fieldMappings, ok := config[constants.ConfigFieldMappings].(map[string]interface{})

		if ok {
			// If this is a BEFORE trigger, we update the record in-memory
			// This modifies the Record map which is a reference type, so changes persist
			if isBeforeTrigger {
				formulaCtx := fe.createFormulaContext(payload)
				for fieldName, expr := range fieldMappings {
					// Evaluate value if it's a formula
					var val interface{} = expr
					if strVal, ok := expr.(string); ok && strings.HasPrefix(strVal, "=") {
						res, err := fe.formula.Evaluate(strVal[1:], formulaCtx)
						if err != nil {
							return fmt.Errorf("formula evaluation failed for field %s: %w", fieldName, err)
						}
						val = res
					}
					payload.Record[fieldName] = val
				}
				return nil
			}

			// Otherwise (AFTER trigger), we must perform a database update via ActionService
			targetObject := payload.ObjectAPIName
			recordID := payload.Record.GetString(constants.FieldID)
			action := &models.ActionMetadata{
				Type:         constants.ActionTypeUpdateRecord,
				TargetObject: &targetObject,
				Config: map[string]interface{}{
					constants.ConfigTargetObject:  targetObject,
					constants.ConfigRecordID:      recordID,
					constants.ConfigFieldMappings: fieldMappings,
				},
			}
			return fe.actionSvc.ExecuteActionDirect(ctx, action, payload.Record, payload.CurrentUser)
		}
		return fmt.Errorf("flow updateRecord missing field_mappings in config")
	}

	if strings.EqualFold(actionType, constants.ActionTypeCreateRecord) {
		// Create a new record
		if targetObject, ok := config[constants.ConfigTargetObject].(string); ok {
			action := &models.ActionMetadata{
				Type:         constants.ActionTypeCreateRecord,
				TargetObject: &targetObject,
				Config:       config,
			}
			return fe.actionSvc.ExecuteActionDirect(ctx, action, payload.Record, payload.CurrentUser)
		}
		return fmt.Errorf("flow createRecord missing target_object in config")
	}

	if strings.EqualFold(actionType, constants.ActionTypeSendEmail) {
		// Send email
		action := &models.ActionMetadata{
			Type:   constants.ActionTypeSendEmail,
			Config: config,
		}
		return fe.actionSvc.ExecuteActionDirect(ctx, action, payload.Record, payload.CurrentUser)
	}

	if strings.EqualFold(actionType, constants.ActionTypeCallWebhook) {
		// Call webhook
		action := &models.ActionMetadata{
			Type:   constants.ActionTypeCallWebhook,
			Config: config,
		}
		return fe.actionSvc.ExecuteActionDirect(ctx, action, payload.Record, payload.CurrentUser)
	}

	if strings.EqualFold(actionType, constants.ActionTypeSubmitForApproval) {
		return fe.executeApprovalLogic(ctx, config, flowID, payload)
	}

	if actionType == "" {
		// For multi-step root flow, do nothing here (handled by executeMultiStepFlow caller)
		return nil
	}

	return fmt.Errorf("unknown flow action type: %s", actionType)
}

// executeMultiStepFlow starts a multi-step flow and executes the first step
func (fe *FlowExecutor) executeMultiStepFlow(ctx context.Context, flow *models.Flow, payload RecordEventPayload) error {
	if fe.flowInstanceManager == nil {
		return fmt.Errorf("flow instance manager not configured")
	}

	// Create Instance
	instance, err := fe.flowInstanceManager.CreateInstance(ctx, flow, payload.ObjectAPIName, payload.Record.GetString(constants.FieldID), payload.CurrentUser)
	if err != nil {
		return fmt.Errorf("failed to create flow instance: %w", err)
	}

	// Get steps to find the first one
	steps, err := fe.flowInstanceManager.GetFlowSteps(ctx, flow.ID, payload.CurrentUser)
	if err != nil {
		return fmt.Errorf("failed to get flow steps: %w", err)
	}

	if len(steps) == 0 {
		return fmt.Errorf("no steps found for flow %s", flow.Name)
	}

	// Find step with lowest order (or StepOrder=1)
	var firstStep *models.FlowStep
	// Simple assumption: StepOrder 1 is first. Or find min.
	minOrder := 999999
	for _, s := range steps {
		if s.StepOrder < minOrder {
			minOrder = s.StepOrder
			firstStep = s
		}
	}

	if firstStep == nil {
		return fmt.Errorf("could not determine first step")
	}

	// Execute the step
	return fe.executeStep(ctx, instance, firstStep, payload)
}

// ExecuteInstanceStep executes a specific step of a running instance
func (fe *FlowExecutor) ExecuteInstanceStep(ctx context.Context, instance *models.FlowInstance, step *models.FlowStep, payload RecordEventPayload) error {
	return fe.executeStep(ctx, instance, step, payload)
}

// executeStep executes a specific step within a flow instance
func (fe *FlowExecutor) executeStep(ctx context.Context, instance *models.FlowInstance, step *models.FlowStep, payload RecordEventPayload) error {
	switch strings.ToLower(step.StepType) {
	case constants.FlowStepTypeApproval:
		return fe.executeApprovalStep(ctx, instance, step, payload)
	case constants.FlowStepTypeAction:
		// Execute Generic Action
		// For Multi-step steps, TriggerType is technically "flow", but effectively "after..." because it runs async/deferred.
		// We pass empty TriggerType or "afterUpdate" if original trigger was that.
		// Use "automated" to imply backend execution.
		var actionType string
		if step.ActionType != nil {
			actionType = *step.ActionType
		}
		return fe.executeActionLogic(ctx, actionType, step.ActionConfig, instance.FlowID, constants.TriggerAfterUpdate, payload)
	default:
		return fmt.Errorf("unsupported step type: %s", step.StepType)
	}
}

// executeApprovalStep creates an approval work item for a multi-step flow step
func (fe *FlowExecutor) executeApprovalStep(ctx context.Context, instance *models.FlowInstance, step *models.FlowStep, payload RecordEventPayload) error {
	if fe.approvalPersistence == nil {
		return fmt.Errorf("approval persistence not configured")
	}

	config := step.ActionConfig
	if config == nil {
		config = make(map[string]interface{})
	}

	// Resolve approver using shared helper
	approverID := fe.resolveApproverID(config, payload)
	recordID := payload.Record.GetString(constants.FieldID)

	// Build work item with flow context
	workItem := fe.buildApprovalWorkItem(payload.ObjectAPIName, recordID, approverID, config, payload.CurrentUser)
	workItem[constants.FieldSysApprovalWorkItem_ProcessID] = instance.FlowID
	workItem[constants.FieldSysApprovalWorkItem_FlowInstanceID] = instance.ID
	workItem[constants.FieldSysApprovalWorkItem_FlowStepID] = step.ID

	// Pause instance before creating work item
	if err := fe.flowInstanceManager.PauseInstance(ctx, instance.ID, step.ID, payload.CurrentUser); err != nil {
		log.Printf("⚠️ Failed to pause flow instance: %v", err)
	}

	// Insert work item
	if _, err := fe.approvalPersistence.Insert(ctx, constants.TableApprovalWorkItem, workItem, payload.CurrentUser); err != nil {
		return fmt.Errorf("failed to create approval work item: %w", err)
	}

	log.Printf("✅ Started approval step %s for flow %s", step.StepName, instance.ID)
	return nil
}

// executeApprovalLogic handles submitForApproval action type for simple flows
// or creates a new multi-step flow instance if the flow is multistep
func (fe *FlowExecutor) executeApprovalLogic(ctx context.Context, config map[string]interface{}, flowID string, payload RecordEventPayload) error {
	if fe.approvalPersistence == nil {
		return fmt.Errorf("approval persistence not configured for approval actions")
	}

	// Resolve approver using shared helper
	approverID := fe.resolveApproverID(config, payload)
	recordID := payload.Record.GetString(constants.FieldID)
	if recordID == "" {
		return fmt.Errorf("cannot submit for approval: record has no ID")
	}

	// Build base work item
	workItem := fe.buildApprovalWorkItem(payload.ObjectAPIName, recordID, approverID, config, payload.CurrentUser)
	workItem[constants.FieldSysApprovalWorkItem_ProcessID] = flowID

	// For multistep flows, create instance and find approval step
	if flowID != "" && fe.flowInstanceManager != nil {
		flow := fe.metadata.GetFlow(ctx, flowID)
		if flow != nil && flow.FlowType == constants.FlowTypeMultistep {
			instance, err := fe.flowInstanceManager.CreateInstance(ctx, flow, payload.ObjectAPIName, recordID, payload.CurrentUser)
			if err != nil {
				return fmt.Errorf("failed to create flow instance: %w", err)
			}
			workItem[constants.FieldSysApprovalWorkItem_FlowInstanceID] = instance.ID

			// Find and link to approval step, then pause
			steps, err := fe.flowInstanceManager.GetFlowSteps(ctx, flow.ID, payload.CurrentUser)
			if err == nil {
				for _, step := range steps {
					if step.StepType == constants.FlowStepTypeApproval {
						workItem[constants.FieldSysApprovalWorkItem_FlowStepID] = step.ID
						if err := fe.flowInstanceManager.PauseInstance(ctx, instance.ID, step.ID, payload.CurrentUser); err != nil {
							log.Printf("⚠️ Failed to pause flow instance: %v", err)
						}
						break
					}
				}
			}
		}
	}

	// Insert work item
	created, err := fe.approvalPersistence.Insert(ctx, constants.TableApprovalWorkItem, workItem, payload.CurrentUser)
	if err != nil {
		return fmt.Errorf("failed to create approval work item: %w", err)
	}

	log.Printf("✅ Created approval work item %s for %s/%s", created.GetString(constants.FieldID), payload.ObjectAPIName, recordID)
	return nil
}

// resolveApproverID extracts approver ID from config using formula or static value
func (fe *FlowExecutor) resolveApproverID(config map[string]interface{}, payload RecordEventPayload) *string {
	// Try formula first
	formulaStr := GetConfigString(config, constants.ConfigApproverFormula)
	if formulaStr != "" {
		formulaCtx := fe.createFormulaContext(payload)
		if result, err := fe.formula.Evaluate(formulaStr, formulaCtx); err == nil {
			if resStr, ok := result.(string); ok && resStr != "" {
				return &resStr
			}
		}
	}

	// Fall back to static approver ID
	staticApprover := GetConfigString(config, constants.ConfigApproverID)
	if staticApprover != "" {
		return &staticApprover
	}

	return nil
}

// buildApprovalWorkItem creates base work item SObject with common fields
func (fe *FlowExecutor) buildApprovalWorkItem(objectAPIName, recordID string, approverID *string, config map[string]interface{}, user *models.UserSession) models.SObject {
	workItem := models.SObject{
		constants.FieldSysApprovalWorkItem_ObjectAPIName: objectAPIName,
		constants.FieldSysApprovalWorkItem_RecordID:      recordID,
		constants.FieldSysApprovalWorkItem_Status:        constants.ApprovalStatusPending,
		constants.FieldSysApprovalWorkItem_SubmittedDate: time.Now(),
	}
	if user != nil {
		workItem[constants.FieldSysApprovalWorkItem_SubmittedByID] = user.ID
	}
	if approverID != nil {
		workItem[constants.FieldSysApprovalWorkItem_ApproverID] = *approverID
	}

	if comments := GetConfigString(config, constants.ConfigComments); comments != "" {
		workItem[constants.FieldSysApprovalWorkItem_Comments] = comments
	}
	return workItem
}
