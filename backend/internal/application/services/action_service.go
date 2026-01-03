package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ActionService handles execution of metadata-driven actions
type ActionService struct {
	metadata    *MetadataService
	persistence *PersistenceService
	permissions *PermissionService
	txManager   *TransactionManager
	formula     *formula.Engine
}

// NewActionService creates a new ActionService
func NewActionService(metadata *MetadataService, persistence *PersistenceService, permissions *PermissionService, txManager *TransactionManager) *ActionService {
	return &ActionService{
		metadata:    metadata,
		persistence: persistence,
		permissions: permissions,
		txManager:   txManager,
		formula:     formula.NewEngine(),
	}
}

// ActionContext holds the state of an action execution, including results from previous steps
type ActionContext struct {
	Record  models.SObject
	User    *models.UserSession
	Results map[string]map[string]interface{}
}

// ExecuteAction executes an action by ID with the given context record and user
func (as *ActionService) ExecuteAction(ctx context.Context, actionID string, contextRecord models.SObject, user *models.UserSession) error {
	// Find the action in metadata cache
	action := as.findAction(actionID)
	if action == nil {
		return fmt.Errorf("action not found: %s", actionID)
	}

	// Create context
	actionCtx := &ActionContext{
		Record:  contextRecord,
		User:    user,
		Results: make(map[string]map[string]interface{}),
	}

	return as.executeActionFromMetadata(ctx, action, actionCtx)
}

// executeActionFromMetadata executes an action definition with the given context
func (as *ActionService) executeActionFromMetadata(ctx context.Context, action *models.ActionMetadata, actionCtx *ActionContext) error {
	// Execute based on action type
	switch action.Type {
	case constants.ActionTypeCreateRecord:
		return as.executeCreateRecord(ctx, action, actionCtx)
	case constants.ActionTypeUpdateRecord:
		return as.executeUpdateRecord(ctx, action, actionCtx)
	case constants.ActionTypeSendEmail:
		return as.executeSendEmail(ctx, action, actionCtx)
	case constants.ActionTypeCallWebhook:
		return as.executeCallWebhook(ctx, action, actionCtx)
	case constants.ActionTypeComposite:
		return as.executeComposite(ctx, action, actionCtx)
	default:
		return fmt.Errorf("unsupported action type: %s", action.Type)
	}
}

// ExecuteActionDirect executes an action metadata directly without ID lookup (used by FlowExecutor)
func (as *ActionService) ExecuteActionDirect(ctx context.Context, action *models.ActionMetadata, record models.SObject, user *models.UserSession) error {
	actionCtx := &ActionContext{
		Record:  record,
		User:    user,
		Results: make(map[string]map[string]interface{}),
	}
	return as.executeActionFromMetadata(ctx, action, actionCtx)
}

// findAction searches for an action by ID using MetadataService
func (as *ActionService) findAction(actionID string) *models.ActionMetadata {
	return as.metadata.GetActionByID(actionID)
}

// executeCreateRecord creates a new record based on action configuration
func (as *ActionService) executeCreateRecord(ctx context.Context, action *models.ActionMetadata, actionCtx *ActionContext) error {
	// Extract configuration
	targetObject, err := GetConfigStringRequired(action.Config, constants.ConfigTargetObject)
	if err != nil {
		return err
	}

	fieldMappings, ok := GetConfigMap(action.Config, constants.ConfigFieldMappings)
	if !ok {
		return fmt.Errorf("field_mappings not specified in action config")
	}

	// Build record from field mappings
	record := make(models.SObject)
	for fieldName, formulaOrValue := range fieldMappings {
		value, err := as.evaluateRef(formulaOrValue, actionCtx, action.ObjectAPIName)
		if err != nil {
			return fmt.Errorf("failed to evaluate field %s: %w", fieldName, err)
		}
		record[fieldName] = value
	}

	// Insert the record (context propagates the transaction)
	result, err := as.persistence.Insert(ctx, targetObject, record, actionCtx.User)
	if err != nil {
		return err
	}

	// Store result for future steps
	actionCtx.Results[action.ID] = result

	return nil
}

