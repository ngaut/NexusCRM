package expression

import (
	"fmt"
	"strings"

	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
)

// SQLWalker converts an expr AST to SQL
type SQLWalker struct {
	builder strings.Builder
	args    []interface{}
	err     error
}

// isNilNode checks if a node represents a null/nil value
// In expr-lang, null can be either a NilNode or an IdentifierNode with value "null", "nil", or "NULL"
func isNilNode(node ast.Node) bool {
	if _, ok := node.(*ast.NilNode); ok {
		return true
	}
	if id, ok := node.(*ast.IdentifierNode); ok {
		val := strings.ToLower(id.Value)
		return val == "null" || val == "nil"
	}
	return false
}

// ToSQL converts an expression string to a SQL WHERE clause and arguments
func ToSQL(expression string) (string, []interface{}, error) {
	// Parse the expression using expr parser directly to get AST
	// We use standard parser.Parse because we just need the tree structure
	tree, err := parser.Parse(expression)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse expression: %w", err)
	}

	walker := &SQLWalker{
		args: make([]interface{}, 0),
	}

	// Start manual traversal
	walker.walk(&tree.Node)

	if walker.err != nil {
		return "", nil, walker.err
	}

	return walker.builder.String(), walker.args, nil
}

// Visit implements ast.Visitor
// Note: Based on probe test, signature must be Visit(node *ast.Node)
func (w *SQLWalker) walk(node *ast.Node) {
	if w.err != nil {
		return
	}
	if node == nil || *node == nil {
		return
	}

	n := *node

	switch v := n.(type) {
	case *ast.BinaryNode:
		w.visitBinary(v)
	case *ast.IdentifierNode:
		w.builder.WriteString(v.Value)
	case *ast.IntegerNode:
		w.builder.WriteString("?")
		w.args = append(w.args, v.Value)
	case *ast.FloatNode:
		w.builder.WriteString("?")
		w.args = append(w.args, v.Value)
	case *ast.StringNode:
		w.builder.WriteString("?")
		w.args = append(w.args, v.Value)
	case *ast.BoolNode:
		w.builder.WriteString("?")
		w.args = append(w.args, v.Value)
	case *ast.NilNode:
		w.builder.WriteString("NULL")
	case *ast.CallNode:
		w.visitCall(v)
	default:
		w.err = fmt.Errorf("unsupported node type: %T", n)
	}
}

func (w *SQLWalker) visitBinary(node *ast.BinaryNode) {
	// Check for null comparisons which need special SQL syntax
	// Note: In expr-lang, null can be either NilNode or IdentifierNode with value "null"/"nil"
	rightIsNil := isNilNode(node.Right)
	leftIsNil := isNilNode(node.Left)

	if rightIsNil || leftIsNil {
		// Handle IS NULL / IS NOT NULL
		var fieldNode ast.Node
		if rightIsNil {
			fieldNode = node.Left
		} else {
			fieldNode = node.Right
		}

		w.builder.WriteString("(")
		w.walk(&fieldNode)
		switch node.Operator {
		case "==":
			w.builder.WriteString(" IS NULL")
		case "!=":
			w.builder.WriteString(" IS NOT NULL")
		default:
			w.err = fmt.Errorf("unsupported operator for null comparison: %s", node.Operator)
		}
		w.builder.WriteString(")")
		return
	}

	w.builder.WriteString("(")

	// Left
	w.walk(&node.Left)

	w.builder.WriteString(" ")

	// Operator mapping
	switch node.Operator {
	case "==":
		w.builder.WriteString("=")
	case "&&":
		w.builder.WriteString("AND")
	case "||":
		w.builder.WriteString("OR")
	default:
		w.builder.WriteString(node.Operator)
	}

	w.builder.WriteString(" ")

	// Right
	w.walk(&node.Right)

	w.builder.WriteString(")")
}

