package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/pkg/constants"
)

// NormalizeSObject keys to match schema (case-insensitive) and coerce values to correct type.
// Returns a new SObject with normalized data.
func NormalizeSObject(schema *models.ObjectMetadata, input models.SObject) models.SObject {
	normalized := make(models.SObject)

	for k, v := range input {
		found := false
		field := FindField(schema, k)
		if field != nil {
			f := *field
			// Key match found

			// Type Coercion
			var val interface{} = v
			if v != nil {
				// Handle Number/Currency/Percent
				isNumber := f.Type == constants.FieldTypeNumber ||
					f.Type == constants.FieldTypeCurrency ||
					f.Type == constants.FieldTypePercent

				if isNumber {
					if s, ok := v.(string); ok && s != "" {
						// Attempt parse
						if fVal, err := strconv.ParseFloat(s, 64); err == nil {
							val = fVal
						}
					}
				}
				// Handle Boolean
				isBoolean := f.Type == constants.FieldTypeBoolean
				if isBoolean {
					if s, ok := v.(string); ok {
						if s == "true" {
							val = true
						} else if s == "false" {
							val = false
						}
					}
				}
			}

			normalized[f.APIName] = val
			found = true
		}

		if !found {
			// Keep original if not found in schema (preserving extra data like 'id' if needed)
			normalized[k] = v
		}
	}
	return normalized
}

// ToStorageRecord prepares an SObject for database insertion/update.
// Converts types (Boolean -> 1/0, JSON -> string) and filters virtual fields.
func ToStorageRecord(schema *models.ObjectMetadata, data models.SObject) models.SObject {
	result := make(models.SObject)

	for key, val := range data {
		// Find field metadata (case-insensitive)
		fieldMeta := FindField(schema, key)

		// PersistenceService ensures system fields are populated.
		// We strictly adhere to the schema; however, ID is always allowed to pass through if present.
		if fieldMeta == nil {
			// ID is special, always pass if present
			if strings.EqualFold(key, constants.FieldID) {
				result[constants.FieldID] = val
			} else if strings.HasSuffix(key, "_type") {
				// Polymorphic type fields are also allowed to pass through if they follow the pattern <field>_type
				// Check if the base field exists and is polymorphic
				baseField := strings.TrimSuffix(key, "_type")
				baseMeta := FindField(schema, baseField)
				if baseMeta != nil && baseMeta.IsPolymorphic {
					result[key] = val
				}
			}
			continue
		}

		// Skip virtual fields
		if fieldMeta.Type == constants.FieldTypeFormula || fieldMeta.Type == constants.FieldTypeRollupSummary {
			continue
		}

		// Use the correct column name from metadata (snake_case)
		columnName := fieldMeta.APIName

		// Convert boolean to int
		if fieldMeta.Type == constants.FieldTypeBoolean {
			if b, ok := val.(bool); ok {
				if b {
					result[columnName] = 1
				} else {
					result[columnName] = 0
				}
			} else {
				// Fallback safety check if normalization missed a type (e.g. nil).
				if val == nil {
					result[columnName] = 0
				} else {
					// Fallback to 0 if not bool
					result[columnName] = 0
				}
			}
			continue
		}

		// Convert JSON to string (for database driver support)
		if fieldMeta.Type == constants.FieldTypeJSON {
			if val == nil {
				result[columnName] = nil
				continue
			}
			// If it's already a string, leave it
			if _, ok := val.(string); ok {
				result[columnName] = val
				continue
			}
			// Marshal map/slice to JSON string
			if bytes, err := json.Marshal(val); err == nil {
				result[columnName] = string(bytes)
			}
			continue
		}

		result[columnName] = val
	}
	return result
}

// FromStorageRecord hydrates an SObject from a database row (converting 1/0 to bool, etc).
// Formula hydration is handled by QueryService. This function focuses on static type conversion.
func FromStorageRecord(schema *models.ObjectMetadata, record models.SObject, visibleFields []string) models.SObject {
	// Filter visible fields if provided
	if len(visibleFields) > 0 {
		filtered := make(models.SObject)
		for _, f := range visibleFields {
			if val, ok := record[f]; ok {
				filtered[f] = val
			}
		}
		// Also keep ID if not in visible list but in record?
		if val, ok := record[constants.FieldID]; ok {
			filtered[constants.FieldID] = val
		}
		// Also keep polymorphic _type fields for visible polymorphic fields
		for _, f := range visibleFields {
			fieldMeta := FindField(schema, f)
			if fieldMeta != nil && fieldMeta.IsPolymorphic {
				typeCol := GetPolymorphicTypeColumnName(f)
				if val, ok := record[typeCol]; ok {
					filtered[typeCol] = val
				}
			}
		}
		record = filtered
	}

	// Iterate to convert types
	for _, field := range schema.Fields {
		// Skip if not in record
		val, exists := record[field.APIName]
		if !exists || val == nil {
			continue
		}

		if field.Type == constants.FieldTypeBoolean {
			// Convert int/int64 to bool
			switch v := val.(type) {
			case int64:
				record[field.APIName] = v != 0
			case int:
				record[field.APIName] = v != 0
			case bool:
				record[field.APIName] = v
			case []uint8: // MySQL sometimes returns bytes for bits/bools
				// simplified check
				record[field.APIName] = len(v) > 0 && v[0] != 0 // very rough, depends on driver
			default:
				// Try string "1" or "0"
				s := fmt.Sprintf("%v", v)
				record[field.APIName] = s == "1" || s == "true"
			}
		}

		// Handle Numeric types which might be bytes or strings
		if field.Type == constants.FieldTypeNumber ||
			field.Type == constants.FieldTypeCurrency ||
			field.Type == constants.FieldTypePercent {

			// Check for []uint8 (bytes from DB)
			if b, ok := val.([]uint8); ok {
				if s := string(b); s != "" {
					if f, err := strconv.ParseFloat(s, 64); err == nil {
						record[field.APIName] = f
					}
				}
			} else if s, ok := val.(string); ok && s != "" {
				if f, err := strconv.ParseFloat(s, 64); err == nil {
					record[field.APIName] = f
				}
			}
		}

		// Boolean hydration is the primary requirement here.
	}

	return record
}

// FindField finds a field in the schema by name (case-insensitive)
func FindField(schema *models.ObjectMetadata, fieldName string) *models.FieldMetadata {
	if schema == nil || schema.Fields == nil {
		return nil
	}
	for _, field := range schema.Fields {
		if strings.EqualFold(field.APIName, fieldName) {
			return &field
		}
	}
	return nil
}

// GetPolymorphicTypeColumnName returns the standardized column name for the type discriminator of a polymorphic field
func GetPolymorphicTypeColumnName(fieldAPIName string) string {
	return fieldAPIName + "_type"
}
