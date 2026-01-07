package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	pkgErrors "github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/backend/pkg/query"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// Helper methods

func (qs *QueryService) hydrateVirtualFields(
	ctx context.Context,
	rows []models.SObject,
	schema *models.ObjectMetadata,
	visibleFields []string,
	currentUser *models.UserSession,
) []models.SObject {
	// Find formula fields and lookup fields
	formulaFields := make([]models.FieldMetadata, 0)
	lookupFields := make([]models.FieldMetadata, 0)

	for _, field := range schema.Fields {
		if !ContainsString(visibleFields, field.APIName) {
			continue
		}

		isFormula := strings.EqualFold(string(field.Type), string(constants.FieldTypeFormula))
		isLookup := strings.EqualFold(string(field.Type), string(constants.FieldTypeLookup))

		if isFormula && field.Formula != nil {
			formulaFields = append(formulaFields, field)
		}
		if isLookup && field.ReferenceTo != nil && len(field.ReferenceTo) > 0 {
			lookupFields = append(lookupFields, field)
		}
	}

	// Hydrate each row
	for i := range rows {
		record := rows[i]

		// Hydrate static types (Boolean, etc.)
		record = FromStorageRecord(schema, record, visibleFields)

		// Hydrate formula fields
		for _, field := range formulaFields {
			formulaCtx := &formula.Context{
				Record: record,
			}
			if currentUser != nil {
				formulaCtx.User = map[string]interface{}{
					constants.FieldID:    currentUser.ID,
					constants.FieldName:  currentUser.Name,
					constants.FieldEmail: currentUser.Email,
				}
			}

			result, err := qs.formula.Evaluate(*field.Formula, formulaCtx)
			if err != nil {
				// Log the error for debugging/monitoring instead of silently failing
				log.Printf("⚠️ Formula evaluation error on field '%s': %v", field.APIName, err)
				record[field.APIName] = nil
			} else {
				// Coerce result based on ReturnType if specified
				record[field.APIName] = coerceFormulaResult(result, field.ReturnType)
			}
		}

		rows[i] = record
	}

	// Hydrate lookup field display names (batch query for efficiency)
	if len(lookupFields) > 0 {
		rows = qs.hydrateLookupNames(ctx, rows, lookupFields)
	}

	return rows
}

// hydrateLookupNames resolves lookup field UUIDs to display names
// It batches queries per referenced object for efficiency
func (qs *QueryService) hydrateLookupNames(ctx context.Context, rows []models.SObject, lookupFields []models.FieldMetadata) []models.SObject {
	if len(rows) == 0 || len(lookupFields) == 0 {
		return rows
	}

	// For each lookup field, collect all unique IDs and their reference object
	// Map: referenceObject -> Set of IDs
	refObjectIDs := make(map[string]map[string]bool)

	for _, field := range lookupFields {
		refObject := field.ReferenceTo[0] // Primary reference object (for polymorphic, we use first)
		if _, exists := refObjectIDs[refObject]; !exists {
			refObjectIDs[refObject] = make(map[string]bool)
		}

		for _, row := range rows {
			if val, ok := row[field.APIName]; ok && val != nil {
				if id, ok := val.(string); ok && id != "" {
					refObjectIDs[refObject][id] = true
				}
			}
		}
	}

	// For each reference object, fetch the names of all IDs
	// Map: ID -> Name
	idToName := make(map[string]string)

	for refObject, idSet := range refObjectIDs {
		if len(idSet) == 0 {
			continue
		}

		// Get the schema for the referenced object to find the name field
		refSchema := qs.metadata.GetSchema(ctx, refObject)
		if refSchema == nil {
			continue
		}

		// Find the name field
		nameField := constants.FieldName
		for _, f := range refSchema.Fields {
			if f.IsNameField {
				nameField = f.APIName
				break
			}
		}

		// Collect IDs into slice
		ids := make([]string, 0, len(idSet))
		for id := range idSet {
			ids = append(ids, id)
		}

		// Build IN query
		placeholders := make([]string, len(ids))
		params := make([]interface{}, len(ids))
		for i, id := range ids {
			placeholders[i] = "?"
			params[i] = id
		}

		sql := fmt.Sprintf("SELECT `id`, `%s` FROM `%s` WHERE `id` IN (%s)",
			nameField, refObject, strings.Join(placeholders, ","))

		q := query.QueryResult{SQL: sql, Params: params}
		results, err := ExecuteQuery(ctx, qs.db, q)
		if err != nil {
			log.Printf("⚠️ Failed to hydrate lookup names for %s: %v", refObject, err)
			continue
		}

		// Build ID -> Name map
		for _, rec := range results {
			if id, ok := rec[constants.FieldID].(string); ok {
				if name, ok := rec[nameField]; ok && name != nil {
					idToName[id] = fmt.Sprintf("%v", name)
				}
			}
		}
	}

	// Populate _Name fields on each row
	for i, row := range rows {
		for _, field := range lookupFields {
			if val, ok := row[field.APIName]; ok && val != nil {
				if id, ok := val.(string); ok {
					if name, found := idToName[id]; found {
						row[field.APIName+"_Name"] = name
					}
				}
			}
		}
		rows[i] = row
	}

	return rows
}

