package services

import (
	"testing"

	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestValidationService_ValidateRecord(t *testing.T) {
	vs := NewValidationService(formula.NewEngine())

	schema := &models.ObjectMetadata{
		Fields: []models.FieldMetadata{
			{APIName: "name", Required: true, Type: constants.FieldTypeText},
			{APIName: "age", Type: constants.FieldTypeNumber, MinValue: floatPtr(18)},
			{APIName: "code", Type: constants.FieldTypeText, Regex: strPtr("^[A-Z]{3}$")},
		},
	}

	tests := []struct {
		name      string
		record    models.SObject
		rules     []*models.ValidationRule
		expectErr bool
	}{
		{
			name: "Valid Record",
			record: models.SObject{
				"name": "John",
				"age":  25,
				"code": "ABC",
			},
			expectErr: false,
		},
		{
			name: "Missing Required",
			record: models.SObject{
				"age": 25,
			},
			expectErr: true,
		},
		{
			name: "Too Young",
			record: models.SObject{
				"name": "Kid",
				"age":  10,
			},
			expectErr: true,
		},
		{
			name: "Invalid Regex",
			record: models.SObject{
				"name": "BadCode",
				"code": "abc", // lower case
			},
			expectErr: true,
		},
		{
			name: "Validation Rule Failure",
			record: models.SObject{
				"name": "RuleBreaker",
				"age":  20,
			},
			rules: []*models.ValidationRule{
				{
					Active:       true,
					Condition:    "age < 21",
					ErrorMessage: "Must be 21 or older",
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vs.ValidateRecord(tt.record, schema, tt.rules, nil)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func floatPtr(v float64) *float64 { return &v }
func strPtr(v string) *string     { return &v }

// ============================================================================
// Null Keyword Support Tests
// ============================================================================

func TestValidationService_NullKeyword(t *testing.T) {
	vs := NewValidationService(formula.NewEngine())

	schema := &models.ObjectMetadata{
		APIName: "test_object",
		Fields: []models.FieldMetadata{
			{APIName: "name", Type: constants.FieldTypeText},
			{APIName: "close_date", Type: constants.FieldTypeDate},
		},
	}

	tests := []struct {
		name      string
		record    models.SObject
		rules     []*models.ValidationRule
		expectErr bool
		errMsg    string
	}{
		{
			name: "Null keyword works with nil value",
			record: models.SObject{
				"name":       "Test",
				"close_date": nil,
			},
			rules: []*models.ValidationRule{
				{
					Active:       true,
					Condition:    "close_date == null",
					ErrorMessage: "Close date is required",
				},
			},
			expectErr: true, // Rule triggers because close_date IS null
			errMsg:    "Close date is required",
		},
		{
			name: "Null keyword works with present value",
			record: models.SObject{
				"name":       "Test",
				"close_date": "2026-01-15",
			},
			rules: []*models.ValidationRule{
				{
					Active:       true,
					Condition:    "close_date == null",
					ErrorMessage: "Close date is required",
				},
			},
			expectErr: false, // Rule does NOT trigger because close_date has a value
		},
		{
			name: "Not null check works",
			record: models.SObject{
				"name":       "Test",
				"close_date": "2026-01-15",
			},
			rules: []*models.ValidationRule{
				{
					Active:       true,
					Condition:    "close_date != null",
					ErrorMessage: "Close date must be empty",
				},
			},
			expectErr: true, // Rule triggers because close_date is NOT null
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vs.ValidateRecord(tt.record, schema, tt.rules, nil)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Empty String Normalization Tests
// ============================================================================

func TestValidationService_EmptyStringNormalization(t *testing.T) {
	vs := NewValidationService(formula.NewEngine())

	schema := &models.ObjectMetadata{
		APIName: "test_object",
		Fields: []models.FieldMetadata{
			{APIName: "name", Type: constants.FieldTypeText},
			{APIName: "close_date", Type: constants.FieldTypeDate},
			{APIName: "amount", Type: constants.FieldTypeNumber},
			{APIName: "is_active", Type: constants.FieldTypeBoolean},
			{APIName: "description", Type: constants.FieldTypeText}, // Text fields should NOT be normalized
		},
	}

	tests := []struct {
		name      string
		record    models.SObject
		rules     []*models.ValidationRule
		expectErr bool
		errMsg    string
	}{
		{
			name: "Empty date string treated as null",
			record: models.SObject{
				"name":       "Test",
				"close_date": "", // Empty string should become nil
			},
			rules: []*models.ValidationRule{
				{
					Active:       true,
					Condition:    "close_date == null",
					ErrorMessage: "Close date is empty",
				},
			},
			expectErr: true, // Rule triggers because empty string -> nil -> null
		},
		{
			name: "Empty number string treated as null",
			record: models.SObject{
				"name":   "Test",
				"amount": "", // Empty string should become nil
			},
			rules: []*models.ValidationRule{
				{
					Active:       true,
					Condition:    "amount == null",
					ErrorMessage: "Amount is empty",
				},
			},
			expectErr: true,
		},
		{
			name: "Empty text string NOT treated as null (preserved)",
			record: models.SObject{
				"name":        "Test",
				"description": "", // Empty string should be PRESERVED for text
			},
			rules: []*models.ValidationRule{
				{
					Active:       true,
					Condition:    "description == null",
					ErrorMessage: "Description is null",
				},
			},
			expectErr: false, // Rule does NOT trigger because text "" is NOT converted to nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vs.ValidateRecord(tt.record, schema, tt.rules, nil)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Fail-Closed Behavior Tests
// ============================================================================

func TestValidationService_FailClosed(t *testing.T) {
	vs := NewValidationService(formula.NewEngine())

	schema := &models.ObjectMetadata{
		APIName: "test_object",
		Fields: []models.FieldMetadata{
			{APIName: "name", Type: constants.FieldTypeText},
		},
	}

	tests := []struct {
		name      string
		record    models.SObject
		rules     []*models.ValidationRule
		expectErr bool
		errType   string // "validation" or "internal"
	}{
		{
			name:   "Syntax error causes internal error",
			record: models.SObject{"name": "Test"},
			rules: []*models.ValidationRule{
				{
					Active:       true,
					Condition:    "name ==== 'Test'", // Invalid syntax
					ErrorMessage: "Should never see this",
				},
			},
			expectErr: true,
			errType:   "internal",
		},
		{
			name:   "Inactive rules are skipped even with bad syntax",
			record: models.SObject{"name": "Test"},
			rules: []*models.ValidationRule{
				{
					Active:       false,              // Inactive
					Condition:    "name ==== 'Test'", // Invalid but won't be evaluated
					ErrorMessage: "Should never see this",
				},
			},
			expectErr: false,
		},
		{
			name:   "Valid rule that passes",
			record: models.SObject{"name": "Test"},
			rules: []*models.ValidationRule{
				{
					Active:       true,
					Condition:    "name == 'SomeOtherName'", // Condition is false, so no error
					ErrorMessage: "Name mismatch",
				},
			},
			expectErr: false,
		},
		{
			name:   "Valid rule that triggers",
			record: models.SObject{"name": "Test"},
			rules: []*models.ValidationRule{
				{
					Active:       true,
					Condition:    "name == 'Test'", // Condition is true, triggers validation error
					ErrorMessage: "Name cannot be Test",
				},
			},
			expectErr: true,
			errType:   "validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vs.ValidateRecord(tt.record, schema, tt.rules, nil)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errType == "internal" {
					assert.Contains(t, err.Error(), "failed to evaluate")
				} else if tt.errType == "validation" {
					assert.Contains(t, err.Error(), tt.rules[0].ErrorMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationService_ValidateFlow(t *testing.T) {
	vs := NewValidationService(formula.NewEngine())

	existing := []*models.Flow{
		{ID: "f1", TriggerObject: "account", TriggerType: constants.TriggerTypeRecordCreated, Status: constants.FlowStatusActive},
	}

	tests := []struct {
		name      string
		flow      *models.Flow
		existing  []*models.Flow
		expectErr bool
	}{
		{
			name:      "Allow different trigger type",
			flow:      &models.Flow{ID: "f2", TriggerObject: "account", TriggerType: constants.TriggerTypeRecordUpdated, Status: constants.FlowStatusActive},
			existing:  existing,
			expectErr: false,
		},
		{
			name:      "Allow different object",
			flow:      &models.Flow{ID: "f3", TriggerObject: "contact", TriggerType: constants.TriggerTypeRecordCreated, Status: constants.FlowStatusActive},
			existing:  existing,
			expectErr: false,
		},
		{
			name:      "Deny duplicate trigger",
			flow:      &models.Flow{ID: "f4", TriggerObject: "account", TriggerType: constants.TriggerTypeRecordCreated, Status: constants.FlowStatusActive},
			existing:  existing,
			expectErr: true,
		},
		{
			name:      "Allow self update",
			flow:      &models.Flow{ID: "f1", TriggerObject: "account", TriggerType: constants.TriggerTypeRecordCreated, Status: constants.FlowStatusActive},
			existing:  existing,
			expectErr: false,
		},
		{
			name:      "Allow duplicate if inactive",
			flow:      &models.Flow{ID: "f5", TriggerObject: "account", TriggerType: constants.TriggerTypeRecordCreated, Status: constants.FlowStatusDraft},
			existing:  existing,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vs.ValidateFlow(tt.flow, tt.existing)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationService_Naming(t *testing.T) {
	vs := NewValidationService(formula.NewEngine())

	tests := []struct {
		name      string
		apiName   string
		expectErr bool
	}{
		{"Valid", "valid_name_123", false},
		{"CamelCase", "BadName", true},
		{"Spaces", "bad name", true},
		{"Dashes", "bad-name", true},
		{"StartNum", "1bad", true},
	}

	for _, tt := range tests {
		t.Run("Object_"+tt.name, func(t *testing.T) {
			err := vs.ValidateObjectMetadata(&models.ObjectMetadata{APIName: tt.apiName})
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
		t.Run("Field_"+tt.name, func(t *testing.T) {
			err := vs.ValidateFieldMetadata(&models.FieldMetadata{APIName: tt.apiName})
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
