package rest

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	appErrors "github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

type UIHandler struct {
	svc *services.ServiceManager
}

func NewUIHandler(svc *services.ServiceManager) *UIHandler {
	return &UIHandler{svc: svc}
}

// ==================== App Handlers ====================

// GetApps handles GET /api/metadata/apps
func (h *UIHandler) GetApps(c *gin.Context) {
	HandleGetEnvelope(c, "apps", func() (interface{}, error) {
		return h.svc.UIMetadata.GetApps(c.Request.Context()), nil
	})
}

// CreateApp handles POST /api/metadata/apps
func (h *UIHandler) CreateApp(c *gin.Context) {
	// requireSystemAdmin handled by middleware

	var req models.AppConfig

	HandleCreateEnvelope(c, "app", "App created successfully", &req, func() error {
		// Set default values
		if req.NavigationItems == nil {
			req.NavigationItems = []models.NavigationItem{}
		}

		return h.svc.UIMetadata.CreateApp(c.Request.Context(), &req)
	})
}

// UpdateApp handles PATCH /api/metadata/apps/:id
func (h *UIHandler) UpdateApp(c *gin.Context) {
	// requireSystemAdmin handled by middleware

	id := c.Param(constants.FieldID)
	var updates models.AppConfig

	HandleUpdateEnvelope(c, "", "App updated successfully", &updates, func() error {
		return h.svc.UIMetadata.UpdateApp(c.Request.Context(), id, &updates)
	})
}

// DeleteApp handles DELETE /api/metadata/apps/:id
func (h *UIHandler) DeleteApp(c *gin.Context) {
	// requireSystemAdmin handled by middleware

	id := c.Param(constants.FieldID)

	HandleDeleteEnvelope(c, "App deleted successfully", func() error {
		// Prevent deleting standard apps if needed
		// For now, allow admin to delete any app except maybe "standard" ones.
		// Service layer should handle specific business rules.
		return h.svc.UIMetadata.DeleteApp(c.Request.Context(), id)
	})
}

// GetActiveTheme handles GET /api/metadata/theme
func (h *UIHandler) GetActiveTheme(c *gin.Context) {
	HandleGetEnvelope(c, "theme", func() (interface{}, error) {
		return h.svc.UIMetadata.GetActiveTheme(c.Request.Context())
	})
}

// CreateTheme handles POST /api/metadata/themes
func (h *UIHandler) CreateTheme(c *gin.Context) {
	// requireSystemAdmin handled by middleware

	var req models.Theme

	HandleCreateEnvelope(c, "theme", "Theme created successfully", &req, func() error {
		return h.svc.UIMetadata.UpsertTheme(c.Request.Context(), &req)
	})
}

// ActivateTheme handles PUT /api/metadata/themes/:id/activate
func (h *UIHandler) ActivateTheme(c *gin.Context) {
	// requireSystemAdmin handled by middleware

	id := c.Param(constants.FieldID)
	var req struct{}
	HandleUpdateEnvelope(c, "", "Theme activated successfully", &req, func() error {
		return h.svc.UIMetadata.ActivateTheme(c.Request.Context(), id)
	})
}

// ==================== Layout Handlers ====================

// GetLayout handles GET /api/metadata/layouts/:objectName
func (h *UIHandler) GetLayout(c *gin.Context) {
	user := GetUserFromContext(c)
	objectName := c.Param("objectName")
	// Get layout with profile-specific logic
	var profileID *string
	if user != nil {
		profileID = &user.ProfileID
	}
	layout := h.svc.UIMetadata.GetLayout(c.Request.Context(), objectName, profileID)
	if layout == nil {
		c.JSON(http.StatusNotFound, gin.H{
			constants.FieldMessage: "Layout for '" + objectName + "' not found",
		})
		return
	}

	// Filter by type if specified
	// if layoutType != "" && layout.Type != layoutType {
	// 	// Try to find a layout with matching type
	// 	// For now, just return the default layout regardless of type
	// }

	c.JSON(http.StatusOK, gin.H{constants.ResponseLayout: layout})
}

// SaveLayout handles POST /api/metadata/layouts
func (h *UIHandler) SaveLayout(c *gin.Context) {
	var layout models.PageLayout
	HandleCreateEnvelope(c, "layout", "Layout saved successfully", &layout, func() error {
		// Validate required fields
		if layout.ID == "" {
			return appErrors.NewValidationError(constants.FieldID, "Layout ID is required")
		}
		if layout.ObjectAPIName == "" {
			return appErrors.NewValidationError(constants.FieldObjectAPIName, "Object API name is required")
		}
		return h.svc.UIMetadata.SaveLayout(c.Request.Context(), &layout)
	})
}

// DeleteLayout handles DELETE /api/metadata/layouts/:id
func (h *UIHandler) DeleteLayout(c *gin.Context) {
	layoutID := c.Param(constants.FieldID)
	HandleDeleteEnvelope(c, "Layout deleted successfully", func() error {
		return h.svc.UIMetadata.DeleteLayout(c.Request.Context(), layoutID)
	})
}

// AssignLayoutToProfile handles POST /api/metadata/layouts/assign
func (h *UIHandler) AssignLayoutToProfile(c *gin.Context) {
	var req struct {
		ProfileID     string `json:"profile_id" binding:"required"`
		ObjectAPIName string `json:"object_api_name" binding:"required"`
		LayoutID      string `json:"layout_id" binding:"required"`
	}

	if !BindJSON(c, &req) {
		return
	}

	if err := h.svc.UIMetadata.AssignLayoutToProfile(c.Request.Context(), req.ProfileID, req.ObjectAPIName, req.LayoutID); err != nil {
		RespondError(c, http.StatusInternalServerError, "Failed to assign layout: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		constants.FieldMessage: "Layout assigned to profile successfully",
	})
}

