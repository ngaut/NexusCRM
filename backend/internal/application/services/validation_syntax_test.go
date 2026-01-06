package services_test

import (
	"testing"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestValidationRule_EmptyStringSyntax(t *testing.T) {
	fEngine := formula.NewEngine()
	vs := services.NewValidationService(fEngine)

	tests := []struct {
		name      string
		condition string
		record    models.SObject
		expectErr bool
	}{
		{
			name:      "Empty String Check (== \"\")",
			condition: `record.field == ""`,
			record:    models.SObject{"field": ""},
			expectErr: true,
		},
		{
			name:      "Nil Check (== nil) with String",
			condition: `record.field == nil`,
			record:    models.SObject{"field": ""},
			expectErr: false, // "" is not nil?
		},
		{
			name:      "Nil Check (== nil) with Nil",
			condition: `record.field == nil`,
			record:    models.SObject{"field": nil},
			expectErr: true,
		},
		{
			name:      "Combined Check",
			condition: `record.field == nil || record.field == ""`,
			record:    models.SObject{"field": ""},
			expectErr: true,
		},
		{
			name:      "Len Check",
			condition: `len(record.field) == 0`,
			record:    models.SObject{"field": ""},
			expectErr: true,
			// expr defaults include len?
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := []*models.ValidationRule{{
				Name:         "TestRule",
				Condition:    tt.condition,
				ErrorMessage: "Error",
				Active:       true,
			}}

			err := vs.ValidateRecord(tt.record, &models.ObjectMetadata{}, rules, nil)

			// We want the rule to return TRUE (Validation Error)
			if tt.expectErr {
				assert.Error(t, err)
				if err != nil {
					// Assert it's not a panic/internal error
					assert.Contains(t, err.Error(), "Error")
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				// Check for "failed to evaluate" which means syntax error
				if assert.NotContains(t, err.Error(), "failed to evaluate") {
					// OK
				} else {
					t.Logf("Syntax/Runtime Error: %v", err)
				}
			}
		})
	}
}
