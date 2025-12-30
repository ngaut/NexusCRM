package services

import (
	"encoding/json"
	"regexp"
	"strconv"

	"fmt"

	"github.com/nexuscrm/shared/pkg/models"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/backend/pkg/fieldtypes"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/backend/pkg/validator"
)

// ValidationService handles record validation logic
type ValidationService struct {
	formula   *formula.Engine
	validator *validator.Registry
}

// NewValidationService creates a new ValidationService
func NewValidationService(formula *formula.Engine) *ValidationService {
	return &ValidationService{
		formula:   formula,
		validator: validator.GetRegistry(),
	}
}

// ValidateRecord performs comprehensive validation on a record
func (vs *ValidationService) ValidateRecord(
	record models.SObject,
	schema *models.ObjectMetadata,
	rules []*models.ValidationRule,
	oldRecord *models.SObject,
) error {
	// 1. Static Schema Constraints
	for _, field := range schema.Fields {
		// Skip system fields for required checks (unless input provided, but typically server-set)
		if field.IsSystem {
			continue
		}

		val, exists := record[field.APIName]

		// Required Check
		if field.Required {
			if !exists || val == nil || val == "" {
				return errors.NewValidationError(field.APIName, "is required")
			}
		}

		if val == nil {
			continue
		}

		// Type Validation & Constraints - Metadata-driven approach
		fieldType := string(field.Type)

		// Check if this field type has a validation pattern defined in metadata
		pattern, message := fieldtypes.GetValidationPattern(fieldType)
		if pattern != "" {
			if strVal, ok := val.(string); ok {
				matched, err := regexp.MatchString(pattern, strVal)
				if err == nil && !matched {
					if message == "" {
						message = "invalid format"
					}
					return errors.NewValidationError(field.APIName, message)
				}
			} else {
				return errors.NewValidationError(field.APIName, "expected string for "+fieldType)
			}
		}

		// Type-specific validations that can't be expressed as regex
		switch fieldType {
		case string(constants.FieldTypeBoolean):
			if _, ok := val.(bool); !ok {
				// Also allow string "true"/"false" or "0"/"1"
				if strVal, ok := val.(string); ok {
					if _, err := strconv.ParseBool(strVal); err != nil {
						return errors.NewValidationError(field.APIName, "expected boolean")
					}
				} else if intVal, ok := val.(int); ok {
					// Allow 0 or 1
					if intVal != 0 && intVal != 1 {
						return errors.NewValidationError(field.APIName, "expected boolean (0/1)")
					}
				} else if int64Val, ok := val.(int64); ok {
					// Allow 0 or 1
					if int64Val != 0 && int64Val != 1 {
						return errors.NewValidationError(field.APIName, "expected boolean (0/1)")
					}
				} else {
					return errors.NewValidationError(field.APIName, "expected boolean")
				}
			}
		case string(constants.FieldTypeNumber), string(constants.FieldTypeCurrency), string(constants.FieldTypePercent):
			// Numeric checks below handles values, but here we ensure type compatibility
			switch v := val.(type) {
			case float64, int, int64:
				// OK
			case string:
				if _, err := strconv.ParseFloat(v, 64); err != nil {
					return errors.NewValidationError(field.APIName, "expected numeric value")
				}
			default:
				return errors.NewValidationError(field.APIName, "expected numeric value")
			}
		}

		// Length Checks for String types
		if strVal, ok := val.(string); ok {
			if field.MinLength != nil && len(strVal) < *field.MinLength {
				return errors.NewValidationError(field.APIName, "is too short")
			}
			if field.MaxLength != nil && len(strVal) > *field.MaxLength {
				return errors.NewValidationError(field.APIName, "is too long")
			}

			// Regex Check
			if field.Regex != nil && *field.Regex != "" {
				matched, err := regexp.MatchString(*field.Regex, strVal)
				if err == nil && !matched {
					msg := "invalid format"
					if field.RegexMessage != nil {
						msg = *field.RegexMessage
					}
					return errors.NewValidationError(field.APIName, msg)
				}
			}
		}

		// Numeric Checks
		if field.MinValue != nil || field.MaxValue != nil {
			// Try to interpret as number
			var numVal float64
			isNum := false

			switch v := val.(type) {
			case float64:
				numVal = v
				isNum = true
			case int:
				numVal = float64(v)
				isNum = true
			case int64:
				numVal = float64(v)
				isNum = true
			case string:
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					numVal = f
					isNum = true
				}
			}

			if isNum {
				if field.MinValue != nil && numVal < *field.MinValue {
					return errors.NewValidationError(field.APIName, "value is too small")
				}
				if field.MaxValue != nil && numVal > *field.MaxValue {
					return errors.NewValidationError(field.APIName, "value is too large")
				}
			}
		}

		// Pluggable Validator Check
		// Uses the validator registry for extensible validation
		if field.Validator != nil && *field.Validator != "" {
			validatorName := *field.Validator
			var config map[string]interface{}

			// Parse validator config if provided
			if field.ValidatorConfig != nil && *field.ValidatorConfig != "" {
				if err := json.Unmarshal([]byte(*field.ValidatorConfig), &config); err != nil {
					// Invalid config, skip validator
					config = nil
				}
			}

			if err := vs.validator.Validate(validatorName, val, config); err != nil {
				return errors.NewValidationError(field.APIName, err.Error())
			}
		}
	}

	// 2. Custom Validation Rules
	if rules != nil {
		ctx := &formula.Context{Record: record}
		// If we have access to oldRecord (update context), we should potentially pass it to formula engine?
		// Currently formula engine context assumes Record only? Assuming minimal for now.

		for _, rule := range rules {
			if !rule.Active {
				continue
			}

			result, err := vs.formula.Evaluate(rule.Condition, ctx)
			if err != nil {
				// Log warning? For now skip invalid rules so they don't block
				continue
			}

			if isTrue, ok := result.(bool); ok && isTrue {
				return errors.NewValidationError("", rule.ErrorMessage)
			}
		}
	}

	return nil
}

