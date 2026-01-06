package services_test

import (
	"testing"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestValidationRule_MissingField(t *testing.T) {
	// Setup Formula Engine
	fEngine := formula.NewEngine()
	vs := services.NewValidationService(fEngine)

	// Define Schema
	schema := &models.ObjectMetadata{
		APIName: "Opportunity",
		Fields: []models.FieldMetadata{
			{APIName: "Stage", Type: "Text"},
			{APIName: "Close_Date", Type: "Date"},
		},
	}

	// Define Validation Rule
	// Error if Stage is 'Closed' and Close_Date is missing (null)
	rules := []*models.ValidationRule{
		{
			Condition:    "Stage == 'Closed' && Close_Date == nil",
			ErrorMessage: "Close Date required",
			Active:       true,
		},
	}

	// Case 1: Field present as nil (Explicit null)
	recordExplicit := models.SObject{
		"Stage":      "Closed",
		"Close_Date": nil,
	}
	err1 := vs.ValidateRecord(recordExplicit, schema, rules, nil)
	assert.Error(t, err1, "Should fail validation when field is explicitly nil")

	// Case 2: Field missing entirely (Implicit null)
	// This simulates the UI behavior where empty fields are omitted from JSON
	recordMissing := models.SObject{
		"Stage": "Closed",
		// Close_Date is missing
	}
	err2 := vs.ValidateRecord(recordMissing, schema, rules, nil)
	if err2 != nil {
		t.Logf("Validation error (as expected?): %v", err2)
	} else {
		t.Log("No validation error returned")
	}
	assert.Error(t, err2, "Should fail validation when field is missing (implicit nil)")
}

func TestValidationRule_Malformed(t *testing.T) {
	// Setup Formula Engine
	fEngine := formula.NewEngine()
	vs := services.NewValidationService(fEngine)

	// Define Schema
	schema := &models.ObjectMetadata{
		APIName: "Opportunity",
		Fields:  []models.FieldMetadata{},
	}

	// Define Malformed Validation Rule
	rules := []*models.ValidationRule{
		{
			Condition:    "NON_EXISTENT_FUNC() == true",
			ErrorMessage: "This should fail compilation/eval",
			Active:       true,
		},
	}

	record := models.SObject{}

	// Expectation: ValidateRecord should return an error because the rule is active and malformed
	err := vs.ValidateRecord(record, schema, rules, nil)
	assert.Error(t, err, "Should return error when validation rule is malformed (Fail Closed)")
	if err != nil {
		t.Logf("Got expected system error: %v", err)
	}
}
