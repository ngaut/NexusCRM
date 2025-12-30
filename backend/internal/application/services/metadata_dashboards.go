package services

import (
	"fmt"
	"log"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Dashboard Methods ====================

// GetDashboards returns all dashboards
func (ms *MetadataService) GetDashboards(user *models.UserSession) []*models.DashboardConfig {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	dashboards, err := ms.queryDashboards()
	if err != nil {
		log.Printf("Failed to get dashboards: %v", err)
		return []*models.DashboardConfig{}
	}

	return dashboards
}

// GetDashboard returns a dashboard by ID
func (ms *MetadataService) GetDashboard(id string) *models.DashboardConfig {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	dashboard, err := ms.queryDashboard(id)
	if err != nil {
		return nil
	}
	return dashboard
}

// CreateDashboard creates a new dashboard
func (ms *MetadataService) CreateDashboard(dashboard *models.DashboardConfig) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if dashboard.Label == "" {
		return fmt.Errorf("dashboard label is required")
	}

	if dashboard.ID == "" {
		dashboard.ID = GenerateID()
	}

	existing, _ := ms.queryDashboard(dashboard.ID)
	if existing != nil {
		return fmt.Errorf("dashboard with ID '%s' already exists", dashboard.ID)
	}

	widgetsJSON, err := MarshalJSONOrDefault(dashboard.Widgets, "[]")
	if err != nil {
		return fmt.Errorf("failed to marshal widgets: %w", err)
	}

	layout := dashboard.Layout
	if layout == "" {
		layout = "two-column"
	}

	description := dashboard.Description
	if description == nil {
		empty := ""
		description = &empty
	}

	query := fmt.Sprintf("INSERT INTO %s (id, name, description, layout, widgets) VALUES (?, ?, ?, ?, ?)", constants.TableDashboard)
	_, err = ms.db.Exec(query, dashboard.ID, dashboard.Label, *description, layout, widgetsJSON)
	if err != nil {
		return fmt.Errorf("failed to insert dashboard: %w", err)
	}

	return nil
}

// UpsertDashboard creates a dashboard if it doesn't exist, or updates it if it does
func (ms *MetadataService) UpsertDashboard(dashboard *models.DashboardConfig) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if dashboard.Label == "" {
		return fmt.Errorf("dashboard label is required")
	}

	if dashboard.ID == "" {
		dashboard.ID = GenerateID()
	}

	widgetsJSON, err := MarshalJSONOrDefault(dashboard.Widgets, "[]")
	if err != nil {
		return fmt.Errorf("failed to marshal widgets: %w", err)
	}

	layout := "two-column"
	description := ""
	if dashboard.Description != nil {
		description = *dashboard.Description
	}

	// Check if dashboard exists
	existing, _ := ms.queryDashboard(dashboard.ID)
	if existing != nil {
		// Update
		query := fmt.Sprintf("UPDATE %s SET name = ?, description = ?, layout = ?, widgets = ? WHERE id = ?", constants.TableDashboard)
		if _, err := ms.db.Exec(query, dashboard.Label, description, layout, string(widgetsJSON), dashboard.ID); err != nil {
			return fmt.Errorf("failed to update dashboard: %w", err)
		}
	} else {
		// Insert
		query := fmt.Sprintf("INSERT INTO %s (id, name, description, layout, widgets) VALUES (?, ?, ?, ?, ?)", constants.TableDashboard)
		_, err = ms.db.Exec(query, dashboard.ID, dashboard.Label, description, layout, widgetsJSON)
		if err != nil {
			return fmt.Errorf("failed to insert dashboard: %w", err)
		}
	}

	return nil
}

// UpdateDashboard updates an existing dashboard
func (ms *MetadataService) UpdateDashboard(id string, updates *models.DashboardConfig) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	existing, err := ms.queryDashboard(id)
	if err != nil || existing == nil {
		return fmt.Errorf("dashboard with ID '%s' not found", id)
	}

	if updates.Label != "" {
		existing.Label = updates.Label
	}
	if updates.Widgets != nil {
		existing.Widgets = updates.Widgets
	}
	if updates.Layout != "" {
		existing.Layout = updates.Layout
	}

	existing.ID = id

	widgetsJSON, err := MarshalJSONOrDefault(existing.Widgets, "[]")
	if err != nil {
		return fmt.Errorf("failed to marshal widgets: %w", err)
	}

	query := fmt.Sprintf("UPDATE %s SET name = ?, description = ?, layout = ?, widgets = ? WHERE id = ?", constants.TableDashboard)
	if _, err := ms.db.Exec(query, existing.Label, existing.Description, existing.Layout, string(widgetsJSON), id); err != nil {
		return fmt.Errorf("failed to update dashboard: %w", err)
	}

	return nil
}

// DeleteDashboard deletes a dashboard
func (ms *MetadataService) DeleteDashboard(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	existing, _ := ms.queryDashboard(id)
	if existing == nil {
		return fmt.Errorf("dashboard with ID '%s' not found", id)
	}

	_, err := ms.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", constants.TableDashboard), id)
	if err != nil {
		return fmt.Errorf("failed to delete dashboard: %w", err)
	}

	return nil
}
