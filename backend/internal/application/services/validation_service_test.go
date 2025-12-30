package services

import (
	"testing"

	"github.com/nexuscrm/shared/pkg/models"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/stretchr/testify/assert"
)

func TestValidationService_ValidateRecord(t *testing.T) {
	vs := NewValidationService(formula.NewEngine())

	schema := &models.ObjectMetadata{
		Fields: []models.FieldMetadata{
			{APIName: "name", Required: true, Type: "Text"},
			{APIName: "age", Type: "Number", MinValue: floatPtr(18)},
			{APIName: "code", Type: "Text", Regex: strPtr("^[A-Z]{3}$")},
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

func TestValidationService_ValidateFlow(t *testing.T) {
	vs := NewValidationService(formula.NewEngine())

	existing := []*models.Flow{
		{ID: "f1", TriggerObject: "account", TriggerType: constants.TriggerTypeRecordCreated, Status: "Active"},
	}

	tests := []struct {
		name      string
		flow      *models.Flow
		existing  []*models.Flow
		expectErr bool
	}{
		{
			name:      "Allow different trigger type",
			flow:      &models.Flow{ID: "f2", TriggerObject: "account", TriggerType: constants.TriggerTypeRecordUpdated, Status: "Active"},
			existing:  existing,
			expectErr: false,
		},
		{
			name:      "Allow different object",
			flow:      &models.Flow{ID: "f3", TriggerObject: "contact", TriggerType: constants.TriggerTypeRecordCreated, Status: "Active"},
			existing:  existing,
			expectErr: false,
		},
		{
			name:      "Deny duplicate trigger",
			flow:      &models.Flow{ID: "f4", TriggerObject: "account", TriggerType: constants.TriggerTypeRecordCreated, Status: "Active"},
			existing:  existing,
			expectErr: true,
		},
		{
			name:      "Allow self update",
			flow:      &models.Flow{ID: "f1", TriggerObject: "account", TriggerType: constants.TriggerTypeRecordCreated, Status: "Active"},
			existing:  existing,
			expectErr: false,
		},
		{
			name:      "Allow duplicate if inactive",
			flow:      &models.Flow{ID: "f5", TriggerObject: "account", TriggerType: constants.TriggerTypeRecordCreated, Status: "Draft"},
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
