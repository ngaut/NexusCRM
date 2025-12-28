package formula

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormulaEngine_BasicArithmetic(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name       string
		expression string
		context    *Context
		expected   interface{}
	}{
		{
			name:       "addition",
			expression: "1+2",
			context:    &Context{},
			expected:   3,
		},
		{
			name:       "subtraction",
			expression: "10-5",
			context:    &Context{},
			expected:   5,
		},
		{
			name:       "multiplication",
			expression: "4*5",
			context:    &Context{},
			expected:   20,
		},
		{
			name:       "division",
			expression: "100/4",
			context:    &Context{},
			expected:   float64(25),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Evaluate(tt.expression, tt.context)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormulaEngine_RecordFields(t *testing.T) {
	engine := NewEngine()

	ctx := &Context{
		Record: map[string]interface{}{
			"Amount":    100.0,
			"Discount":  10.0,
			"FirstName": "John",
			"LastName":  "Doe",
		},
	}

	tests := []struct {
		name       string
		expression string
		expected   interface{}
	}{
		{
			name:       "simple field access",
			expression: "record.Amount",
			expected:   100.0,
		},
		{
			name:       "field arithmetic",
			expression: "record.Amount*1.1",
			expected:   110.0,
		},
		{
			name:       "string concatenation",
			expression: "record.FirstName+record.LastName",
			expected:   "JohnDoe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Evaluate(tt.expression, ctx)
			assert.NoError(t, err)
			// Use InDelta for float comparisons to handle floating-point precision
			if expectedFloat, ok := tt.expected.(float64); ok {
				if resultFloat, ok := result.(float64); ok {
					assert.InDelta(t, expectedFloat, resultFloat, 0.0001)
					return
				}
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormulaEngine_Functions(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name       string
		expression string
		checkType  bool // If true, just check type, not exact value
	}{
		{
			name:       "TODAY function",
			expression: "TODAY()",
			checkType:  true,
		},
		{
			name:       "NOW function",
			expression: "NOW()",
			checkType:  true,
		},
		{
			name:       "UPPER function",
			expression: "UPPER('hello')",
		},
		{
			name:       "LOWER function",
			expression: "LOWER('WORLD')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Evaluate(tt.expression, &Context{})
			assert.NoError(t, err)
			assert.NotNil(t, result)

			if tt.name == "UPPER function" {
				assert.Equal(t, "HELLO", result)
			}
			if tt.name == "LOWER function" {
				assert.Equal(t, "world", result)
			}
		})
	}
}

func BenchmarkFormulaEvaluate(b *testing.B) {
	engine := NewEngine()
	ctx := &Context{
		Record: map[string]interface{}{
			"Amount": 100.0,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Evaluate("record.Amount*1.1", ctx)
	}
}