// Calculate evaluates formula fields for a given record
func (qs *QueryService) Calculate(
	ctx context.Context,
	objectName string,
	record models.SObject,
	user *models.UserSession,
) (models.SObject, error) {
	schema := qs.metadata.GetSchema(ctx, objectName)
	if schema == nil {
		return nil, pkgErrors.NewNotFoundError("Object", objectName)
	}

	// Prepare context
	userMap := map[string]interface{}{
		constants.FieldID:   user.ID,
		constants.FieldName: user.Name,
	}

	// Normalize record keys to match schema field names (case-insensitive match)
	// AND coerce types (e.g. string "123" to float64 for Number fields) to ensure formula engine works
	record = NormalizeSObject(schema, record)

	formulaCtx := &formula.Context{
		Record: record,
		User:   userMap,
		// Fetcher could be added here if we want relationships
	}

	// Iterate over all fields to find formulas
	for _, field := range schema.Fields {
		// Check Formula field
		expr := ""
		if field.Formula != nil {
			expr = *field.Formula
		}

		if strings.EqualFold(string(field.Type), string(constants.FieldTypeFormula)) && expr != "" {
			val, err := qs.formula.Evaluate(expr, formulaCtx)
			if err != nil {
				// We log or return? Return error for API visibility.
				// But we might want partial success?
				// For now, fail fast.
				return nil, fmt.Errorf("failed to evaluate formula %s: %v", field.APIName, err)
			}
			record[field.APIName] = val
			formulaCtx.Record[field.APIName] = val
		}
	}

	return record, nil
}

// coerceFormulaResult converts a formula result to the specified return type
func coerceFormulaResult(result interface{}, returnType *models.FieldType) interface{} {
	if result == nil || returnType == nil {
		return result
	}

	switch *returnType {
	case constants.FieldTypeNumber, constants.FieldTypeCurrency, constants.FieldTypePercent:
		// Ensure numeric types are properly formatted as float64
		switch v := result.(type) {
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case float32:
			return float64(v)
		case float64:
			return v
		case string:
			var f float64
			if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
				return f
			}
		}
	case constants.FieldTypeBoolean:
		// Ensure boolean types
		switch v := result.(type) {
		case bool:
			return v
		case int:
			return v != 0
		case int64:
			return v != 0
		case float64:
			return v != 0
		case string:
			lower := strings.ToLower(v)
			return lower == "true" || lower == "1" || lower == "yes"
		}
	case constants.FieldTypeText:
		// Convert to string - result is guaranteed non-nil at this point (checked at function start)
		return fmt.Sprintf("%v", result)
	}

	return result
}
