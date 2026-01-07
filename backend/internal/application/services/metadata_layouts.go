package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Layout Methods ====================

// GetLayout returns the layout for an object
func (ms *MetadataService) GetLayout(ctx context.Context, apiName string, profileID *string) *models.PageLayout {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var layout *models.PageLayout

	// 1. If profileID provided, try to find assigned layout
	if profileID != nil {
		layoutID, err := ms.repo.GetLayoutIDForProfile(ctx, *profileID, apiName)
		if err == nil && layoutID != "" {
			l, err := ms.repo.GetLayout(ctx, layoutID)
			if err == nil && l != nil {
				layout = l
			}
		}
	}

	// 2. Fallback: get all layouts for object and pick first
	if layout == nil {
		layouts, err := ms.repo.GetLayouts(ctx, apiName)
		if err == nil && len(layouts) > 0 {
			layout = layouts[0]
		}
	}

	// 3. Last resort: generate default layout from schema
	if layout == nil {
		obj, err := ms.repo.GetSchemaByAPIName(ctx, apiName)
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
		ms.augmentLayoutWithRelatedLists(ctx, layout)
	}

	return layout
}

// augmentLayoutWithRelatedLists finds objects that lookup to the current object and adds them as related lists
func (ms *MetadataService) augmentLayoutWithRelatedLists(ctx context.Context, layout *models.PageLayout) {
	results, err := ms.repo.GetRelatedListConfigs(ctx, layout.ObjectAPIName)
	if err != nil {
		log.Printf("⚠️ Failed to query child relationships for %s: %v", layout.ObjectAPIName, err)
		return
	}

	for _, res := range results {
		// Avoid duplicates if any exist
		exists := false
		for _, rl := range layout.RelatedLists {
			if rl.ObjectAPIName == res.ChildObjectAPI && rl.LookupField == res.LookupFieldAPI {
				exists = true
				break
			}
		}

		if !exists {
			// Default columns
			columns := []string{constants.FieldName, constants.FieldCreatedDate}

			// Override if relationship defines columns
			if res.RelatedListFields.Valid && res.RelatedListFields.String != "" {
				var customColumns []string
				if err := json.Unmarshal([]byte(res.RelatedListFields.String), &customColumns); err == nil && len(customColumns) > 0 {
					columns = customColumns
				}
			}

			layout.RelatedLists = append(layout.RelatedLists, models.RelatedListConfig{
				ID:            fmt.Sprintf("%s-%s", res.ChildObjectAPI, res.LookupFieldAPI),
				Label:         res.ChildPluralLabel,
				ObjectAPIName: res.ChildObjectAPI,
				LookupField:   res.LookupFieldAPI,
				Fields:        columns,
			})
		}
	}
}

// SaveLayout saves or updates a page layout
func (ms *MetadataService) SaveLayout(ctx context.Context, layout *models.PageLayout) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.repo.SaveLayout(ctx, layout)
}

// DeleteLayout soft-deletes a layout
func (ms *MetadataService) DeleteLayout(ctx context.Context, layoutID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.repo.DeleteLayout(ctx, layoutID)
}

// AssignLayoutToProfile assigns a layout to a profile
func (ms *MetadataService) AssignLayoutToProfile(ctx context.Context, profileID, objectAPIName, layoutID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.repo.AssignLayoutToProfile(ctx, profileID, objectAPIName, layoutID)
}

// addFieldToLayout adds a new field to the first section of the object's default layout
// NOTE: Assumes ms.mu is already locked
func (ms *MetadataService) addFieldToLayout(ctx context.Context, objectAPIName, fieldAPIName string) error {
	// Lock held by caller

	// 1. Find the default layout. We use SQL because GetLayouts fetches all.
	// We could use Repo methods. Repo.GetLayouts(context.Background(), objectAPIName)
	// Or define a new Repo method GetDefaultLayout?
	// For now, let's use GetLayouts and filter or just GetLayouts and pick default.
	// But `GetLayouts` returns all.
	// `metadata_layouts.go` original had `err := ms.db.QueryRow...` to get one.

	// Ideally we add `GetLayouts` usage.
	layouts, err := ms.repo.GetLayouts(ctx, objectAPIName)
	if err != nil {
		return fmt.Errorf("failed to query layout: %w", err)
	}
	if len(layouts) == 0 {
		return nil
	}
	// Pick first or specifically default. Assuming first is fine as per logic.
	layout := layouts[0]

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
			if err := ms.repo.SaveLayout(ctx, layout); err != nil {
				return fmt.Errorf("failed to update layout: %w", err)
			}
		}
	}

	return nil
}
