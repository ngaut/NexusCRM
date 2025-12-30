package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/nexuscrm/shared/pkg/models"
	"github.com/nexuscrm/shared/pkg/constants"
)

// RollupService handles rollup summary field calculations
type RollupService struct {
	db        *sql.DB
	metadata  *MetadataService
	txManager *TransactionManager
}

// NewRollupService creates a new RollupService
func NewRollupService(db *sql.DB, metadata *MetadataService, txManager *TransactionManager) *RollupService {
	return &RollupService{
		db:        db,
		metadata:  metadata,
		txManager: txManager,
	}
}

// ProcessRollups orchestration method.
// The tx parameter ensures rollup updates participate in the caller's transaction for ACID compliance.
// Pass nil for tx to execute outside a transaction (not recommended for data consistency).
func (rs *RollupService) ProcessRollups(ctx context.Context, tx *sql.Tx, childObjName string, childRecord models.SObject) error {
	affected, err := rs.FindAffectedRollups(childObjName, childRecord)
	if err != nil {
		return err
	}

	for _, item := range affected {
		newVal, err := rs.CalculateRollup(ctx, item, tx)
		if err != nil {
			return fmt.Errorf("failed to calculate rollup %s.%s: %w", item.ParentObjName, item.RollupField.APIName, err)
		}

		// Direct Update of Parent
		// Using raw SQL to avoid circular dependency with PersistenceService.Update
		// This skips validation logic on the parent, which is usually acceptable for system-calculated rollups.
		log.Printf("🔄 Updating Rollup %s.%s on %s = %v", item.ParentObjName, item.RollupField.APIName, item.ParentID, newVal)

		// Using backticks to quote identifiers (prevents issues with reserved words)
		updateQuery := fmt.Sprintf("UPDATE `%s` SET `%s` = ? WHERE `id` = ?", item.ParentObjName, item.RollupField.APIName)

		var execErr error
		if tx != nil {
			_, execErr = tx.ExecContext(ctx, updateQuery, newVal, item.ParentID)
		} else {
			_, execErr = rs.db.ExecContext(ctx, updateQuery, newVal, item.ParentID)
		}

		if execErr != nil {
			return fmt.Errorf("failed to update parent rollup %s: %w", item.ParentID, execErr)
		}
	}
	return nil
}

// AffectedRollup represents a rollup calculation that needs to be performed
type AffectedRollup struct {
	ParentObjName string
	ParentID      string
	RollupField   models.FieldMetadata
}

// FindAffectedRollups identifies parent records that need rollup recalculation
// based on a change to a child record.
func (rs *RollupService) FindAffectedRollups(childObjName string, childRecord models.SObject) ([]AffectedRollup, error) {
	var affected []AffectedRollup

	// iterate over all schemas to find objects that have a rollup summary pointing to this child object
	schemas := rs.metadata.GetSchemas()
	for _, parentSchema := range schemas {
		for _, field := range parentSchema.Fields {
			// Check if it's a Rollup Summary field
			if field.RollupConfig == nil {
				continue
			}

			// Check if it matches this child object
			if field.RollupConfig.SummaryObject != childObjName {
				continue
			}

			// Find the relationship field value on the child record
			relField := field.RollupConfig.RelationshipField
			if relField == "" {
				// Fallback: If there is only one lookup to Parent, use it?
				// For now, strict: must have relationship_field.
				continue
			}

			// Get Parent ID from Child Record
			parentIDVal, ok := childRecord[relField]
			if !ok || parentIDVal == nil {
				continue // No parent linked, nothing to roll up
			}

			parentID, ok := parentIDVal.(string)
			if !ok || parentID == "" {
				continue
			}

			affected = append(affected, AffectedRollup{
				ParentObjName: parentSchema.APIName,
				ParentID:      parentID,
				RollupField:   field,
			})
		}
	}

	return affected, nil
}

// CalculateRollup performs the aggregation query and returns the value
func (rs *RollupService) CalculateRollup(ctx context.Context, rollup AffectedRollup, tx *sql.Tx) (interface{}, error) {
	config := rollup.RollupField.RollupConfig

	// Default value for no rows
	var result interface{} = 0
	if rollup.RollupField.Type == constants.FieldTypeDate || rollup.RollupField.Type == constants.FieldTypeDateTime {
		result = nil
	}

	// 1. Build Aggregation Query
	// SELECT FUNC(field) FROM child WHERE rel_field = ? AND is_deleted = false

	funcMap := map[string]string{
		"COUNT": "COUNT",
		"SUM":   "SUM",
		"MIN":   "MIN",
		"MAX":   "MAX",
		"AVG":   "AVG",
	}

	aggFunc, ok := funcMap[strings.ToUpper(config.CalcType)]
	if !ok {
		return nil, fmt.Errorf("unsupported rollup type: %s", config.CalcType)
	}

	aggExpression := "*"
	if config.CalcType != "COUNT" {
		aggExpression = fmt.Sprintf("`%s`", config.SummaryField)
	}
	// For SUM/AVG/MIN/MAX, if field is null, we should ignore? SQL does this naturally.

	// Using backticks to quote all identifiers
	baseQuery := fmt.Sprintf("SELECT %s(%s) FROM `%s` WHERE `%s` = ? AND `%s` = false",
		aggFunc, aggExpression, config.SummaryObject, config.RelationshipField, constants.FieldIsDeleted)

	// Add optional filter
	// SECURITY NOTE: The filter is a SQL WHERE clause fragment stored in metadata.
	// It is set by admins during field configuration, not runtime user input.
	// However, proper validation should be added to MetadataService.CreateField to prevent injection.
	if config.Filter != nil && *config.Filter != "" {
		baseQuery += " AND (" + *config.Filter + ")"
	}

	// 2. Execute Query
	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, baseQuery, rollup.ParentID)
	} else {
		row = rs.db.QueryRowContext(ctx, baseQuery, rollup.ParentID)
	}

	var scanDest interface{}
	// For SUM/AVG on empty set, SQL returns NULL. We need to handle that.
	var nullFloat sql.NullFloat64
	var nullString sql.NullString // For MIN/MAX on dates/text

	// Decide scan destination based on expected return type
	switch config.CalcType {
	case "COUNT":
		// Count always returns a number, 0 if empty usually
		if err := row.Scan(&scanDest); err != nil {
			return nil, err
		}
		return scanDest, nil
	case "SUM", "AVG":
		// Return float/decimal
		if err := row.Scan(&nullFloat); err != nil {
			return nil, err
		}
		if nullFloat.Valid {
			return nullFloat.Float64, nil
		}
		return 0, nil // Default for Sum/Avg is 0
	case "MIN", "MAX":
		// Could be number, date, or text
		// Let's try scanning as string first for MAX/MIN or interface
		if err := row.Scan(&nullString); err != nil {
			return nil, err
		}
		if nullString.Valid {
			return nullString.String, nil
		}
		return nil, nil // Default for Min/Max is null
	}

	return result, nil
}
