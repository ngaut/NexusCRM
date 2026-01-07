package services

import (
	"context"
	"fmt"
	"log"

	"github.com/nexuscrm/shared/pkg/models"
)

// normalizeWidgets ensures all widgets have IDs
func normalizeWidgets(widgets []models.WidgetConfig) []models.WidgetConfig {
	for i := range widgets {
		if widgets[i].ID == "" {
			widgets[i].ID = GenerateID()
		}
	}
	return widgets
}

// ==================== Dashboard Methods ====================

// GetDashboards returns all dashboards
func (ms *MetadataService) GetDashboards(ctx context.Context, user *models.UserSession) []*models.DashboardConfig {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	dashboards, err := ms.repo.GetAllDashboards(ctx)
	if err != nil {
		log.Printf("Failed to get dashboards: %v", err)
		return []*models.DashboardConfig{}
	}

	return dashboards
}

// GetDashboard returns a dashboard by ID
func (ms *MetadataService) GetDashboard(ctx context.Context, id string) *models.DashboardConfig {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	dashboard, err := ms.repo.GetDashboard(ctx, id)
	if err != nil {
		return nil
	}
	return dashboard
}

// CreateDashboard creates a new dashboard
func (ms *MetadataService) CreateDashboard(ctx context.Context, dashboard *models.DashboardConfig) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if dashboard.Label == "" {
		return fmt.Errorf("dashboard label is required")
	}

	if dashboard.ID == "" {
		dashboard.ID = GenerateID()
	}

	existing, _ := ms.repo.GetDashboard(ctx, dashboard.ID)
	if existing != nil {
		return fmt.Errorf("dashboard with ID '%s' already exists", dashboard.ID)
	}

	// Normalize widgets: ensure IDs
	dashboard.Widgets = normalizeWidgets(dashboard.Widgets)

	// Layout default
	if dashboard.Layout == "" {
		dashboard.Layout = "two-column"
	}

	// Insert into DB via Repo
	if err := ms.repo.CreateDashboard(ctx, dashboard); err != nil {
		return fmt.Errorf("failed to insert dashboard: %w", err)
	}

	return nil
}

// UpdateDashboard updates an existing dashboard
func (ms *MetadataService) UpdateDashboard(ctx context.Context, id string, updates *models.DashboardConfig) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	existing, err := ms.repo.GetDashboard(ctx, id)
	if err != nil || existing == nil {
		return fmt.Errorf("dashboard with ID '%s' not found", id)
	}

	// Update existing fields with provided updates
	if updates.Label != "" {
		existing.Label = updates.Label
	}
	if updates.Widgets != nil {
		existing.Widgets = updates.Widgets
	}
	if updates.Layout != "" {
		existing.Layout = updates.Layout
	}
	if updates.Description != nil {
		existing.Description = updates.Description
	}

	existing.ID = id
	existing.Widgets = normalizeWidgets(existing.Widgets)

	// Update DB via Repo
	if err := ms.repo.UpdateDashboard(ctx, id, existing); err != nil {
		return fmt.Errorf("failed to update dashboard: %w", err)
	}

	return nil
}

// DeleteDashboard deletes a dashboard
func (ms *MetadataService) DeleteDashboard(ctx context.Context, id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	existing, _ := ms.repo.GetDashboard(ctx, id)
	if existing == nil {
		return fmt.Errorf("dashboard with ID '%s' not found", id)
	}

	return ms.repo.DeleteDashboard(ctx, id)
}
