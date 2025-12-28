package bootstrap

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/backend/pkg/constants"
)

// AssertionViolation represents a single design violation
type AssertionViolation struct {
	Category    string // e.g., "SystemFields", "DuplicateAction"
	Severity    string // "error" or "warning"
	Object      string // Table/Object name affected
	Description string
}

// AssertionResult contains all violations found during assertion checks
type AssertionResult struct {
	Violations []AssertionViolation
	Passed     bool
}

// RunAssertions executes all design assertions and returns violations
// By default, violations are logged as warnings. Use strictMode=true to fail on violations.
func RunAssertions(db *database.TiDBConnection, strictMode bool) (*AssertionResult, error) {
	log.Println("🔍 Running startup assertions...")

	result := &AssertionResult{
		Violations: []AssertionViolation{},
		Passed:     true,
	}

	// Run all assertions
	assertSystemFieldsConsistency(db.DB(), result)
	assertNoDuplicateActions(db.DB(), result)
	assertNoDuplicateFlows(db.DB(), result)
	assertNoOrphanedSharingRules(db.DB(), result)
	assertDefaultAppExists(db.DB(), result)
	assertDefaultThemeExists(db.DB(), result)

	assertStandardActionsExist(db.DB(), result)
	assertSystemAdminProfileExists(db.DB(), result)
	assertSystemAdminUserExists(db.DB(), result)
	assertCriticalTablesExist(db.DB(), result)
	assertObjectTableMapping(db.DB(), result)
	assertFieldColumnMapping(db.DB(), result)
	assertRelationshipIntegrity(db.DB(), result)
	assertConstraintConsistency(db.DB(), result)
	assertNamingConventions(db.DB(), result)
	assertFormulaValidity(db.DB(), result)

	// Report results
	if len(result.Violations) == 0 {
		log.Println("✅ All assertions passed")
		return result, nil
	}

	result.Passed = false
	log.Printf("⚠️  Found %d assertion violation(s):", len(result.Violations))
	for i, v := range result.Violations {
		log.Printf("   %d. [%s] %s: %s", i+1, v.Severity, v.Category, v.Description)
	}

	if strictMode {
		return result, fmt.Errorf("assertion failures in strict mode: %d violation(s)", len(result.Violations))
	}

	return result, nil
}

// assertSystemFieldsConsistency checks that all tables have required system fields
func assertSystemFieldsConsistency(db *sql.DB, result *AssertionResult) {
	log.Println("   📋 Checking system fields consistency...")

	// Required system fields for all tables
	requiredFields := constants.GetSystemFieldNames()

	// Get all tables from metadata
	rows, err := db.Query(`
		SELECT api_name FROM ` + "`" + constants.TableObject + "`" + `
		WHERE is_deleted = false OR is_deleted IS NULL
	`)
	if err != nil {
		log.Printf("   ⚠️  Could not query objects: %v", err)
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			tables = append(tables, name)
		}
	}

	// Check each table has all required fields in metadata
	for _, tableName := range tables {
		// Query fields for this object
		fieldRows, err := db.Query(`
			SELECT f.api_name 
			FROM `+"`"+constants.TableField+"`"+` f
			JOIN `+"`"+constants.TableObject+"`"+` o ON f.object_id = o.id
			WHERE o.api_name = ? AND (f.is_deleted = false OR f.is_deleted IS NULL)
		`, tableName)
		if err != nil {
			continue
		}

		existingFields := make(map[string]bool)
		for fieldRows.Next() {
			var fieldName string
			if err := fieldRows.Scan(&fieldName); err == nil {
				existingFields[fieldName] = true
			}
		}
		_ = fieldRows.Close()

		// Check for missing required fields
		var missing []string
		for _, required := range requiredFields {
			if !existingFields[required] {
				missing = append(missing, required)
			}
		}

		if len(missing) > 0 {
			result.Violations = append(result.Violations, AssertionViolation{
				Category:    "SystemFields",
				Severity:    "error",
				Object:      tableName,
				Description: fmt.Sprintf("Table '%s' missing system fields in metadata: %s", tableName, strings.Join(missing, ", ")),
			})
		}
	}
}

