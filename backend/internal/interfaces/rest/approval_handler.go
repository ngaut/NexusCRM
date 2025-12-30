package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/shared/pkg/constants"
)

// ApprovalHandler handles approval process API endpoints
type ApprovalHandler struct {
	svc *services.ServiceManager
}

// NewApprovalHandler creates a new ApprovalHandler
func NewApprovalHandler(svc *services.ServiceManager) *ApprovalHandler {
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

	workItem, err := h.svc.Approval.Submit(c.Request.Context(), serviceReq, user)
	if err != nil {
		// Map common errors to status codes
		if err.Error() == "no active approval process found for this object" ||
			err.Error() == "record already has a pending approval" {
			RespondError(c, 400, err.Error())
			return
		}
		if err.Error() == "you don't have permission to submit this record for approval" {
			RespondError(c, 403, err.Error())
			return
		}
		RespondError(c, 500, "Failed to submit for approval: "+err.Error())
		return
	}

	c.JSON(200, gin.H{
		"success":      true,
		constants.FieldMessage:      "Record submitted for approval",
		"work_item_id": workItem[constants.FieldID],
	})
}

// Approve handles POST /api/approvals/:workItemId/approve
func (h *ApprovalHandler) Approve(c *gin.Context) {
	workItemID := c.Param("workItemId")
	user := GetUserFromContext(c)

	var req ApprovalActionRequest
	_ = c.ShouldBindJSON(&req) // Optional comments

	err := h.svc.Approval.Approve(c.Request.Context(), workItemID, req.Comments, user)
	if err != nil {
		handleApprovalError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		constants.FieldMessage: "Approval granted",
	})
}

// Reject handles POST /api/approvals/:workItemId/reject
func (h *ApprovalHandler) Reject(c *gin.Context) {
	workItemID := c.Param("workItemId")
	user := GetUserFromContext(c)

	var req ApprovalActionRequest
	_ = c.ShouldBindJSON(&req) // Optional comments

	err := h.svc.Approval.Reject(c.Request.Context(), workItemID, req.Comments, user)
	if err != nil {
		handleApprovalError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		constants.FieldMessage: "Approval rejected",
	})
}

// GetPending handles GET /api/approvals/pending
func (h *ApprovalHandler) GetPending(c *gin.Context) {
	user := GetUserFromContext(c)
	// Delegate to service
	items, err := h.svc.Approval.GetPending(c.Request.Context(), user)
	if err != nil {
		RespondError(c, 500, "Failed to fetch pending approvals: "+err.Error())
		return
	}

	c.JSON(200, gin.H{
		"work_items": items,
	})
}

// GetHistory handles GET /api/approvals/history/:objectApiName/:recordId
func (h *ApprovalHandler) GetHistory(c *gin.Context) {
	user := GetUserFromContext(c)
	objectAPIName := c.Param("objectApiName")
	recordID := c.Param("recordId")
	// Delegate to service
	items, err := h.svc.Approval.GetHistory(c.Request.Context(), objectAPIName, recordID, user)
	if err != nil {
		RespondError(c, 500, "Failed to fetch approval history: "+err.Error())
		return
	}

	c.JSON(200, gin.H{
		"work_items": items,
	})
}

// CheckProcess handles GET /api/approvals/check/:objectApiName
func (h *ApprovalHandler) CheckProcess(c *gin.Context) {
	user := GetUserFromContext(c)
	objectAPIName := c.Param("objectApiName")
	// Delegate to service
	process, err := h.svc.Approval.CheckProcess(c.Request.Context(), objectAPIName, user)
	if err != nil {
		// Log error but generally we just want to know if it exists or not
	}

	processName := ""
	if process != nil {
		if name, ok := process[constants.FieldName].(string); ok {
			processName = name
		}
	}

	c.JSON(200, gin.H{
		"has_process":  process != nil,
		"process_name": processName,
	})
}

// GetFlowProgress handles GET /api/approvals/flow-progress/:instanceId
func (h *ApprovalHandler) GetFlowProgress(c *gin.Context) {
	instanceID := c.Param("instanceId")
	user := GetUserFromContext(c)

	progress, err := h.svc.Approval.GetFlowProgress(c.Request.Context(), instanceID, user)
	if err != nil {
		RespondError(c, 500, "Failed to get flow progress: "+err.Error())
		return
	}
	if progress == nil {
		RespondError(c, 404, "Flow instance not found")
		return
	}

	c.JSON(200, progress)
}

// Private helpers

func handleApprovalError(c *gin.Context, err error) {
	msg := err.Error()
	if msg == "approval work item not found" {
		RespondError(c, 404, msg)
		return
	}
	if msg == "work item is not pending" {
		RespondError(c, 400, msg)
		return
	}
	if msg == "you are not authorized to approved this item" ||
		msg == "you are not authorized to rejected this item" {
		RespondError(c, 403, msg)
		return
	}
	RespondError(c, 500, msg)
}
