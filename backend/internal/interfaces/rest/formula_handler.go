package rest

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	appErrors "github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/shared/pkg/constants"
)

type FormulaHandler struct {
	engine *formula.Engine
}

func NewFormulaHandler() *FormulaHandler {
	return &FormulaHandler{
		engine: formula.NewEngine(),
	}
}

// EvaluateRequest represents a formula evaluation request
type EvaluateRequest struct {
	Expression string           `json:"expression" binding:"required"`
	Context    *formula.Context `json:"context"`
}

// ValidateRequest represents a formula validation request
type ValidateRequest struct {
	Expression string `json:"expression" binding:"required"`
}

// EvaluateConditionRequest represents a condition evaluation request (returns boolean)
type EvaluateConditionRequest struct {
	Expression string                 `json:"expression" binding:"required"`
	Record     map[string]interface{} `json:"record"`
}

// SubstituteRequest represents a template substitution request
type SubstituteRequest struct {
	Template string                 `json:"template" binding:"required"`
	Record   map[string]interface{} `json:"record"`
}

// Evaluate handles POST /api/formula/evaluate
func (h *FormulaHandler) Evaluate(c *gin.Context) {
	var req EvaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondAppError(c, appErrors.NewValidationError("json", err.Error()))
		return
	}

	if req.Context == nil {
		req.Context = &formula.Context{}
	}

	result, err := h.engine.Evaluate(req.Expression, req.Context)
	if err != nil {
		// Evaluate failed -> User Error (Bad Request usually)
		RespondAppError(c, appErrors.NewValidationError("expression", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"result":     result,
			"expression": req.Expression,
		},
	})
}

// EvaluateCondition handles POST /api/formula/condition
// Evaluates a condition expression and returns a boolean result
// This is used by the frontend for visibility conditions and validation rules
func (h *FormulaHandler) EvaluateCondition(c *gin.Context) {
	var req EvaluateConditionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondAppError(c, appErrors.NewValidationError("json", err.Error()))
		return
	}

	ctx := &formula.Context{
		Record: req.Record,
	}

	result, err := h.engine.Evaluate(req.Expression, ctx)
	if err != nil {
		// Condition evaluation error -> return false in data, OR return error?
		// Previous logic: "success: true, result: false, error: ...".
		// Strict consistency: Return "data: false" but maybe log warning?
		// Or return proper error?
		// Use RespondAppError effectively kills the visibility.
		// Let's return data: false as strict value.
		c.JSON(http.StatusOK, gin.H{
			"data": false,
		})
		return
	}

	// Convert result to boolean
	boolResult := toBool(result)

	c.JSON(http.StatusOK, gin.H{
		"data": boolResult,
	})
}

// Substitute handles POST /api/formula/substitute
// Replaces {fieldName} placeholders with values from the record
// Used by the frontend for template strings like "{Name} - {Email}"
func (h *FormulaHandler) Substitute(c *gin.Context) {
	var req SubstituteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	result := substituteTemplate(req.Template, req.Record)

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

// Validate handles POST /api/formula/validate
func (h *FormulaHandler) Validate(c *gin.Context) {
	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	_, err := h.engine.Compile(req.Expression)
	if err != nil {
		// Valid: false is a successful check result, not an HTTP error
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"valid": false,
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{"valid": true},
	})
}

// GetFunctions handles GET /api/formula/functions
func (h *FormulaHandler) GetFunctions(c *gin.Context) {
	functions := h.engine.GetFunctionDefinitions()

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"functions": functions,
			"count":     len(functions),
		},
	})
}

// ClearCache handles DELETE /api/formula/cache
func (h *FormulaHandler) ClearCache(c *gin.Context) {
	h.engine.ClearCache()

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			constants.FieldMessage: "Formula cache cleared successfully",
		},
	})
}

// substituteTemplate replaces {fieldName} placeholders with values from the record
func substituteTemplate(template string, record map[string]interface{}) string {
	if record == nil {
		return template
	}

	// Match {fieldName} patterns
	re := regexp.MustCompile(`\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		field := strings.Trim(match, "{}")
		if val, ok := record[field]; ok {
			if val == nil {
				return ""
			}
			return fmt.Sprintf("%v", val)
		}
		return match // Leave unchanged if field not found
	})
}

// toBool converts any value to boolean
func toBool(val interface{}) bool {
	switch v := val.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	case string:
		lower := strings.ToLower(v)
		return lower == "true" || lower == "1" || lower == "yes"
	case nil:
		return false
	default:
		return val != nil
	}
}
