package formula

import (
	"encoding/json"
	"fmt"

	"github.com/nexuscrm/backend/pkg/auth"
	"github.com/nexuscrm/backend/pkg/expression"
)

// Context holds the evaluation context for formulas
type Context struct {
	Record    map[string]interface{}                                                                   `json:"record"`
	Prior     map[string]interface{}                                                                   `json:"prior"`
	User      map[string]interface{}                                                                   `json:"user"`
	Env       map[string]interface{}                                                                   `json:"env"`
	Fields    map[string]interface{}                                                                   `json:"-"` // Catch-all for unmapped fields
	Extra     map[string]json.RawMessage                                                               `json:"-"` // For custom unmarshaling
	Fetcher   func(record map[string]interface{}, relationName string) (map[string]interface{}, error) `json:"-"`
	IsVisible func(fieldName string) bool                                                              `json:"-"` // Optional FLS check
}

// UnmarshalJSON implements custom JSON unmarshaling to capture extra fields
func (c *Context) UnmarshalJSON(data []byte) error {
	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if c.Fields == nil {
		c.Fields = make(map[string]interface{})
	}

	for key, value := range temp {
		switch key {
		case "record":
			if recordMap, ok := value.(map[string]interface{}); ok {
				c.Record = recordMap
			}
		case "prior":
			if priorMap, ok := value.(map[string]interface{}); ok {
				c.Prior = priorMap
			}
		case "user":
			if userMap, ok := value.(map[string]interface{}); ok {
				c.User = userMap
			}
		case "env":
			if envMap, ok := value.(map[string]interface{}); ok {
				c.Env = envMap
			}
		default:
			c.Fields[key] = value
		}
	}
	return nil
}

// CompiledFormula represents a compiled formula ready for evaluation
// In this new engine, it's just a placeholder as caching is handled by expression engine
type CompiledFormula struct {
	Expression string
}

// Engine is the formula evaluation engine backed by expression engine
type Engine struct {
	exprEngine *expression.Engine
}

// FunctionDefinition represents a formula function definition for API responses
type FunctionDefinition struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
}

// NewEngine creates a new formula engine
func NewEngine() *Engine {
	engine := &Engine{
		exprEngine: expression.NewEngine(),
	}
	engine.registerBuiltinFunctions()
	return engine
}

// Compile compiles a formula expression - effectively a no-op/validate for now
func (e *Engine) Compile(expression string) (*CompiledFormula, error) {
	// We can validate syntax here if we want, but Evaluate handles it.
	// For API compatibility we just return a valid obj.
	return &CompiledFormula{Expression: expression}, nil
}

// ToSQL converts an expression string to a SQL WHERE clause and arguments
func ToSQL(expr string) (string, []interface{}, error) {
	return expression.ToSQL(expr)
}

// Validate validates a formula expression syntax
func (e *Engine) Validate(expression string, env map[string]interface{}) error {
	return e.exprEngine.Validate(expression, env)
}

// Evaluate evaluates a formula expression with the given context
func (e *Engine) Evaluate(expression string, ctx *Context) (interface{}, error) {
	// Flatten context for generic expression engine
	env := make(map[string]interface{})

	// 1. Add top-level objects
	if ctx.Record != nil {
		// Apply FLS filtering if IsVisible is provided
		var recordMap map[string]interface{}
		if ctx.IsVisible != nil {
			recordMap = make(map[string]interface{})
			for k, v := range ctx.Record {
				if ctx.IsVisible(k) {
					recordMap[k] = v
				}
			}
		} else {
			recordMap = ctx.Record
		}

		env["record"] = recordMap
		// Also flatten record fields into top level for convenience (e.g. Amount instead of record.Amount)
		// This matches previous behavior where plain identifiers looked in record
		for k, v := range recordMap {
			env[k] = v
		}
	}
	if ctx.Prior != nil {
		env["prior"] = ctx.Prior
	}
	if ctx.User != nil {
		env["user"] = ctx.User
	}
	if ctx.Env != nil {
		env["env"] = ctx.Env
	}

	// 2. Add extra fields (overrides)
	if ctx.Fields != nil {
		for k, v := range ctx.Fields {
			// Apply FLS to extra fields too if they are considered fields
			if ctx.IsVisible != nil && !ctx.IsVisible(k) {
				continue
			}
			env[k] = v
		}
	}

	// 3. Evaluate
	// Note: Expression engine returns generic interface{}, we might need strict typing if callers expect it?
	// The original engine returned interface{}, so this is fine.
	// We might need to ensure certain types (float64 vs int) are handled consistently but expr does a good job.
	return e.exprEngine.Evaluate(expression, env)
}

// GetFunctionDefinitions returns all registered function definitions
func (e *Engine) GetFunctionDefinitions() []FunctionDefinition {
	// Note: these are manually synced with expression/engine.go standard functions.
	// The frontend uses this for auto-complete.
	return []FunctionDefinition{
		{Name: "TODAY", Category: "Date", Description: "Returns today's date (YYYY-MM-DD)", Usage: "TODAY()"},
		{Name: "NOW", Category: "Date", Description: "Returns current date/time", Usage: "NOW()"},
		{Name: "DATE_ADD", Category: "Date", Description: "Adds days to a date", Usage: "DATE_ADD(date, days)"},
		{Name: "LEN", Category: "Text", Description: "Length of string", Usage: "LEN(text)"},
		{Name: "UPPER", Category: "Text", Description: "Converts to uppercase", Usage: "UPPER(text)"},
		{Name: "LOWER", Category: "Text", Description: "Converts to lowercase", Usage: "LOWER(text)"},
		{Name: "ROUND", Category: "Math", Description: "Rounds a number to specified precision", Usage: "ROUND(number, precision)"},
		{Name: "IF", Category: "Logic", Description: "Conditional logic", Usage: "IF(condition, true_val, false_val)"},
	}
}

// RegisterFunction delegates to expression engine
func (e *Engine) RegisterFunction(name string, fn func(params ...interface{}) (interface{}, error)) {
	e.exprEngine.RegisterFunction(name, fn)
}

// registerBuiltinFunctions registers formula-specific functions
func (e *Engine) registerBuiltinFunctions() {
	// BCRYPT for password hashing
	e.RegisterFunction("BCRYPT", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("BCRYPT requires 1 argument")
		}
		password, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("BCRYPT argument must be string")
		}
		// Hash the password using auth package
		return auth.HashPassword(password)
	})
}

// ClearCache clears the formula cache
func (e *Engine) ClearCache() {

}