// ValidateFlow checks for duplicate active triggers
func (vs *ValidationService) ValidateFlow(flow *models.Flow, existingFlows []*models.Flow) error {
	if flow.Status != constants.FlowStatusActive {
		return nil
	}

	// Check for duplicates
	// We only allow one active flow per trigger type per object (simplified rule for stability)
	if flow.TriggerType == constants.TriggerTypeRecordCreated ||
		flow.TriggerType == constants.TriggerTypeRecordUpdated ||
		flow.TriggerType == constants.TriggerTypeRecordDeleted {

		count := 0
		for _, existing := range existingFlows {
			if existing.ID != flow.ID && // Skip self
				existing.TriggerObject == flow.TriggerObject &&
				existing.TriggerType == flow.TriggerType &&
				existing.Status == constants.FlowStatusActive {
				count++
			}
		}

		if count > 0 {
			return errors.NewValidationError(constants.FieldTriggerType, fmt.Sprintf("Object '%s' already has an active flow for trigger '%s'", flow.TriggerObject, flow.TriggerType))
		}
	}
	return nil
}

// ValidateObjectMetadata enforces naming conventions
func (vs *ValidationService) ValidateObjectMetadata(obj *models.ObjectMetadata) error {
	if !isSnakeCase(obj.APIName) {
		return errors.NewValidationError(constants.FieldAPIName, fmt.Sprintf("Invalid API name '%s': must be in snake_case (lowercase letters, numbers, underscores)", obj.APIName))
	}
	return nil
}

// ValidateFieldMetadata enforces naming conventions
func (vs *ValidationService) ValidateFieldMetadata(field *models.FieldMetadata) error {
	if !isSnakeCase(field.APIName) {
		return errors.NewValidationError(constants.FieldAPIName, fmt.Sprintf("Invalid API name '%s': must be in snake_case (lowercase letters, numbers, underscores)", field.APIName))
	}
	return nil
}

func isSnakeCase(s string) bool {
	validName := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	return validName.MatchString(s)
}
