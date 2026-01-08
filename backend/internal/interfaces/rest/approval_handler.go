package rest

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	appErrors "github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ApprovalService defines the interface for approval operations
type ApprovalService interface {
	Submit(ctx context.Context, req services.SubmitRequest, user *models.UserSession) (models.SObject, error)
	Approve(ctx context.Context, workItemID, comments string, user *models.UserSession) error
	Reject(ctx context.Context, workItemID, comments string, user *models.UserSession) error
	GetPending(ctx context.Context, user *models.UserSession) ([]models.SObject, error)
	GetHistory(ctx context.Context, objectAPIName, recordID string, user *models.UserSession) ([]models.SObject, error)
	CheckProcess(ctx context.Context, objectAPIName string, user *models.UserSession) (models.SObject, error)
	GetFlowProgress(ctx context.Context, instanceID string, user *models.UserSession) (*services.FlowInstanceProgress, error)
}

// ApprovalHandler handles approval process API endpoints
type ApprovalHandler struct {
	svc ApprovalService
}

// NewApprovalHandler creates a new ApprovalHandler
func NewApprovalHandler(svc ApprovalService) *ApprovalHandler {
	return &ApprovalHandler{svc: svc}
}

// ============================================================================
// Request/Response Types
// ============================================================================

// SubmitRequest represents a request to submit a record for approval
type SubmitRequest struct {
	ObjectAPIName string `json:"object_api_name" binding:"required"`
	RecordID      string `json:"record_id" binding:"required"`
	Comments      string `json:"comments"`
}

// ApprovalActionRequest represents an approve/reject request
type ApprovalActionRequest struct {
	Comments string `json:"comments"`
}

// ============================================================================
// Public Endpoints
// ============================================================================

// Submit handles POST /api/approvals/submit
func (h *ApprovalHandler) Submit(c *gin.Context) {
	user := GetUserFromContext(c)

	var req SubmitRequest
	if !BindJSON(c, &req) {
		return
	}

	serviceReq := services.SubmitRequest{
		ObjectAPIName: req.ObjectAPIName,
		RecordID:      req.RecordID,
		Comments:      req.Comments,
	}

	workItem, err := h.svc.Submit(c.Request.Context(), serviceReq, user)
	if err != nil {
		// Map common errors to status codes
		if err.Error() == "no active approval process found for this object" ||
			err.Error() == "record already has a pending approval" {
			RespondAppError(c, appErrors.NewValidationError("approval", err.Error()))
			return
		}
		if err.Error() == "you don't have permission to submit this record for approval" {
			RespondAppError(c, appErrors.NewPermissionError("submit_approval", req.RecordID))
			return
		}
		RespondAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"success":              true,
			constants.FieldMessage: "Record submitted for approval",
			"work_item_id":         workItem[constants.FieldID],
		},
	})
}

// Approve handles POST /api/approvals/:workItemId/approve
func (h *ApprovalHandler) Approve(c *gin.Context) {
	workItemID := c.Param("workItemId")
	user := GetUserFromContext(c)

	var req ApprovalActionRequest
	_ = c.ShouldBindJSON(&req) // Optional comments

	err := h.svc.Approve(c.Request.Context(), workItemID, req.Comments, user)
	if err != nil {
		RespondAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"success":              true,
			constants.FieldMessage: "Approval granted",
		},
	})
}

// Reject handles POST /api/approvals/:workItemId/reject
func (h *ApprovalHandler) Reject(c *gin.Context) {
	workItemID := c.Param("workItemId")
	user := GetUserFromContext(c)

	var req ApprovalActionRequest
	_ = c.ShouldBindJSON(&req) // Optional comments

	err := h.svc.Reject(c.Request.Context(), workItemID, req.Comments, user)
	if err != nil {
		RespondAppError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"success":              true,
			constants.FieldMessage: "Approval rejected",
		},
	})
}

// GetPending handles GET /api/approvals/pending
func (h *ApprovalHandler) GetPending(c *gin.Context) {
	user := GetUserFromContext(c)
	HandleGetEnvelope(c, "data", func() (interface{}, error) {
		return h.svc.GetPending(c.Request.Context(), user)
	})
}

// GetHistory handles GET /api/approvals/history/:objectApiName/:recordId
func (h *ApprovalHandler) GetHistory(c *gin.Context) {
	user := GetUserFromContext(c)
	objectAPIName := c.Param("objectApiName")
	recordID := c.Param("recordId")
	HandleGetEnvelope(c, "data", func() (interface{}, error) {
		return h.svc.GetHistory(c.Request.Context(), objectAPIName, recordID, user)
	})
}

// CheckProcess handles GET /api/approvals/check/:objectApiName
func (h *ApprovalHandler) CheckProcess(c *gin.Context) {
	user := GetUserFromContext(c)
	objectAPIName := c.Param("objectApiName")

	HandleGetEnvelope(c, "data", func() (interface{}, error) {
		process, err := h.svc.CheckProcess(c.Request.Context(), objectAPIName, user)
		if err != nil {
			log.Printf("Warning: failed to check approval process for %s: %v", objectAPIName, err)
			// Don't error out, just return false
		}

		processName := ""
		if process != nil {
			if name, ok := process[constants.FieldName].(string); ok {
				processName = name
			}
		}

		return gin.H{
			"has_process":  process != nil,
			"process_name": processName,
		}, nil
	})
}

// GetFlowProgress handles GET /api/approvals/flow-progress/:instanceId
func (h *ApprovalHandler) GetFlowProgress(c *gin.Context) {
	instanceID := c.Param("instanceId")
	user := GetUserFromContext(c)

	HandleGetEnvelope(c, "data", func() (interface{}, error) {
		progress, err := h.svc.GetFlowProgress(c.Request.Context(), instanceID, user)
		if err != nil {
			return nil, err
		}
		if progress == nil {
			return nil, appErrors.NewNotFoundError("Flow Instance", instanceID)
		}
		return progress, nil
	})
}

// Private helpers

func handleApprovalError(c *gin.Context, err error) {
	msg := err.Error()
	if msg == "approval work item not found" {
		RespondAppError(c, appErrors.NewNotFoundError("Approval Work Item", "unknown"))
		return
	}
	if msg == "work item is not pending" {
		RespondAppError(c, appErrors.NewValidationError("status", msg))
		return
	}
	if msg == "you are not authorized to approve this item" ||
		msg == "you are not authorized to reject this item" {
		RespondAppError(c, appErrors.NewPermissionError("approve/reject", "Work Item"))
		return
	}
	RespondAppError(c, appErrors.NewInternalError(msg, err))
}
