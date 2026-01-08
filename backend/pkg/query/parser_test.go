package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFormulaQuery(t *testing.T) {
	tests := []struct {
		name           string
		term           string
		objectName     string
		expectedSQL    string
		expectedParams []interface{}
		expectedOk     bool
	}{
		{
			name:           "Simple Equality",
			term:           "status = Open",
			objectName:     "Ticket",
			expectedSQL:    "`Ticket`.`status` = ?",
			expectedParams: []interface{}{"Open"},
			expectedOk:     true,
		},
		{
			name:           "Inequality (!=)",
			term:           "priority != Low",
			objectName:     "Ticket",
			expectedSQL:    "`Ticket`.`priority` != ?",
			expectedParams: []interface{}{"Low"},
			expectedOk:     true,
		},
		{
			name:           "Inequality (<>)",
			term:           "priority <> Low",
			objectName:     "Ticket",
			expectedSQL:    "`Ticket`.`priority` != ?",
			expectedParams: []interface{}{"Low"},
			expectedOk:     true,
		},
		{
			name:           "Greater Than",
			term:           "amount > 100",
			objectName:     "Opportunity",
			expectedSQL:    "`Opportunity`.`amount` > ?",
			expectedParams: []interface{}{"100"},
			expectedOk:     true,
		},
		{
			name:           "Less Than or Equal",
			term:           "score <= 50",
			objectName:     "Lead",
			expectedSQL:    "`Lead`.`score` <= ?",
			expectedParams: []interface{}{"50"},
			expectedOk:     true,
		},
		{
			name:           "Invalid Field Name (Injection Attempt)",
			term:           "field;DROP TABLE = value",
			objectName:     "User",
			expectedSQL:    "",
			expectedParams: nil,
			expectedOk:     false,
		},
		{
			name:           "Invalid Field Name (Space)",
			term:           "field name = value",
			objectName:     "User",
			expectedSQL:    "",
			expectedParams: nil,
			expectedOk:     false,
		},
		{
			name:           "No Operator",
			term:           "justtext",
			objectName:     "User",
			expectedSQL:    "",
			expectedParams: nil,
			expectedOk:     false,
		},
		{
			name:           "Empty Term",
			term:           "",
			objectName:     "User",
			expectedSQL:    "",
			expectedParams: nil,
			expectedOk:     false,
		},
		{
			name:           "Underscore in Field Name (Valid)",
			term:           "first_name = John",
			objectName:     "Contact",
			expectedSQL:    "`Contact`.`first_name` = ?",
			expectedParams: []interface{}{"John"},
			expectedOk:     true,
		},
		{
			name:           "Numeric Field Name (Valid)",
			term:           "field1 = value",
			objectName:     "Custom",
			expectedSQL:    "`Custom`.`field1` = ?",
			expectedParams: []interface{}{"value"},
			expectedOk:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params, ok := ParseFormulaQuery(tt.term, tt.objectName)

			assert.Equal(t, tt.expectedOk, ok, "Success status mismatch")
			if tt.expectedOk {
				assert.Equal(t, tt.expectedSQL, sql, "SQL mismatch")
				assert.Equal(t, tt.expectedParams, params, "Params mismatch")
			}
		})
	}
}
