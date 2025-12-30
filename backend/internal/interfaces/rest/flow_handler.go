package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/pkg/constants"
	"github.com/nexuscrm/backend/pkg/errors"
)

// FlowHandler handles flow management API endpoints
type FlowHandler struct {
	svc *services.ServiceManager
}

// NewFlowHandler creates a new FlowHandler
func NewFlowHandler(svc *services.ServiceManager) *FlowHandler {
	return &FlowHandler{svc: svc}
}

// GetAllFlows handles GET /api/metadata/flows
func (h *FlowHandler) GetAllFlows(c *gin.Context) {
	HandleGetEnvelope(c, "flows", func() (interface{}, error) {
		return h.svc.Metadata.GetFlows(), nil
	})
}

// GetFlow handles GET /api/metadata/flows/:flowId
func (h *FlowHandler) GetFlow(c *gin.Context) {
	flowID := c.Param("flowId")
	HandleGetEnvelope(c, "", func() (interface{}, error) {
		flow := h.svc.Metadata.GetFlow(flowID)
		if flow == nil {
			return nil, errors.NewNotFoundError("Flow", flowID)
		}
		return flow, nil
	})
}

// CreateFlow handles POST /api/metadata/flows
func (h *FlowHandler) CreateFlow(c *gin.Context) {
	var flow models.Flow
	HandleCreateEnvelope(c, "data", "Flow created successfully", &flow, func() error {
		if flow.Name == "" {
			return errors.NewValidationError("name", "Flow name is required")
		}
		if flow.TriggerObject == "" {
			return errors.NewValidationError("trigger_object", "Trigger object is required")
		}
		if flow.ActionType == "" && flow.FlowType != constants.FlowTypeMultistep {
			return errors.NewValidationError("action_type", "Action type is required for simple flows")
		}
		return h.svc.Metadata.CreateFlow(&flow)
	})
}

// UpdateFlow handles PATCH /api/metadata/flows/:flowId
func (h *FlowHandler) UpdateFlow(c *gin.Context) {
	flowID := c.Param("flowId")
	var updates models.Flow

	HandleUpdateEnvelope(c, "", "Flow updated successfully", &updates, func() error {
		if err := h.svc.Metadata.UpdateFlow(flowID, &updates); err != nil {
			return err
		}
		// Re-fetch to ensure response reflects server state
		if updated := h.svc.Metadata.GetFlow(flowID); updated != nil {
			updates = *updated
		}
		return nil
	})
}

// DeleteFlow handles DELETE /api/metadata/flows/:flowId
func (h *FlowHandler) DeleteFlow(c *gin.Context) {
	flowID := c.Param("flowId")
	HandleDeleteEnvelope(c, "Flow deleted successfully", func() error {
		return h.svc.Metadata.DeleteFlow(flowID)
	})
}

// ============================================================================
// Auto-Launched Flow Execution
// ============================================================================

// ExecuteFlowRequest represents a request to execute an auto-launched flow
type ExecuteFlowRequest struct {
	RecordID      string                 `json:"record_id"`       // For update_record actions
	ObjectAPIName string                 `json:"object_api_name"` // Optional: Override trigger object
	Context       map[string]interface{} `json:"context"`         // Additional field values
}

// ExecuteFlowResponse represents the result of flow execution
type ExecuteFlowResponse struct {
	Success bool                   `json:"success"`
	FlowID  string                 `json:"flow_id"`
	Message string                 `json:"message"`
	Result  map[string]interface{} `json:"result,omitempty"`
}

// ExecuteFlow handles POST /api/flows/:flowId/execute
// Allows auto-launched flows to be invoked via REST API (admin only)
func (h *FlowHandler) ExecuteFlow(c *gin.Context) {
	flowID := c.Param("flowId")
	user := GetUserFromContext(c)

	var req ExecuteFlowRequest
	if !BindJSON(c, &req) {
		return
	}

	// Validate and get flow
	flow, err := h.validateFlowForExecution(flowID, user)
	if err != nil {
		RespondError(c, err.code, err.message)
		return
	}

	// Execute based on action type
	result, execErr := h.executeFlowAction(c, flow, &req, user)
	if execErr != nil {
		RespondError(c, 500, "Flow execution failed: "+execErr.Error())
		return
	}

	c.JSON(200, ExecuteFlowResponse{
		Success: true,
		FlowID:  flowID,
		Message: "Flow executed successfully",
		Result:  result,
	})
}