// executeUpdateRecord updates an existing record based on action configuration
func (as *ActionService) executeUpdateRecord(ctx context.Context, action *models.ActionMetadata, actionCtx *ActionContext) error {
	// Extract configuration
	targetObjectName, err := GetConfigStringRequired(action.Config, constants.ConfigTargetObject)
	if err != nil {
		return err
	}

	recordIDVal, err := as.getConfigValue(action.Config, constants.ConfigRecordID, actionCtx, action.ObjectAPIName)
	if err != nil {
		return fmt.Errorf("failed to get record_id: %w", err)
	}
	recordID := fmt.Sprintf("%v", recordIDVal)

	fieldMappings, ok := GetConfigMap(action.Config, constants.ConfigFieldMappings)
	if !ok {
		return fmt.Errorf("field_mappings not specified in action config")
	}

	// Build updates from field mappings
	updates := make(models.SObject)
	for fieldName, formulaOrValue := range fieldMappings {
		value, err := as.evaluateRef(formulaOrValue, actionCtx, action.ObjectAPIName)
		if err != nil {
			return fmt.Errorf("failed to evaluate field %s: %w", fieldName, err)
		}
		updates[fieldName] = value
	}

	// Update the record
	return as.persistence.Update(ctx, targetObjectName, recordID, updates, actionCtx.User)
}

// executeComposite executes a sequence of actions within a transaction
func (as *ActionService) executeComposite(ctx context.Context, action *models.ActionMetadata, actionCtx *ActionContext) error {
	// composite action config should have a "steps" array
	stepsInterface, ok := action.Config[constants.ConfigKeySteps]
	if !ok {
		return fmt.Errorf("steps not specified in composite action")
	}

	stepsList, ok := stepsInterface.([]interface{})
	if !ok {
		return fmt.Errorf("steps must be an array")
	}

	// EXECUTE WITHIN TRANSACTION
	return as.txManager.WithTransaction(func(tx *sql.Tx) error {
		// Inject transaction into context
		txCtx := as.txManager.InjectTx(ctx, tx)

		// Execute steps using the transactional context
		return as.executeSteps(txCtx, stepsList, actionCtx, action.ObjectAPIName)
	})
}

func (as *ActionService) executeSteps(ctx context.Context, steps []interface{}, actionCtx *ActionContext, sourceObjectName string) error {
	for _, stepInterface := range steps {
		stepConfig, ok := stepInterface.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid step config format")
		}

		stepID, _ := stepConfig[constants.FieldID].(string)
		stepType, _ := stepConfig[constants.FieldSysAction_Type].(string)

		if stepID == "" || stepType == "" {
			return fmt.Errorf("step must have id and type")
		}

		stepAction := &models.ActionMetadata{
			ID:            stepID,
			ObjectAPIName: sourceObjectName, // Inherit source object context
			Type:          stepType,
			Config:        stepConfig, // Pass full config
		}

		// Recursively execute
		if err := as.executeActionFromMetadata(ctx, stepAction, actionCtx); err != nil {
			return fmt.Errorf("step %s failed: %w", stepID, err)
		}
	}
	return nil
}

// executeSendEmail sends an email based on action configuration
func (as *ActionService) executeSendEmail(ctx context.Context, action *models.ActionMetadata, actionCtx *ActionContext) error {
	// Extract email configuration
	toEmail, err := as.getConfigValue(action.Config, constants.ConfigTo, actionCtx, action.ObjectAPIName)
	if err != nil {
		return fmt.Errorf("failed to get 'to' email: %w", err)
	}

	subject, err := as.getConfigValue(action.Config, constants.ConfigSubject, actionCtx, action.ObjectAPIName)
	if err != nil {
		return fmt.Errorf("failed to get email subject: %w", err)
	}

	body, err := as.getConfigValue(action.Config, constants.ConfigBody, actionCtx, action.ObjectAPIName)
	if err != nil {
		return fmt.Errorf("failed to get email body: %w", err)
	}

	// Optional fields (used when email integration is implemented)
	ccEmail, _ := as.getConfigValue(action.Config, constants.ConfigCc, actionCtx, action.ObjectAPIName)
	bccEmail, _ := as.getConfigValue(action.Config, constants.ConfigBcc, actionCtx, action.ObjectAPIName)
	_, _, _ = body, ccEmail, bccEmail // Silence unused warnings - will be used when email is implemented

	// User details for logging
	userName := constants.DefaultUserName
	userEmail := constants.DefaultUserEmail
	if actionCtx.User != nil {
		userName = actionCtx.User.Name
		if actionCtx.User.Email != nil {
			userEmail = *actionCtx.User.Email
		}
	}

	// For now, log the email details
	// In production, this would integrate with an SMTP server or email service
	log.Printf("📧 EMAIL ACTION TRIGGERED: To=%v Subject=%v Triggered by: %s (%s)", toEmail, subject, userName, userEmail)

	return nil
}

