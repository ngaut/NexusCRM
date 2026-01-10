package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nexuscrm/backend/pkg/auth"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// BulkInsertOptions configures bulk insert behavior
type BulkInsertOptions struct {
	BatchSize       int  // Records per batch (default 100)
	SkipValidation  bool // Skip record validation (for trusted imports)
	SkipFlows       bool // Skip BeforeCreate flows (faster import)
	SkipAutoNumbers bool // Skip auto-number generation (if pre-assigned)
}

// BulkInsertResult contains the result of a bulk insert operation
type BulkInsertResult struct {
	SuccessCount int      `json:"success_count"`
	FailedCount  int      `json:"failed_count"`
	Errors       []string `json:"errors,omitempty"`
}

// BulkInsert creates multiple records in a single transaction
func (ps *PersistenceService) BulkInsert(
	ctx context.Context,
	objectName string,
	records []models.SObject,
	currentUser *models.UserSession,
	options BulkInsertOptions,
) (BulkInsertResult, error) {
	result := BulkInsertResult{}

	if len(records) == 0 {
		return result, nil
	}

	// 1. Permission check (once for all records)
	schema, err := ps.prepareOperation(ctx, objectName, constants.PermCreate, currentUser)
	if err != nil {
		return result, err
	}

	// Calculate comprehensive column list from schema
	storageColumns := getStorageColumns(schema)

	// Identify unique fields for batch-level uniqueness check
	uniqueFields := make([]string, 0)
	for _, f := range schema.Fields {
		if f.IsUnique {
			uniqueFields = append(uniqueFields, f.APIName)
		}
	}
	// Map to track unique values within this batch: FieldName -> Value -> True
	batchUniqueValues := make(map[string]map[string]bool)
	for _, f := range uniqueFields {
		batchUniqueValues[f] = make(map[string]bool)
	}

	// 2. Get validation rules (once, cached)
	validationRules := ps.metadata.GetValidationRules(ctx, objectName)

	// 3. Pre-flight validation and preparation
	preparedRecords := make([]models.SObject, 0, len(records))
	for i, record := range records {
		// Apply defaults (and Generate System Fields logic - respecting input Audit fields)
		prepared := ps.applyDefaults(ctx, record, schema, currentUser)

		// Validate polymorphic lookups if not skipped
		if !options.SkipValidation {
			resolvedTypes, err := ps.validatePolymorphicLookups(ctx, prepared, schema)
			if err != nil {
				result.FailedCount++
				result.Errors = append(result.Errors, fmt.Sprintf("record %d: %v", i, err))
				continue
			}
			for fieldName, objType := range resolvedTypes {
				prepared[GetPolymorphicTypeColumnName(fieldName)] = objType
			}

			// Validate static rules
			if err := ps.validator.ValidateRecord(prepared, schema, validationRules, nil); err != nil {
				result.FailedCount++
				result.Errors = append(result.Errors, fmt.Sprintf("record %d: %v", i, err))
				continue
			}

			// Check uniqueness (Database)
			if err := ps.checkUniqueness(ctx, objectName, prepared, schema, ""); err != nil {
				result.FailedCount++
				result.Errors = append(result.Errors, fmt.Sprintf("record %d: %v", i, err))
				continue
			}

			// Check uniqueness (Batch-level)
			batchDupFound := false
			for _, fieldName := range uniqueFields {
				if val, ok := prepared[fieldName]; ok && val != nil {
					key := fmt.Sprintf("%v", val)
					if batchUniqueValues[fieldName][key] {
						result.FailedCount++
						result.Errors = append(result.Errors, fmt.Sprintf("record %d: duplicate value '%s' for unique field %s within batch", i, key, fieldName))
						batchDupFound = true
						break
					}
					batchUniqueValues[fieldName][key] = true
				}
			}
			if batchDupFound {
				continue
			}
		}

		// Generate system fields (Explicit call for any remaining logic not covered by applyDefaults?
		// Actually applyDefaults calls generateSystemFields. We already did it above.)
		// But in previous version I called it again. Let's verify applyDefaults.
		// applyDefaults calls generateSystemFields. So we don't need to call it again.
		// Wait, did I merge correct values? "applyDefaults" returns NEW OBJECT.
		// "prepared" is that object.
		// So we are good.

		// CRITICAL: Hash passwords for _System_User bulk inserts
		if strings.EqualFold(objectName, constants.TableUser) {
			if pwd, ok := prepared[constants.FieldPassword].(string); ok && pwd != "" {
				if !strings.HasPrefix(pwd, "$2") { // Not already hashed
					hashed, err := auth.HashPassword(pwd)
					if err != nil {
						result.FailedCount++
						result.Errors = append(result.Errors, fmt.Sprintf("record %d: failed to hash password: %v", i, err))
						continue
					}
					prepared[constants.FieldPassword] = string(hashed)
				}
			}
		}

		preparedRecords = append(preparedRecords, prepared)
	}

	// Check if all records failed validation
	if len(preparedRecords) == 0 && result.FailedCount > 0 {
		return result, fmt.Errorf("all %d records failed validation", result.FailedCount)
	}

	// 4. Execute transactional bulk insert
	err = ps.RunInTransaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		// Generate AutoNumbers for all records in batch (optimized)
		if !options.SkipAutoNumbers {
			if err := ps.generateAutoNumbersBatch(txCtx, tx, objectName, preparedRecords); err != nil {
				return fmt.Errorf("failed to generate auto-numbers: %w", err)
			}
		}

		// Convert to storage format
		storageRecords := make([]models.SObject, len(preparedRecords))
		for i, record := range preparedRecords {
			storageRecords[i] = ToStorageRecord(schema, record)
		}

		// Determine batch size
		batchSize := options.BatchSize
		if batchSize <= 0 {
			batchSize = 100
		}

		// Execute bulk insert with explicit columns
		if err := ps.repo.BulkInsert(txCtx, tx, objectName, storageRecords, storageColumns, batchSize); err != nil {
			return fmt.Errorf("bulk insert failed: %w", err)
		}

		// NOTE: Rollups are skipped for bulk insert for performance.
		// If rollups are needed, use single Insert or run a separate rollup recalculation job.

		return nil
	})

	if err != nil {
		return result, err
	}

	result.SuccessCount = len(preparedRecords)
	log.Printf("✨ Bulk created %d records in %s (User: %s)", result.SuccessCount, objectName, getUserID(currentUser))

	return result, nil
}

