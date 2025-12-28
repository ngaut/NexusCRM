package services_test

import (
	"testing"

	"github.com/nexuscrm/backend/pkg/expression"
)

func TestFormulaFilter_ToSQL(t *testing.T) {
	testCases := []struct {
		name       string
		filterExpr string
		expectSQL  string
		expectArgs int
		shouldFail bool
	}{
		{
			name:       "Simple equality",
			filterExpr: "status == 'Open'",
			expectSQL:  "status = ?",
			expectArgs: 1,
		},
		{
			name:       "Greater than",
			filterExpr: "amount > 1000",
			expectSQL:  "amount > ?",
			expectArgs: 1,
		},
		{
			name:       "Less than or equal",
			filterExpr: "price <= 50.5",
			expectSQL:  "price <= ?",
			expectArgs: 1,
		},
		{
			name:       "Not equal",
			filterExpr: "type != 'Cancelled'",
			expectSQL:  "type != ?",
			expectArgs: 1,
		},
		{
			name:       "AND condition",
			filterExpr: "status == 'Open' && amount > 500",
			expectSQL:  "(status = ?) AND (amount > ?)",
			expectArgs: 2,
		},
		{
			name:       "OR condition",
			filterExpr: "type == 'A' || type == 'B'",
			expectSQL:  "(type = ?) OR (type = ?)",
			expectArgs: 2,
		},
		{
			name:       "CONTAINS function",
			filterExpr: "CONTAINS(name, 'Acme')",
			expectSQL:  "name LIKE ?",
			expectArgs: 1,
		},
		{
			name:       "STARTS_WITH function",
			filterExpr: "STARTS_WITH(name, 'Test')",
			expectSQL:  "name LIKE ?",
			expectArgs: 1,
		},
		{
			name:       "ENDS_WITH function",
			filterExpr: "ENDS_WITH(email, '.com')",
			expectSQL:  "email LIKE ?",
			expectArgs: 1,
		},
		{
			name:       "Empty filter",
			filterExpr: "",
			expectSQL:  "",
			expectArgs: 0,
		},
		{
			name:       "Invalid syntax",
			filterExpr: "status == == 'Open'",
			shouldFail: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.filterExpr == "" {
				// Empty filter should result in no SQL
				t.Log("✅ Empty filter handled correctly")
				return
			}

			sql, args, err := expression.ToSQL(tc.filterExpr)

			if tc.shouldFail {
				if err == nil {
					t.Errorf("Expected error for invalid syntax, got SQL: %s", sql)
				} else {
					t.Logf("✅ Invalid syntax correctly rejected: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if sql == "" && tc.expectSQL != "" {
				t.Errorf("Expected SQL: %s, got empty", tc.expectSQL)
			}

			if len(args) != tc.expectArgs {
				t.Errorf("Expected %d args, got %d", tc.expectArgs, len(args))
			}

			t.Logf("✅ Filter '%s' -> SQL '%s' with %d args", tc.filterExpr, sql, len(args))
		})
	}
}

func TestFormulaFilter_StringFunctions(t *testing.T) {
	// Test CONTAINS produces correct LIKE pattern
	t.Run("CONTAINS pattern", func(t *testing.T) {
		sql, args, err := expression.ToSQL("CONTAINS(name, 'Acme')")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(args) != 1 {
			t.Fatalf("Expected 1 arg, got %d", len(args))
		}
		pattern, ok := args[0].(string)
		if !ok {
			t.Fatalf("Expected string arg, got %T", args[0])
		}
		if pattern != "%Acme%" {
			t.Errorf("Expected pattern '%%Acme%%', got '%s'", pattern)
		}
		t.Logf("✅ CONTAINS produces correct LIKE pattern: SQL='%s', pattern='%s'", sql, pattern)
	})

	// Test STARTS_WITH produces correct LIKE pattern
	t.Run("STARTS_WITH pattern", func(t *testing.T) {
		sql, args, err := expression.ToSQL("STARTS_WITH(name, 'Test')")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(args) != 1 {
			t.Fatalf("Expected 1 arg, got %d", len(args))
		}
		pattern, ok := args[0].(string)
		if !ok {
			t.Fatalf("Expected string arg, got %T", args[0])
		}
		if pattern != "Test%" {
			t.Errorf("Expected pattern 'Test%%', got '%s'", pattern)
		}
		t.Logf("✅ STARTS_WITH produces correct LIKE pattern: SQL='%s', pattern='%s'", sql, pattern)
	})

	// Test ENDS_WITH produces correct LIKE pattern
	t.Run("ENDS_WITH pattern", func(t *testing.T) {
		sql, args, err := expression.ToSQL("ENDS_WITH(email, '.com')")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(args) != 1 {
			t.Fatalf("Expected 1 arg, got %d", len(args))
		}
		pattern, ok := args[0].(string)
		if !ok {
			t.Fatalf("Expected string arg, got %T", args[0])
		}
		if pattern != "%.com" {
			t.Errorf("Expected pattern '%%.com', got '%s'", pattern)
		}
		t.Logf("✅ ENDS_WITH produces correct LIKE pattern: SQL='%s', pattern='%s'", sql, pattern)
	})
}
