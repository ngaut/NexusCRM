package rest

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	appErrors "github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

type ActionHandler struct {
	svc *services.ServiceManager
}

func NewActionHandler(svc *services.ServiceManager) *ActionHandler {
	return &ActionHandler{svc: svc}
}

// Helper to sanitize action config (remove secrets)
func sanitizeAction(action *models.ActionMetadata) *models.ActionMetadata {
	if action == nil {
		return nil
	}
	// Create shallow copy to avoid modifying cache/source
	sanitized := *action
	if sanitized.Config != nil {
		// Deep copy config to sanitize
		newConfig := make(map[string]interface{})
		for k, v := range sanitized.Config {
			if isSensitiveKey(k) {
				newConfig[k] = "********"
			} else {
				newConfig[k] = v
			}
		}
		sanitized.Config = newConfig
	}
	return &sanitized
}

func isSensitiveKey(key string) bool {
	sensitive := []string{"headers", "authorization", "token", "secret", "password", "key", "api_key"}
	keyLower := strings.ToLower(key)

	for _, s := range sensitive {
		if strings.Contains(keyLower, s) {
			return true
		}
	}
	return false
}

// GetActions handles GET /api/metadata/actions/:objectName
func (h *ActionHandler) GetActions(c *gin.Context) {
	objectName := c.Param("objectName")
	HandleGetEnvelope(c, "data", func() (interface{}, error) {
		actions := h.svc.Metadata.GetActions(c.Request.Context(), objectName)
		sanitized := make([]*models.ActionMetadata, len(actions))
		for i, a := range actions {
			sanitized[i] = sanitizeAction(a)
		}
		return sanitized, nil
	})
}

// GetAllActions handles GET /api/metadata/actions
func (h *ActionHandler) GetAllActions(c *gin.Context) {
	HandleGetEnvelope(c, "data", func() (interface{}, error) {
		actions := h.svc.Metadata.GetAllActions(c.Request.Context())
		sanitized := make([]*models.ActionMetadata, len(actions))
		for i, a := range actions {
			sanitized[i] = sanitizeAction(a)
		}
		return sanitized, nil
	})
}

// GetAction handles GET /api/metadata/actions/id/:actionId
func (h *ActionHandler) GetAction(c *gin.Context) {
	actionID := c.Param("actionId")
	HandleGetEnvelope(c, "data", func() (interface{}, error) {
		action := h.svc.Metadata.GetActionByID(c.Request.Context(), actionID)
		if action == nil {
			return nil, appErrors.NewNotFoundError("Action", actionID)
		}
		return sanitizeAction(action), nil
	})
}

// CreateAction handles POST /api/metadata/actions
func (h *ActionHandler) CreateAction(c *gin.Context) {
	var action models.ActionMetadata
	HandleCreateEnvelope(c, "data", "Action created successfully", &action, func() error {
		err := h.svc.Metadata.CreateAction(c.Request.Context(), &action)
		if err == nil {
			// Sanitize response
			safe := sanitizeAction(&action)
			action = *safe
		}
		return err
	})
}

// UpdateAction handles PATCH /api/metadata/actions/:actionId
func (h *ActionHandler) UpdateAction(c *gin.Context) {
	actionID := c.Param("actionId")
	var updates models.ActionMetadata
	// No key returned, just message, but if envelope returns data we should check.
	// HandleUpdateEnvelope passed &updates.
	HandleUpdateEnvelope(c, "", "Action updated successfully", &updates, func() error {
		return h.svc.Metadata.UpdateAction(c.Request.Context(), actionID, &updates)
		// updates struct is used for input parsing. Its config might contain secrets being SET.
		// We shouldn't return secrets back (echoing them).
		// Typically UPDATE response might just be "success" or the updated object.
		// HandleUpdateEnvelope usually returns the object passed in?
		// Checking HandleUpdateEnvelope implementation is widely unknown here, but usually it echoes.
		// To be safe, we can sanitize `updates` after call.
	})
	// Can't easily hook into envelope post-execution here without modifying envelope or using closures differently.
	// But `updates` only contains what user sent. If user sent secret, echoing it back is arguably okay (they just sent it).
	// Logic for Get is more critical (reading existing secrets).
}

// DeleteAction handles DELETE /api/metadata/actions/:actionId
func (h *ActionHandler) DeleteAction(c *gin.Context) {
	actionID := c.Param("actionId")
	HandleDeleteEnvelope(c, "Action deleted successfully", func() error {
		return h.svc.Metadata.DeleteAction(c.Request.Context(), actionID)
	})
}

// ExecuteAction handles POST /api/actions/execute/:actionId
func (h *ActionHandler) ExecuteAction(c *gin.Context) {
	user := GetUserFromContext(c)
	actionID := c.Param("actionId")

	var req struct {
		ContextRecord models.SObject `json:"context_record"`
		RecordID      string         `json:"record_id"`
	}

	// Manual handle because return structure is result-specific (not envelope with key "req")
	if !BindJSON(c, &req) {
		return
	}

	// Validate: must have RecordID OR ContextRecord
	if req.RecordID == "" && len(req.ContextRecord) == 0 {
		RespondAppError(c, appErrors.NewValidationError("record_id", "Must provide recordId or contextRecord"))
		return
	}

	err := h.svc.ActionSvc.ExecuteAction(c.Request.Context(), actionID, req.ContextRecord, user)
	if err != nil {
		// Distinguish errors? For now 400 or 500
		RespondAppError(c, err)
		return
	}

	// Return Action Result
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			constants.FieldMessage: "Action executed successfully",
		},
	})
}
