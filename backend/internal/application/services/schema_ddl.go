package services

import (
	"fmt"
	"strings"

	"github.com/nexuscrm/backend/internal/domain/schema"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/shared/pkg/constants"
)

// buildColumnDDL generates DDL for a single column
func (sm *SchemaManager) buildColumnDDL(col schema.ColumnDefinition) string {
	var sb strings.Builder

	// Check if this is a Formula field with an expression
	if col.Formula != "" && strings.EqualFold(col.LogicalType, string(constants.FieldTypeFormula)) {
		// Use generated column syntax for Formula fields
		// Determine SQL type from return_type
		sqlType := sm.MapFieldTypeToSQL(col.ReturnType)
		if sqlType == "" || sqlType == "VARCHAR(255)" {
			// Default to DECIMAL for numeric formulas, or VARCHAR for text
			if col.ReturnType == "" || col.ReturnType == string(constants.FieldTypeNumber) || col.ReturnType == string(constants.FieldTypeCurrency) || col.ReturnType == string(constants.FieldTypePercent) {
				sqlType = "DECIMAL(18,6)"
			}
		}

		// Convert formula expression to SQL (inline, no placeholders)
		formulaSQL := sm.convertFormulaToSQL(col.Formula)

		sb.WriteString(fmt.Sprintf("`%s` %s GENERATED ALWAYS AS (%s) VIRTUAL",
			col.Name, sqlType, formulaSQL))
		return sb.String()
	}

	// Standard column DDL
	sqlType := sm.MapFieldTypeToSQL(col.Type)
	sb.WriteString(fmt.Sprintf("`%s` %s", col.Name, sqlType))

	if !col.Nullable {
		sb.WriteString(" NOT NULL")
	}
	if col.Default != "" {
		sb.WriteString(fmt.Sprintf(" DEFAULT %s", col.Default))
	}
	if col.AutoIncrement {
		sb.WriteString(" AUTO_INCREMENT")
	}
	if col.PrimaryKey {
		sb.WriteString(" PRIMARY KEY")
	}
	if col.Unique {
		sb.WriteString(" UNIQUE")
	}

	return sb.String()
}

// convertFormulaToSQL converts a formula expression to inline SQL for generated columns
// This is different from ToSQL which uses placeholders - here we need inline values
func (sm *SchemaManager) convertFormulaToSQL(formula string) string {
	// For simple arithmetic expressions, the formula can often be used directly
	// as column references match SQL column names
	// Examples: "price * quantity" -> "price * quantity"
	//           "amount * 0.1" -> "amount * 0.1"

	// Replace common functions with SQL equivalents
	result := formula

	// TODAY() -> CURDATE()
	result = strings.ReplaceAll(result, "TODAY()", "CURDATE()")
	result = strings.ReplaceAll(result, "today()", "CURDATE()")

	// NOW() stays as NOW()

	// LEN() -> CHAR_LENGTH()
	result = strings.ReplaceAll(result, "LEN(", "CHAR_LENGTH(")
	result = strings.ReplaceAll(result, "len(", "CHAR_LENGTH(")

	// UPPER/LOWER are the same in SQL

	// For IF(), the syntax is the same: IF(cond, true_val, false_val)

	return result
}

// buildForeignKeyDDL generates DDL for a foreign key constraint
func (sm *SchemaManager) buildForeignKeyDDL(fk schema.ForeignKeyDefinition) string {
	ddl := fmt.Sprintf("FOREIGN KEY (`%s`) REFERENCES %s", fk.Column, fk.References)
	if fk.OnDelete != "" {
		ddl += fmt.Sprintf(" ON DELETE %s", fk.OnDelete)
	}
	if fk.OnUpdate != "" {
		ddl += fmt.Sprintf(" ON UPDATE %s", fk.OnUpdate)
	}
	return ddl
}

// buildIndexDDL generates inline index DDL for CREATE TABLE statement
func (sm *SchemaManager) buildIndexDDL(tableName string, idx schema.IndexDefinition) string {
	// Generate index name if not provided
	indexName := idx.Name
	if indexName == "" {
		indexName = fmt.Sprintf("idx_%s_%s", tableName, strings.Join(idx.Columns, "_"))
	}

	columnList := strings.Join(idx.Columns, "`, `")

	if idx.Unique {
		return fmt.Sprintf("UNIQUE KEY `%s` (`%s`)", indexName, columnList)
	}
	return fmt.Sprintf("KEY `%s` (`%s`)", indexName, columnList)
}

// ValidateFormula validates a formula expression syntax by attempting to compile it
func (sm *SchemaManager) ValidateFormula(formulaStr string, env map[string]interface{}) error {
	// Use the higher-level formula engine to validate (includes built-ins like BCRYPT)
	engine := formula.NewEngine()
	return engine.Validate(formulaStr, env)
}
