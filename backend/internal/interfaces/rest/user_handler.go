package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

type UserHandler struct {
	svcMgr *services.ServiceManager
}

func NewUserHandler(svcMgr *services.ServiceManager) *UserHandler {
	return &UserHandler{
		svcMgr: svcMgr,
	}
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	ProfileId string `json:"profile_id"`
}

// Register handles POST /api/auth/register
// SECURITY: Should be restricted to System Admins only (handled by middleware)
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	// We manually handle binding and response here because the RegisterRequest input
	// differs significantly from the UserSession output, making it hard to use the
	// standard HandleCreateEnvelope helper which expects a unified model.

	if !BindJSON(c, &req) {
		return
	}

	user, err := h.svcMgr.Auth.CreateUser(c.Request.Context(), services.CreateUserRequest{
		Name:      req.Name,
		Email:     req.Email,
		Password:  req.Password,
		ProfileID: req.ProfileId,
	})

	if err != nil {
		RespondError(c, errors.GetHTTPStatus(err), err.Error())
		return
	}

	// We can manually construct the success response to match envelope style
	c.JSON(http.StatusCreated, gin.H{
		constants.FieldMessage: "User created successfully",
		"user": gin.H{
			constants.FieldID:        user.ID,
			constants.FieldName:      user.Name,
			constants.FieldEmail:     user.Email,
			constants.FieldProfileID: user.ProfileID,
		},
	})
}

// UpdateUserRequest represents update user request
type UpdateUserRequest struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	ProfileId string `json:"profile_id"`
	IsActive  *bool  `json:"is_active"`
}

// UpdateUser handles PUT /api/auth/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param(constants.FieldID)

	var req UpdateUserRequest
	HandleUpdateEnvelope(c, "", "User updated successfully", &req, func() error {
		if userID == "" {
			return errors.NewValidationError(constants.FieldID, "is required")
		}
		return h.svcMgr.Auth.UpdateUser(c.Request.Context(), userID, services.UpdateUserRequest{
			Name:      req.Name,
			Email:     req.Email,
			Password:  req.Password,
			ProfileID: req.ProfileId,
			IsActive:  req.IsActive,
		})
	})
}

// DeleteUser handles DELETE /api/auth/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param(constants.FieldID)
	HandleDeleteEnvelope(c, "User deleted successfully", func() error {
		if userID == "" {
			return errors.NewValidationError(constants.FieldID, "is required")
		}
		return h.svcMgr.Auth.DeleteUser(c.Request.Context(), userID)
	})
}

// GetUsers handles GET /api/auth/users
func (h *UserHandler) GetUsers(c *gin.Context) {
	HandleGetEnvelope(c, "users", func() (interface{}, error) {
		return h.svcMgr.Auth.GetUsers(c.Request.Context())
	})
}

// GetProfiles handles GET /api/auth/profiles
func (h *UserHandler) GetProfiles(c *gin.Context) {
	HandleGetEnvelope(c, "profiles", func() (interface{}, error) {
		return h.svcMgr.Auth.GetProfiles(c.Request.Context())
	})
}

// GetProfilePermissions handles GET /api/auth/profiles/:id/permissions
func (h *UserHandler) GetProfilePermissions(c *gin.Context) {
	profileID := c.Param(constants.FieldID)
	HandleGetEnvelope(c, "permissions", func() (interface{}, error) {
		return h.svcMgr.Permissions.GetObjectPermissions(profileID)
	})
}

