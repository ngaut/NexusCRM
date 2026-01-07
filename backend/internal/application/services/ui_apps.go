package services

import (
	"database/sql"
	"fmt"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== App Methods ====================

// GetApps delegates to MetadataService
func (s *UIMetadataService) GetApps() []*models.AppConfig {
	return s.metadata.GetApps()
}

// CreateApp delegates to MetadataService
func (s *UIMetadataService) CreateApp(app *models.AppConfig) error {
	return s.metadata.CreateApp(app)
}

// UpdateApp delegates to MetadataService
func (s *UIMetadataService) UpdateApp(appID string, updates *models.AppConfig) error {
	return s.metadata.UpdateApp(appID, updates)
}

// DeleteApp delegates to MetadataService
func (s *UIMetadataService) DeleteApp(appID string) error {
	return s.metadata.DeleteApp(appID)
}

// UpdateAppTx updates an app within a transaction
func (s *UIMetadataService) UpdateAppTx(tx *sql.Tx, appID string, updates *models.AppConfig) error {
	updates.ID = appID
	navigationItemsJSON, err := MarshalJSONOrDefault(updates.NavigationItems, "[]")
	if err != nil {
		return fmt.Errorf("failed to stringify navigationItems: %w", err)
	}

	query := fmt.Sprintf(`
		UPDATE %s 
		SET name = ?, label = ?, description = ?, icon = ?, navigation_items = ?, last_modified_date = NOW()
		WHERE id = ?
	`, constants.TableApp)
	// Use ID as Name
	_, err = tx.Exec(query, updates.ID, updates.Label, updates.Description, updates.Icon, navigationItemsJSON, appID)
	if err != nil {
		return fmt.Errorf("failed to update app in transaction: %w", err)
	}
	return nil
}

// AddAppNavigationItemTx adds a new navigation item to an existing app within a transaction
func (s *UIMetadataService) AddAppNavigationItemTx(tx *sql.Tx, appID string, item models.NavigationItem) error {
	// Get current state from DB (outside of Tx, but acceptable for this operation)
	app := s.metadata.GetApp(appID)
	if app == nil {
		return fmt.Errorf("app with ID '%s' not found", appID)
	}

	// Appending to the slice in the new struct
	app.NavigationItems = append(app.NavigationItems, item)

	return s.UpdateAppTx(tx, appID, app)
}

// RemoveObjectFromAllAppsTx removes an object from all app navigation items within a transaction
// This is called when an object is deleted to maintain referential integrity
func (s *UIMetadataService) RemoveObjectFromAllAppsTx(tx *sql.Tx, objectAPIName string) error {
	apps := s.metadata.GetApps()

	for _, app := range apps {
		updated := false
		newItems := []models.NavigationItem{}

		for _, item := range app.NavigationItems {
			// Skip navigation items that point to the deleted object
			if item.Type == "object" && item.ObjectAPIName == objectAPIName {
				updated = true
				continue
			}
			newItems = append(newItems, item)
		}

		// If this app had the object, update it
		if updated {
			app.NavigationItems = newItems
			if err := s.UpdateAppTx(tx, app.ID, app); err != nil {
				return err
			}
		}
	}

	return nil
}

// AddAppNavigationItem adds a new navigation item to an existing app
func (s *UIMetadataService) AddAppNavigationItem(appID string, item models.NavigationItem) error {
	app := s.metadata.GetApp(appID)
	if app == nil {
		return fmt.Errorf("app with ID '%s' not found", appID)
	}

	app.NavigationItems = append(app.NavigationItems, item)
	return s.metadata.UpdateApp(appID, app)
}
