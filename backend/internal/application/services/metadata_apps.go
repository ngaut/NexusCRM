package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== App Methods ====================

func (ms *MetadataService) GetApps(ctx context.Context) []*models.AppConfig {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	apps, err := ms.repo.GetAllApps(ctx)
	if err != nil {
		log.Printf("Failed to get apps: %v", err)
		return []*models.AppConfig{}
	}
	return apps
}

func (ms *MetadataService) GetApp(ctx context.Context, id string) *models.AppConfig {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	app, err := ms.repo.GetApp(ctx, id)
	if err != nil {
		log.Printf("Warning: Failed to query app %s: %v", id, err)
		return nil
	}
	return app
}

func (ms *MetadataService) GetAppWithTx(ctx context.Context, tx *sql.Tx, id string) (*models.AppConfig, error) {
	// No invalidation lock needed for read in Tx usually, but logic depends on consistency needs.
	// We delegate to repo.
	return ms.repo.GetAppWithTx(ctx, tx, id)
}

// CreateApp creates a new app configuration
func (ms *MetadataService) CreateApp(ctx context.Context, app *models.AppConfig) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Validate
	if app.ID == "" {
		return fmt.Errorf("app ID is required")
	}
	if app.Label == "" {
		return fmt.Errorf("app label is required")
	}

	// Check if exists
	existing, err := ms.repo.GetApp(ctx, app.ID)
	if err != nil {
		return fmt.Errorf("failed to check app existence: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("app with ID '%s' already exists", app.ID)
	}

	// Insert into DB via Repo
	app.CreatedDate = time.Now()
	app.LastModifiedDate = time.Now()
	if err := ms.repo.CreateApp(ctx, app); err != nil {
		return fmt.Errorf("failed to insert app: %w", err)
	}

	return nil
}

// UpdateApp updates an existing app configuration
func (ms *MetadataService) UpdateApp(ctx context.Context, appID string, updates *models.AppConfig) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Check if exists
	existing, err := ms.repo.GetApp(ctx, appID)
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
	if err := ms.repo.UpdateApp(ctx, appID, existing); err != nil {
		return fmt.Errorf("failed to update app: %w", err)
	}

	return nil
}

// UpdateAppTx updates an existing app configuration within a transaction
func (ms *MetadataService) UpdateAppTx(ctx context.Context, tx *sql.Tx, appID string, updates *models.AppConfig) error {
	// Note: We don't lock mutex here because caller handles transaction and concurrency usually means specialized flow.
	// But strictly, we should be careful. Since it's a DB transaction, DB locks apply.
	// We should update the DB via Repo.

	// Ensure ID doesn't change
	updates.ID = appID
	updates.LastModifiedDate = time.Now()

	// Update DB via Repo with Tx
	if err := ms.repo.UpdateAppWithTx(ctx, tx, appID, updates); err != nil {
		return fmt.Errorf("failed to update app in tx: %w", err)
	}

	return nil
}

// DeleteApp deletes an app configuration
func (ms *MetadataService) DeleteApp(ctx context.Context, appID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.repo.DeleteApp(ctx, appID)
}
