package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/shared/pkg/constants"
)

type RoleHandler struct {
	svcMgr *services.ServiceManager
}

func NewRoleHandler(svcMgr *services.ServiceManager) *RoleHandler {
	return &RoleHandler{
		svcMgr: svcMgr,
	}
}

// CreateRoleRequest represents the payload for creating a role
type CreateRoleRequest struct {
	Name         string  `json:"name" binding:"required"`
	Description  string  `json:"description"`
	ParentRoleID *string `json:"parent_role_id"`
}

// CreateRole handles POST /api/auth/roles
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	// Note: HandleCreateEnvelope handles binding, but we need the result.
	// We can use a modified approach or just manual with RespondAppError?
	// HandleCreateEnvelope signature: (c, key, msg, obj, action)
	// Action returns error only.
	// We need 'createdRole' returned.
	// Let's use manual binding + RespondAppError + Standard Envelope manually for now to be safe,
	// OR better: Update CreateRole service to populate the req object? No.

	if !BindJSON(c, &req) {
		return
	}

	createdRole, err := h.svcMgr.Permissions.CreateRole(c.Request.Context(), req.Name, req.Description, req.ParentRoleID)
	if err != nil {
		RespondAppError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		constants.FieldMessage: "Role created successfully",
		"data":                 createdRole,
	})
}

// GetRoles handles GET /api/auth/roles
func (h *RoleHandler) GetRoles(c *gin.Context) {
	HandleGetEnvelope(c, "data", func() (interface{}, error) {
		return h.svcMgr.Permissions.GetAllRoles(c.Request.Context())
	})
}

// GetRole handles GET /api/auth/roles/:id
func (h *RoleHandler) GetRole(c *gin.Context) {
	id := c.Param(constants.FieldID)
	HandleGetEnvelope(c, "data", func() (interface{}, error) {
		return h.svcMgr.Permissions.GetRole(c.Request.Context(), id)
	})
}

// UpdateRoleRequest represents the payload for updating a role
type UpdateRoleRequest struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	ParentRoleID *string `json:"parent_role_id"`
}

// UpdateRole handles PUT /api/auth/roles/:id
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id := c.Param(constants.FieldID)
	var req UpdateRoleRequest
	HandleUpdateEnvelope(c, "", "Role updated successfully", &req, func() error {
		_, err := h.svcMgr.Permissions.UpdateRole(c.Request.Context(), id, req.Name, req.Description, req.ParentRoleID)
		return err
	})
}

// DeleteRole handles DELETE /api/auth/roles/:id
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id := c.Param(constants.FieldID)
	HandleDeleteEnvelope(c, "Role deleted successfully", func() error {
		return h.svcMgr.Permissions.DeleteRole(c.Request.Context(), id)
	})
}
