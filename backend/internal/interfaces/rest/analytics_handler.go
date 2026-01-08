package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/pkg/auth"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
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
		RespondAppError(c, errors.NewValidationError("sql", "SQL query cannot be empty"))
		return
	}

	// Retrieve user from context (set by requireAuth middleware)
	userInterface, exists := c.Get(constants.ContextKeyUser)
	if !exists {
		RespondAppError(c, errors.NewUnauthorizedError("User session not found"))
		return
	}

	// Convert pkg/auth.UserSession to domain/models.UserSession
	authSession, ok := userInterface.(auth.UserSession)
	if !ok {
		RespondAppError(c, errors.NewInternalError("Invalid user session type", nil))
		return
	}

	userSession := &models.UserSession{
		ID:        authSession.ID,
		Name:      authSession.Name,
		Email:     &authSession.Email, // auth uses string, models uses *string
		ProfileID: authSession.ProfileId,
		RoleID:    authSession.RoleId,
	}

	HandleGetEnvelope(c, "data", func() (interface{}, error) {
		return h.svc.QuerySvc.ExecuteRawSQL(
			c.Request.Context(),
			req.SQL,
			req.Params,
			userSession,
		)
	})
}
