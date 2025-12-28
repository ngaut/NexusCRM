package expression

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Engine is a wrapper around antonmedv/expr
type Engine struct {
	programCache map[string]*vm.Program
	functions    map[string]func(params ...interface{}) (interface{}, error)
	mu           sync.RWMutex
}

// NewEngine creates a new expression engine
func NewEngine() *Engine {
	return &Engine{
		programCache: make(map[string]*vm.Program),
		functions:    make(map[string]func(params ...interface{}) (interface{}, error)),
	}
}

// Evaluate compiles (if needed) and runs an expression against the given environment
func (e *Engine) Evaluate(expression string, env map[string]interface{}) (interface{}, error) {
	program, err := e.getProgram(expression, env)
	if err != nil {
		return nil, err
	}

	output, err := expr.Run(program, env)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// RegisterFunction registers a custom function
func (e *Engine) RegisterFunction(name string, fn func(params ...interface{}) (interface{}, error)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.functions == nil {
		e.functions = make(map[string]func(params ...interface{}) (interface{}, error))
	}
	e.functions[name] = fn
	// Clear cache as available functions changed
	e.programCache = make(map[string]*vm.Program)
}

func (e *Engine) getProgram(expression string, env map[string]interface{}) (*vm.Program, error) {
	e.mu.RLock()
	if prog, ok := e.programCache[expression]; ok {
		e.mu.RUnlock()
		return prog, nil
	}
	e.mu.RUnlock()

	e.mu.Lock()
	defer e.mu.Unlock()

	// Double check
	if prog, ok := e.programCache[expression]; ok {
		return prog, nil
	}

	// Define standard functions
	options := []expr.Option{
		expr.Env(env),
		expr.Function("TODAY", func(params ...interface{}) (interface{}, error) {
			return time.Now().Format("2006-01-02"), nil
		}),
		expr.Function("NOW", func(params ...interface{}) (interface{}, error) {
			return time.Now().Format("2006-01-02 15:04:05"), nil
		}),
		expr.Function("LEN", func(params ...interface{}) (interface{}, error) {
			if len(params) != 1 {
				return nil, fmt.Errorf("LEN requires 1 argument")
			}
			s, ok := params[0].(string)
			if !ok {
				return nil, fmt.Errorf("LEN argument must be string")
			}
			return len(s), nil
		}),
		expr.Function("UPPER", func(params ...interface{}) (interface{}, error) {
			if len(params) != 1 {
				return nil, fmt.Errorf("UPPER requires 1 argument")
			}
			s, ok := params[0].(string)
			if !ok {
				return nil, fmt.Errorf("UPPER argument must be string")
			}
			return strings.ToUpper(s), nil
		}),
		expr.Function("LOWER", func(params ...interface{}) (interface{}, error) {
			if len(params) != 1 {
				return nil, fmt.Errorf("LOWER requires 1 argument")
			}
			s, ok := params[0].(string)
			if !ok {
				return nil, fmt.Errorf("LOWER argument must be string")
			}
			return strings.ToLower(s), nil
		}),
		expr.Function("ROUND", func(params ...interface{}) (interface{}, error) {
			if len(params) != 2 {
				return nil, fmt.Errorf("ROUND requires 2 arguments")
			}
			// expr handles types well, but we might get float or int
			val, err := toFloat(params[0])
			if err != nil {
				return nil, fmt.Errorf("ROUND arg 1 must be number")
			}
			prec, err := toInt(params[1])
			if err != nil {
				return nil, fmt.Errorf("ROUND arg 2 must be integer")
			}

			// Simple rounding logic
			// Standard logic: Multiply by 10^prec, round, divide
			mult := 1.0
			for i := 0; i < prec; i++ {
				mult *= 10
			}
			return float64(int(val*mult+0.5)) / mult, nil
		}),
		expr.Function("IF", func(params ...interface{}) (interface{}, error) {
			if len(params) != 3 {
				return nil, fmt.Errorf("IF requires 3 arguments (condition, true_value, false_value)")
			}
			cond, ok := params[0].(bool)
			if !ok {
				return nil, fmt.Errorf("IF condition must be boolean")
			}
			if cond {
				return params[1], nil
			}
			return params[2], nil
		}),
		expr.Function("DATE_ADD", func(params ...interface{}) (interface{}, error) {
			if len(params) != 2 {
				return nil, fmt.Errorf("DATE_ADD requires 2 arguments (date, days)")
			}
			// Parse date string
			dateStr, ok := params[0].(string)
			if !ok {
				return nil, fmt.Errorf("DATE_ADD date must be string")
			}
			days, err := toInt(params[1])
			if err != nil {
				return nil, fmt.Errorf("DATE_ADD days must be integer")
			}
			t, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				// Try datetime format
				t, err = time.Parse("2006-01-02 15:04:05", dateStr)
				if err != nil {
					return nil, fmt.Errorf("DATE_ADD date format invalid")
				}
			}
			result := t.AddDate(0, 0, days)
			return result.Format("2006-01-02"), nil
		}),
	}

	// Add custom functions
	for name, fn := range e.functions {
		options = append(options, expr.Function(name, fn))
	}

	// Compile
	program, err := expr.Compile(expression, options...)
	if err != nil {
		return nil, err
	}

	e.programCache[expression] = program
	return program, nil
}

// Validation helper
func (e *Engine) Validate(expression string, env map[string]interface{}) error {
	_, err := e.getProgram(expression, env)
	return err
}

func toFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case string:
		var f float64
		_, err := fmt.Sscanf(val, "%f", &f)
		return f, err
	}
	return 0, fmt.Errorf("cannot convert %T to float", v)
}

func toInt(v interface{}) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case float64:
		return int(val), nil
	case int64:
		return int(val), nil
	case float32:
		return int(val), nil
	case string:
		var i int
		_, err := fmt.Sscanf(val, "%d", &i)
		return i, err
	}
	return 0, fmt.Errorf("cannot convert %T to int", v)
}
