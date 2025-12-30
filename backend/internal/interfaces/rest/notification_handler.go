package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/shared/pkg/constants"
)

type NotificationHandler struct {
	svcMgr *services.ServiceManager
}

func NewNotificationHandler(svcMgr *services.ServiceManager) *NotificationHandler {
	return &NotificationHandler{svcMgr: svcMgr}
}

// GetNotifications handles GET /api/notifications
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	user := GetUserFromContext(c)

	HandleGetEnvelope(c, "notifications", func() (interface{}, error) {
		return h.svcMgr.Notification.GetMyNotifications(c.Request.Context(), user)
	})
}

// MarkAsRead handles POST /api/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	user := GetUserFromContext(c)
	id := c.Param("id")

	if err := h.svcMgr.Notification.MarkAsRead(c.Request.Context(), id, user); err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{constants.ResponseSuccess: true})
}
