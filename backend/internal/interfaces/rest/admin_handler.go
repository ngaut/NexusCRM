package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
)

// AdminHandler handles administrative endpoints
type AdminHandler struct {
	svc *services.ServiceManager
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(svc *services.ServiceManager) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// GetTableRegistry returns all registered tables
func (h *AdminHandler) GetTableRegistry(c *gin.Context) {
	HandleGetEnvelope(c, "data", func() (interface{}, error) {
		tables, err := h.svc.Schema.GetTableRegistry()
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"total":  len(tables),
			"tables": tables,
		}, nil
	})
}

// ValidateSchema runs validation and returns health status
func (h *AdminHandler) ValidateSchema(c *gin.Context) {
	HandleGetEnvelope(c, "data", func() (interface{}, error) {
		return h.svc.Schema.ValidateSchemaRegistry()
	})
}
