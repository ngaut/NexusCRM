package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/internal/domain/models"
)

type FeedHandler struct {
	svcMgr *services.ServiceManager
}

func NewFeedHandler(svcMgr *services.ServiceManager) *FeedHandler {
	return &FeedHandler{svcMgr: svcMgr}
}

// CreateComment handles POST /api/feed/comments
func (h *FeedHandler) CreateComment(c *gin.Context) {
	user := GetUserFromContext(c)
	var comment models.SystemComment

	HandleCreateEnvelope(c, "comment", "Comment added successfully", &comment, func() error {
		created, err := h.svcMgr.Feed.CreateComment(c.Request.Context(), comment, user)
		if err != nil {
			return err
		}
		comment = *created
		return nil
	})
}

// GetComments handles GET /api/feed/:recordId
func (h *FeedHandler) GetComments(c *gin.Context) {
	user := GetUserFromContext(c)
	recordID := c.Param("recordId")

	HandleGetEnvelope(c, "comments", func() (interface{}, error) {
		return h.svcMgr.Feed.GetComments(c.Request.Context(), recordID, user)
	})
}
