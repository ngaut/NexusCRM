package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== App Methods ====================

func (ms *MetadataService) GetApps() []*models.AppConfig {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	apps, err := ms.queryAllApps()
	if err != nil {
		log.Printf("Failed to get apps: %v", err)
		return []*models.AppConfig{}
	}
	return apps
}

func (ms *MetadataService) GetApp(id string) *models.AppConfig {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	app, err := ms.queryApp(id)
	if err != nil {
		log.Printf("Warning: Failed to query app %s: %v", id, err)
		return nil
	}
	return app
}

// CreateApp creates a new app configuration
func (ms *MetadataService) CreateApp(app *models.AppConfig) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Validate
	if app.ID == "" || app.Label == "" {
		return fmt.Errorf("app ID and label are required")
	}

	// Check if exists
	existing, _ := ms.queryApp(app.ID)
	if existing != nil {
		return fmt.Errorf("app with ID '%s' already exists", app.ID)
	}

	// Use helper for JSON
	navItemsJSON, err := MarshalJSONOrDefault(app.NavigationItems, "[]")
	if err != nil {
		return fmt.Errorf("failed to marshal navigation items: %w", err)
	}

	// Insert into DB
	query := fmt.Sprintf(`INSERT INTO %s (
		id, name, label, description, icon, color, is_default, navigation_items
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, constants.TableApp)
	_, err = ms.db.Exec(query, app.ID, app.ID, app.Label, app.Description, app.Icon, app.Color, app.IsDefault, navItemsJSON)
	if err != nil {
		return fmt.Errorf("failed to insert app: %w", err)
	}

	return nil
}

// UpdateApp updates an existing app configuration
func (ms *MetadataService) UpdateApp(appID string, updates *models.AppConfig) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Check if exists
	existing, _ := ms.queryApp(appID)
	if existing == nil {
		return fmt.Errorf("app with ID '%s' not found", appID)
	}

	// Ensure ID doesn't change
	updates.ID = appID

	// Merge updates with existing values - only update fields that are provided
	label := updates.Label
	if label == "" {
		label = existing.Label
	}
	description := updates.Description
	if description == "" {
		description = existing.Description
	}
	icon := updates.Icon
	if icon == "" {
		icon = existing.Icon
	}
	color := updates.Color
	if color == "" {
		color = existing.Color
	}

	var navItemsJSON string
	var err error
	if updates.NavigationItems != nil {
		navItemsJSON, err = MarshalJSONOrDefault(updates.NavigationItems, "[]")
	} else {
		// Keep existing if update is nil
		navItemsJSON, err = MarshalJSONOrDefault(existing.NavigationItems, "[]")
	}
	if err != nil {
		return fmt.Errorf("failed to marshal navigation items: %w", err)
	}

	// Update DB with merged values
	query := fmt.Sprintf("UPDATE %s SET label = ?, description = ?, icon = ?, color = ?, is_default = ?, navigation_items = ? WHERE id = ?", constants.TableApp)
	_, err = ms.db.Exec(query, label, description, icon, color, updates.IsDefault, navItemsJSON, appID)
	if err != nil {
		return fmt.Errorf("failed to update app: %w", err)
	}

	return nil
}

// DeleteApp deletes an app configuration
func (ms *MetadataService) DeleteApp(appID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.deleteMetadataRecord(constants.TableApp, appID, "app")
}

// CreateObjectInApp creates a new object and automatically calculates navigation
func (ms *MetadataService) CreateObjectInApp(appID string, schema *models.ObjectMetadata) error {
	// 1. Set AppID on schema
	schema.AppID = &appID

	// 2. Create the Schema
	if err := ms.CreateSchema(schema); err != nil {
		return err
	}

	// 3. Add to App Navigation
	ms.mu.Lock() // Re-acquire lock for App update (CreateSchema has its own lock)
	defer ms.mu.Unlock()

	app, err := ms.queryApp(appID)
	if err != nil || app == nil {
		// Just log warning if app not found, main object creation succeeded
		log.Printf("⚠️ Warning: Created object %s but failed to find app %s to add navigation", schema.APIName, appID)
		return nil
	}

	// Create navigation item
	newItem := models.NavigationItem{
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

	// Persist App Update
	navItemsJSON, err := json.Marshal(app.NavigationItems)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to marshal navigation items for app %s: %v", appID, err)
		return nil
	}
	if _, err := ms.db.Exec(fmt.Sprintf("UPDATE %s SET navigation_items = ? WHERE id = ?", constants.TableApp), string(navItemsJSON), appID); err != nil {
		log.Printf("⚠️ Warning: Created object %s but failed to update app navigation: %v", schema.APIName, err)
	}

	// 4. Auto-grant Permissions using centralized helper
	if err := GrantInitialObjectPermissions(ms.db, schema.APIName, constants.TableProfile, constants.TableObjectPerms, constants.ProfileSystemAdmin); err != nil {
		log.Printf("⚠️ Warning: Created object %s but failed to grant permissions: %v", schema.APIName, err)
	}

	return nil
}
