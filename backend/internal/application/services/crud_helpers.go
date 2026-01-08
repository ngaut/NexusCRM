package services

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/nexuscrm/backend/pkg/utils"
)

// Scannable is an interface for something that can scan into a destination (sql.Row or sql.Rows)
type Scannable interface {
	Scan(dest ...interface{}) error
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

// ScanRows has been moved to pkg/query/scanner.go as ScanRowsToSObjects

// ==================== Execution Helpers ====================

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

// ==================== Permission Helpers ====================
