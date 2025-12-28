package expression

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEngine_Evaluate(t *testing.T) {
	e := NewEngine()

	tests := []struct {
		name     string
		expr     string
		env      map[string]interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "Simple Math",
			expr:     "1 + 1",
			env:      nil,
			expected: 2,
		},
		{
			name:     "Variable Access",
			expr:     "user.Age > 18",
			env:      map[string]interface{}{"user": map[string]interface{}{"Age": 20}},
			expected: true,
		},
		{
			name:     "Nested Access",
			expr:     "record.Account.Name",
			env:      map[string]interface{}{"record": map[string]interface{}{"Account": map[string]interface{}{"Name": "Acme"}}},
			expected: "Acme",
		},
		{
			name:     "Date Function",
			expr:     "TODAY()",
			env:      nil,
			expected: time.Now().Format("2006-01-02"),
		},
		{
			name:     "String Function",
			expr:     "LEN(name)",
			env:      map[string]interface{}{"name": "Nexus"},
			expected: 5,
		},
		{
			name:     "Ternary",
			expr:     "score > 50 ? 'Pass' : 'Fail'",
			env:      map[string]interface{}{"score": 80},
			expected: "Pass",
		},
		{
			name:     "Complex Logic",
			expr:     "(amount * 0.1) > 100",
			env:      map[string]interface{}{"amount": 2000},
			expected: true, // 200 > 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := e.Evaluate(tt.expr, tt.env)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
