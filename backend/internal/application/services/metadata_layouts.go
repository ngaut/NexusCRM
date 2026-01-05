package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Layout Methods ====================

// GetLayout returns the layout for an object
func (ms *MetadataService) GetLayout(apiName string, profileID *string) *models.PageLayout {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var layout *models.PageLayout

	// 1. If profileID provided, try to find assigned layout
	if profileID != nil {
		var layoutID string
		err := ms.db.QueryRow(fmt.Sprintf("SELECT layout_id FROM %s WHERE profile_id = ? AND object_api_name = ?", constants.TableProfileLayout), *profileID, apiName).Scan(&layoutID)
		if err == nil {
			l, err := ms.queryLayout(layoutID)
			if err == nil && l != nil {
				layout = l
			}
		}
	}

	// 2. Fallback: get all layouts for object and pick first
	if layout == nil {
		layouts, err := ms.queryLayouts(apiName)
		if err == nil && len(layouts) > 0 {
			layout = layouts[0]
		}
	}

	// 3. Last resort: generate default layout from schema
	if layout == nil {
		obj, err := ms.querySchemaByAPIName(apiName)
		if err != nil || obj == nil {
			return nil
		}

		// Create default layout
		newLayout := &models.PageLayout{
			ID:            "default_" + apiName,
			ObjectAPIName: apiName,
			LayoutName:    "Default Layout",
			Type:          "Detail",
			Sections: []models.PageSection{
				{
					ID:      GenerateID(),
					Label:   "Information",
					Columns: 2,
					Fields:  []string{},
				},
				{
					ID:      GenerateID(),
					Label:   "System Information",
					Columns: 2,
					Fields:  []string{constants.FieldCreatedByID, constants.FieldCreatedDate, constants.FieldLastModifiedByID, constants.FieldLastModifiedDate},
				},
			},
		}

		// Add fields to layout
		for _, f := range obj.Fields {
			if !f.IsSystem && f.APIName != "Id" {
				newLayout.Sections[0].Fields = append(newLayout.Sections[0].Fields, f.APIName)
			}
		}
		layout = newLayout
	}

	// 4. Augment with Related Lists (Auto-Discovery) if missing
	if layout != nil && len(layout.RelatedLists) == 0 {
		// We use a helper to avoid complex logic inside the main flow
		// Note: We are under RLock, but Query is safe.
		ms.augmentLayoutWithRelatedLists(layout)
	}

	return layout
}

// augmentLayoutWithRelatedLists finds objects that lookup to the current object and adds them as related lists
func (ms *MetadataService) augmentLayoutWithRelatedLists(layout *models.PageLayout) {
	// Find all fields that lookup TO this object
	// Join with _System_Object to get the child object's details
	// Left Join with _System_Relationship to get configured related list columns
	query := fmt.Sprintf(`
		SELECT f.api_name, o.api_name, o.plural_label, r.related_list_fields
		FROM %s f
		JOIN %s o ON f.object_id = o.id
		LEFT JOIN %s r ON r.child_object_api_name = o.api_name AND r.field_api_name = f.api_name
		WHERE f.reference_to = ? AND f.type = 'Lookup'
	`, constants.TableField, constants.TableObject, constants.TableRelationship)

	rows, err := ms.db.Query(query, layout.ObjectAPIName)
	if err != nil {
		log.Printf("⚠️ Failed to query child relationships for %s: %v\nQuery: %s", layout.ObjectAPIName, err, query)
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var lookupFieldAPI, childObjectAPI, childPluralLabel string
		var relatedListFields sql.NullString

		if err := rows.Scan(&lookupFieldAPI, &childObjectAPI, &childPluralLabel, &relatedListFields); err != nil {
			log.Printf("Warning: Failed to scan child relationship row: %v", err)
			continue
		}

		// Avoid duplicates if any exist
		exists := false
		for _, rl := range layout.RelatedLists {
			if rl.ObjectAPIName == childObjectAPI && rl.LookupField == lookupFieldAPI {
				exists = true
				break
			}
		}

		if !exists {
			// Default columns
			columns := []string{constants.FieldName, constants.FieldCreatedDate}

			// Override if relationship defines columns
			if relatedListFields.Valid && relatedListFields.String != "" {
				var customColumns []string
				if err := json.Unmarshal([]byte(relatedListFields.String), &customColumns); err == nil && len(customColumns) > 0 {
					columns = customColumns
				}
			}

			layout.RelatedLists = append(layout.RelatedLists, models.RelatedListConfig{
				ID:            fmt.Sprintf("%s-%s", childObjectAPI, lookupFieldAPI),
				Label:         childPluralLabel,
				ObjectAPIName: childObjectAPI,
				LookupField:   lookupFieldAPI,
				Fields:        columns,
			})
		}
	}
}

