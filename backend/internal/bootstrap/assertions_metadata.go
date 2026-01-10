package bootstrap

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/shared/pkg/constants"
)

// ==================== Metadata/Schema Assertions ====================

// assertObjectTableMapping verifies that every Object metadata entry has a corresponding DB table
func assertObjectTableMapping(db *sql.DB, result *AssertionResult) {
	log.Println("   📋 Checking object-table mapping verification...")

	// 1. Get all objects
	rows, err := db.Query("SELECT api_name FROM " + constants.TableObject)
	if err != nil {
		log.Printf("   ⚠️  Could not query objects: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}

		// 2. Check if table exists
		_, err := db.Exec("SELECT 1 FROM `" + tableName + "` LIMIT 1")
		if err != nil {
			result.Violations = append(result.Violations, AssertionViolation{
				Category:    "SchemaMismatch",
				Severity:    constants.SeverityError,
				Object:      tableName,
				Description: fmt.Sprintf("Metadata exists for object '%s', but the database table is missing.", tableName),
			})
		}
	}
}

// assertFieldColumnMapping verifies that every metadata field has a corresponding physical column
func assertFieldColumnMapping(db *sql.DB, result *AssertionResult) {
	log.Println("   📋 Checking field-column mapping verification...")

	// 1. Get all objects
	rows, err := db.Query("SELECT id, api_name FROM " + constants.TableObject)
	if err != nil {
		log.Printf("   ⚠️  Could not query objects: %v", err)
		return
	}
	defer rows.Close()

	type ObjInfo struct{ ID, Name string }
	var objects []ObjInfo
	for rows.Next() {
		var o ObjInfo
		if err := rows.Scan(&o.ID, &o.Name); err == nil {
			objects = append(objects, o)
		}
	}

	for _, obj := range objects {
		// Get expected columns from Field metadata
		fieldRows, err := db.Query("SELECT api_name FROM "+constants.TableField+" WHERE object_id = ? AND (`+constants.FieldIsDeleted+` = false OR `+constants.FieldIsDeleted+` IS NULL)", obj.ID)
		if err != nil {
			continue
		}
		expectedCols := make(map[string]bool)
		for fieldRows.Next() {
			var colName string
			if err := fieldRows.Scan(&colName); err == nil {
				expectedCols[colName] = true
			}
		}
		_ = fieldRows.Close()

		// Get actual columns from DB
		cols, err := db.Query("SELECT column_name FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ?", obj.Name)
		if err != nil {
			log.Printf("   ⚠️  Could not query schema for table %s: %v", obj.Name, err)
			continue
		}

		actualCols := make(map[string]bool)
		for cols.Next() {
			var colName string
			if err := cols.Scan(&colName); err == nil {
				actualCols[colName] = true
			}
		}
		_ = cols.Close()

		// Diff
		var missing []string
		for expected := range expectedCols {
			if !actualCols[expected] {
				missing = append(missing, expected)
			}
		}

		if len(missing) > 0 {
			result.Violations = append(result.Violations, AssertionViolation{
				Category:    "SchemaMismatch",
				Severity:    constants.SeverityError,
				Object:      obj.Name,
				Description: fmt.Sprintf("Table '%s' is missing physical columns for fields: %s", obj.Name, strings.Join(missing, ", ")),
			})
		}
	}
}

// assertRelationshipIntegrity ensures Lookup/MasterDetail fields point to valid objects
func assertRelationshipIntegrity(db *sql.DB, result *AssertionResult) {
	log.Println("   📋 Checking relationship integrity...")

	// Get all object API names for validation
	rows, err := db.Query("SELECT api_name FROM " + constants.TableObject)
	if err != nil {
		log.Printf("   ⚠️  Could not query objects: %v", err)
		return
	}
	defer rows.Close()
	validObjects := make(map[string]bool)
	for rows.Next() {
		var name string
		// Scan only the name
		if err := rows.Scan(&name); err != nil {
			continue
		}
		validObjects[strings.ToLower(name)] = true
	}

	// Check Lookup/MasterDetail fields
	fRows, err := db.Query(`
		SELECT o.api_name, f.api_name, f.reference_to, f.type
		FROM ` + constants.TableField + ` f
		JOIN ` + constants.TableObject + ` o ON f.object_id = o.id
		WHERE f.type IN ('Lookup', 'MasterDetail') AND (f.`+constants.FieldIsDeleted+` = false OR f.`+constants.FieldIsDeleted+` IS NULL)
	`)
	if err != nil {
		log.Printf("   ⚠️  Could not query fields: %v", err)
		return
	}
	defer func() { fRows.Close() }()

	for fRows.Next() {
		var objName, fieldName, refToJSON, fType string
		if err := fRows.Scan(&objName, &fieldName, &refToJSON, &fType); err != nil {
			continue
		}

		// reference_to is stored as JSON array
		if strings.HasPrefix(refToJSON, "[") {
			refs := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(refToJSON, "[", ""), "]", ""), "\"", "")
			targets := strings.Split(refs, ",")
			for _, target := range targets {
				target = strings.TrimSpace(target)
				if target != "" && !validObjects[strings.ToLower(target)] {
					result.Violations = append(result.Violations, AssertionViolation{
						Category:    "BrokenRelationship",
						Severity:    constants.SeverityError,
						Object:      objName,
						Description: fmt.Sprintf("Field '%s' references non-existent object '%s'", fieldName, target),
					})
				}
			}
		} else {
			// Invalid format - should be JSON array
			result.Violations = append(result.Violations, AssertionViolation{
				Category:    "InvalidFormat",
				Severity:    constants.SeverityError,
				Object:      objName,
				Description: fmt.Sprintf("Field '%s' has invalid reference_to format (expected JSON array): %s", fieldName, refToJSON),
			})
		}
	}
}

