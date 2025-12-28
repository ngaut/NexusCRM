package rest

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/pkg/auth"
)

type AnalyticsHandler struct {
	svc *services.ServiceManager
}

func NewAnalyticsHandler(svc *services.ServiceManager) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

type AdminQueryRequest struct {
	SQL    string        `json:"sql"`
	Params []interface{} `json:"params"`
}

func (h *AnalyticsHandler) ExecuteAdminQuery(c *gin.Context) {
	// Parse request
	var req AdminQueryRequest
	if !BindJSON(c, &req) {
		return
	}

	// Security check: This endpoint is strictly for System Admins.
	// We rely on the middleware "RequireSystemAdmin" to ensure this.
	// If the middleware is missing, this is a massive hole.
	// We could re-verify here, but middleware is the pattern.

	if req.SQL == "" {
		RespondError(c, http.StatusBadRequest, "SQL query cannot be empty")
		return
	}

	// Retrieve user from context (set by requireAuth middleware)
	userInterface, exists := c.Get("user")
	if !exists {
		RespondError(c, http.StatusUnauthorized, "User session not found")
		return
	}

	// Convert pkg/auth.UserSession to domain/models.UserSession
	authSession, ok := userInterface.(auth.UserSession)
	if !ok {
		RespondError(c, http.StatusInternalServerError, "Invalid user session type")
		return
	}

	userSession := &models.UserSession{
		ID:        authSession.ID,
		Name:      authSession.Name,
		Email:     &authSession.Email, // auth uses string, models uses *string
		ProfileID: authSession.ProfileId,
		RoleID:    authSession.RoleId,
	}

	// Execute raw SQL using QueryService for security/safety
	results, err := h.svc.QuerySvc.ExecuteRawSQL(
		c.Request.Context(),
		req.SQL,
		req.Params,
		userSession,
	)
	if err != nil {
		RespondError(c, http.StatusBadRequest, fmt.Sprintf("Query execution failed: %v", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"results": results,
	})
}
