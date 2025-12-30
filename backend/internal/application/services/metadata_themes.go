package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/nexuscrm/shared/pkg/models"
	"github.com/nexuscrm/shared/pkg/constants"
)

// scanTheme scans a row into a Theme struct
func (ms *MetadataService) scanTheme(row Scannable) (*models.Theme, error) {
	var theme models.Theme
	var colorsJSON, logoURL sql.NullString

	if err := row.Scan(&theme.ID, &theme.Name, &theme.IsActive, &colorsJSON, &theme.Density, &logoURL); err != nil {
		return nil, err
	}

	theme.LogoURL = models.NullStringToPtr(logoURL)
	if colorsJSON.Valid {
		if err := models.ParseJSON(colorsJSON.String, &theme.Colors); err != nil {
			log.Printf("⚠️ Failed to parse theme colors: %v", err)
		}
	}

	return &theme, nil
}

// GetActiveTheme returns the currently active theme
func (ms *MetadataService) GetActiveTheme() (*models.Theme, error) {
	query := fmt.Sprintf("SELECT id, name, is_active, colors, density, logo_url FROM %s WHERE is_active = true LIMIT 1", constants.TableTheme)
	row := ms.db.QueryRow(query)
	theme, err := ms.scanTheme(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Return nil if no active theme found, not error
		}
		return nil, err
	}
	return theme, nil
}

// UpsertTheme creates or updates a theme
func (ms *MetadataService) UpsertTheme(theme *models.Theme) error {
	// Check if exists by Name
	var id string
	query := fmt.Sprintf("SELECT id FROM %s WHERE name = ?", constants.TableTheme)
	err := ms.db.QueryRow(query, theme.Name).Scan(&id)

	colorsJSON, marshalErr := MarshalJSONOrDefault(theme.Colors, "{}")
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal colors: %w", marshalErr)
	}

	if errors.Is(err, sql.ErrNoRows) {
		// Insert
		if theme.ID == "" {
			theme.ID = GenerateID()
		}
		insert := fmt.Sprintf("INSERT INTO %s (id, name, is_active, colors, density, logo_url) VALUES (?, ?, ?, ?, ?, ?)", constants.TableTheme)
		_, err := ms.db.Exec(insert, theme.ID, theme.Name, theme.IsActive, colorsJSON, theme.Density, theme.LogoURL)
		return err
	} else if err != nil {
		return err
	}

	// Update
	update := fmt.Sprintf("UPDATE %s SET is_active = ?, colors = ?, density = ?, logo_url = ? WHERE id = ?", constants.TableTheme)
	_, err = ms.db.Exec(update, theme.IsActive, colorsJSON, theme.Density, theme.LogoURL, id)
	return err
}

// ActivateTheme sets a specific theme as active and deactivates all others
func (ms *MetadataService) ActivateTheme(themeID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	tx, err := ms.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Deactivate all
	_, err = tx.Exec(fmt.Sprintf("UPDATE %s SET is_active = false", constants.TableTheme))
	if err != nil {
		return fmt.Errorf("failed to deactivate themes: %w", err)
	}

	// 2. Activate target
	result, err := tx.Exec(fmt.Sprintf("UPDATE %s SET is_active = true WHERE id = ?", constants.TableTheme), themeID)
	if err != nil {
		return fmt.Errorf("failed to activate theme: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("theme not found: %s", themeID)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("✅ Theme activated: %s", themeID)
	return nil
}
