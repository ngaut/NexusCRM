package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// ToBool safely converts various types to boolean
// Handles bool, int, int64, float64, string ("1", "true", "yes", "on")
func ToBool(val interface{}) bool {
	if val == nil {
		return false
	}

	switch v := val.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int64:
		return v != 0
	case int32:
		return v != 0
	case float64:
		return v != 0
	case float32:
		return v != 0
	case []byte:
		// Handle raw DB bytes often returned for TINYINT
		str := string(v)
		return parseBoolString(str)
	case string:
		return parseBoolString(v)
	default:
		// Fallback: try string conversion
		str := fmt.Sprintf("%v", v)
		return parseBoolString(str)
	}
}

// parseBoolString parses boolean from string representation
func parseBoolString(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	if lower == "1" || lower == "true" || lower == "yes" || lower == "on" || lower == "t" {
		return true
	}
	// Try strconv for edge cases
	if b, err := strconv.ParseBool(lower); err == nil {
		return b
	}
	return false
}
