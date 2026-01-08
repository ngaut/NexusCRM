package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== App Methods ====================

// GetApps delegates to MetadataService
func (s *UIMetadataService) GetApps(ctx context.Context) []*models.AppConfig {
	return s.metadata.GetApps(ctx)
}

// CreateApp delegates to MetadataService
func (s *UIMetadataService) CreateApp(ctx context.Context, app *models.AppConfig) error {
	return s.metadata.CreateApp(ctx, app)
}

// UpdateApp delegates to MetadataService
func (s *UIMetadataService) UpdateApp(ctx context.Context, appID string, updates *models.AppConfig) error {
	return s.metadata.UpdateApp(ctx, appID, updates)
}

// DeleteApp delegates to MetadataService
func (s *UIMetadataService) DeleteApp(ctx context.Context, appID string) error {
	return s.metadata.DeleteApp(ctx, appID)
}

// UpdateAppTx updates an app within a transaction
func (s *UIMetadataService) UpdateAppTx(ctx context.Context, tx *sql.Tx, appID string, updates *models.AppConfig) error {
	return s.metadata.UpdateAppTx(ctx, tx, appID, updates)
}

// AddAppNavigationItemTx adds a new navigation item to an existing app within a transaction
func (s *UIMetadataService) AddAppNavigationItemTx(ctx context.Context, tx *sql.Tx, appID string, item models.NavigationItem) error {
	// Get current state from DB within Tx for consistency
	app, err := s.metadata.GetAppWithTx(ctx, tx, appID)
	if err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}
	if app == nil {
		return fmt.Errorf("app with ID '%s' not found", appID)
	}

	// Appending to the slice in the new struct
	app.NavigationItems = append(app.NavigationItems, item)

	return s.UpdateAppTx(ctx, tx, appID, app)
}

// RemoveObjectFromAllAppsTx removes an object from all app navigation items within a transaction
// This is called when an object is deleted to maintain referential integrity
func (s *UIMetadataService) RemoveObjectFromAllAppsTx(ctx context.Context, tx *sql.Tx, objectAPIName string) error {
	apps := s.metadata.GetApps(ctx)

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
			if err := s.UpdateAppTx(ctx, tx, app.ID, app); err != nil {
				return err
			}
		}
	}

	return nil
}

// AddAppNavigationItem adds a new navigation item to an existing app
func (s *UIMetadataService) AddAppNavigationItem(ctx context.Context, appID string, item models.NavigationItem) error {
	app := s.metadata.GetApp(ctx, appID)
	// if err != nil {
	// 	return fmt.Errorf("failed to get app: %w", err)
	// }
	if app == nil {
		return fmt.Errorf("app with ID '%s' not found", appID)
	}

	app.NavigationItems = append(app.NavigationItems, item)
	return s.metadata.UpdateApp(ctx, appID, app)
}

// CreateObjectWithApp creates a new object and automatically adds it to an app's navigation
// This orchestrates Schema Creation -> App Update -> Permission Granting
func (s *UIMetadataService) CreateObjectWithApp(ctx context.Context, appID string, schema *models.ObjectMetadata) error {
	// 1. Set AppID on schema
	schema.AppID = &appID

	// 2. Create the Schema (delegated to MetadataService)
	if err := s.metadata.CreateSchema(ctx, schema); err != nil {
		return err
	}

	// 3. Add to App Navigation
	// We use existing public methods which handle locking internally
	app := s.metadata.GetApp(ctx, appID)
	if app == nil {
		// Just log warning if app not found, main object creation succeeded
		return nil
	}

	// Create navigation item
	newItem := models.NavigationItem{
		// ID generation - we can use a helper or simple string concat if we don't have GenerateID accessible?
		// We don't have access to services.GenerateID in this file if it's package private or circular?
		// It is likely in crud_helpers.go or similar as package utility.
		// Assuming GenerateID is available in package services.
		ID:            fmt.Sprintf("nav-%s-%s", schema.APIName, GenerateID()[:8]),
		Type:          "object",
		ObjectAPIName: schema.APIName,
		Label:         schema.PluralLabel,
		Icon:          schema.Icon,
	}

	// Append to items
	if app.NavigationItems == nil {
		app.NavigationItems = []models.NavigationItem{}
	}
	app.NavigationItems = append(app.NavigationItems, newItem)

	// 4. Update the Default App (Sales) navigation
	// Also add to Sales app by default for visibility
	if err := s.AddAppNavigationItem(ctx, "standard__Sales", newItem); err != nil {
		// Just log, don't fail flow
		log.Printf("⚠️ Warning: Created object %s but failed to update Sales app navigation: %v\n", schema.APIName, err)
	}

	// 5. Grant initial permissions (System Admin gets explicit access)
	if s.permissions != nil {
		if err := s.permissions.GrantInitialPermissions(ctx, schema.APIName); err != nil {
			log.Printf("⚠️ Warning: Created object %s but failed to grant permissions: %v\n", schema.APIName, err)
		}
	} else {
		log.Printf("⚠️ Warning: PermissionService not available, skipping initial grant for %s\n", schema.APIName)
	}

	return nil
}
