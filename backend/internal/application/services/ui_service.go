package services

import (
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
)

// UIMetadataService manages UI-related metadata (Apps, Tabs, Layouts, Dashboards)
type UIMetadataService struct {
	db       *database.TiDBConnection
	metadata *MetadataService
}

// NewUIMetadataService creates a new UIMetadataService
func NewUIMetadataService(db *database.TiDBConnection, metadata *MetadataService) *UIMetadataService {
	return &UIMetadataService{
		db:       db,
		metadata: metadata,
	}
}

// RefreshCache reloads all UI metadata from the database
func (s *UIMetadataService) RefreshCache() error {
	// No-op as caches are removed
	return nil
}

// ==================== Dashboard Methods ====================

// GetDashboards delegates to MetadataService
func (s *UIMetadataService) GetDashboards(user *models.UserSession) []*models.DashboardConfig {
	return s.metadata.GetDashboards(user)
}

// GetDashboard delegates to MetadataService
func (s *UIMetadataService) GetDashboard(id string) *models.DashboardConfig {
	return s.metadata.GetDashboard(id)
}

// CreateDashboard delegates to MetadataService
func (s *UIMetadataService) CreateDashboard(dashboard *models.DashboardConfig) error {
	return s.metadata.CreateDashboard(dashboard)
}

// UpdateDashboard delegates to MetadataService
func (s *UIMetadataService) UpdateDashboard(id string, updates *models.DashboardConfig) error {
	return s.metadata.UpdateDashboard(id, updates)
}

// DeleteDashboard delegates to MetadataService
func (s *UIMetadataService) DeleteDashboard(id string) error {
	return s.metadata.DeleteDashboard(id)
}

// ==================== Layout Methods ====================

func (s *UIMetadataService) GetLayout(apiName string, profileID *string) *models.PageLayout {
	// MetadataService.GetLayout takes (apiName, profileID *string) and returns *models.PageLayout
	// It internally handles augmenting with related lists.
	return s.metadata.GetLayout(apiName, profileID)
}

func (s *UIMetadataService) SaveLayout(layout *models.PageLayout) error {
	return s.metadata.SaveLayout(layout)
}

func (s *UIMetadataService) DeleteLayout(layoutID string) error {
	return s.metadata.DeleteLayout(layoutID)
}

func (s *UIMetadataService) AssignLayoutToProfile(profileID, objectAPIName, layoutID string) error {
	return s.metadata.AssignLayoutToProfile(profileID, objectAPIName, layoutID)
}

// ==================== List View Methods ====================

func (s *UIMetadataService) GetListViews(objectAPIName string) []*models.ListView {
	return s.metadata.GetListViews(objectAPIName)
}

// App methods are in ui_apps.go (kept separate due to larger logic)

func (s *UIMetadataService) GetActiveTheme() (*models.Theme, error) {
	return s.metadata.GetActiveTheme()
}

// UpsertTheme delegates to MetadataService
func (s *UIMetadataService) UpsertTheme(theme *models.Theme) error {
	return s.metadata.UpsertTheme(theme)
}

// ActivateTheme delegates to MetadataService
func (s *UIMetadataService) ActivateTheme(id string) error {
	return s.metadata.ActivateTheme(id)
}