// flowError represents a flow execution error with HTTP status code
type flowError struct {
	code    int
	message string
}

// validateFlowForExecution validates flow exists, is active, and user has permission
func (h *FlowHandler) validateFlowForExecution(flowID string, user *models.UserSession) (*models.Flow, *flowError) {
	flow := h.svc.Metadata.GetFlow(flowID)
	if flow == nil {
		return nil, &flowError{404, "Flow not found: " + flowID}
	}

	if flow.Status != constants.FlowStatusActive {
		return nil, &flowError{400, "Flow is not active (status: " + flow.Status + ")"}
	}

	// Security: Only system admins can execute flows via API
	if !constants.IsSuperUser(user.ProfileID) {
		return nil, &flowError{403, "Only system administrators can execute flows via API"}
	}

	return flow, nil
}

// executeFlowAction executes the appropriate action based on flow's ActionType
func (h *FlowHandler) executeFlowAction(
	c *gin.Context,
	flow *models.Flow,
	req *ExecuteFlowRequest,
	user *models.UserSession,
) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	switch flow.ActionType {
	case "create_record":
		return h.executeCreateRecord(c, flow, req, user)

	case "update_record":
		return h.executeUpdateRecord(c, flow, req, user)

	default:
		// No-op for unsupported action types
		result["action_type"] = flow.ActionType
		result["message"] = "Action type not implemented for REST invocation"
		return result, nil
	}
}

// executeCreateRecord handles create_record flow action
func (h *FlowHandler) executeCreateRecord(
	c *gin.Context,
	flow *models.Flow,
	req *ExecuteFlowRequest,
	user *models.UserSession,
) (map[string]interface{}, error) {
	if flow.ActionConfig == nil {
		return nil, nil
	}

	targetObject := getStringFromConfig(flow.ActionConfig, constants.ConfigTargetObject)
	if targetObject == "" {
		return nil, nil
	}

	// Validate target object exists
	if h.svc.Metadata.GetSchema(targetObject) == nil {
		RespondError(c, 400, "Target object does not exist: "+targetObject)
		return nil, nil
	}

	// Build record: context first, then field_mappings (flow config takes precedence)
	newRecord := buildRecordFromConfig(req.Context, flow.ActionConfig)

	created, err := h.svc.Persistence.Insert(c.Request.Context(), targetObject, newRecord, user)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"created_id":    created[constants.FieldID],
		"target_object": targetObject,
	}, nil
}

// executeUpdateRecord handles update_record flow action
func (h *FlowHandler) executeUpdateRecord(
	c *gin.Context,
	flow *models.Flow,
	req *ExecuteFlowRequest,
	user *models.UserSession,
) (map[string]interface{}, error) {
	if flow.ActionConfig == nil || req.RecordID == "" {
		return nil, nil
	}

	objectName := req.ObjectAPIName
	if objectName == "" {
		objectName = flow.TriggerObject
	}

	// Validate object exists
	if h.svc.Metadata.GetSchema(objectName) == nil {
		RespondError(c, 400, "Object does not exist: "+objectName)
		return nil, nil
	}

	// Build updates: context first, then field_mappings (flow config takes precedence)
	updates := buildRecordFromConfig(req.Context, flow.ActionConfig)

	err := h.svc.Persistence.Update(c.Request.Context(), objectName, req.RecordID, updates, user)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"updated_id": req.RecordID,
	}, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// getStringFromConfig safely extracts a string value from action config
func getStringFromConfig(config map[string]interface{}, key string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return ""
}

// buildRecordFromConfig builds an SObject from context and field_mappings
// Context values are applied first, then field_mappings override (flow config takes precedence)
func buildRecordFromConfig(context map[string]interface{}, actionConfig map[string]interface{}) models.SObject {
	record := models.SObject{}

	// Apply context values first
	for k, v := range context {
		record[k] = v
	}

	// Apply field_mappings (overrides context)
	if fieldMap, ok := actionConfig[constants.ConfigFieldMappings].(map[string]interface{}); ok {
		for field, value := range fieldMap {
			record[field] = value
		}
	}

	return record
}
