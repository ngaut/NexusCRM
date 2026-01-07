package query

import (
	"fmt"
	"strings"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// QueryType represents the type of SQL query
type QueryType string

const (
	QueryTypeSelect QueryType = "SELECT"
	QueryTypeInsert QueryType = "INSERT"
	QueryTypeUpdate QueryType = "UPDATE"
	QueryTypeDelete QueryType = "DELETE"
)

// QueryResult represents the built SQL query and parameters
type QueryResult struct {
	SQL    string
	Params []interface{}
}

// Builder is a fluent SQL query builder
type Builder struct {
	queryType    QueryType
	table        string
	fields       []string
	joins        []string
	whereClauses []string
	params       []interface{}
	orderBy      string
	groupBy      string
	limit        *int
	values       map[string]interface{}

	// Metadata context for smart features
	schema *models.ObjectMetadata
}

// From creates a new SELECT query builder
func From(table string) *Builder {
	return &Builder{
		queryType:    QueryTypeSelect,
		table:        table,
		fields:       make([]string, 0),
		joins:        make([]string, 0),
		whereClauses: make([]string, 0),
		params:       make([]interface{}, 0),
	}
}

// Insert creates a new INSERT query builder
func Insert(table string, data map[string]interface{}) *Builder {
	return &Builder{
		queryType: QueryTypeInsert,
		table:     table,
		values:    data,
		params:    make([]interface{}, 0),
	}
}

// Update creates a new UPDATE query builder
func Update(table string) *Builder {
	return &Builder{
		queryType:    QueryTypeUpdate,
		table:        table,
		values:       make(map[string]interface{}),
		whereClauses: make([]string, 0),
		params:       make([]interface{}, 0),
	}
}

// Delete creates a new DELETE query builder
func Delete(table string) *Builder {
	return &Builder{
		queryType:    QueryTypeDelete,
		table:        table,
		whereClauses: make([]string, 0),
		params:       make([]interface{}, 0),
	}
}

// WithMetadata attaches object metadata for smart features
func (b *Builder) WithMetadata(schema *models.ObjectMetadata) *Builder {
	b.schema = schema
	return b
}

// Select specifies which fields to select
func (b *Builder) Select(fields []string) *Builder {
	if b.queryType != QueryTypeSelect {
		return b
	}

	if len(fields) == 1 && fields[0] == "*" {
		b.fields = append(b.fields, "*")
		return b
	}

	for _, field := range fields {
		// Add field with table prefix if not already prefixed
		if !strings.Contains(field, ".") && field != "*" {
			b.fields = append(b.fields, fmt.Sprintf("`%s`.`%s`", b.table, field))
		} else {
			b.fields = append(b.fields, field)
		}
	}

	// Always ensure ID is present
	hasID := false
	for _, f := range b.fields {
		if strings.Contains(f, ".`id`") || f == "id" || f == "*" {
			hasID = true
			break
		}
	}
	if !hasID {
		b.fields = append([]string{fmt.Sprintf("`%s`.`id`", b.table)}, b.fields...)
	}

	return b
}

// AddSelectRaw adds a raw select expression
func (b *Builder) AddSelectRaw(expression string, alias ...string) *Builder {
	if b.queryType != QueryTypeSelect {
		return b
	}

	if len(alias) > 0 && alias[0] != "" {
		b.fields = append(b.fields, fmt.Sprintf("%s as `%s`", expression, alias[0]))
	} else {
		b.fields = append(b.fields, expression)
	}
	return b
}

// Join adds a JOIN clause
func (b *Builder) Join(joinType string, table string, alias string, on string) *Builder {
	if b.queryType != QueryTypeSelect {
		return b
	}

	b.joins = append(b.joins, fmt.Sprintf("%s JOIN `%s` as `%s` ON %s", joinType, table, alias, on))
	return b
}

// Where adds a WHERE condition
func (b *Builder) Where(condition string, value ...interface{}) *Builder {
	b.whereClauses = append(b.whereClauses, condition)
	if len(value) > 0 {
		b.params = append(b.params, value...)
	}
	return b
}

// WhereRaw adds a raw WHERE condition with parameters
func (b *Builder) WhereRaw(sql string, params []interface{}) *Builder {
	if sql != "" {
		b.whereClauses = append(b.whereClauses, sql)
		b.params = append(b.params, params...)
	}
	return b
}

// ExcludeDeleted adds is_deleted = 0 condition
func (b *Builder) ExcludeDeleted() *Builder {
	return b.Where(fmt.Sprintf("`%s`.%s = %d", b.table, constants.FieldIsDeleted, constants.IsDeletedFalse))
}

// Set sets values for UPDATE query
func (b *Builder) Set(data map[string]interface{}) *Builder {
	if b.queryType != QueryTypeUpdate {
		return b
	}

	b.values = data
	return b
}

// OrderBy adds ORDER BY clause
func (b *Builder) OrderBy(field string, direction string) *Builder {
	if b.queryType != QueryTypeSelect {
		return b
	}

	// Add table prefix if not present
	col := field
	if !strings.Contains(field, ".") && !strings.Contains(field, "`") {
		col = fmt.Sprintf("`%s`.`%s`", b.table, field)
	}

	b.orderBy = fmt.Sprintf("ORDER BY %s %s", col, direction)
	return b
}

// GroupBy adds GROUP BY clause
func (b *Builder) GroupBy(field string) *Builder {
	if b.queryType != QueryTypeSelect {
		return b
	}

	col := field
	if !strings.Contains(field, ".") && !strings.Contains(field, "`") {
		col = fmt.Sprintf("`%s`.`%s`", b.table, field)
	}

	if b.groupBy == "" {
		b.groupBy = fmt.Sprintf("GROUP BY %s", col)
	} else {
		b.groupBy += fmt.Sprintf(", %s", col)
	}
	return b
}

// GroupByRaw adds a raw GROUP BY clause
func (b *Builder) GroupByRaw(sql string) *Builder {
	if b.queryType != QueryTypeSelect {
		return b
	}

	b.groupBy = fmt.Sprintf("GROUP BY %s", sql)
	return b
}

// Limit adds LIMIT clause
func (b *Builder) Limit(n int) *Builder {
	if b.queryType != QueryTypeSelect {
		return b
	}

	b.limit = &n
	return b
}

// Build constructs the final SQL query
func (b *Builder) Build() QueryResult {
	var sql string
	var params []interface{}

	switch b.queryType {
	case QueryTypeSelect:
		sql = b.buildSelect()
		params = b.params

	case QueryTypeInsert:
		sql, params = b.buildInsert()

	case QueryTypeUpdate:
		sql, params = b.buildUpdate()

	case QueryTypeDelete:
		sql = b.buildDelete()
		params = b.params
	}

	return QueryResult{
		SQL:    sql,
		Params: params,
	}
}

func (b *Builder) buildSelect() string {
	var parts []string

	// SELECT
	fields := "*"
	if len(b.fields) > 0 {
		fields = strings.Join(b.fields, ", ")
	}
	parts = append(parts, fmt.Sprintf("SELECT %s FROM `%s`", fields, b.table))

	// JOINs
	if len(b.joins) > 0 {
		parts = append(parts, strings.Join(b.joins, " "))
	}

	// WHERE
	if len(b.whereClauses) > 0 {
		parts = append(parts, fmt.Sprintf("WHERE %s", strings.Join(b.whereClauses, " AND ")))
	}

	// GROUP BY
	if b.groupBy != "" {
		parts = append(parts, b.groupBy)
	}

	// ORDER BY
	if b.orderBy != "" {
		parts = append(parts, b.orderBy)
	}

	// LIMIT
	if b.limit != nil {
		parts = append(parts, fmt.Sprintf("LIMIT %d", *b.limit))
	}

	return strings.Join(parts, " ")
}

func (b *Builder) buildInsert() (string, []interface{}) {
	var cols []string
	var placeholders []string
	var params []interface{}

	for key, val := range b.values {
		cols = append(cols, fmt.Sprintf("`%s`", key))
		placeholders = append(placeholders, "?")
		params = append(params, val)
	}

	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
		b.table,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "))

	return sql, params
}

func (b *Builder) buildUpdate() (string, []interface{}) {
	var setClauses []string
	var params []interface{}

	for key, val := range b.values {
		setClauses = append(setClauses, fmt.Sprintf("`%s` = ?", key))
		params = append(params, val)
	}

	sql := fmt.Sprintf("UPDATE `%s` SET %s", b.table, strings.Join(setClauses, ", "))

	if len(b.whereClauses) > 0 {
		sql += fmt.Sprintf(" WHERE %s", strings.Join(b.whereClauses, " AND "))
		params = append(params, b.params...)
	}

	return sql, params
}

func (b *Builder) buildDelete() string {
	sql := fmt.Sprintf("DELETE FROM `%s`", b.table)

	if len(b.whereClauses) > 0 {
		sql += fmt.Sprintf(" WHERE %s", strings.Join(b.whereClauses, " AND "))
	}

	return sql
}

// ApplySecurity applies row-level security based on sharing model
// This will be implemented by the SharingService
func (b *Builder) ApplySecurity(securitySQL string, securityParams []interface{}) *Builder {
	if securitySQL != "" {
		b.whereClauses = append(b.whereClauses, securitySQL)
		b.params = append(b.params, securityParams...)
	}
	return b
}