// assertNoDuplicateActions checks that no object has duplicate action names
func assertNoDuplicateActions(db *sql.DB, result *AssertionResult) {
	log.Println("   📋 Checking for duplicate actions...")

	rows, err := db.Query(`
		SELECT object_api_name, name, COUNT(*) as cnt
		FROM ` + "`" + constants.TableAction + "`" + `
		WHERE is_deleted = false OR is_deleted IS NULL
		GROUP BY object_api_name, name
		HAVING cnt > 1
	`)
	if err != nil {
		// Table might not have is_deleted column, try without it
		rows, err = db.Query(`
			SELECT object_api_name, name, COUNT(*) as cnt
			FROM ` + "`" + constants.TableAction + "`" + `
			GROUP BY object_api_name, name
			HAVING cnt > 1
		`)
		if err != nil {
			log.Printf("   ⚠️  Could not query actions: %v", err)
			return
		}
	}
	defer rows.Close()

	for rows.Next() {
		var objectName, actionName string
		var count int
		if err := rows.Scan(&objectName, &actionName, &count); err == nil {
			result.Violations = append(result.Violations, AssertionViolation{
				Category:    "DuplicateAction",
				Severity:    "error",
				Object:      objectName,
				Description: fmt.Sprintf("Object '%s' has %d duplicate actions named '%s'", objectName, count, actionName),
			})
		}
	}
}

// assertNoDuplicateFlows checks that no object has duplicate flow triggers
func assertNoDuplicateFlows(db *sql.DB, result *AssertionResult) {
	log.Println("   📋 Checking for duplicate flows...")

	rows, err := db.Query(`
		SELECT trigger_object, trigger_type, COUNT(*) as cnt
		FROM ` + "`" + constants.TableFlow + "`" + `
		WHERE status = 'Active' AND (is_deleted = false OR is_deleted IS NULL)
		GROUP BY trigger_object, trigger_type
		HAVING cnt > 1
	`)
	if err != nil {
		// Try without is_deleted if column doesn't exist
		rows, err = db.Query(`
			SELECT trigger_object, trigger_type, COUNT(*) as cnt
			FROM ` + "`" + constants.TableFlow + "`" + `
			WHERE status = 'Active'
			GROUP BY trigger_object, trigger_type
			HAVING cnt > 1
		`)
		if err != nil {
			log.Printf("   ⚠️  Could not query flows: %v", err)
			return
		}
	}
	defer rows.Close()

	for rows.Next() {
		var objectName, triggerType string
		var count int
		if err := rows.Scan(&objectName, &triggerType, &count); err == nil {
			result.Violations = append(result.Violations, AssertionViolation{
				Category:    "DuplicateFlow",
				Severity:    "warning",
				Object:      objectName,
				Description: fmt.Sprintf("Object '%s' has %d active flows with trigger '%s'", objectName, count, triggerType),
			})
		}
	}
}

// assertNoOrphanedSharingRules checks that sharing rules point to valid groups/users
func assertNoOrphanedSharingRules(db *sql.DB, result *AssertionResult) {
	log.Println("   📋 Checking for orphaned sharing rules...")

	// Check for sharing rules pointing to non-existent groups
	rows, err := db.Query(`
		SELECT sr.id, sr.name, sr.share_with_group_id
		FROM ` + "`" + constants.TableSharingRule + "`" + ` sr
		LEFT JOIN ` + "`" + constants.TableGroup + "`" + ` g ON sr.share_with_group_id = g.id
		WHERE sr.share_with_group_id IS NOT NULL 
		  AND g.id IS NULL
		  AND (sr.is_deleted = false OR sr.is_deleted IS NULL)
	`)
	if err != nil {
		// Try without is_deleted
		rows, err = db.Query(`
			SELECT sr.id, sr.name, sr.share_with_group_id
			FROM ` + "`" + constants.TableSharingRule + "`" + ` sr
			LEFT JOIN ` + "`" + constants.TableGroup + "`" + ` g ON sr.share_with_group_id = g.id
			WHERE sr.share_with_group_id IS NOT NULL AND g.id IS NULL
		`)
		if err != nil {
			log.Printf("   ⚠️  Could not query sharing rules: %v", err)
			return
		}
	}
	defer rows.Close()

	for rows.Next() {
		var ruleID, ruleName, groupID string
		if err := rows.Scan(&ruleID, &ruleName, &groupID); err == nil {
			result.Violations = append(result.Violations, AssertionViolation{
				Category:    "OrphanedSharingRule",
				Severity:    "warning",
				Object:      ruleName,
				Description: fmt.Sprintf("Sharing rule '%s' references non-existent group '%s'", ruleName, groupID),
			})
		}
	}
}