// assertConstraintConsistency checks if is_unique fields have actual DB indexes
func assertConstraintConsistency(db *sql.DB, result *AssertionResult) {
	log.Println("   📋 Checking uniqueness constraints...")

	rows, err := db.Query(`
		SELECT o.api_name, f.api_name
		FROM ` + constants.TableField + ` f
		JOIN ` + constants.TableObject + ` o ON f.object_id = o.id
		WHERE f.` + "`unique`" + ` = true AND (f.`+constants.FieldIsDeleted+` = false OR f.`+constants.FieldIsDeleted+` IS NULL)
	`)
	if err != nil {
		log.Printf("   ⚠️  Could not query unique fields: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tableName, colName string
		if err := rows.Scan(&tableName, &colName); err != nil {
			continue
		}

		// Check information_schema.statistics for non_unique = 0
		idxRows, err := db.Query(`
			SELECT index_name FROM information_schema.statistics 
			WHERE table_schema = DATABASE() 
			  AND table_name = ? 
			  AND column_name = ? 
			  AND non_unique = 0
		`, tableName, colName)

		if err != nil {
			continue
		}

		hasIndex := false
		if idxRows.Next() {
			hasIndex = true
		}
		_ = idxRows.Close()

		if !hasIndex {
			result.Violations = append(result.Violations, AssertionViolation{
				Category:    "MissingConstraint",
				Severity:    constants.SeverityWarning,
				Object:      tableName,
				Description: fmt.Sprintf("Field '%s' is marked unique but has no unique index in DB.", colName),
			})
		}
	}
}

// assertNamingConventions enforces snake_case for all API names
func assertNamingConventions(db *sql.DB, result *AssertionResult) {
	log.Println("   📋 Checking naming conventions...")

	validName := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

	// Check Objects
	rows, err := db.Query("SELECT api_name FROM " + constants.TableObject + " WHERE is_custom = true AND (`+constants.FieldIsDeleted+` = false OR `+constants.FieldIsDeleted+` IS NULL)")
	if err == nil {
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				continue
			}
			if !validName.MatchString(name) {
				result.Violations = append(result.Violations, AssertionViolation{
					Category:    "NamingConvention",
					Severity:    constants.SeverityWarning,
					Object:      name,
					Description: fmt.Sprintf("Object API name '%s' should be snake_case (lowercase entry starting with letter).", name),
				})
			}
		}
		rows.Close()
	}

	// Check Fields
	fRows, err := db.Query(`
		SELECT o.api_name, f.api_name 
		FROM ` + constants.TableField + ` f
		JOIN ` + constants.TableObject + ` o ON f.object_id = o.id
		WHERE f.is_custom = true AND (f.`+constants.FieldIsDeleted+` = false OR f.`+constants.FieldIsDeleted+` IS NULL)
	`)
	if err == nil {
		for fRows.Next() {
			var objName, fieldName string
			if err := fRows.Scan(&objName, &fieldName); err != nil {
				continue
			}
			if !validName.MatchString(fieldName) {
				result.Violations = append(result.Violations, AssertionViolation{
					Category:    "NamingConvention",
					Severity:    constants.SeverityWarning,
					Object:      objName,
					Description: fmt.Sprintf("Custom field '%s' should be snake_case.", fieldName),
				})
			}
		}
		fRows.Close()
	}
}

// assertFormulaValidity checks if formula expressions can be parsed
func assertFormulaValidity(db *sql.DB, result *AssertionResult) {
	log.Println("   📋 Checking formula validity...")

	engine := formula.NewEngine()

	rows, err := db.Query(`
		SELECT o.api_name, f.api_name, f.formula_expression
		FROM ` + constants.TableField + ` f
		JOIN ` + constants.TableObject + ` o ON f.object_id = o.id
		WHERE f.formula_expression IS NOT NULL AND f.formula_expression != '' AND (f.`+constants.FieldIsDeleted+` = false OR f.`+constants.FieldIsDeleted+` IS NULL)
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var objName, fieldName, expression string
		if err := rows.Scan(&objName, &fieldName, &expression); err != nil {
			log.Printf("Failed to scan formula row: %v", err)
			return
		}

		// Validate syntax (dry run)
		if err := engine.Validate(expression, map[string]interface{}{}); err != nil {
			result.Violations = append(result.Violations, AssertionViolation{
				Category:    "InvalidFormula",
				Severity:    constants.SeverityError,
				Object:      objName,
				Description: fmt.Sprintf("Field '%s' has invalid formula: %v", fieldName, err),
			})
		}
	}
}
