package rest

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/pkg/auth"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/constants"
)

type AuthHandler struct {
	svcMgr *services.ServiceManager
}

func NewAuthHandler(svcMgr *services.ServiceManager) *AuthHandler {
	return &AuthHandler{
		svcMgr: svcMgr,
	}
}

// LoginRequest represents login request body
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Success   bool                   `json:"success"`
	Token     string                 `json:"token,omitempty"`
	User      map[string]interface{} `json:"user,omitempty"`
	ExpiresAt string                 `json:"expires_at,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Message   string                 `json:"message,omitempty"`
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if !BindJSON(c, &req) {
		return
	}

	// Validate email format
	if !auth.IsValidEmail(req.Email) {
		RespondError(c, http.StatusBadRequest, "Invalid email format")
		return
	}

	// Delegate to AuthService
	result, err := h.svcMgr.Auth.Login(req.Email, req.Password, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		RespondError(c, errors.GetHTTPStatus(err), err.Error())
		return
	}

	// Return response
	userData := map[string]interface{}{
		constants.FieldID:        result.User.ID,
		constants.FieldName:      result.User.Name,
		constants.FieldEmail:     result.User.Email,
		constants.FieldProfileID: result.User.ProfileId,
	}

	// Always include roleId for consistent API contract (value or null)
	if result.User.RoleId != nil {
		userData[constants.FieldRoleID] = *result.User.RoleId
	} else {
		userData[constants.FieldRoleID] = nil
	}

	c.JSON(http.StatusOK, LoginResponse{
		Success:   true,
		Token:     result.Token,
		User:      userData,
		ExpiresAt: result.ExpiresAt.Format(time.RFC3339),
	})
}

// Logout handles POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get token from context (set by auth middleware)
	tokenString, exists := c.Get(constants.ContextKeyToken)
	if !exists {
		RespondError(c, http.StatusUnauthorized, "No token provided")
		return
	}

	HandleDeleteEnvelope(c, "Logged out successfully", func() error {
		return h.svcMgr.Auth.Logout(tokenString.(string))
	})
}

// GetMe handles GET /api/auth/me
func (h *AuthHandler) GetMe(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userInterface, exists := c.Get(constants.ContextKeyUser)
	if !exists {
		RespondError(c, http.StatusUnauthorized, "User not found")
		return
	}

	user := userInterface.(auth.UserSession)

	// Use HandleGetEnvelope? It expects action to return (interface{}, error)
	// Here we already have the data locally.
	// But to be consistent with envelope format { "user": ... }
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			constants.FieldID:        user.ID,
			constants.FieldName:      user.Name,
			constants.FieldEmail:     user.Email,
			constants.FieldProfileID: user.ProfileId,
			constants.FieldRoleID:    user.RoleId,
		},
	})
}

// ChangePasswordRequest represents change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// ChangePassword handles POST /api/auth/change-password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	HandleUpdateEnvelope(c, "", "Password changed successfully", &req, func() error {
		// Get user from context
		userInterface, exists := c.Get(constants.ContextKeyUser)
		if !exists {
			return errors.NewUnauthorizedError("User not found")
		}
		user := userInterface.(auth.UserSession)

		return h.svcMgr.Auth.ChangePassword(user.ID, req.CurrentPassword, req.NewPassword)
	})
}

// GetMyPermissions handles GET /api/auth/permissions/me
func (h *AuthHandler) GetMyPermissions(c *gin.Context) {
	userInterface, exists := c.Get(constants.ContextKeyUser)
	if !exists {
		RespondError(c, http.StatusUnauthorized, "User not found")
		return
	}
	user := userInterface.(auth.UserSession)

	// Refresh permissions first to ensure we have latest
	if err := h.svcMgr.Permissions.RefreshPermissions(); err != nil {
		log.Printf("Failed to refresh permissions: %v", err)
	}

	perms, err := h.svcMgr.Permissions.GetEffectiveObjectPermissions(user.ID)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	fieldPerms, err := h.svcMgr.Permissions.GetEffectiveFieldPermissions(user.ID)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Return flat JSON response (not using HandleGetEnvelope which wraps in a key)
	c.JSON(http.StatusOK, gin.H{
		"objectPermissions": perms,
		"fieldPermissions":  fieldPerms,
	})
}