// assertDefaultAppExists checks if there is at least one active default app
func assertDefaultAppExists(db *sql.DB, result *AssertionResult) {
	log.Println("   📋 Checking for default application...")

	var count int
	// Check for any default app (is_default = 1)
	err := db.QueryRow("SELECT COUNT(*) FROM " + constants.TableApp + " WHERE is_default = 1").Scan(&count)
	if err != nil {
		log.Printf("   ⚠️  Could not query apps: %v", err)
		return
	}

	if count == 0 {
		// If no default, check if ANY app exists
		if err = db.QueryRow("SELECT COUNT(*) FROM " + constants.TableApp).Scan(&count); err != nil {
			log.Printf("   ⚠️  Could not count apps: %v", err)
			return
		}
		if count == 0 {
			result.Violations = append(result.Violations, AssertionViolation{
				Category:    "MissingData",
				Severity:    "warning",
				Object:      "_System_App",
				Description: "No applications defined. System will be unusable.",
			})
		} else {
			result.Violations = append(result.Violations, AssertionViolation{
				Category:    "Configuration",
				Severity:    "warning",
				Object:      "_System_App",
				Description: "No default application set.",
			})
		}
	}
}

// assertDefaultThemeExists checks if a theme exists
func assertDefaultThemeExists(db *sql.DB, result *AssertionResult) {
	log.Println("   📋 Checking for themes...")

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM " + constants.TableTheme).Scan(&count)
	if err != nil {
		log.Printf("   ⚠️  Could not query themes: %v", err)
		return
	}

	if count == 0 {
		result.Violations = append(result.Violations, AssertionViolation{
			Category:    "MissingData",
			Severity:    "warning",
			Object:      "_System_Theme",
			Description: "No themes defined. UI will not load correctly.",
		})
	}
}

// assertStandardActionsExist checks that core objects have minimal actions
func assertStandardActionsExist(db *sql.DB, result *AssertionResult) {
	log.Println("   📋 Checking for standard actions...")

	coreObjects := []string{"Account", "Contact", "Opportunity"}
	requiredActions := []string{"Edit", "Delete"}

	for _, obj := range coreObjects {
		for _, action := range requiredActions {
			var count int
			query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE object_api_name = ? AND name = ?", constants.TableAction)
			err := db.QueryRow(query, obj, action).Scan(&count)
			if err != nil {
				continue
			}

			if count == 0 {
				result.Violations = append(result.Violations, AssertionViolation{
					Category:    "MissingAction",
					Severity:    "warning",
					Object:      obj,
					Description: fmt.Sprintf("Standard object '%s' is missing '%s' action.", obj, action),
				})
			}
		}
	}
}

// assertSystemAdminProfileExists ensures the admin profile is present
func assertSystemAdminProfileExists(db *sql.DB, result *AssertionResult) {
	log.Println("   📋 Checking for System Admin profile...")

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM "+constants.TableProfile+" WHERE id = ?", constants.ProfileSystemAdmin).Scan(&count)
	if err != nil {
		log.Printf("   ⚠️  Could not query profiles: %v", err)
		return
	}

	if count == 0 {
		result.Violations = append(result.Violations, AssertionViolation{
			Category:    "MissingSystemData",
			Severity:    "error",
			Object:      "_System_Profile",
			Description: fmt.Sprintf("Critical: System Admin profile '%s' is missing.", constants.ProfileSystemAdmin),
		})
	}
}

// assertSystemAdminUserExists ensures there is at least one active admin user
func assertSystemAdminUserExists(db *sql.DB, result *AssertionResult) {
	log.Println("   📋 Checking for active System Admin user...")

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM "+constants.TableUser+" WHERE profile_id = ? AND is_active = true", constants.ProfileSystemAdmin).Scan(&count)
	if err != nil {
		log.Printf("   ⚠️  Could not query users: %v", err)
		return
	}

	if count == 0 {
		result.Violations = append(result.Violations, AssertionViolation{
			Category:    "MissingData",
			Severity:    "error",
			Object:      "_System_User",
			Description: "No active System Admin users found. You may be locked out.",
		})
	}
}

// assertCriticalTablesExist checks for presence of core system tables
func assertCriticalTablesExist(db *sql.DB, result *AssertionResult) {
	log.Println("   📋 Checking for critical system tables...")

	tables := []string{
		constants.TableObject,
		constants.TableField,
		constants.TableProfile,
		constants.TableUser,
		constants.TableApp,
	}

	for _, t := range tables {
		// Try to select 1 to verify table existence
		_, err := db.Exec("SELECT 1 FROM " + t + " LIMIT 1")
		if err != nil {
			result.Violations = append(result.Violations, AssertionViolation{
				Category:    "MissingTable",
				Severity:    "error",
				Object:      t,
				Description: fmt.Sprintf("Critical system table '%s' is missing from the database.", t),
			})
		}
	}
}

// Metadata/Schema assertion functions are in assertions_metadata.go:
// - assertObjectTableMapping, assertFieldColumnMapping
// - assertRelationshipIntegrity, assertConstraintConsistency
// - assertNamingConventions, assertFormulaValidity
