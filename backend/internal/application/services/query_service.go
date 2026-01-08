package services

import (
	"context"
	"fmt"

	"strings"

	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	pkgErrors "github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// QueryService handles all query operations with formula hydration
type QueryService struct {
	repo        *persistence.QueryRepository
	metadata    *MetadataService
	permissions *PermissionService
	validator   *SecurityValidator
	formula     *formula.Engine
}

// NewQueryService creates a new QueryService
func NewQueryService(
	repo *persistence.QueryRepository,
	metadata *MetadataService,
	permissions *PermissionService,
) *QueryService {
	return &QueryService{
		repo:        repo,
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
	if !qs.permissions.CheckObjectPermissionWithUser(ctx, req.ObjectAPIName, constants.PermRead, currentUser) {
		return nil, fmt.Errorf("insufficient permissions to read %s", req.ObjectAPIName)
	}

	schema := qs.metadata.GetSchema(ctx, req.ObjectAPIName)
	if schema == nil {
		return nil, pkgErrors.NewNotFoundError("Object", req.ObjectAPIName)
	}

	// Build visible field list
	visibleFields := qs.metadata.GetSystemFields(ctx, req.ObjectAPIName)

	// Add custom fields that are visible
	for _, field := range schema.Fields {
		isSystem := field.IsSystem || field.IsNameField
		if !isSystem && qs.permissions.CheckFieldVisibilityWithUser(ctx, req.ObjectAPIName, field.APIName, currentUser) {
			visibleFields = append(visibleFields, field.APIName)
			if field.IsPolymorphic {
				visibleFields = append(visibleFields, GetPolymorphicTypeColumnName(field.APIName))
			}
		}
	}

	// Delegate to Repository
	results, err := qs.repo.Find(ctx, schema, req, visibleFields)
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
	if !qs.permissions.CheckObjectPermissionWithUser(ctx, objectName, constants.PermRead, currentUser) {
		return []models.SObject{}, nil
	}

	schema := qs.metadata.GetSchema(ctx, objectName)
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
				qs.permissions.CheckFieldVisibilityWithUser(ctx, objectName, field.APIName, currentUser) {
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

	// NOTE: Row-level security deferred (see line 71 for details)

	// Delegate to Repository
	results, err := qs.repo.Search(ctx, objectName, term, searchFields, fieldsToSelect, 20)
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
	schemas := qs.metadata.GetSchemas(ctx)
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

	if !qs.permissions.CheckObjectPermissionWithUser(ctx, objectName, constants.PermRead, currentUser) {
		return nil, fmt.Errorf("access denied: cannot read %s", objectName)
	}

	schema := qs.metadata.GetSchema(ctx, objectName)
	if schema == nil {
		return nil, pkgErrors.NewNotFoundError("Object", objectName)
	}

	// Delegate to Repository
	val, err := qs.repo.RunAnalytics(ctx, objectName, analyticsQuery)
	if err != nil {
		return nil, err
	}

	if analyticsQuery.Operation == "group_by" {
		return val, nil // val is already []SObject
	}

	// val is scalar (float64, int64, etc)
	return val, nil
}

// ExecuteRawSQL executes a raw SQL query with parameters (Validated and Secured)
func (qs *QueryService) ExecuteRawSQL(ctx context.Context, sql string, params []interface{}, user *models.UserSession) ([]models.SObject, error) {
	// Validate and Rewrite SQL (RLS/FLS)
	// This ensures users can only see what they are allowed to see.
	safeSQL, safeParams, err := qs.validator.ValidateAndRewrite(ctx, sql, params, user)
	if err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	// Delegate to Repository
	return qs.repo.ExecuteRawSQL(ctx, safeSQL, safeParams)
}
