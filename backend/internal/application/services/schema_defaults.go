package services

import (
	"github.com/nexuscrm/backend/internal/domain/models"
)

// EnrichWithSystemFields adds standard system fields to an object schema if they are missing.
// This ensures that all custom objects created via the API have the required infrastructure fields.
// EnrichWithSystemFields adds standard system fields to an object schema if they are missing.
// This ensures that all custom objects created via the API have the required infrastructure fields.
// standardFields should be provided by SchemaManager (SSOT).
func EnrichWithSystemFields(schema *models.ObjectMetadata, standardFields []models.FieldMetadata) {
	// Helper to check if field exists
	hasField := func(apiName string) bool {
		for _, f := range schema.Fields {
			if f.APIName == apiName {
				return true
			}
		}
		return false
	}

	var newFields []models.FieldMetadata

	// Iterate over standard SSOT fields and add if missing
	for _, stdField := range standardFields {
		if !hasField(stdField.APIName) {
			// Create a copy to avoid pointer issues if standardFields elements are reused (safety)
			fieldCopy := stdField
			newFields = append(newFields, fieldCopy)
		}
	}

	// Prepend system fields to ensure they appear first
	schema.Fields = append(newFields, schema.Fields...)
}
