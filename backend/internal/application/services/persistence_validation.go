package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/pkg/constants"
	appErrors "github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/backend/pkg/query"
)

// checkUniqueness checks if unique fields are unique
func (s *PersistenceService) checkUniqueness(ctx context.Context, objectName string, data models.SObject, schema *models.ObjectMetadata, excludeID string) error {
	for _, field := range schema.Fields {
		if !field.Unique {
			continue
		}

		val, ok := data[field.APIName]
		if !ok || val == nil {
			continue // Skip if not present or null
		}

		// Skip if string is empty
		if strVal, ok := val.(string); ok && strVal == "" {
			continue
		}

		// Check database using builder and ExecuteQuery
		builder := query.From(objectName).
			Select([]string{constants.FieldID}).
			Where(fmt.Sprintf("%s = ?", field.APIName), val).
			Limit(1)

		if excludeID != "" {
			builder.Where(fmt.Sprintf("%s != ?", constants.FieldID), excludeID)
		}

		q := builder.Build()
		results, err := ExecuteQuery(ctx, s.db, q)
		if err != nil {
			return err
		}

		if len(results) > 0 {
			// Found duplicate
			return appErrors.NewConflictError(objectName, field.APIName, fmt.Sprintf("%v", val))
		}
	}
	return nil
}

// areValuesEqual checks if two interface values are equal
func (s *PersistenceService) areValuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// generateSystemFields creates system field values based on metadata
// This applies system field values based on metadata definitions
func (s *PersistenceService) generateSystemFields(
	objectName string,
	data models.SObject,
	currentUser *models.UserSession,
	isInsert bool,
) models.SObject {
	// 1. Get Schema
	schema := s.metadata.GetSchema(objectName)
	if schema == nil {
		// Should not happen as validated before
		return data
	}

	result := make(models.SObject)
	// Copy existing data
	for k, v := range data {
		result[k] = v
	}

	// 2. Iterate fields and apply system/default values
	for _, field := range schema.Fields {
		fieldName := field.APIName

		// Skip if value already exists and we are not forcing system fields
		// (For system fields, we overwrite)
		_, exists := result[fieldName]

		if field.IsSystem {
			// ID
			if fieldName == constants.FieldID && isInsert && !exists {
				result[constants.FieldID] = GenerateID()
				continue
			}

			// Owner
			if fieldName == constants.FieldOwnerID {
				if isInsert && !exists {
					if currentUser != nil {
						result[constants.FieldOwnerID] = currentUser.ID
					}
				}
				continue
			}

			// Created By
			if fieldName == constants.FieldCreatedByID && isInsert {
				if currentUser != nil {
					result[constants.FieldCreatedByID] = currentUser.ID
				}
				continue
			}

			// Created Date
			if fieldName == constants.FieldCreatedDate && isInsert {
				result[constants.FieldCreatedDate] = NowTimestamp()
				continue
			}

			// Last Modified By
			if fieldName == constants.FieldLastModifiedByID {
				if currentUser != nil {
					result[constants.FieldLastModifiedByID] = currentUser.ID
				}
				continue
			}

			// Last Modified Date
			if fieldName == constants.FieldLastModifiedDate {
				result[constants.FieldLastModifiedDate] = NowTimestamp()
				continue
			}

			// Is Deleted - default to false on insert
			if fieldName == constants.FieldIsDeleted && isInsert && !exists {
				result[constants.FieldIsDeleted] = false
				continue
			}
		}

		// Defaults for non-system fields on insert (or system fields not handled above)
		if isInsert && !exists && field.DefaultValue != nil {
			val := *field.DefaultValue
			if strings.EqualFold(val, "CURRENT_TIMESTAMP") {
				result[fieldName] = NowTimestamp()
			} else {
				result[fieldName] = val
			}
		}
	}

	return result
}

// applyDefaults applies default values to fields that are missing
func (s *PersistenceService) applyDefaults(data models.SObject, schema *models.ObjectMetadata, currentUser *models.UserSession) models.SObject {
	// Wrapper around generateSystemFields for clarity
	return s.generateSystemFields(schema.APIName, data, currentUser, true)
}

// mergeRecords merges updates into base record
func (s *PersistenceService) mergeRecords(base, updates models.SObject) models.SObject {
	merged := make(models.SObject)
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range updates {
		merged[k] = v
	}
	return merged
}

// getRecordName returns the name/label of a record
func (s *PersistenceService) getRecordName(record models.SObject, schema *models.ObjectMetadata) string {
	// Find name field
	nameField := constants.FieldID // Fallback
	for _, f := range schema.Fields {
		if f.IsNameField {
			nameField = f.APIName
			break
		}
	}

	if val, ok := record[nameField]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}
