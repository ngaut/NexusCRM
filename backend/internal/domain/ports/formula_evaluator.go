package ports

import (
	"github.com/nexuscrm/backend/pkg/formula"
)

// FormulaEvaluator provides formula evaluation capabilities.
// This interface enables testing components that use formulas without
// requiring a real formula engine.
type FormulaEvaluator interface {
	// Evaluate evaluates a formula expression with the given context.
	// Returns the result or an error if evaluation fails.
	Evaluate(expression string, ctx *formula.Context) (interface{}, error)
}
