package services

import (
	"context"
	"fmt"

	"strings"

	"github.com/nexuscrm/backend/internal/infrastructure/database"
	pkgErrors "github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/backend/pkg/query"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// QueryService handles all query operations with formula hydration
type QueryService struct {
	db          *database.TiDBConnection
	metadata    *MetadataService
	permissions *PermissionService
	validator   *SecurityValidator
	formula     *formula.Engine
}

// NewQueryService creates a new QueryService
func NewQueryService(
	db *database.TiDBConnection,
	metadata *MetadataService,
	permissions *PermissionService,
) *QueryService {
	return &QueryService{
		db:          db,
		metadata:    metadata,
		permissions: permissions,
		validator:   NewSecurityValidator(permissions, metadata),
		formula:     formula.NewEngine(),
	}
}

// Query executes a query based on a QueryRequest
func (qs *QueryService) Query(
	ctx context.Context,
	req models.QueryRequest,
	currentUser *models.UserSession,
) ([]models.SObject, error) {
	// Check permissions
	if !qs.permissions.CheckObjectPermissionWithUser(req.ObjectAPIName, constants.PermRead, currentUser) {
		return nil, fmt.Errorf("insufficient permissions to read %s", req.ObjectAPIName)
	}

	schema := qs.metadata.GetSchema(req.ObjectAPIName)
	if schema == nil {
		return nil, pkgErrors.NewNotFoundError("Object", req.ObjectAPIName)
	}

	// Build query
	builder := query.From(req.ObjectAPIName).WithMetadata(schema)

	// Start with system fields from metadata
	visibleFields := qs.metadata.GetSystemFields(req.ObjectAPIName)

	// Add custom fields that are visible
	for _, field := range schema.Fields {
		isSystem := field.IsSystem || field.IsNameField
		if !isSystem && qs.permissions.CheckFieldVisibilityWithUser(req.ObjectAPIName, field.APIName, currentUser) {
			visibleFields = append(visibleFields, field.APIName)
			if field.IsPolymorphic {
				visibleFields = append(visibleFields, GetPolymorphicTypeColumnName(field.APIName))
			}
		}
	}

	builder.Select(visibleFields)

	// Exclude deleted (only if field exists)
	hasIsDeleted := false
	for _, f := range schema.Fields {
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
			// Quote field name and use provided operator
			condition := fmt.Sprintf("`%s`.`%s` %s ?", req.ObjectAPIName, c.Field, c.Op)
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

	results, err := ExecuteQuery(ctx, qs.db, q)
	if err != nil {
		return nil, err
	}

	// Hydrate virtual fields (formulas, booleans)
	results = qs.hydrateVirtualFields(ctx, results, schema, visibleFields, currentUser)

	return results, nil
}

// QueryWithFilter executes a query with a formula expression filter
func (qs *QueryService) QueryWithFilter(
	ctx context.Context,
	objectName string,
	filterExpr string,
	currentUser *models.UserSession,
	orderBy string,
	orderDirection string,
	limit int,
) ([]models.SObject, error) {
	req := models.QueryRequest{
		ObjectAPIName: objectName,
		FilterExpr:    filterExpr,
		SortField:     orderBy,
		SortDirection: orderDirection,
		Limit:         limit,
	}
	return qs.Query(ctx, req, currentUser)
}

// SearchSingleObject searches within a single object
func (qs *QueryService) SearchSingleObject(ctx context.Context, objectName string, term string, currentUser *models.UserSession) ([]models.SObject, error) {
	if !qs.permissions.CheckObjectPermissionWithUser(objectName, constants.PermRead, currentUser) {
		return []models.SObject{}, nil
	}

	schema := qs.metadata.GetSchema(objectName)
	if schema == nil {
		return nil, pkgErrors.NewNotFoundError("Object", objectName)
	}

	// Find searchable fields
	searchFields := make([]string, 0)
	for _, field := range schema.Fields {
		isText := strings.EqualFold(string(field.Type), string(constants.FieldTypeText))
		isTextArea := strings.EqualFold(string(field.Type), string(constants.FieldTypeTextArea))
		isEmail := strings.EqualFold(string(field.Type), string(constants.FieldTypeEmail))
		isFormula := strings.EqualFold(string(field.Type), string(constants.FieldTypeFormula))
		isRollup := strings.EqualFold(string(field.Type), string(constants.FieldTypeRollupSummary))

		if isText || isTextArea || isEmail || field.APIName == constants.FieldName {

			if !isFormula && !isRollup &&
				qs.permissions.CheckFieldVisibilityWithUser(objectName, field.APIName, currentUser) {
				searchFields = append(searchFields, field.APIName)
			}
		}
	}

	if len(searchFields) == 0 {
		return []models.SObject{}, nil
	}

	// Find the display name field (Name Field or fallback)
	nameField := ""
	for _, f := range schema.Fields {
		if f.IsNameField {
			nameField = f.APIName
			break
		}
	}
	if nameField == "" {
		// Fallback: look for 'name' column
		for _, f := range schema.Fields {
			if strings.EqualFold(f.APIName, constants.FieldName) {
				nameField = f.APIName
				break
			}
		}
	}
	if nameField == "" {
		// Last fallback: first Text field
		for _, f := range schema.Fields {
			if strings.EqualFold(string(f.Type), string(constants.FieldTypeText)) {
				nameField = f.APIName
				break
			}
		}
	}

	fieldsToSelect := []string{constants.FieldID}
	if nameField != "" {
		fieldsToSelect = append(fieldsToSelect, nameField)
	}

	if schema.ListFields != nil {
		fieldsToSelect = append(fieldsToSelect, schema.ListFields...)
	}

	builder := query.From(objectName).Select(fieldsToSelect).ExcludeDeleted()

	// Support for listing all records with "*" wildcard
	if term != "*" && term != "" {
		// Try to parse as formula expression (field op value)
		if formulaCondition, formulaParams, ok := parseFormulaQuery(term, objectName); ok {
			builder.WhereRaw(formulaCondition, formulaParams)
		} else {
			// Fallback to text search
			searchConditions := make([]string, 0)
			searchParams := make([]interface{}, 0)
			for _, field := range searchFields {
				searchConditions = append(searchConditions, fmt.Sprintf("`%s`.`%s` LIKE ?", objectName, field))
				searchParams = append(searchParams, fmt.Sprintf("%%%s%%", term))
			}
			builder.WhereRaw(fmt.Sprintf("(%s)", strings.Join(searchConditions, " OR ")), searchParams)
		}
	}
	// If term is "*", no WHERE clause is added - returns all records

	// NOTE: Row-level security deferred (see line 71 for details)

	builder.Limit(20)

	// Execute
	q := builder.Build()
	results, err := ExecuteQuery(ctx, qs.db, q) // Passed ctx
	if err != nil {
		return []models.SObject{}, err
	}

	// Post-process to ensure 'name' property exists for frontend compatibility
	if nameField != "" && nameField != constants.FieldName {
		for _, row := range results {
			if val, ok := row[nameField]; ok {
				row[constants.FieldName] = val
			}
		}
	}

	return results, nil
}

// GlobalSearch searches across all objects
func (qs *QueryService) GlobalSearch(ctx context.Context, term string, currentUser *models.UserSession) ([]models.SearchResult, error) { // Added ctx
	schemas := qs.metadata.GetSchemas()
	results := make([]models.SearchResult, 0)

	for _, schema := range schemas {
		if !schema.Searchable {
			continue
		}

		matches, err := qs.SearchSingleObject(ctx, schema.APIName, term, currentUser) // Passed ctx
		if err != nil || len(matches) == 0 {
			continue
		}

		if len(matches) > 5 {
			matches = matches[:5]
		}

		results = append(results, models.SearchResult{
			ObjectAPIName: schema.APIName,
			ObjectLabel:   schema.PluralLabel,
			Icon:          schema.Icon,
			Matches:       matches,
		})
	}

	return results, nil
}

// RunAnalytics executes an analytics query
func (qs *QueryService) RunAnalytics(ctx context.Context, analyticsQuery models.AnalyticsQuery, currentUser *models.UserSession) (interface{}, error) {
	objectName := analyticsQuery.ObjectAPIName

	if !qs.permissions.CheckObjectPermissionWithUser(objectName, constants.PermRead, currentUser) {
		return nil, fmt.Errorf("access denied: cannot read %s", objectName)
	}

	schema := qs.metadata.GetSchema(objectName)
	if schema == nil {
		return nil, pkgErrors.NewNotFoundError("Object", objectName)
	}

	builder := query.From(objectName)

	// Check for is_deleted field
	hasIsDeleted := false
	for _, f := range schema.Fields {
		if strings.EqualFold(f.APIName, constants.FieldIsDeleted) {
			hasIsDeleted = true
			break
		}
	}
	if hasIsDeleted {
		builder.ExcludeDeleted()
	}

	// Apply filter expression using formula engine
	if analyticsQuery.FilterExpr != "" {
		sqlWhere, args, err := formula.ToSQL(analyticsQuery.FilterExpr)
		if err != nil {
			return nil, fmt.Errorf("invalid filter expression: %w", err)
		}
		builder.WhereRaw(sqlWhere, args)
	}

	// NOTE: Row-level security deferred (see line 71 for details)

	// Build aggregation
	switch analyticsQuery.Operation {
	case "count":
		builder.AddSelectRaw("COUNT(*) as val")

	case "group_by":
		if analyticsQuery.GroupBy == nil {
			return nil, fmt.Errorf("group_by field missing")
		}

		agg := "COUNT(*)"
		if analyticsQuery.Field != nil {
			agg = fmt.Sprintf("SUM(`%s`)", *analyticsQuery.Field)
		}

		builder.AddSelectRaw(fmt.Sprintf("`%s` as name", *analyticsQuery.GroupBy))
		builder.AddSelectRaw(fmt.Sprintf("%s as value", agg))
		builder.GroupByRaw(fmt.Sprintf("`%s`", *analyticsQuery.GroupBy))
		builder.Limit(20)

	default: // sum, avg
		if analyticsQuery.Field == nil {
			return nil, fmt.Errorf("field missing for aggregation")
		}

		builder.AddSelectRaw(fmt.Sprintf("%s(`%s`) as val", strings.ToUpper(analyticsQuery.Operation), *analyticsQuery.Field))
	}

	// Execute
	q := builder.Build()
	results, err := ExecuteQuery(ctx, qs.db, q)
	if err != nil {
		return nil, err
	}

	if analyticsQuery.Operation == "group_by" {
		return results, nil
	}

	if len(results) > 0 {
		return results[0]["val"], nil
	}

	return 0, nil
}

// ExecuteRawSQL executes a raw SQL query with parameters (Validated and Secured)
func (qs *QueryService) ExecuteRawSQL(ctx context.Context, sql string, params []interface{}, user *models.UserSession) ([]models.SObject, error) {
	// Validate and Rewrite SQL (RLS/FLS)
	// This ensures users can only see what they are allowed to see.
	safeSQL, safeParams, err := qs.validator.ValidateAndRewrite(sql, params, user)
	if err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	q := query.QueryResult{
		SQL:    safeSQL,
		Params: safeParams,
	}

	return ExecuteQuery(ctx, qs.db, q)
}

// Helper methods moved to query_helpers.go (findField, applyCriteriaToBuilder, hydrateVirtualFields, filterMemoryCriteria)

// parseFormulaQuery parses simple formula expressions like "field = value" or "field > 100"
// Returns SQL condition, params, and true if successfully parsed as formula
func parseFormulaQuery(term, objectName string) (string, []interface{}, bool) {
	// Supported operators in order of specificity (check multi-char first)
	operators := []struct {
		symbol string
		sqlOp  string
	}{
		{"!=", "!="}, {"<>", "!="}, {">=", ">="}, {"<=", "<="},
		{"=", "="}, {">", ">"}, {"<", "<"},
	}

	for _, op := range operators {
		if idx := strings.Index(term, op.symbol); idx > 0 {
			field := strings.TrimSpace(term[:idx])
			value := strings.TrimSpace(term[idx+len(op.symbol):])

			// Validate field name (basic sanity check - alphanumeric + underscore)
			if field == "" || value == "" {
				continue
			}
			for _, c := range field {
				if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
					return "", nil, false // Invalid field name
				}
			}

			// Build SQL condition
			condition := fmt.Sprintf("`%s`.`%s` %s ?", objectName, field, op.sqlOp)
			return condition, []interface{}{value}, true
		}
	}

	return "", nil, false
}
