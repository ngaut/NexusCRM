package query

import (
	"fmt"
	"strings"
)

// ParseFormulaQuery parses simple formula expressions like "field = value" or "field > 100"
// Returns SQL condition, params, and true if successfully parsed as formula
func ParseFormulaQuery(term, objectName string) (string, []interface{}, bool) {
	// Supported operators in order of specificity (check multi-char first)
	operators := []struct {
		symbol string
		sqlOp  string
	}{
		{"!=", "!="}, {"<>", "!="}, {">=", ">="}, {"<=", "<="},
		{"=", "="}, {">", ">"}, {"<", "<"},
	}

	for _, op := range operators {
		if idx := strings.Index(term, op.symbol); idx > 0 {
			field := strings.TrimSpace(term[:idx])
			value := strings.TrimSpace(term[idx+len(op.symbol):])

			// Validate field name (basic sanity check - alphanumeric + underscore)
			if field == "" || value == "" {
				continue
			}
			for _, c := range field {
				if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
					return "", nil, false // Invalid field name
				}
			}

			// Build SQL condition
			condition := fmt.Sprintf("`%s`.`%s` %s ?", objectName, field, op.sqlOp)
			return condition, []interface{}{value}, true
		}
	}

	return "", nil, false
}