// executeCallWebhook calls a webhook based on action configuration
func (as *ActionService) executeCallWebhook(ctx context.Context, action *models.ActionMetadata, actionCtx *ActionContext) error {
	// Extract webhook configuration
	urlValue, err := as.getConfigValue(action.Config, constants.ConfigURL, actionCtx, action.ObjectAPIName)
	if err != nil {
		return fmt.Errorf("failed to get webhook URL: %w", err)
	}
	url := fmt.Sprintf("%v", urlValue)

	methodValue, err := as.getConfigValue(action.Config, constants.ConfigMethod, actionCtx, action.ObjectAPIName)
	if err != nil {
		// Default to POST if not specified
		methodValue = "POST"
	}
	method := strings.ToUpper(fmt.Sprintf("%v", methodValue))

	// Validate method
	if method != "GET" && method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE" {
		return fmt.Errorf("invalid HTTP method: %s", method)
	}

	// Build request body from payload config
	var bodyReader io.Reader
	if payload, err := as.getConfigValue(action.Config, constants.ConfigPayload, actionCtx, action.ObjectAPIName); err == nil && payload != nil {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to serialize webhook payload: %w", err)
		}
		bodyReader = bytes.NewReader(payloadBytes)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// Set default content type
	req.Header.Set("Content-Type", "application/json")

	// Add custom headers if specified
	if headers, ok := GetConfigMap(action.Config, constants.ConfigHeaders); ok {
		for headerName, headerValue := range headers {
			req.Header.Set(headerName, fmt.Sprintf("%v", headerValue))
		}
	}

	// Execute the webhook with a reasonable timeout
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("⚠️ WEBHOOK FAILED: URL=%s Method=%s Error=%v", url, method, err)
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response status
	if resp.StatusCode >= 400 {
		log.Printf("⚠️ WEBHOOK ERROR RESPONSE: URL=%s Status=%d", url, resp.StatusCode)
		return fmt.Errorf("webhook returned error status: %d", resp.StatusCode)
	}

	log.Printf("✅ WEBHOOK SUCCESS: URL=%s Method=%s Status=%d", url, method, resp.StatusCode)
	return nil
}

// getConfigValue extracts a value from action config and evaluates it if it's a formula
func (as *ActionService) getConfigValue(config map[string]interface{}, key string, actionCtx *ActionContext, sourceObjectName string) (interface{}, error) {
	value, exists := config[key]
	if !exists {
		return nil, fmt.Errorf("%s not specified in action config", key)
	}

	// Evaluate if it's a formula
	return as.evaluateRef(value, actionCtx, sourceObjectName)
}

// evaluateRef evaluates a value using the metadata-driven Formula Engine
func (as *ActionService) evaluateRef(value interface{}, actionCtx *ActionContext, sourceObjectName string) (interface{}, error) {
	// If it's a string, check if it's a formula
	strValue, ok := value.(string)
	if !ok {
		// Return literal value as-is
		return value, nil
	}

	// Check if it's a formula (starts with {! and ends with })
	if !strings.HasPrefix(strValue, "{!") || !strings.HasSuffix(strValue, "}") {
		// Return literal string
		return strValue, nil
	}

	// Extract formula content
	expression := strings.TrimSpace(strValue[2 : len(strValue)-1])

	// Prepare Formula Engine Context from ActionContext
	formulaCtx := &formula.Context{
		Record: actionCtx.Record,
		Env:    make(map[string]interface{}), // Add Env if needed
	}

	// Add FLS check if user is present
	if actionCtx.User != nil {
		formulaCtx.IsVisible = func(fieldName string) bool {
			// Use provided source object name
			if sourceObjectName != "" {
				return as.permissions.CheckFieldVisibilityWithUser(sourceObjectName, fieldName, actionCtx.User)
			}
			// Fallback: If we can't determine object, we default to hidden for safety
			return false
		}
	}

	// Add User context
	if actionCtx.User != nil {
		email := ""
		if actionCtx.User.Email != nil {
			email = *actionCtx.User.Email
		}
		formulaCtx.User = map[string]interface{}{
			constants.FieldID:        actionCtx.User.ID,
			constants.FieldName:      actionCtx.User.Name,
			constants.FieldEmail:     email,
			constants.FieldProfileID: actionCtx.User.ProfileID,
		}
	} else {
		// Ensure User map exists to prevent panic if referenced
		formulaCtx.User = make(map[string]interface{})
	}

	// Add Results from previous steps as custom fields.
	// Flatten Results: results.stepId.field -> accessing via formula context.
	// We convert map[string]map[string]interface{} to map[string]interface{} for the "results" key.
	resultsMap := make(map[string]interface{})
	for k, v := range actionCtx.Results {
		resultsMap[k] = v
	}

	if formulaCtx.Fields == nil {
		formulaCtx.Fields = make(map[string]interface{})
	}
	formulaCtx.Fields[constants.ConfigKeyResults] = resultsMap

	// Evaluate using shared engine
	return as.formula.Evaluate(expression, formulaCtx)
}
