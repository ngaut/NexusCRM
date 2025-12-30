package expression

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSQL(t *testing.T) {
	tests := []struct {
		name         string
		expression   string
		expectedSQL  string
		expectedArgs []interface{}
		expectError  bool
	}{
		{
			name:         "simple equality",
			expression:   "Amount == 1000",
			expectedSQL:  "(Amount = ?)",
			expectedArgs: []interface{}{1000},
		},
		{
			name:         "simple greater than",
			expression:   "Amount > 500",
			expectedSQL:  "(Amount > ?)",
			expectedArgs: []interface{}{500},
		},
		{
			name:         "string literal",
			expression:   "Stage == 'Closed Won'",
			expectedSQL:  "(Stage = ?)",
			expectedArgs: []interface{}{"Closed Won"},
		},
		{
			name:         "logical AND",
			expression:   "Amount > 1000 && Stage == 'Closed Won'",
			expectedSQL:  "((Amount > ?) AND (Stage = ?))",
			expectedArgs: []interface{}{1000, "Closed Won"},
		},
		{
			name:         "logical OR",
			expression:   "Stage == 'New' || Stage == 'Prospecting'",
			expectedSQL:  "((Stage = ?) OR (Stage = ?))",
			expectedArgs: []interface{}{"New", "Prospecting"},
		},
		{
			name:         "mixed logic",
			expression:   "(Amount > 1000 || Probability > 0.5) && Stage != 'Lost'",
			expectedSQL:  "(((Amount > ?) OR (Probability > ?)) AND (Stage != ?))",
			expectedArgs: []interface{}{1000, 0.5, "Lost"},
		},
		{
			name:         "function UPPER",
			expression:   "UPPER(Name) == 'ACME'",
			expectedSQL:  "(UPPER(Name) = ?)",
			expectedArgs: []interface{}{"ACME"},
		},
		{
			name:         "function LOWER",
			expression:   "LOWER(City) == 'new york'",
			expectedSQL:  "(LOWER(City) = ?)",
			expectedArgs: []interface{}{"new york"},
		},
		{
			name:         "function LEN",
			expression:   "LEN(Code) > 5",
			expectedSQL:  "(CHAR_LENGTH(Code) > ?)",
			expectedArgs: []interface{}{5},
		},
		{
			name:         "function IF",
			expression:   "IF(Amount > 1000, 'High', 'Low') == 'High'",
			expectedSQL:  "(IF((Amount > ?), ?, ?) = ?)",
			expectedArgs: []interface{}{1000, "High", "Low", "High"},
		},
		{
			name:         "function TODAY",
			expression:   "CloseDate > TODAY()",
			expectedSQL:  "(CloseDate > CURDATE())",
			expectedArgs: []interface{}{},
		},
		{
			name:         "function NOW",
			expression:   "CreatedDate < NOW()",
			expectedSQL:  "(CreatedDate < NOW())",
			expectedArgs: []interface{}{},
		},
		{
			name:         "function DATE_ADD",
			expression:   "CloseDate < DATE_ADD(TODAY(), 30)",
			expectedSQL:  "(CloseDate < DATE_ADD(CURDATE(), INTERVAL ? DAY))",
			expectedArgs: []interface{}{30},
		},
		{
			name:         "null comparison IS NOT NULL",
			expression:   "id != null",
			expectedSQL:  "(id IS NOT NULL)",
			expectedArgs: []interface{}{},
		},
		{
			name:         "null comparison IS NULL",
			expression:   "name == null",
			expectedSQL:  "(name IS NULL)",
			expectedArgs: []interface{}{},
		},
		{
			name:         "null comparison combined with other conditions",
			expression:   "status == 'Open' && assignee_id != null",
			expectedSQL:  "((status = ?) AND (assignee_id IS NOT NULL))",
			expectedArgs: []interface{}{"Open"},
		},
		{
			name:        "unsupported AND keyword",
			expression:  "status == 'Open' AND assignee_id != null",
			expectError: true, // Should fail - use && instead of AND
		},
		{
			name:        "unsupported OR keyword",
			expression:  "status == 'New' OR status == 'Prospecting'",
			expectError: true, // Should fail - use || instead of OR
		},
		{
			name:        "unsupported node",
			expression:  "map(Items, {.Price})", // map/lambda not supported in SQL walker yet
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := ToSQL(tt.expression)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSQL, sql)
				assert.Equal(t, tt.expectedArgs, args)
			}
		})
	}
}