// UpdateProfilePermissions handles PUT /api/auth/profiles/:id/permissions
func (h *UserHandler) UpdateProfilePermissions(c *gin.Context) {
	profileID := c.Param(constants.FieldID)

	var perms []struct {
		ObjectAPIName string `json:"object_api_name"`
		AllowRead     bool   `json:"allow_read"`
		AllowCreate   bool   `json:"allow_create"`
		AllowEdit     bool   `json:"allow_edit"`
		AllowDelete   bool   `json:"allow_delete"`
		ViewAll       bool   `json:"view_all"`
		ModifyAll     bool   `json:"modify_all"`
	}

	// Helper handles binding
	HandleUpdateEnvelope(c, "", "Permissions updated successfully", &perms, func() error {
		for _, p := range perms {
			perm := models.SystemObjectPerms{
				ProfileID:     &profileID,
				ObjectAPIName: p.ObjectAPIName,
				AllowRead:     p.AllowRead,
				AllowCreate:   p.AllowCreate,
				AllowEdit:     p.AllowEdit,
				AllowDelete:   p.AllowDelete,
				ViewAll:       p.ViewAll,
				ModifyAll:     p.ModifyAll,
			}
			if err := h.svcMgr.Permissions.UpdateObjectPermission(perm); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetProfileFieldPermissions handles GET /api/auth/profiles/:id/permissions/fields
func (h *UserHandler) GetProfileFieldPermissions(c *gin.Context) {
	profileID := c.Param(constants.FieldID)
	HandleGetEnvelope(c, "permissions", func() (interface{}, error) {
		return h.svcMgr.Permissions.GetFieldPermissions(profileID)
	})
}

// UpdateProfileFieldPermissions handles PUT /api/auth/profiles/:id/permissions/fields
func (h *UserHandler) UpdateProfileFieldPermissions(c *gin.Context) {
	profileID := c.Param(constants.FieldID)

	var perms []struct {
		ObjectAPIName string `json:"object_api_name"`
		FieldAPIName  string `json:"field_api_name"`
		AllowRead     bool   `json:"allow_read"`
		AllowEdit     bool   `json:"allow_edit"`
	}

	HandleUpdateEnvelope(c, "", "Field permissions updated successfully", &perms, func() error {
		for _, p := range perms {
			perm := models.SystemFieldPerms{
				ProfileID:     &profileID,
				ObjectAPIName: p.ObjectAPIName,
				FieldAPIName:  p.FieldAPIName,
				Readable:      p.AllowRead,
				Editable:      p.AllowEdit,
			}
			if err := h.svcMgr.Permissions.UpdateFieldPermission(perm); err != nil {
				return err
			}
		}
		return nil
	})
}

// ==================== Permission Set Permissions ====================

// GetPermissionSetPermissions handles GET /api/auth/permission-sets/:id/permissions
func (h *UserHandler) GetPermissionSetPermissions(c *gin.Context) {
	permSetID := c.Param(constants.FieldID)
	HandleGetEnvelope(c, "permissions", func() (interface{}, error) {
		return h.svcMgr.Permissions.GetPermissionSetObjectPermissions(permSetID)
	})
}

// UpdatePermissionSetPermissions handles PUT /api/auth/permission-sets/:id/permissions
func (h *UserHandler) UpdatePermissionSetPermissions(c *gin.Context) {
	permSetID := c.Param(constants.FieldID)

	var perms []struct {
		ObjectAPIName string `json:"object_api_name"`
		AllowRead     bool   `json:"allow_read"`
		AllowCreate   bool   `json:"allow_create"`
		AllowEdit     bool   `json:"allow_edit"`
		AllowDelete   bool   `json:"allow_delete"`
		ViewAll       bool   `json:"view_all"`
		ModifyAll     bool   `json:"modify_all"`
	}

	HandleUpdateEnvelope(c, "", "Permissions updated successfully", &perms, func() error {
		for _, p := range perms {
			perm := models.SystemObjectPerms{
				PermissionSetID: &permSetID,
				ObjectAPIName:   p.ObjectAPIName,
				AllowRead:       p.AllowRead,
				AllowCreate:     p.AllowCreate,
				AllowEdit:       p.AllowEdit,
				AllowDelete:     p.AllowDelete,
				ViewAll:         p.ViewAll,
				ModifyAll:       p.ModifyAll,
			}
			if err := h.svcMgr.Permissions.UpdatePermissionSetObjectPermission(perm); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetPermissionSetFieldPermissions handles GET /api/auth/permission-sets/:id/permissions/fields
func (h *UserHandler) GetPermissionSetFieldPermissions(c *gin.Context) {
	permSetID := c.Param(constants.FieldID)
	HandleGetEnvelope(c, "permissions", func() (interface{}, error) {
		return h.svcMgr.Permissions.GetPermissionSetFieldPermissions(permSetID)
	})
}

// UpdatePermissionSetFieldPermissions handles PUT /api/auth/permission-sets/:id/permissions/fields
func (h *UserHandler) UpdatePermissionSetFieldPermissions(c *gin.Context) {
	permSetID := c.Param(constants.FieldID)

	var perms []struct {
		ObjectAPIName string `json:"object_api_name"`
		FieldAPIName  string `json:"field_api_name"`
		AllowRead     bool   `json:"allow_read"`
		AllowEdit     bool   `json:"allow_edit"`
	}

	HandleUpdateEnvelope(c, "", "Field permissions updated successfully", &perms, func() error {
		for _, p := range perms {
			perm := models.SystemFieldPerms{
				PermissionSetID: &permSetID,
				ObjectAPIName:   p.ObjectAPIName,
				FieldAPIName:    p.FieldAPIName,
				Readable:        p.AllowRead,
				Editable:        p.AllowEdit,
			}
			if err := h.svcMgr.Permissions.UpdatePermissionSetFieldPermission(perm); err != nil {
				return err
			}
		}
		return nil
	})
}

// ==================== Effective Permissions (Admin View) ====================

// GetUserEffectivePermissions handles GET /api/auth/users/:id/permissions/effective
func (h *UserHandler) GetUserEffectivePermissions(c *gin.Context) {
	userID := c.Param(constants.FieldID)
	HandleGetEnvelope(c, "permissions", func() (interface{}, error) {
		return h.svcMgr.Permissions.GetEffectiveObjectPermissions(userID)
	})
}

// GetUserEffectiveFieldPermissions handles GET /api/auth/users/:id/permissions/fields/effective
func (h *UserHandler) GetUserEffectiveFieldPermissions(c *gin.Context) {
	userID := c.Param(constants.FieldID)
	HandleGetEnvelope(c, "permissions", func() (interface{}, error) {
		return h.svcMgr.Permissions.GetEffectiveFieldPermissions(userID)
	})
}

// ==================== Permission Set Management (CRUD) ====================

type CreatePermissionSetRequest struct {
	Name        string `json:"name" binding:"required"`
	Label       string `json:"label" binding:"required"`
	Description string `json:"description"`
}

type UpdatePermissionSetRequest struct {
	Name        string `json:"name" binding:"required"`
	Label       string `json:"label" binding:"required"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

// CreatePermissionSet handles POST /api/auth/permission-sets
func (h *UserHandler) CreatePermissionSet(c *gin.Context) {
	var req CreatePermissionSetRequest
	if !BindJSON(c, &req) {
		return
	}

	id, err := h.svcMgr.Permissions.CreatePermissionSet(req.Name, req.Label, req.Description)
	if err != nil {
		RespondError(c, errors.GetHTTPStatus(err), err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		constants.FieldMessage: "Permission Set created successfully",
		"permission_set": gin.H{
			constants.FieldID:   id,
			constants.FieldName: req.Name,
			"label":             req.Label,
			"description":       req.Description,
			"is_active":         true,
		},
	})
}

// UpdatePermissionSet handles PUT /api/auth/permission-sets/:id
func (h *UserHandler) UpdatePermissionSet(c *gin.Context) {
	id := c.Param("id")
	var req UpdatePermissionSetRequest
	HandleUpdateEnvelope(c, "", "Permission Set updated successfully", &req, func() error {
		return h.svcMgr.Permissions.UpdatePermissionSet(id, req.Name, req.Label, req.Description, req.IsActive)
	})
}

// DeletePermissionSet handles DELETE /api/auth/permission-sets/:id
func (h *UserHandler) DeletePermissionSet(c *gin.Context) {
	id := c.Param("id")
	HandleDeleteEnvelope(c, "Permission Set deleted successfully", func() error {
		return h.svcMgr.Permissions.DeletePermissionSet(id)
	})
}
