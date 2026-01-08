package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/nexuscrm/shared/pkg/constants"
)

// RollupRepository handles the dynamic SQL execution required for calculating and applying rollups.
// It bypasses the standard PersistenceService to avoid circular dependencies and overhead for system calculations.
type RollupRepository struct {
	db *sql.DB
}

// NewRollupRepository creates a new RollupRepository
func NewRollupRepository(db *sql.DB) *RollupRepository {
	return &RollupRepository{
		db: db,
	}
}

// CalculateRollup executes the aggregation query and returns the raw result.
// It supports COUNT, SUM, MIN, MAX, AVG.
// The query is constructed dynamically based on metadata configuration.
func (r *RollupRepository) CalculateRollup(ctx context.Context, tx *sql.Tx, calcType, summaryObject, summaryField, relationField, parentID string, filter *string) (interface{}, error) {
	// 1. Build Aggregation Query
	// SELECT FUNC(field) FROM child WHERE rel_field = ? AND is_deleted = false

	funcMap := map[string]string{
		"COUNT": "COUNT",
		"SUM":   "SUM",
		"MIN":   "MIN",
		"MAX":   "MAX",
		"AVG":   "AVG",
	}

	aggFunc, ok := funcMap[strings.ToUpper(calcType)]
	if !ok {
		return nil, fmt.Errorf("unsupported rollup type: %s", calcType)
	}

	aggExpression := "*"
	if calcType != "COUNT" {
		aggExpression = fmt.Sprintf("`%s`", summaryField)
	}

	// Using backticks to quote all identifiers
	baseQuery := fmt.Sprintf("SELECT %s(%s) FROM `%s` WHERE `%s` = ? AND `%s` = false",
		aggFunc, aggExpression, summaryObject, relationField, constants.FieldIsDeleted)

	// Add optional filter
	if filter != nil && *filter != "" {
		// Note: Filter safety is validated by Service before passed here,
		// but we trust the inputs at Repository layer generally.
		// Service adds validation.
		baseQuery += " AND (" + *filter + ")"
	}

	// 2. Execute Query
	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, baseQuery, parentID)
	} else {
		row = r.db.QueryRowContext(ctx, baseQuery, parentID)
	}

	var scanDest interface{}
	// For SUM/AVG on empty set, SQL returns NULL. We need to handle that.
	var nullFloat sql.NullFloat64
	var nullString sql.NullString // For MIN/MAX on dates/text

	// Decide scan destination based on expected return type from DB perspective
	switch calcType {
	case "COUNT":
		// Count always returns a number (int)
		if err := row.Scan(&scanDest); err != nil {
			return nil, err
		}
		// Convert byte array or other types to int/float if needed, but usually driver handles scalar
		// TiDB/MySQL driver usually returns int64 for COUNT
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
		if err := row.Scan(&nullString); err != nil {
			return nil, err
		}
		if nullString.Valid {
			return nullString.String, nil
		}
		return nil, nil // Default for Min/Max is null
	}

	return 0, nil
}

// UpdateParentRollup updates the target field on the parent record with the calculated value.
func (r *RollupRepository) UpdateParentRollup(ctx context.Context, tx *sql.Tx, parentObjName, parentID, targetField string, value interface{}) error {
	// Update query: UPDATE parent SET field = ? WHERE id = ?
	updateQuery := fmt.Sprintf("UPDATE `%s` SET `%s` = ? WHERE `id` = ?", parentObjName, targetField)

	// value can be nil, int, float, string
	var execErr error
	if tx != nil {
		_, execErr = tx.ExecContext(ctx, updateQuery, value, parentID)
	} else {
		_, execErr = r.db.ExecContext(ctx, updateQuery, value, parentID)
	}

	if execErr != nil {
		return execErr
	}
	return nil
}
