package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
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

	HandleGetEnvelope(c, "data", func() (interface{}, error) {
		return h.svcMgr.Notification.GetMyNotifications(c.Request.Context(), user)
	})
}

// MarkAsRead handles POST /api/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	user := GetUserFromContext(c)
	id := c.Param("id")

	var req struct{}
	HandleUpdateEnvelope(c, "", "Notification marked as read", &req, func() error {
		return h.svcMgr.Notification.MarkAsRead(c.Request.Context(), id, user)
	})
}
