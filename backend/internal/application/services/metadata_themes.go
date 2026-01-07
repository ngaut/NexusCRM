package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nexuscrm/shared/pkg/models"
)

// GetActiveTheme returns the currently active theme
func (ms *MetadataService) GetActiveTheme() (*models.Theme, error) {
	// No lock needed for simple read (consistent with previous implementation)
	return ms.repo.GetActiveTheme(context.Background())
}

// UpsertTheme creates or updates a theme
func (ms *MetadataService) UpsertTheme(theme *models.Theme) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ctx := context.Background()

	// Check if exists by Name
	existing, err := ms.repo.GetThemeByName(ctx, theme.Name)
	if err != nil {
		return fmt.Errorf("failed to check theme existence: %w", err)
	}

	if existing != nil {
		// Found, update it (Update ID to match existing record)
		theme.ID = existing.ID
		theme.LastModifiedDate = time.Now()
		if err := ms.repo.UpdateTheme(ctx, theme); err != nil {
			return fmt.Errorf("failed to update theme: %w", err)
		}
	} else {
		// Not found, Insert
		if theme.ID == "" {
			theme.ID = GenerateID()
		}
		theme.CreatedDate = time.Now()
		theme.LastModifiedDate = time.Now()
		if err := ms.repo.CreateTheme(ctx, theme); err != nil {
			return fmt.Errorf("failed to create theme: %w", err)
		}
	}
	return nil
}

// ActivateTheme sets a specific theme as active and deactivates all others
func (ms *MetadataService) ActivateTheme(themeID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if err := ms.repo.ActivateTheme(context.Background(), themeID); err != nil {
		return err
	}

	log.Printf("✅ Theme activated: %s", themeID)
	return nil
}
