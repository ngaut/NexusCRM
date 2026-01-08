package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/backend/pkg/query"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// QueryRepository handles complex read operations (filtering, searching, analytics)
type QueryRepository struct {
	db *sql.DB
}

// NewQueryRepository creates a new QueryRepository
func NewQueryRepository(db *sql.DB) *QueryRepository {
	return &QueryRepository{db: db}
}

// GetExecutor returns the DB connection (Queries are usually not transactional, but could be)
func (r *QueryRepository) GetExecutor() Executor {
	return r.db
}

// Find executes a structured query request
func (r *QueryRepository) Find(ctx context.Context, tableSchema *models.ObjectMetadata, req models.QueryRequest, visibleFields []string) ([]models.SObject, error) {
	// Build query
	builder := query.From(tableSchema.APIName).WithMetadata(tableSchema)
	builder.Select(visibleFields)

	// Exclude deleted (only if field exists)
	hasIsDeleted := false
	for _, f := range tableSchema.Fields {
		if strings.EqualFold(f.APIName, constants.FieldIsDeleted) {
			hasIsDeleted = true
			break
		}
	}
	if hasIsDeleted {
		builder.ExcludeDeleted()
	}

	// Apply criteria
	if len(req.Criteria) > 0 {
		for _, c := range req.Criteria {
			// Validate Field Name (simple alphanumeric check)
			if !isValidFieldName(c.Field) {
				return nil, fmt.Errorf("invalid field name in criteria: %s", c.Field)
			}
			// Validate Operator
			validOps := map[string]bool{
				"=": true, "!=": true, "<": true, ">": true, "<=": true, ">=": true, "LIKE": true, "IN": true,
			}
			if !validOps[strings.ToUpper(c.Op)] {
				return nil, fmt.Errorf("invalid operator in criteria: %s", c.Op)
			}

			condition := fmt.Sprintf("`%s`.`%s` %s ?", tableSchema.APIName, c.Field, c.Op)
			builder.Where(condition, c.Val)
		}
	}

	// Apply formula expression filter
	if req.FilterExpr != "" {
		sqlWhere, args, err := formula.ToSQL(req.FilterExpr)
		if err != nil {
			return nil, fmt.Errorf("invalid filter expression: %w", err)
		}
		builder.WhereRaw(sqlWhere, args)
	}

	// Apply sorting
	if req.SortField != "" {
		builder.OrderBy(req.SortField, req.SortDirection)
	}

	// Apply limit
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	builder.Limit(limit)

	// Build and execute
	q := builder.Build()

	exec := r.GetExecutor()
	rows, err := exec.QueryContext(ctx, q.SQL, q.Params...)
	if err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	return query.ScanRowsToSObjects(rows)
}

// Search performs a text search on specific fields
func (r *QueryRepository) Search(ctx context.Context, tableName string, term string, searchFields []string, selectFields []string, limit int) ([]models.SObject, error) {
	builder := query.From(tableName).Select(selectFields).ExcludeDeleted()

	// Support for listing all records with "*" wildcard
	if term != "*" && term != "" {
		// Try to parse as formula expression (field op value)
		if formulaCondition, formulaParams, ok := query.ParseFormulaQuery(term, tableName); ok {
			builder.WhereRaw(formulaCondition, formulaParams)
		} else {
			// Fallback to text search
			searchConditions := make([]string, 0)
			searchParams := make([]interface{}, 0)
			for _, field := range searchFields {
				searchConditions = append(searchConditions, fmt.Sprintf("`%s`.`%s` LIKE ?", tableName, field))
				searchParams = append(searchParams, fmt.Sprintf("%%%s%%", term))
			}
			builder.WhereRaw(fmt.Sprintf("(%s)", strings.Join(searchConditions, " OR ")), searchParams)
		}
	}

	if limit <= 0 {
		limit = 20
	}
	builder.Limit(limit)

	q := builder.Build()

	exec := r.GetExecutor()
	rows, err := exec.QueryContext(ctx, q.SQL, q.Params...)
	if err != nil {
		return nil, fmt.Errorf("search execution error: %w", err)
	}
	defer rows.Close()

	return query.ScanRowsToSObjects(rows)
}

// RunAnalytics executes an aggregation query
func (r *QueryRepository) RunAnalytics(ctx context.Context, tableName string, q models.AnalyticsQuery) (interface{}, error) {
	builder := query.From(tableName)

	// Assuming is_deleted check is handled by caller passing correct schema or we check it here?
	// Consistent with Find, let's assume we want to exclude deleted if possible.
	// Simpler: Just build what's asked.
	// NOTE: In strict checking, we should verify if IsDeleted exists.
	// For now, sticking to the core logic extraction.
	builder.ExcludeDeleted()

	if q.FilterExpr != "" {
		sqlWhere, args, err := formula.ToSQL(q.FilterExpr)
		if err != nil {
			return nil, fmt.Errorf("invalid filter expression: %w", err)
		}
		builder.WhereRaw(sqlWhere, args)
	}

	switch q.Operation {
	case "count":
		builder.AddSelectRaw("COUNT(*) as val")

	case "group_by":
		agg := "COUNT(*)"
		if q.Field != nil {
			agg = fmt.Sprintf("SUM(`%s`)", *q.Field)
		}

		builder.AddSelectRaw(fmt.Sprintf("`%s` as name", *q.GroupBy))
		builder.AddSelectRaw(fmt.Sprintf("%s as value", agg))
		builder.GroupByRaw(fmt.Sprintf("`%s`", *q.GroupBy))
		builder.Limit(20)

	default: // sum, avg
		builder.AddSelectRaw(fmt.Sprintf("%s(`%s`) as val", strings.ToUpper(q.Operation), *q.Field))
	}

	queryP := builder.Build()

	exec := r.GetExecutor()
	rows, err := exec.QueryContext(ctx, queryP.SQL, queryP.Params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := query.ScanRowsToSObjects(rows)
	if err != nil {
		return nil, err
	}

	if q.Operation == "group_by" {
		return results, nil
	}

	if len(results) > 0 {
		return results[0]["val"], nil
	}

	return 0, nil
}

// ExecuteRawSQL executes a raw SQL string (Validated by Service Layer)
func (r *QueryRepository) ExecuteRawSQL(ctx context.Context, sql string, params []interface{}) ([]models.SObject, error) {
	exec := r.GetExecutor()
	rows, err := exec.QueryContext(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("raw query error: %w", err)
	}
	defer rows.Close()

	return query.ScanRowsToSObjects(rows)
}

// GetLookupNames retrieves specific ID and Name fields for a list of IDs
func (r *QueryRepository) GetLookupNames(ctx context.Context, tableName string, ids []string, nameField string) ([]models.SObject, error) {
	if len(ids) == 0 {
		return []models.SObject{}, nil
	}

	placeholders := make([]string, len(ids))
	params := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		params[i] = id
	}

	sql := fmt.Sprintf("SELECT `id`, `%s` FROM `%s` WHERE `id` IN (%s)",
		nameField, tableName, strings.Join(placeholders, ","))

	exec := r.GetExecutor()
	rows, err := exec.QueryContext(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("lookup hydration error: %w", err)
	}
	defer rows.Close()

	return query.ScanRowsToSObjects(rows)
}

// isValidFieldName checks if a field name is safe (alphanumeric + underscore)
func isValidFieldName(name string) bool {
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}