// ==================== Dashboard Handlers ====================

// GetDashboards handles GET /api/metadata/dashboards
func (h *UIHandler) GetDashboards(c *gin.Context) {
	user := GetUserFromContext(c)
	HandleGetEnvelope(c, "dashboards", func() (interface{}, error) {
		return h.svc.UIMetadata.GetDashboards(c.Request.Context(), user), nil
	})
}

// GetDashboard handles GET /api/metadata/dashboards/:id
func (h *UIHandler) GetDashboard(c *gin.Context) {
	id := c.Param("id")
	// Custom 404
	dashboard := h.svc.UIMetadata.GetDashboard(c.Request.Context(), id)
	if dashboard == nil {
		RespondError(c, http.StatusNotFound, appErrors.ErrNotFound.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{constants.ResponseDashboard: dashboard})
}

// CreateDashboard handles POST /api/metadata/dashboards
func (h *UIHandler) CreateDashboard(c *gin.Context) {
	var dashboard models.DashboardConfig
	HandleCreateEnvelope(c, "dashboard", "Dashboard created successfully", &dashboard, func() error {
		// Strict Simplification: Do not allow widgets during creation.
		// Agents/Clients must use add_dashboard_widget or update flow.
		if len(dashboard.Widgets) > 0 {
			return appErrors.NewValidationError("widgets", "Dashboard creation with widgets is not supported. Please create the dashboard first, then add widgets.")
		}
		return h.svc.UIMetadata.CreateDashboard(c.Request.Context(), &dashboard)
	})
}

// UpdateDashboard handles PATCH /api/metadata/dashboards/:id
func (h *UIHandler) UpdateDashboard(c *gin.Context) {
	id := c.Param("id")
	var updates models.DashboardConfig
	HandleUpdateEnvelope(c, "dashboard", "Dashboard updated successfully", &updates, func() error {
		return h.svc.UIMetadata.UpdateDashboard(c.Request.Context(), id, &updates)
	})
}

// DeleteDashboard handles DELETE /api/metadata/dashboards/:id
func (h *UIHandler) DeleteDashboard(c *gin.Context) {
	id := c.Param("id")
	HandleDeleteEnvelope(c, "Dashboard deleted successfully", func() error {
		return h.svc.UIMetadata.DeleteDashboard(c.Request.Context(), id)
	})
}

// ==================== List View Handlers ====================

// GetListViews handles GET /api/metadata/listviews?objectApiName=X
func (h *UIHandler) GetListViews(c *gin.Context) {
	objectAPIName := c.Query("objectApiName")
	if objectAPIName == "" {
		RespondError(c, http.StatusBadRequest, "objectApiName query parameter is required")
		return
	}
	HandleGetEnvelope(c, "views", func() (interface{}, error) {
		return h.svc.UIMetadata.GetListViews(c.Request.Context(), objectAPIName), nil
	})
}

// CreateListView handles POST /api/metadata/listviews
func (h *UIHandler) CreateListView(c *gin.Context) {
	var view models.ListView
	HandleCreateEnvelope(c, "view", "List view created successfully", &view, func() error {
		if view.ObjectAPIName == "" {
			return appErrors.NewValidationError(constants.FieldObjectAPIName, "is required")
		}
		if view.Label == "" {
			return appErrors.NewValidationError("label", "is required")
		}
		return h.svc.Metadata.CreateListView(c.Request.Context(), &view)
	})
}

// UpdateListView handles PATCH /api/metadata/listviews/:id
func (h *UIHandler) UpdateListView(c *gin.Context) {
	id := c.Param("id")
	var updates models.ListView
	HandleUpdateEnvelope(c, "view", "List view updated successfully", &updates, func() error {
		return h.svc.Metadata.UpdateListView(c.Request.Context(), id, &updates)
	})
}

// DeleteListView handles DELETE /api/metadata/listviews/:id
func (h *UIHandler) DeleteListView(c *gin.Context) {
	id := c.Param("id")
	HandleDeleteEnvelope(c, "List view deleted successfully", func() error {
		return h.svc.Metadata.DeleteListView(c.Request.Context(), id)
	})
}

// GetSetupPages handles GET /api/setup/pages
func (h *UIHandler) GetSetupPages(c *gin.Context) {
	log.Printf("Hit GetSetupPages. Filter: %s", c.Query("filter"))
	HandleGetEnvelope(c, "pages", func() (interface{}, error) {
		pages, err := h.svc.Metadata.GetSetupPages(c.Request.Context())
		if err != nil {
			log.Printf("Error getting setup pages: %v", err)
			return nil, err
		}
		log.Printf("Retrieved %d total pages", len(pages))

		// Filter logic
		filter := c.Query("filter")
		if filter != "" {
			// Basic filter parsing "key:value"
			parts := strings.Split(filter, ":")
			if len(parts) == 2 {
				key, val := parts[0], parts[1]
				if key == "is_enabled" {
					var filtered []models.SetupPage
					wantEnabled := val == "true"
					log.Printf("Debug Filter: key=%s val=%s want=%v total_pages=%d", key, val, wantEnabled, len(pages))
					for _, p := range pages {
						log.Printf("  Page %s IsEnabled=%v", p.ID, p.IsEnabled)
						if p.IsEnabled == wantEnabled {
							filtered = append(filtered, p)
						}
					}
					return filtered, nil
				}
			}
		}

		return pages, nil
	})
}
