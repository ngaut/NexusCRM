package services

import (
	"context"

	"github.com/nexuscrm/shared/pkg/models"
)

// UIMetadataService manages UI-related metadata (Apps, Tabs, Layouts, Dashboards)
type UIMetadataService struct {
	metadata    *MetadataService
	permissions *PermissionService
}

// NewUIMetadataService creates a new UIMetadataService
func NewUIMetadataService(metadata *MetadataService, permissions *PermissionService) *UIMetadataService {
	return &UIMetadataService{
		metadata:    metadata,
		permissions: permissions,
	}
}

// RefreshCache reloads all UI metadata from the database
func (s *UIMetadataService) RefreshCache() error {
	// No-op as caches are removed
	return nil
}

// ==================== Dashboard Methods ====================

// GetDashboards delegates to MetadataService
func (s *UIMetadataService) GetDashboards(ctx context.Context, user *models.UserSession) []*models.DashboardConfig {
	return s.metadata.GetDashboards(ctx, user)
}

// GetDashboard delegates to MetadataService
func (s *UIMetadataService) GetDashboard(ctx context.Context, id string) *models.DashboardConfig {
	return s.metadata.GetDashboard(ctx, id)
}

// CreateDashboard delegates to MetadataService
func (s *UIMetadataService) CreateDashboard(ctx context.Context, dashboard *models.DashboardConfig) error {
	return s.metadata.CreateDashboard(ctx, dashboard)
}

// UpdateDashboard delegates to MetadataService
func (s *UIMetadataService) UpdateDashboard(ctx context.Context, id string, updates *models.DashboardConfig) error {
	return s.metadata.UpdateDashboard(ctx, id, updates)
}

// DeleteDashboard delegates to MetadataService
func (s *UIMetadataService) DeleteDashboard(ctx context.Context, id string) error {
	return s.metadata.DeleteDashboard(ctx, id)
}

// ==================== Layout Methods ====================

func (s *UIMetadataService) GetLayout(ctx context.Context, apiName string, profileID *string) *models.PageLayout {
	// MetadataService.GetLayout takes (apiName, profileID *string) and returns *models.PageLayout
	// It internally handles augmenting with related lists.
	return s.metadata.GetLayout(ctx, apiName, profileID)
}

func (s *UIMetadataService) SaveLayout(ctx context.Context, layout *models.PageLayout) error {
	return s.metadata.SaveLayout(ctx, layout)
}

func (s *UIMetadataService) DeleteLayout(ctx context.Context, layoutID string) error {
	return s.metadata.DeleteLayout(ctx, layoutID)
}

func (s *UIMetadataService) AssignLayoutToProfile(ctx context.Context, profileID, objectAPIName, layoutID string) error {
	return s.metadata.AssignLayoutToProfile(ctx, profileID, objectAPIName, layoutID)
}

// ==================== List View Methods ====================

func (s *UIMetadataService) GetListViews(ctx context.Context, objectAPIName string) []*models.ListView {
	return s.metadata.GetListViews(ctx, objectAPIName)
}

// App methods are in ui_apps.go (kept separate due to larger logic)

func (s *UIMetadataService) GetActiveTheme(ctx context.Context) (*models.Theme, error) {
	return s.metadata.GetActiveTheme(ctx)
}

// UpsertTheme delegates to MetadataService
func (s *UIMetadataService) UpsertTheme(ctx context.Context, theme *models.Theme) error {
	return s.metadata.UpsertTheme(ctx, theme)
}

// ActivateTheme delegates to MetadataService
func (s *UIMetadataService) ActivateTheme(ctx context.Context, id string) error {
	return s.metadata.ActivateTheme(ctx, id)
}
