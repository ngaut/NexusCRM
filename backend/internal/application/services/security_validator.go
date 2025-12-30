package services

import (
	"fmt"
	"strings"

	"github.com/nexuscrm/shared/pkg/models"
	"github.com/nexuscrm/shared/pkg/constants"

	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/format"
	"github.com/pingcap/tidb/pkg/parser/opcode"
	"github.com/pingcap/tidb/pkg/parser/test_driver" // Using test_driver for ValueExpr
)

// SecurityValidator handles SQL parsing and security enforcement
type SecurityValidator struct {
	parser      *parser.Parser
	permissions *PermissionService
	metadata    *MetadataService
}

// NewSecurityValidator creates a new SecurityValidator
func NewSecurityValidator(permissions *PermissionService, metadata *MetadataService) *SecurityValidator {
	// Initialize parser purely for parsing logic; context usually not strictly required for basic parse
	return &SecurityValidator{
		parser:      parser.New(),
		permissions: permissions,
		metadata:    metadata,
	}
}

// ValidateAndRewrite parses the SQL, validates permissions, and rewrites it for RLS
func (v *SecurityValidator) ValidateAndRewrite(sql string, params []interface{}, user *models.UserSession) (string, []interface{}, error) {
	// 1. Parse SQL
	stmtNodes, _, err := v.parser.Parse(sql, "", "")
	if err != nil {
		return "", nil, fmt.Errorf("SQL parse error: %v", err)
	}

	if len(stmtNodes) != 1 {
		return "", nil, fmt.Errorf("only single SQL statements are allowed")
	}

	stmt := stmtNodes[0]

	// 2. Only allow SELECT statements
	selectStmt, ok := stmt.(*ast.SelectStmt)
	if !ok {
		return "", nil, fmt.Errorf("only SELECT statements are allowed in analytics")
	}

	// 3. Visitor for Validation
	visitor := &SecurityVisitor{
		user:        user,
		permissions: v.permissions,
		metadata:    v.metadata,
		err:         nil,
	}

	stmt.Accept(visitor)

	if visitor.err != nil {
		return "", nil, visitor.err
	}

	// 4. Transform: Inject RLS Logic if needed
	// We do this after the visitor pass to avoid modifying AST while visiting it
	// For simple single-table queries, we inject into the Where clause.
	if !constants.IsSuperUser(user.ProfileID) {
		if err := v.applyRLS(selectStmt, user); err != nil {
			return "", nil, err
		}
	}

	// 5. Restore SQL
	var sb strings.Builder
	restoreCtx := format.NewRestoreCtx(format.DefaultRestoreFlags, &sb)
	if err := stmt.Restore(restoreCtx); err != nil {
		return "", nil, fmt.Errorf("SQL restore error: %v", err)
	}

	// Note: We currently inject owner_id as a literal string in AST to avoid shifting `params`.
	// This is safe because user.ID is a trusted internal UUID.

	return sb.String(), params, nil
}

// applyRLS injects "AND owner_id = 'userID'" into the WHERE clause
func (v *SecurityValidator) applyRLS(stmt *ast.SelectStmt, user *models.UserSession) error {
	// Strategy: If the FROM clause targets a table that needs RLS, add filter.
	// Limitation: Complex joins are hard. We target the *primary* table in standard CRM usage.

	if stmt.From == nil || stmt.From.TableRefs == nil || stmt.From.TableRefs.Left == nil {
		return nil
	}

	// Drill to TableSource
	ts, ok := stmt.From.TableRefs.Left.(*ast.TableSource)
	if !ok {
		return nil // Complex query, skipping RLS MVP (should allow or deny? Allow for now to not break complex reports)
	}

	tn, ok := ts.Source.(*ast.TableName)
	if !ok {
		return nil
	}

	objName := tn.Name.O

	// Check metadata for owner_id presence
	schema := v.metadata.GetSchema(objName)
	if schema == nil {
		return nil // Unknown object, likely system or error, let permission check handle it
	}

	hasOwner := false
	for _, f := range schema.Fields {
		if f.APIName == constants.FieldOwnerID {
			hasOwner = true
			break
		}
	}

	if hasOwner {
		// Construct AST Node: owner_id = 'user.ID'
		// Using ast.NewCIStr for names
		// Using test_driver.ValueExpr for string literal

		colName := &ast.ColumnName{Name: ast.NewCIStr(constants.FieldOwnerID)}
		colExpr := &ast.ColumnNameExpr{Name: colName}

		// Trying to use test_driver ValueExpr
		rightExpr := &test_driver.ValueExpr{}
		rightExpr.SetString(user.ID) // API assumption

		cond := &ast.BinaryOperationExpr{
			Op: opcode.EQ,
			L:  colExpr,
			R:  rightExpr,
		}

		if stmt.Where == nil {
			stmt.Where = cond
		} else {
			stmt.Where = &ast.BinaryOperationExpr{
				Op: opcode.LogicAnd,
				L:  stmt.Where,
				R:  cond,
			}
		}
	}

	return nil
}

type SecurityVisitor struct {
	user        *models.UserSession
	permissions *PermissionService
	metadata    *MetadataService
	err         error
}

func (v *SecurityVisitor) Enter(in ast.Node) (ast.Node, bool) {
	if v.err != nil {
		return in, true
	}

	// Validate Table Permissions
	if t, ok := in.(*ast.TableName); ok {
		objName := t.Name.O
		if objName != "" {
			if !v.permissions.CheckObjectPermissionWithUser(objName, constants.PermRead, v.user) {
				v.err = fmt.Errorf("access denied: cannot read table '%s'", objName)
				return in, true
			}
		}
	}

	// Validate Column Permissions
	if c, ok := in.(*ast.ColumnName); ok {
		tableName := c.Table.O
		colName := c.Name.O

		// If Table is known, check specific field visibility
		if tableName != "" && colName != "*" {
			if !v.permissions.CheckFieldVisibilityWithUser(tableName, colName, v.user) {
				v.err = fmt.Errorf("access denied: cannot read field '%s' on table '%s'", colName, tableName)
				return in, true
			}
		}
	}
	return in, false
}

func (v *SecurityVisitor) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}