// generateAutoNumbersBatch optimizes auto-number generation by fetching a block of numbers at once
func (ps *PersistenceService) generateAutoNumbersBatch(ctx context.Context, tx *sql.Tx, objectName string, records []models.SObject) error {
	if len(records) == 0 {
		return nil
	}

	autoNumbers := ps.metadata.GetAutoNumbers(ctx, objectName)
	if len(autoNumbers) == 0 {
		return nil
	}

	for _, an := range autoNumbers {
		// 1. Lock the auto-number record
		autoNumberRecord, err := ps.repo.GetLock(ctx, tx, constants.TableAutoNumber, an.ID)
		if err != nil {
			return fmt.Errorf("failed to lock auto-number %s: %w", an.ID, err)
		}

		var currentValue int
		if autoNumberRecord != nil {
			if val, ok := autoNumberRecord[constants.FieldSysAutoNumber_CurrentNumber]; ok && val != nil {
				// Handle potential float64 from JSON/DB driver
				switch v := val.(type) {
				case int64:
					currentValue = int(v)
				case float64:
					currentValue = int(v)
				case int:
					currentValue = v
				}
			}
		}

		count := len(records)
		newValue := currentValue + count

		// 2. Update DB with new max value (single update)
		anUpdate := models.SObject{
			constants.FieldSysAutoNumber_CurrentNumber: newValue,
			constants.FieldLastModifiedDate:            time.Now().UTC(),
		}

		if err := ps.repo.Update(ctx, tx, constants.TableAutoNumber, an.ID, anUpdate); err != nil {
			return fmt.Errorf("failed to update auto-number %s: %w", an.ID, err)
		}

		// 3. Assign sequential values to records in memory
		for i := 0; i < count; i++ {
			num := currentValue + i + 1
			formatted := ps.formatAutoNumber(an.DisplayFormat, num)
			records[i][an.FieldAPIName] = formatted
		}
	}
	return nil
}

// getStorageColumns determines the list of physical database columns from the schema
// This ensures that even if input records are sparse, the bulk insert statement
// includes all relevant columns (setting NULL for missing values).
func getStorageColumns(schema *models.ObjectMetadata) []string {
	columns := make([]string, 0, len(schema.Fields)+1)
	columns = append(columns, constants.FieldID) // ID always exists
	seen := map[string]bool{constants.FieldID: true}

	for _, field := range schema.Fields {
		if seen[field.APIName] {
			continue
		}

		// Skip virtual fields
		if field.Type == constants.FieldTypeFormula || field.Type == constants.FieldTypeRollupSummary {
			continue
		}

		columns = append(columns, field.APIName)
		seen[field.APIName] = true

		// Add polymorphic type column if applicable
		if field.IsPolymorphic {
			polyCol := field.APIName + constants.PolymorphicTypeSuffix
			if !seen[polyCol] {
				columns = append(columns, polyCol)
				seen[polyCol] = true
			}
		}
	}
	return columns
}
