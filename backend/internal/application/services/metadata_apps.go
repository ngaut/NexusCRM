package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== App Methods ====================

func (ms *MetadataService) GetApps() []*models.AppConfig {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	apps, err := ms.repo.GetAllApps(context.Background())
	if err != nil {
		log.Printf("Failed to get apps: %v", err)
		return []*models.AppConfig{}
	}
	return apps
}

func (ms *MetadataService) GetApp(id string) *models.AppConfig {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	app, err := ms.repo.GetApp(context.Background(), id)
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
	if app.ID == "" {
		return fmt.Errorf("App ID is required")
	}
	if app.Label == "" {
		return fmt.Errorf("App Label is required")
	}

	// Check if exists
	existing, err := ms.repo.GetApp(context.Background(), app.ID)
	if err != nil {
		return fmt.Errorf("failed to check app existence: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("app with ID '%s' already exists", app.ID)
	}

	// Insert into DB via Repo
	app.CreatedDate = time.Now()
	app.LastModifiedDate = time.Now()
	if err := ms.repo.CreateApp(context.Background(), app); err != nil {
		return fmt.Errorf("failed to insert app: %w", err)
	}

	return nil
}

// UpdateApp updates an existing app configuration
func (ms *MetadataService) UpdateApp(appID string, updates *models.AppConfig) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Check if exists
	existing, err := ms.repo.GetApp(context.Background(), appID)
	if err != nil {
		return fmt.Errorf("failed to check app existence: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("app with ID '%s' not found", appID)
	}

	// Ensure ID doesn't change
	updates.ID = appID

	// Merge updates with existing values - only update fields that are provided
	// We update `existing` object with `updates` values, then save `existing`.
	if updates.Label != "" {
		existing.Label = updates.Label
	}
	if updates.Description != "" {
		existing.Description = updates.Description
	}
	if updates.Icon != "" {
		existing.Icon = updates.Icon
	}
	if updates.Color != "" {
		existing.Color = updates.Color
	}
	// Note: IsDefault is a boolean and lacks "not set" state in the struct.
	// If updates.IsDefault is false (default), it will overwrite existing true.
	// This assumes 'updates' contains the authoritative state for IsDefault.
	existing.IsDefault = updates.IsDefault

	if updates.NavigationItems != nil {
		existing.NavigationItems = updates.NavigationItems
	}

	existing.LastModifiedDate = time.Now()

	// Update DB via Repo
	if err := ms.repo.UpdateApp(context.Background(), appID, existing); err != nil {
		return fmt.Errorf("failed to update app: %w", err)
	}

	return nil
}

// DeleteApp deletes an app configuration
func (ms *MetadataService) DeleteApp(appID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.repo.DeleteApp(context.Background(), appID)
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

	app, err := ms.repo.GetApp(context.Background(), appID)
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

	// Persist App Update via Repo
	if err := ms.repo.UpdateApp(context.Background(), appID, app); err != nil {
		log.Printf("⚠️ Warning: Created object %s but failed to update app navigation: %v", schema.APIName, err)
	}

	// 4. Auto-grant Permissions using centralized helper
	if err := GrantInitialObjectPermissions(ms.db, schema.APIName, constants.TableProfile, constants.TableObjectPerms, constants.ProfileSystemAdmin); err != nil {
		log.Printf("⚠️ Warning: Created object %s but failed to grant permissions: %v", schema.APIName, err)
	}

	return nil
}
