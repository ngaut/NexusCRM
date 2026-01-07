package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nexuscrm/backend/pkg/query"
	"github.com/nexuscrm/backend/pkg/utils"
	"github.com/nexuscrm/shared/pkg/models"
)

// Scannable is an interface for something that can scan into a destination (sql.Row or sql.Rows)
type Scannable interface {
	Scan(dest ...interface{}) error
}

// Executor allows executing SQL queries (implemented by *sql.DB, *sql.Tx, *database.TiDBConnection)
type Executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	// Context-aware methods
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// ==================== ID Generation ====================

// GenerateID creates a new UUID string
func GenerateID() string {
	return utils.GenerateID()
}

// ==================== Timestamp Helpers ======================================

// NowTimestamp returns the current time in database-friendly format
func NowTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// NowTime returns current time.Time
func NowTime() time.Time {
	return time.Now()
}

// ==================== Nullable Field Scanners ====================

// ScanNullString safely extracts a value from sql.NullString
func ScanNullString(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// ScanNullStringValue safely extracts a string value from sql.NullString (empty if null)
func ScanNullStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// ScanNullBool safely extracts a value from sql.NullBool
func ScanNullBool(nb sql.NullBool) bool {
	if nb.Valid {
		return nb.Bool
	}
	return false
}

// ScanNullFloat64 safely extracts a value from sql.NullFloat64
func ScanNullFloat64(nf sql.NullFloat64) *float64 {
	if nf.Valid {
		return &nf.Float64
	}
	return nil
}

// ScanNullInt64 safely extracts a value from sql.NullInt64
func ScanNullInt64(ni sql.NullInt64) *int {
	if ni.Valid {
		val := int(ni.Int64)
		return &val
	}
	return nil
}

// ==================== String Helpers =====================================

// StringPtr returns a pointer to a string (useful for optional fields)
func StringPtr(s string) *string {
	return &s
}

// ContainsString checks if a slice contains a string (case-sensitive)
func ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ContainsStringIgnoreCase checks if a slice contains a string (case-insensitive)
// DESIGN ASSUMPTION: Object API names are case-insensitive for relationship matching.
// REST handlers often lowercase object names, but metadata stores original case.
// This function ensures robust matching regardless of case differences.
func ContainsStringIgnoreCase(slice []string, item string) bool {
	lowerItem := strings.ToLower(strings.TrimSpace(item))
	for _, s := range slice {
		if strings.ToLower(strings.TrimSpace(s)) == lowerItem {
			return true
		}
	}
	return false
}

// WrapStringToSlice converts a single string to a slice (for ReferenceTo field conversion)
func WrapStringToSlice(s string) []string {
	if s == "" {
		return nil
	}
	return []string{s}
}

// SliceToNullJSON converts a []string to sql.NullString as JSON array
func SliceToNullJSON(slice []string) sql.NullString {
	if len(slice) == 0 {
		return sql.NullString{Valid: false}
	}
	b, err := json.Marshal(slice)
	if err != nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: string(b), Valid: true}
}

// ==================== SQL Helpers ====================

// ToNullString converts a *string to sql.NullString for database operations
func ToNullString(s *string) sql.NullString {
	if s != nil {
		return sql.NullString{String: *s, Valid: true}
	}
	return sql.NullString{Valid: false}
}

// ToNullInt64 converts a *int or *int64 to sql.NullInt64
func ToNullInt64(i interface{}) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	switch v := i.(type) {
	case *int:
		if v == nil {
			return sql.NullInt64{Valid: false}
		}
		return sql.NullInt64{Int64: int64(*v), Valid: true}
	case *int64:
		if v == nil {
			return sql.NullInt64{Valid: false}
		}
		return sql.NullInt64{Int64: *v, Valid: true}
	case int:
		return sql.NullInt64{Int64: int64(v), Valid: true}
	default:
		return sql.NullInt64{Valid: false}
	}
}

// ScanRows scans SQL rows into a slice of SObject maps
func ScanRows(rows *sql.Rows) ([]models.SObject, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := make([]models.SObject, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		record := make(models.SObject)
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				record[col] = string(b)
			} else {
				record[col] = val
			}
		}

		results = append(results, record)
	}

	return results, nil
}

// ==================== Execution Helpers ====================

// ExecuteQuery executes a built query and returns SObjects.
// It handles query execution, scanning, and error wrapping.
func ExecuteQuery(ctx context.Context, executor Executor, q query.QueryResult) ([]models.SObject, error) {
	rows, err := executor.QueryContext(ctx, q.SQL, q.Params...)
	if err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	defer func() { _ = rows.Close() }()

	results, err := ScanRows(rows)
	if err != nil {
		return nil, fmt.Errorf("row scan error: %w", err)
	}

	return results, nil
}

// UnmarshalJSONField unmarshals a generic JSON field if valid
func UnmarshalJSONField(ns sql.NullString, target interface{}) {
	if ns.Valid && ns.String != "" {
		_ = json.Unmarshal([]byte(ns.String), target)
	}
}

// MarshalJSONOrDefault marshals an interface to string or returns default
func MarshalJSONOrDefault(v interface{}, defaultVal string) (string, error) {
	if v == nil {
		return defaultVal, nil
	}

	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	s := string(b)
	if s == "null" {
		return defaultVal, nil
	}
	return s, nil
}

// ==================== Permission Helpers ====================

// GrantInitialObjectPermissions grants default permissions for a new object to all profiles.
// This is a centralized helper to avoid duplicated permission-granting logic.
// The profileConstantAdmin parameter should be the constant for system_admin profile ID.
func GrantInitialObjectPermissions(exec Executor, objectAPIName string, tableProfile string, tableObjectPerms string, profileSystemAdmin string) error {
	// Get all Profiles
	query := fmt.Sprintf("SELECT id FROM %s", tableProfile)
	rows, err := exec.Query(query)
	if err != nil {
		return fmt.Errorf("failed to fetch profiles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var profileID string
		if err := rows.Scan(&profileID); err != nil {
			continue
		}

		// Check if System Admin for elevated permissions
		isSystemAdmin := profileID == profileSystemAdmin

		// Default permissions: CRUD for everyone, ModifyAll/ViewAll for Admin
		allowRead := true
		allowCreate := true
		allowEdit := true
		allowDelete := true
		viewAll := false
		modifyAll := false

		if isSystemAdmin {
			viewAll = true
			modifyAll = true
		}

		id := GenerateID()
		insertQuery := fmt.Sprintf(`
			INSERT INTO %s (id, profile_id, object_api_name, allow_read, allow_create, allow_edit, allow_delete, view_all, modify_all)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				allow_read = VALUES(allow_read),
				allow_create = VALUES(allow_create),
				allow_edit = VALUES(allow_edit),
				allow_delete = VALUES(allow_delete),
				view_all = VALUES(view_all),
				modify_all = VALUES(modify_all)
		`, tableObjectPerms)

		_, err := exec.Exec(insertQuery, id, profileID, objectAPIName, allowRead, allowCreate, allowEdit, allowDelete, viewAll, modifyAll)
		if err != nil {
			log.Printf("⚠️ Warning: Failed to grant permission for profile %s: %v", profileID, err)
		}
	}
	return nil
}
