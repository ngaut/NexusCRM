package services_test

import (
	"testing"
	"time"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestValidationRule_DateComparison(t *testing.T) {
	// Setup Formula Engine
	fEngine := formula.NewEngine()
	vs := services.NewValidationService(fEngine)

	// Define Schema
	schema := &models.ObjectMetadata{
		APIName: "Opportunity",
		Fields: []models.FieldMetadata{
			{APIName: "Close_Date", Type: "Date"},
			{APIName: "Amount", Type: "Currency"},
		},
	}

	// Rule: Close_Date cannot be in the past
	// Formula: Close_Date < TODAY()
	// Error if True
	rules := []*models.ValidationRule{
		{
			Condition:    "Close_Date < TODAY()",
			ErrorMessage: "Close Date cannot be in the past",
			Active:       true,
		},
	}

	// Case 1: Close_Date is yesterday (String format, as from JSON)
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	recordPast := models.SObject{
		"Close_Date": yesterday,
		"Amount":     1000,
	}

	// This SHOULD return an error because condition "yesterday < today" is TRUE
	err := vs.ValidateRecord(recordPast, schema, rules, nil)
	assert.Error(t, err, "Should fail when date is in the past")
	if err != nil {
		t.Logf("Got expected past date error: %v", err)
	}

	// Case 2: Close_Date is tomorrow (String format)
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	recordFuture := models.SObject{
		"Close_Date": tomorrow,
		"Amount":     1000,
	}

	// This SHOULD NOT return an error
	err2 := vs.ValidateRecord(recordFuture, schema, rules, nil)
	assert.NoError(t, err2, "Should pass when date is in the future")
}