func (w *SQLWalker) visitCall(node *ast.CallNode) {
	// Extract function name from Callee
	callee, ok := node.Callee.(*ast.IdentifierNode)
	if !ok {
		w.err = fmt.Errorf("unsupported callee type: %T", node.Callee)
		return
	}

	fnName := strings.ToUpper(callee.Value)

	switch fnName {
	case "UPPER":
		w.builder.WriteString("UPPER(")
		w.walkArgs(node.Arguments)
		w.builder.WriteString(")")

	case "LOWER":
		w.builder.WriteString("LOWER(")
		w.walkArgs(node.Arguments)
		w.builder.WriteString(")")

	case "LEN":
		w.builder.WriteString("CHAR_LENGTH(")
		w.walkArgs(node.Arguments)
		w.builder.WriteString(")")

	case "IF":
		// IF(cond, true, false) -> IF(cond_sql, true_sql, false_sql)
		if len(node.Arguments) != 3 {
			w.err = fmt.Errorf("IF requires 3 arguments")
			return
		}
		w.builder.WriteString("IF(")
		arg0 := node.Arguments[0]
		w.walk(&arg0)
		w.builder.WriteString(", ")
		arg1 := node.Arguments[1]
		w.walk(&arg1)
		w.builder.WriteString(", ")
		arg2 := node.Arguments[2]
		w.walk(&arg2)
		w.builder.WriteString(")")

	case "TODAY":
		// TODAY() -> CURDATE()
		w.builder.WriteString("CURDATE()")

	case "NOW":
		// NOW() -> NOW()
		w.builder.WriteString("NOW()")

	case "DATE_ADD":
		// DATE_ADD(date, days) -> DATE_ADD(date_sql, INTERVAL ? DAY)
		if len(node.Arguments) != 2 {
			w.err = fmt.Errorf("DATE_ADD requires 2 arguments")
			return
		}
		w.builder.WriteString("DATE_ADD(")
		arg0 := node.Arguments[0]
		w.walk(&arg0)
		w.builder.WriteString(", INTERVAL ")
		arg1 := node.Arguments[1]
		w.walk(&arg1)
		w.builder.WriteString(" DAY)")

	case "CONTAINS":
		// CONTAINS(field, 'text') -> field LIKE '%text%'
		if len(node.Arguments) != 2 {
			w.err = fmt.Errorf("CONTAINS requires 2 arguments")
			return
		}
		arg0 := node.Arguments[0]
		w.walk(&arg0)
		w.builder.WriteString(" LIKE ")
		// Get the string value and wrap with %
		strArg, ok := node.Arguments[1].(*ast.StringNode)
		if !ok {
			w.err = fmt.Errorf("CONTAINS second argument must be a string")
			return
		}
		w.builder.WriteString("?")
		w.args = append(w.args, "%"+strArg.Value+"%")

	case "STARTS_WITH":
		// STARTS_WITH(field, 'text') -> field LIKE 'text%'
		if len(node.Arguments) != 2 {
			w.err = fmt.Errorf("STARTS_WITH requires 2 arguments")
			return
		}
		arg0 := node.Arguments[0]
		w.walk(&arg0)
		w.builder.WriteString(" LIKE ")
		strArg, ok := node.Arguments[1].(*ast.StringNode)
		if !ok {
			w.err = fmt.Errorf("STARTS_WITH second argument must be a string")
			return
		}
		w.builder.WriteString("?")
		w.args = append(w.args, strArg.Value+"%")

	case "ENDS_WITH":
		// ENDS_WITH(field, 'text') -> field LIKE '%text'
		if len(node.Arguments) != 2 {
			w.err = fmt.Errorf("ENDS_WITH requires 2 arguments")
			return
		}
		arg0 := node.Arguments[0]
		w.walk(&arg0)
		w.builder.WriteString(" LIKE ")
		strArg, ok := node.Arguments[1].(*ast.StringNode)
		if !ok {
			w.err = fmt.Errorf("ENDS_WITH second argument must be a string")
			return
		}
		w.builder.WriteString("?")
		w.args = append(w.args, "%"+strArg.Value)

	default:
		w.err = fmt.Errorf("unsupported function: %s", callee.Value)
	}
}

// Helper to walk multiple args with comma separation
func (w *SQLWalker) walkArgs(args []ast.Node) {
	for i, arg := range args {
		if i > 0 {
			w.builder.WriteString(", ")
		}
		argNode := arg
		w.walk(&argNode)
	}
}