// SaveLayout saves or updates a page layout
func (ms *MetadataService) SaveLayout(layout *models.PageLayout) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Serialize layout to JSON
	configJSON, err := json.Marshal(layout)
	if err != nil {
		return fmt.Errorf("failed to marshal layout: %w", err)
	}

	// Use UPSERT
	query := fmt.Sprintf(`
		INSERT INTO %s (id, object_api_name, config) 
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE 
			object_api_name = VALUES(object_api_name),
			config = VALUES(config)
	`, constants.TableLayout)

	_, err = ms.db.Exec(query, layout.ID, layout.ObjectAPIName, string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to save layout: %w", err)
	}

	return nil
}

// DeleteLayout soft-deletes a layout
func (ms *MetadataService) DeleteLayout(layoutID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Hard delete for now
	_, err := ms.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", constants.TableLayout), layoutID)
	if err != nil {
		return fmt.Errorf("failed to delete layout: %w", err)
	}

	return nil
}

// AssignLayoutToProfile assigns a layout to a profile
func (ms *MetadataService) AssignLayoutToProfile(profileID, objectAPIName, layoutID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Use UPSERT (INSERT ON DUPLICATE KEY UPDATE)
	// Assuming unique constraint/PK on (profile_id, object_api_name)
	query := fmt.Sprintf(`
		INSERT INTO %s (profile_id, object_api_name, layout_id) 
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE layout_id = VALUES(layout_id)
	`, constants.TableProfileLayout)

	_, err := ms.db.Exec(query, profileID, objectAPIName, layoutID)
	if err != nil {
		return fmt.Errorf("failed to assign layout: %w", err)
	}

	return nil
}

// addFieldToLayout adds a new field to the first section of the object's default layout
// NOTE: Assumes ms.mu is already locked
func (ms *MetadataService) addFieldToLayout(objectAPIName, fieldAPIName string) error {
	// Lock held by caller

	// 1. Find the default layout
	var layoutID, configJSON string
	err := ms.db.QueryRow(fmt.Sprintf("SELECT id, config FROM %s WHERE object_api_name = ? LIMIT 1", constants.TableLayout), objectAPIName).Scan(&layoutID, &configJSON)
	if err == sql.ErrNoRows {
		// No layout exists, CreateSchema usually creates one. If missing, we ignore or create?
		// For now, ignore.
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to query layout: %w", err)
	}

	// 2. Parse JSON
	var layout models.PageLayout
	if err := json.Unmarshal([]byte(configJSON), &layout); err != nil {
		return fmt.Errorf("failed to parse layout config: %w", err)
	}

	// 3. Add field to first section if not present
	if len(layout.Sections) > 0 {
		exists := false
		for _, f := range layout.Sections[0].Fields {
			if f == fieldAPIName {
				exists = true
				break
			}
		}
		if !exists {
			layout.Sections[0].Fields = append(layout.Sections[0].Fields, fieldAPIName)

			// 4. Save updated layout
			newConfigJSON, err := json.Marshal(layout)
			if err != nil {
				return fmt.Errorf("failed to marshal updated layout: %w", err)
			}

			_, err = ms.db.Exec(fmt.Sprintf("UPDATE %s SET config = ? WHERE id = ?", constants.TableLayout), string(newConfigJSON), layoutID)
			if err != nil {
				return fmt.Errorf("failed to update layout: %w", err)
			}
		}
	}

	return nil
}
