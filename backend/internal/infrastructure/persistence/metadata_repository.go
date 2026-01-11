package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nexuscrm/backend/pkg/utils"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

const (
	// Local constants for fields not yet in shared constants
	fieldRecordTypeIsDefault         = "is_default"
	fieldRecordTypeBusinessProcessID = "business_process_id"
)

type MetadataRepository struct {
	db *sql.DB
}

func NewMetadataRepository(db *sql.DB) *MetadataRepository {
	return &MetadataRepository{db: db}
}

// =================================================================================
// SQL Columns
// =================================================================================

var objectColumns = []string{
	constants.FieldID,
	constants.FieldAPIName,
	constants.FieldLabel,
	constants.FieldPluralLabel,
	constants.FieldSysObject_Icon,
	constants.FieldDescription,
	constants.FieldIsCustom,
	constants.FieldSysObject_PathField,
	constants.FieldSysObject_ListFields,
	constants.FieldSysObject_AppID,
	constants.FieldSysObject_ThemeColor,
}

var fieldColumns = []string{
	constants.FieldID,
	constants.FieldObjectID,
	constants.FieldAPIName,
	constants.FieldLabel,
	"`" + constants.FieldType + "`",
	constants.FieldRequired,
	constants.FieldSysField_IsUnique,
	constants.FieldIsSystem,
	constants.FieldSysField_IsNameField,
	"`" + constants.FieldSysField_Options + "`",
	constants.FieldReferenceTo,
	constants.FieldSysField_DeleteRule,
	constants.FieldSysField_IsMasterDetail,
	constants.FieldSysField_RelationshipName,
	constants.FieldSysField_Formula,
	constants.FieldSysField_ReturnType,
	constants.FieldSysField_DefaultValue,
	constants.FieldSysField_IsPolymorphic,
	constants.FieldSysField_HelpText,
	constants.FieldDescription,
	constants.FieldSysField_TrackHistory,
	constants.FieldSysField_MinValue,
	constants.FieldSysField_MaxValue,
	constants.FieldSysField_MinLength,
	constants.FieldSysField_MaxLength,
	constants.FieldSysField_Regex,
	constants.FieldSysField_RegexMessage,
	constants.FieldSysField_Validator,
	constants.FieldSysField_ControllingField,
	constants.FieldSysField_PicklistDependency,
	constants.FieldSysField_RollupConfig,
}

var actionColumns = []string{
	constants.FieldSysAction_ID,
	constants.FieldSysAction_ObjectAPIName,
	constants.FieldSysAction_Name,
	constants.FieldSysAction_Label,
	constants.FieldSysAction_Type,
	constants.FieldSysAction_Icon,
	constants.FieldSysAction_TargetObject,
	constants.FieldSysAction_Config,
}

var validationRuleColumns = []string{
	constants.FieldSysValidation_ID,
	constants.FieldSysValidation_ObjectAPIName,
	constants.FieldSysValidation_Name,
	constants.FieldSysValidation_Active,
	constants.FieldSysValidation_Condition,
	constants.FieldSysValidation_ErrorMessage,
}

var flowColumns = []string{
	constants.FieldSysFlow_ID,
	constants.FieldSysFlow_Name,
	constants.FieldSysFlow_TriggerObject,
	constants.FieldSysFlow_TriggerType,
	constants.FieldSysFlow_TriggerCondition,
	constants.FieldSysFlow_ActionType,
	constants.FieldSysFlow_ActionConfig,
	constants.FieldSysFlow_Status,
	constants.FieldSysFlow_FlowType,
	constants.FieldSysFlow_Schedule,
	constants.FieldSysFlow_ScheduleTimezone,
	constants.FieldSysFlow_LastRunAt,
	constants.FieldSysFlow_NextRunAt,
	constants.FieldSysFlow_IsRunning,
	constants.FieldSysFlow_LastModifiedDate,
}

var sharingRuleColumns = []string{
	constants.FieldSysSharingRule_ID,
	constants.FieldSysSharingRule_ObjectAPIName,
	constants.FieldSysSharingRule_Name,
	constants.FieldSysSharingRule_Criteria,
	constants.FieldSysSharingRule_AccessLevel,
	constants.FieldSysSharingRule_ShareWithRoleID,
	constants.FieldSysSharingRule_ShareWithGroupID,
}

// =================================================================================
// Schema Queries
// =================================================================================

// GetSchemaByAPIName queries a single schema by API name
func (r *MetadataRepository) GetSchemaByAPIName(ctx context.Context, apiName string) (*models.ObjectMetadata, error) {
	objectQuery := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", strings.Join(objectColumns, ", "), constants.TableObject, constants.FieldAPIName)
	row := r.db.QueryRowContext(ctx, objectQuery, apiName)

	obj, err := r.scanObject(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan object: %w", err)
	}

	obj.Fields, err = r.GetFieldsForObject(ctx, obj.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load fields: %w", err)
	}

	return obj, nil
}

// GetAllSchemas queries all schemas
func (r *MetadataRepository) GetAllSchemas(ctx context.Context) ([]*models.ObjectMetadata, error) {
	objectQuery := fmt.Sprintf("SELECT %s FROM %s", strings.Join(objectColumns, ", "), constants.TableObject)
	rows, err := r.db.QueryContext(ctx, objectQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query objects: %w", err)
	}
	defer rows.Close()

	schemas := make([]*models.ObjectMetadata, 0)
	idToSchema := make(map[string]*models.ObjectMetadata)

	for rows.Next() {
		obj, err := r.scanObject(rows)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, obj)
		idToSchema[strings.ToLower(obj.ID)] = obj
	}

	fieldQuery := fmt.Sprintf("SELECT %s FROM %s", strings.Join(fieldColumns, ", "), constants.TableField)
	fieldRows, err := r.db.QueryContext(ctx, fieldQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query fields: %w", err)
	}
	defer func() { _ = fieldRows.Close() }()

	for fieldRows.Next() {
		field, objectID, err := r.scanField(fieldRows)
		if err != nil {
			log.Printf("⚠️ Failed to scan field: %v", err)
			continue
		}

		if obj, ok := idToSchema[strings.ToLower(objectID)]; ok {
			obj.Fields = append(obj.Fields, *field)
		}
	}

	return schemas, nil
}

// GetFieldsForObject queries fields for a specific object ID
func (r *MetadataRepository) GetFieldsForObject(ctx context.Context, objectID string) ([]models.FieldMetadata, error) {
	fieldQuery := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", strings.Join(fieldColumns, ", "), constants.TableField, constants.FieldObjectID)
	rows, err := r.db.QueryContext(ctx, fieldQuery, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query fields: %w", err)
	}
	defer rows.Close()

	fields := make([]models.FieldMetadata, 0)
	for rows.Next() {
		field, _, err := r.scanField(rows)
		if err != nil {
			log.Printf("⚠️ Failed to scan field: %v", err)
			continue
		}
		fields = append(fields, *field)
	}
	return fields, nil
}

// GetRecordTypes queries record types for an object
func (r *MetadataRepository) GetRecordTypes(ctx context.Context, objectAPIName string) ([]*models.RecordType, error) {
	// 1. Resolve Object ID
	var objectID string
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", constants.FieldID, constants.TableObject, constants.FieldSysObject_APIName), objectAPIName).Scan(&objectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*models.RecordType{}, nil
		}
		return nil, fmt.Errorf("failed to resolve object id: %w", err)
	}

	// 2. Query Record Types by Object ID
	// Note: _System_RecordType has columns: id, object_id, name, description, is_active, is_master
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldObjectID, constants.FieldSysRecordType_Name,
		constants.FieldSysRecordType_Description, constants.FieldSysRecordType_IsActive,
		"is_master", // Missing constant for is_master
		constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", cols, constants.TableRecordType, constants.FieldObjectID)

	rows, err := r.db.QueryContext(ctx, query, objectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	types := make([]*models.RecordType, 0)
	for rows.Next() {
		var rt models.RecordType
		var description sql.NullString
		var isActive, isMaster sql.NullBool
		var objID string

		// Scan matching the selected columns
		if err := rows.Scan(
			&rt.ID, &objID, &rt.Name,
			&description, &isActive, &isMaster,
			&rt.CreatedDate, &rt.LastModifiedDate,
		); err != nil {
			log.Printf("Warning: Failed to scan record type: %v", err)
			continue
		}

		rt.ObjectAPIName = objectAPIName // Set from argument
		rt.Label = rt.Name               // Set Label to Name as table has no separate Label

		desc := ""
		if description.Valid {
			desc = description.String
		}
		rt.Description = &desc

		// BusinessProcessID does not exist in schema, leaving nil

		rt.IsActive = isActive.Bool
		// rt.IsMaster = isMaster.Bool // If struct has IsMaster

		types = append(types, &rt)
	}
	return types, nil
}

// GetAutoNumbers queries auto numbers for an object
func (r *MetadataRepository) GetAutoNumbers(ctx context.Context, objectAPIName string) ([]*models.AutoNumber, error) {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldObjectAPIName, constants.FieldSysAutoNumber_FieldAPIName,
		constants.FieldSysAutoNumber_DisplayFormat, constants.FieldSysAutoNumber_StartingNumber,
		constants.FieldSysAutoNumber_CurrentNumber, constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = ?
	`, cols, constants.TableAutoNumber, constants.FieldObjectAPIName)
	rows, err := r.db.QueryContext(ctx, query, objectAPIName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	anList := make([]*models.AutoNumber, 0)
	for rows.Next() {
		var an models.AutoNumber
		if err := rows.Scan(
			&an.ID, &an.ObjectAPIName, &an.FieldAPIName, &an.DisplayFormat,
			&an.StartingNumber, &an.CurrentValue, &an.CreatedDate, &an.LastModifiedDate,
		); err != nil {
			log.Printf("Warning: Failed to scan auto number: %v", err)
			continue
		}
		anList = append(anList, &an)
	}
	return anList, nil
}

// GetRelationships queries relationships for a child object
func (r *MetadataRepository) GetRelationships(ctx context.Context, childObjectAPIName string) ([]*models.Relationship, error) {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysRelationship_ChildObjectAPIName, constants.FieldSysRelationship_ParentObjectAPIName,
		constants.FieldSysRelationship_FieldAPIName, constants.FieldSysRelationship_RelationshipName,
		constants.FieldSysRelationship_RelationshipType, constants.FieldSysRelationship_CascadeDelete,
		constants.FieldSysRelationship_RestrictedDelete, constants.FieldSysRelationship_RelatedListLabel,
		constants.FieldSysRelationship_RelatedListFields, constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = ?
	`, cols, constants.TableRelationship, constants.FieldSysRelationship_ChildObjectAPIName)
	rows, err := r.db.QueryContext(ctx, query, childObjectAPIName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rels := make([]*models.Relationship, 0)
	for rows.Next() {
		var rel models.Relationship
		var relatedListLabel, relatedListFields sql.NullString
		var cascadeDelete, restrictedDelete sql.NullBool

		if err := rows.Scan(
			&rel.ID, &rel.ChildObjectAPIName, &rel.ParentObjectAPIName, &rel.FieldAPIName,
			&rel.RelationshipName, &rel.RelationshipType, &cascadeDelete, &restrictedDelete,
			&relatedListLabel, &relatedListFields, &rel.CreatedDate, &rel.LastModifiedDate,
		); err != nil {
			log.Printf("Warning: Failed to scan relationship: %v", err)
			continue
		}

		if relatedListLabel.Valid {
			rel.RelatedListLabel = &relatedListLabel.String
		}
		if relatedListFields.Valid {
			rel.RelatedListFields = &relatedListFields.String
		}
		if cascadeDelete.Valid {
			rel.CascadeDelete = cascadeDelete.Bool
		}
		if restrictedDelete.Valid {
			rel.RestrictedDelete = restrictedDelete.Bool
		}

		rels = append(rels, &rel)
	}
	return rels, nil
}

// GetFieldDependencies queries field dependencies for an object
func (r *MetadataRepository) GetFieldDependencies(ctx context.Context, objectAPIName string) ([]*models.FieldDependency, error) {
	// 1. Resolve Object ID
	var objectID string
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", constants.FieldID, constants.TableObject, constants.FieldSysObject_APIName), objectAPIName).Scan(&objectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*models.FieldDependency{}, nil
		}
		return nil, fmt.Errorf("failed to resolve object id: %w", err)
	}

	// 2. Query Field Dependencies via JOIN with _System_Field to filter by Object
	// Schema: id, controlling_field_id, dependent_field_id, controlling_value, dependent_values
	cols := strings.Join([]string{
		"d." + constants.FieldID,
		"d." + constants.FieldSysFieldDependency_ControllingFieldID,
		"d." + constants.FieldSysFieldDependency_DependentFieldID,
		"d." + constants.FieldSysFieldDependency_ControllingValue,
		"d." + constants.FieldSysFieldDependency_DependentValues,
		"d." + constants.FieldCreatedDate,
		"d." + constants.FieldLastModifiedDate,
	}, ", ")

	// We join on dependent_field_id as dependencies are usually part of the dependent field's definition
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s d
		JOIN %s f ON d.%s = f.%s
		WHERE f.%s = ?
	`, cols, constants.TableFieldDependency,
		constants.TableField, constants.FieldSysFieldDependency_DependentFieldID, constants.FieldID, // d.dependent_field_id = f.id
		constants.FieldObjectID)

	rows, err := r.db.QueryContext(ctx, query, objectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deps := make([]*models.FieldDependency, 0)
	for rows.Next() {
		var dep models.FieldDependency
		var controllingValue, dependentValues sql.NullString

		if err := rows.Scan(
			&dep.ID, &dep.ControllingField, &dep.DependentField,
			&controllingValue, &dependentValues,
			&dep.CreatedDate, &dep.LastModifiedDate,
		); err != nil {
			log.Printf("Warning: Failed to scan field dependency: %v", err)
			continue
		}

		dep.ObjectAPIName = objectAPIName // Manual set

		if controllingValue.Valid {
			dep.ControllingValue = controllingValue.String
		}

		if dependentValues.Valid && dependentValues.String != "" {
			if err := json.Unmarshal([]byte(dependentValues.String), &dep.DependentValues); err != nil {
				log.Printf("Warning: Failed to unmarshal dependent values for %s: %v", dep.ID, err)
			}
		}

		deps = append(deps, &dep)
	}
	return deps, nil
}

// =================================================================================
// Logic Queries (Actions, Flows, Validation, Sharing)
// =================================================================================

// GetActions queries actions for an object
func (r *MetadataRepository) GetActions(ctx context.Context, objectAPIName string) ([]*models.ActionMetadata, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE LOWER(%s) = LOWER(?)", strings.Join(actionColumns, ", "), constants.TableAction, constants.FieldSysAction_ObjectAPIName)
	rows, err := r.db.QueryContext(ctx, query, objectAPIName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actions := make([]*models.ActionMetadata, 0)
	for rows.Next() {
		action, err := r.scanAction(rows)
		if err != nil {
			continue
		}
		actions = append(actions, action)
	}
	return actions, nil
}

// GetAction queries a single action by ID
func (r *MetadataRepository) GetAction(ctx context.Context, id string) (*models.ActionMetadata, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", strings.Join(actionColumns, ", "), constants.TableAction, constants.FieldSysAction_ID)
	action, err := r.scanAction(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return action, nil
}

// GetAllActions returns all actions
func (r *MetadataRepository) GetAllActions(ctx context.Context) ([]*models.ActionMetadata, error) {
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(actionColumns, ", "), constants.TableAction)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actions := make([]*models.ActionMetadata, 0)
	for rows.Next() {
		action, err := r.scanAction(rows)
		if err != nil {
			log.Printf("Warning: Failed to scan action: %v", err)
			continue
		}
		actions = append(actions, action)
	}
	return actions, nil
}

// CreateValidationRule creates a new validation rule
func (r *MetadataRepository) CreateValidationRule(ctx context.Context, rule *models.ValidationRule) error {
	query := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, `condition`, %s) VALUES (?, ?, ?, ?, ?, ?)", constants.TableValidation, constants.FieldID, constants.FieldSysValidation_ObjectAPIName, constants.FieldSysValidation_Name, constants.FieldSysValidation_Active, constants.FieldSysValidation_ErrorMessage)
	_, err := r.db.ExecContext(ctx, query, rule.ID, rule.ObjectAPIName, rule.Name, rule.Active, rule.Condition, rule.ErrorMessage)
	return err
}

// UpdateValidationRule updates a validation rule
func (r *MetadataRepository) UpdateValidationRule(ctx context.Context, id string, updates *models.ValidationRule) error {
	// Update the validation rule
	query := fmt.Sprintf("UPDATE %s SET %s = ?, %s = ?, `condition` = ?, %s = ? WHERE %s = ?", constants.TableValidation, constants.FieldSysValidation_Name, constants.FieldSysValidation_Active, constants.FieldSysValidation_ErrorMessage, constants.FieldID)
	_, err := r.db.ExecContext(ctx, query, updates.Name, updates.Active, updates.Condition, updates.ErrorMessage, id)
	return err
}

// GetValidationRule returns a single validation rule by ID
func (r *MetadataRepository) GetValidationRule(ctx context.Context, id string) (*models.ValidationRule, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", strings.Join(validationRuleColumns, ", "), constants.TableValidation, constants.FieldID)
	row := r.db.QueryRowContext(ctx, query, id)
	rule, err := r.scanValidationRule(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return rule, nil
}

// DeleteValidationRule deletes a validation rule
func (r *MetadataRepository) DeleteValidationRule(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = ?", constants.TableValidation, constants.FieldID), id)
	return err
}

// GetValidationRules queries validation rules for an object
func (r *MetadataRepository) GetValidationRules(ctx context.Context, objectAPIName string) ([]*models.ValidationRule, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE LOWER(%s) = LOWER(?)", strings.Join(validationRuleColumns, ", "), constants.TableValidation, constants.FieldSysValidation_ObjectAPIName)
	rows, err := r.db.QueryContext(ctx, query, objectAPIName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]*models.ValidationRule, 0)
	for rows.Next() {
		rule, err := r.scanValidationRule(rows)
		if err != nil {
			continue
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// GetAllFlows queries all flows
func (r *MetadataRepository) GetAllFlows(ctx context.Context) ([]*models.Flow, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = false OR %s IS NULL", strings.Join(flowColumns, ", "), constants.TableFlow, constants.FieldIsDeleted, constants.FieldIsDeleted)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flows := make([]*models.Flow, 0)
	for rows.Next() {
		flow, err := r.scanFlow(rows)
		if err != nil {
			log.Printf("⚠️ Failed to scan flow: %v\n", err)
			continue
		}
		flows = append(flows, flow)
	}
	return flows, nil
}

// GetFlow queries a single flow
func (r *MetadataRepository) GetFlow(ctx context.Context, id string) (*models.Flow, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", strings.Join(flowColumns, ", "), constants.TableFlow, constants.FieldSysFlow_ID)
	flow, err := r.scanFlow(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return flow, nil
}

// GetSharingRules queries sharing rules for an object
func (r *MetadataRepository) GetSharingRules(ctx context.Context, objectAPIName string) ([]*models.SystemSharingRule, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE LOWER(%s) = LOWER(?)", strings.Join(sharingRuleColumns, ", "), constants.TableSharingRule, constants.FieldSysSharingRule_ObjectAPIName)
	rows, err := r.db.QueryContext(ctx, query, objectAPIName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]*models.SystemSharingRule, 0)
	for rows.Next() {
		var rule models.SystemSharingRule
		var roleID, groupID sql.NullString
		if err := rows.Scan(&rule.ID, &rule.ObjectAPIName, &rule.Name, &rule.Criteria, &rule.AccessLevel, &roleID, &groupID); err != nil {
			continue
		}
		if roleID.Valid {
			rule.ShareWithRoleID = &roleID.String
		}
		if groupID.Valid {
			rule.ShareWithGroupID = &groupID.String
		}
		rules = append(rules, &rule)
	}
	return rules, nil
}

// =================================================================================
// Write Methods (Exec)
// =================================================================================

// Exec executes a query without returning result rows
func (r *MetadataRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.db.ExecContext(ctx, query, args...)
}

func (r *MetadataRepository) marshalJSON(v interface{}) (string, error) {
	if v == nil {
		return "{}", nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (r *MetadataRepository) unmarshalJSON(data string, v interface{}) {
	if data == "" {
		return
	}
	_ = json.Unmarshal([]byte(data), v)
}

// =================================================================================
// Write Methods (Actions & Flows)
// =================================================================================

// CreateAction creates a new action
func (r *MetadataRepository) CreateAction(ctx context.Context, action *models.ActionMetadata) error {
	configJSON, err := r.marshalJSON(action.Config)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	var targetObject sql.NullString
	if action.TargetObject != nil {
		targetObject.String = *action.TargetObject
		targetObject.Valid = true
	}

	cols := strings.Join([]string{
		constants.FieldSysAction_ID, constants.FieldSysAction_ObjectAPIName, constants.FieldSysAction_Name,
		constants.FieldSysAction_Label, constants.FieldSysAction_Type, constants.FieldSysAction_Icon,
		constants.FieldSysAction_TargetObject, constants.FieldSysAction_Config,
	}, ", ")
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", constants.TableAction, cols)
	_, err = r.db.ExecContext(ctx, query, action.ID, action.ObjectAPIName, action.Name, action.Label,
		action.Type, action.Icon, targetObject, configJSON)
	return err
}

// UpdateAction updates an existing action
func (r *MetadataRepository) UpdateAction(ctx context.Context, actionID string, updates *models.ActionMetadata) error {
	configJSON, err := r.marshalJSON(updates.Config)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	var targetObject sql.NullString
	if updates.TargetObject != nil {
		targetObject.String = *updates.TargetObject
		targetObject.Valid = true
	}

	updatesCols := strings.Join([]string{
		fmt.Sprintf("%s=?", constants.FieldSysAction_ObjectAPIName),
		fmt.Sprintf("%s=?", constants.FieldSysAction_Name),
		fmt.Sprintf("%s=?", constants.FieldSysAction_Label),
		fmt.Sprintf("%s=?", constants.FieldSysAction_Type),
		fmt.Sprintf("%s=?", constants.FieldSysAction_Icon),
		fmt.Sprintf("%s=?", constants.FieldSysAction_TargetObject),
		fmt.Sprintf("%s=?", constants.FieldSysAction_Config),
	}, ", ")
	query := fmt.Sprintf(`UPDATE %s SET %s WHERE %s=?`, constants.TableAction, updatesCols, constants.FieldSysAction_ID)
	_, err = r.db.ExecContext(ctx, query, updates.ObjectAPIName, updates.Name, updates.Label,
		updates.Type, updates.Icon, targetObject, configJSON, actionID)
	return err
}

// DeleteAction deletes an action
func (r *MetadataRepository) DeleteAction(ctx context.Context, actionID string) error {
	_, err := r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = ?", constants.TableAction, constants.FieldSysAction_ID), actionID)
	return err
}

// CheckActionExists checks for duplicate (object_api_name, name)
func (r *MetadataRepository) CheckActionExists(ctx context.Context, objectAPIName, name string) (bool, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ? AND %s = ?", constants.TableAction, constants.FieldSysAction_ObjectAPIName, constants.FieldSysAction_Name)
	if err := r.db.QueryRowContext(ctx, query, objectAPIName, name).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateFlow creates a new flow
func (r *MetadataRepository) CreateFlow(ctx context.Context, flow *models.Flow) error {
	actionConfigJSON, err := r.marshalJSON(flow.ActionConfig)
	if err != nil {
		return fmt.Errorf("failed to serialize action config: %w", err)
	}

	cols := strings.Join([]string{
		constants.FieldSysFlow_ID, constants.FieldSysFlow_Name, constants.FieldSysFlow_TriggerObject,
		constants.FieldSysFlow_TriggerType, constants.FieldSysFlow_TriggerCondition, constants.FieldSysFlow_ActionType,
		constants.FieldSysFlow_ActionConfig, constants.FieldSysFlow_Status, constants.FieldSysFlow_FlowType,
		constants.FieldSysFlow_Schedule, constants.FieldSysFlow_ScheduleTimezone, constants.FieldSysFlow_LastRunAt,
		constants.FieldSysFlow_NextRunAt, constants.FieldSysFlow_IsRunning, constants.FieldSysFlow_CreatedDate,
		constants.FieldSysFlow_LastModifiedDate,
	}, ", ")
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, constants.TableFlow, cols)

	now := time.Now()
	_, err = r.db.ExecContext(ctx, query,
		flow.ID, flow.Name, flow.TriggerObject, flow.TriggerType, flow.TriggerCondition,
		flow.ActionType, actionConfigJSON, flow.Status, flow.FlowType,
		flow.Schedule, flow.ScheduleTimezone, flow.LastRunAt, flow.NextRunAt, flow.IsRunning,
		now, // created_date
		now, // last_modified_date
	)
	return err
}

// UpdateFlow updates an existing flow
func (r *MetadataRepository) UpdateFlow(ctx context.Context, flowID string, flow *models.Flow) error {
	actionConfigJSON, err := r.marshalJSON(flow.ActionConfig)
	if err != nil {
		return fmt.Errorf("failed to serialize action config: %w", err)
	}

	updatesCols := strings.Join([]string{
		fmt.Sprintf("%s=?", constants.FieldSysFlow_Name),
		fmt.Sprintf("%s=?", constants.FieldSysFlow_TriggerObject),
		fmt.Sprintf("%s=?", constants.FieldSysFlow_TriggerType),
		fmt.Sprintf("%s=?", constants.FieldSysFlow_TriggerCondition),
		fmt.Sprintf("%s=?", constants.FieldSysFlow_ActionType),
		fmt.Sprintf("%s=?", constants.FieldSysFlow_ActionConfig),
		fmt.Sprintf("%s=?", constants.FieldSysFlow_Status),
		fmt.Sprintf("%s=?", constants.FieldSysFlow_FlowType),
		fmt.Sprintf("%s=?", constants.FieldSysFlow_Schedule),
		fmt.Sprintf("%s=?", constants.FieldSysFlow_ScheduleTimezone),
		fmt.Sprintf("%s=?", constants.FieldSysFlow_LastRunAt),
		fmt.Sprintf("%s=?", constants.FieldSysFlow_NextRunAt),
		fmt.Sprintf("%s=?", constants.FieldSysFlow_IsRunning),
		fmt.Sprintf("%s=?", constants.FieldSysFlow_LastModifiedDate),
	}, ", ")
	query := fmt.Sprintf(`UPDATE %s SET %s WHERE %s=?`, constants.TableFlow, updatesCols, constants.FieldSysFlow_ID)

	_, err = r.db.ExecContext(ctx, query,
		flow.Name, flow.TriggerObject, flow.TriggerType, flow.TriggerCondition,
		flow.ActionType, actionConfigJSON, flow.Status, flow.FlowType,
		flow.Schedule, flow.ScheduleTimezone, flow.LastRunAt, flow.NextRunAt, flow.IsRunning,
		time.Now(),
		flowID,
	)
	return err
}

// DeleteFlow deletes a flow
func (r *MetadataRepository) DeleteFlow(ctx context.Context, flowID string) error {
	_, err := r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = ?", constants.TableFlow, constants.FieldSysFlow_ID), flowID)
	return err
}

// SaveFlowSteps saves flow steps
func (r *MetadataRepository) SaveFlowSteps(ctx context.Context, flowID string, steps []models.FlowStep) error {
	cols := strings.Join([]string{
		constants.FieldSysFlowStep_ID, constants.FieldSysFlowStep_FlowID, constants.FieldStepName,
		constants.FieldStepType, constants.FieldStepOrder, constants.FieldSysFlowStep_ActionType,
		constants.FieldSysFlowStep_ActionConfig, constants.FieldSysFlowStep_EntryCondition,
		constants.FieldSysFlowStep_OnSuccessStep, constants.FieldSysFlowStep_OnFailureStep,
	}, ", ")
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, constants.TableFlowStep, cols)

	for _, step := range steps {
		// ID generation should ideally be done by caller or here if needed. Caller usually does validation/id gen.
		// Assuming ID is present.

		actionConfigJSON, err := r.marshalJSON(step.ActionConfig)
		if err != nil {
			return fmt.Errorf("failed to serialize step action config: %w", err)
		}

		_, err = r.db.ExecContext(ctx, query, step.ID, flowID, step.StepName, step.StepType, step.StepOrder,
			step.ActionType, actionConfigJSON, step.EntryCondition, step.OnSuccessStep, step.OnFailureStep)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteFlowSteps deletes all steps for a flow
func (r *MetadataRepository) DeleteFlowSteps(ctx context.Context, flowID string) error {
	_, err := r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = ?", constants.TableFlowStep, constants.FieldSysFlowStep_FlowID), flowID)
	return err
}

// GetFlowsByObject checks if any flow exists for an object
func (r *MetadataRepository) GetFlowsByObject(ctx context.Context, objectName string) ([]*models.Flow, error) {
	cols := strings.Join([]string{
		constants.FieldSysFlow_ID, constants.FieldSysFlow_TriggerObject,
		constants.FieldSysFlow_TriggerType, constants.FieldSysFlow_Status,
	}, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", cols, constants.TableFlow, constants.FieldSysFlow_TriggerObject)
	rows, err := r.db.QueryContext(ctx, query, objectName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flows []*models.Flow
	for rows.Next() {
		var f models.Flow
		if err := rows.Scan(&f.ID, &f.TriggerObject, &f.TriggerType, &f.Status); err != nil {
			return nil, err
		}
		flows = append(flows, &f)
	}
	return flows, nil
}

// =================================================================================
// UI / Layout Methods
// =================================================================================

// GetAllApps queries all apps
func (r *MetadataRepository) GetAllApps(ctx context.Context) ([]*models.AppConfig, error) {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysApp_Name, constants.FieldSysApp_Label,
		constants.FieldSysApp_Description, constants.FieldSysApp_Icon, constants.FieldSysApp_Color,
		constants.FieldSysApp_NavigationItems, constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s", cols, constants.TableApp)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apps := make([]*models.AppConfig, 0)
	for rows.Next() {
		app, err := r.scanApp(rows)
		if err != nil {
			log.Printf("Warning: Failed to scan app: %v", err)
			continue
		}
		apps = append(apps, app)
	}
	return apps, nil
}

// GetApp queries a single app by ID
func (r *MetadataRepository) GetApp(ctx context.Context, id string) (*models.AppConfig, error) {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysApp_Name, constants.FieldSysApp_Label,
		constants.FieldSysApp_Description, constants.FieldSysApp_Icon, constants.FieldSysApp_Color,
		constants.FieldSysApp_NavigationItems, constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", cols, constants.TableApp, constants.FieldID)
	app, err := r.scanApp(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return app, nil
}

// GetAppWithTx queries a single app by ID within a transaction
func (r *MetadataRepository) GetAppWithTx(ctx context.Context, tx *sql.Tx, id string) (*models.AppConfig, error) {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysApp_Name, constants.FieldSysApp_Label,
		constants.FieldSysApp_Description, constants.FieldSysApp_Icon, constants.FieldSysApp_Color,
		constants.FieldSysApp_NavigationItems, constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", cols, constants.TableApp, constants.FieldID)
	app, err := r.scanApp(tx.QueryRowContext(ctx, query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return app, nil
}

// UpdateAppWithTx updates an existing app within a provided transaction
func (r *MetadataRepository) UpdateAppWithTx(ctx context.Context, tx *sql.Tx, appID string, app *models.AppConfig) error {
	navItemsJSON, err := r.marshalJSON(app.NavigationItems)
	if err != nil {
		return fmt.Errorf("failed to marshal navigation items: %w", err)
	}

	updates := strings.Join([]string{
		fmt.Sprintf("%s = ?", constants.FieldSysApp_Label),
		fmt.Sprintf("%s = ?", constants.FieldSysApp_Description),
		fmt.Sprintf("%s = ?", constants.FieldSysApp_Icon),
		fmt.Sprintf("%s = ?", constants.FieldSysApp_Color),
		fmt.Sprintf("%s = ?", constants.FieldSysApp_NavigationItems),
		fmt.Sprintf("%s = ?", constants.FieldSysApp_IsDefault),
		fmt.Sprintf("%s = ?", constants.FieldLastModifiedDate),
	}, ", ")

	query := fmt.Sprintf(`
		UPDATE %s 
		SET %s
		WHERE %s = ?
	`, constants.TableApp, updates, constants.FieldID)

	_, err = tx.ExecContext(ctx, query,
		app.Label, app.Description, app.Icon, app.Color, navItemsJSON, app.IsDefault, time.Now(), appID,
	)
	return err
}

// GetLayouts queries all layouts for an object
func (r *MetadataRepository) GetLayouts(ctx context.Context, objectAPIName string) ([]*models.PageLayout, error) {
	cols := strings.Join([]string{
		constants.FieldSysLayout_Config, constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE LOWER(%s) = LOWER(?)", cols, constants.TableLayout, constants.FieldObjectAPIName)
	rows, err := r.db.QueryContext(ctx, query, objectAPIName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	layouts := make([]*models.PageLayout, 0)
	for rows.Next() {
		layout, err := r.scanLayout(rows)
		if err != nil {
			log.Printf("Warning: Failed to scan layout: %v", err)
			continue
		}
		layouts = append(layouts, layout)
	}
	return layouts, nil
}

// GetLayout queries a single layout by ID
func (r *MetadataRepository) GetLayout(ctx context.Context, layoutID string) (*models.PageLayout, error) {
	cols := strings.Join([]string{
		constants.FieldSysLayout_Config, constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	layout, err := r.scanLayout(r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", cols, constants.TableLayout, constants.FieldID), layoutID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return layout, nil
}

// GetLayoutIDForProfile returns the layout ID assigned to a profile for an object
func (r *MetadataRepository) GetLayoutIDForProfile(ctx context.Context, profileID, objectAPIName string) (string, error) {
	var layoutID string
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT %s FROM %s WHERE %s = ? AND LOWER(%s) = LOWER(?)",
		constants.FieldSysProfileLayout_LayoutID, constants.TableProfileLayout,
		constants.FieldProfileID, constants.FieldObjectAPIName), profileID, objectAPIName).Scan(&layoutID)
	if err == sql.ErrNoRows {
		return "", nil // Not assigned
	}
	if err != nil {
		return "", err
	}
	return layoutID, nil
}

// SaveLayout saves or updates a page layout
func (r *MetadataRepository) SaveLayout(ctx context.Context, layout *models.PageLayout) error {
	configJSON, err := r.marshalJSON(layout)
	if err != nil {
		return fmt.Errorf("failed to marshal layout: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (%s, %s, %s, %s, %s) 
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON DUPLICATE KEY UPDATE 
			%s = VALUES(%s),
			%s = VALUES(%s),
			%s = CURRENT_TIMESTAMP
	`, constants.TableLayout, constants.FieldID, constants.FieldObjectAPIName, constants.FieldSysLayout_Config, constants.FieldCreatedDate, constants.FieldLastModifiedDate, constants.FieldObjectAPIName, constants.FieldObjectAPIName, constants.FieldSysLayout_Config, constants.FieldSysLayout_Config, constants.FieldLastModifiedDate)

	_, err = r.db.ExecContext(ctx, query, layout.ID, layout.ObjectAPIName, configJSON)
	return err
}

// DeleteLayout deletes a layout
func (r *MetadataRepository) DeleteLayout(ctx context.Context, layoutID string) error {
	_, err := r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = ?", constants.TableLayout, constants.FieldID), layoutID)
	return err
}

// AssignLayoutToProfile assigns a layout to a profile
func (r *MetadataRepository) AssignLayoutToProfile(ctx context.Context, profileID, objectAPIName, layoutID string) error {
	cols := strings.Join([]string{
		constants.FieldProfileID, constants.FieldObjectAPIName, constants.FieldSysProfileLayout_LayoutID,
		constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	updates := strings.Join([]string{
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysProfileLayout_LayoutID, constants.FieldSysProfileLayout_LayoutID),
		fmt.Sprintf("%s = NOW()", constants.FieldLastModifiedDate),
	}, ", ")
	query := fmt.Sprintf(`
		INSERT INTO %s (%s) 
		VALUES (?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE %s
	`, constants.TableProfileLayout, cols, updates)

	_, err := r.db.ExecContext(ctx, query, profileID, objectAPIName, layoutID)
	return err
}

// CreateApp creates a new app
func (r *MetadataRepository) CreateApp(ctx context.Context, app *models.AppConfig) error {
	navItemsJSON, err := r.marshalJSON(app.NavigationItems)
	if err != nil {
		return fmt.Errorf("failed to marshal navigation items: %w", err)
	}

	cols := strings.Join([]string{
		constants.FieldSysApp_ID, constants.FieldSysApp_Name, constants.FieldSysApp_Label,
		constants.FieldSysApp_Description, constants.FieldSysApp_Icon, constants.FieldSysApp_Color,
		constants.FieldSysApp_IsDefault, constants.FieldSysApp_NavigationItems,
		constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, constants.TableApp, cols)

	_, err = r.db.ExecContext(ctx, query, app.ID, app.ID, app.Label, app.Description, app.Icon, app.Color, app.IsDefault, navItemsJSON, app.CreatedDate, app.LastModifiedDate)
	return err
}

// UpdateApp updates an existing app
func (r *MetadataRepository) UpdateApp(ctx context.Context, appID string, updates *models.AppConfig) error {

	navItemsJSON, err := r.marshalJSON(updates.NavigationItems)
	if err != nil {
		return fmt.Errorf("failed to marshal navigation items: %w", err)
	}

	updateFields := strings.Join([]string{
		fmt.Sprintf("%s = ?", constants.FieldSysApp_Label),
		fmt.Sprintf("%s = ?", constants.FieldSysApp_Description),
		fmt.Sprintf("%s = ?", constants.FieldSysApp_Icon),
		fmt.Sprintf("%s = ?", constants.FieldSysApp_Color),
		fmt.Sprintf("%s = ?", constants.FieldSysApp_IsDefault),
		fmt.Sprintf("%s = ?", constants.FieldSysApp_NavigationItems),
		fmt.Sprintf("%s = ?", constants.FieldLastModifiedDate),
	}, ", ")
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?", constants.TableApp, updateFields, constants.FieldID)
	_, err = r.db.ExecContext(ctx, query, updates.Label, updates.Description, updates.Icon, updates.Color, updates.IsDefault, navItemsJSON, updates.LastModifiedDate, appID)
	return err
}

// DeleteApp deletes an app
func (r *MetadataRepository) DeleteApp(ctx context.Context, appID string) error {
	_, err := r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = ?", constants.TableApp, constants.FieldID), appID)
	return err
}

// CreateDashboard creates a new dashboard
func (r *MetadataRepository) CreateDashboard(ctx context.Context, dashboard *models.DashboardConfig) error {
	widgetsJSON, err := r.marshalJSON(dashboard.Widgets)
	if err != nil {
		return fmt.Errorf("failed to marshal widgets: %w", err)
	}

	// Handle description pointer or value
	desc := ""
	if dashboard.Description != nil {
		desc = *dashboard.Description
	}

	query := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s) VALUES (?, ?, ?, ?, ?)", constants.TableDashboard, constants.FieldID, constants.FieldSysDashboard_Name, constants.FieldSysDashboard_Description, constants.FieldSysDashboard_Layout, constants.FieldSysDashboard_Widgets)
	_, err = r.db.ExecContext(ctx, query, dashboard.ID, dashboard.Label, desc, dashboard.Layout, widgetsJSON)
	return err
}

// UpdateDashboard updates a dashboard
func (r *MetadataRepository) UpdateDashboard(ctx context.Context, id string, dashboard *models.DashboardConfig) error {
	widgetsJSON, err := r.marshalJSON(dashboard.Widgets)
	if err != nil {
		return fmt.Errorf("failed to marshal widgets: %w", err)
	}

	desc := ""
	if dashboard.Description != nil {
		desc = *dashboard.Description
	}

	query := fmt.Sprintf("UPDATE %s SET %s = ?, %s = ?, %s = ?, %s = ? WHERE %s = ?", constants.TableDashboard, constants.FieldSysDashboard_Name, constants.FieldSysDashboard_Description, constants.FieldSysDashboard_Layout, constants.FieldSysDashboard_Widgets, constants.FieldID)
	_, err = r.db.ExecContext(ctx, query, dashboard.Label, desc, dashboard.Layout, widgetsJSON, id)
	return err
}

// DeleteDashboard deletes a dashboard
func (r *MetadataRepository) DeleteDashboard(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = ?", constants.TableDashboard, constants.FieldID), id)
	return err
}

// GetAllDashboards queries all dashboards
func (r *MetadataRepository) GetAllDashboards(ctx context.Context) ([]*models.DashboardConfig, error) {
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf("SELECT %s, %s, %s, %s, %s FROM %s", constants.FieldID, constants.FieldSysDashboard_Name, constants.FieldSysDashboard_Description, constants.FieldSysDashboard_Layout, constants.FieldSysDashboard_Widgets, constants.TableDashboard))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dashboards := make([]*models.DashboardConfig, 0)
	for rows.Next() {
		db, err := r.scanDashboard(rows)
		if err != nil {
			log.Printf("Warning: Failed to scan dashboard: %v", err)
			continue
		}
		dashboards = append(dashboards, db)
	}
	return dashboards, nil
}

// GetDashboard queries a single dashboard
func (r *MetadataRepository) GetDashboard(ctx context.Context, id string) (*models.DashboardConfig, error) {
	db, err := r.scanDashboard(r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT %s, %s, %s, %s, %s FROM %s WHERE %s = ?", constants.FieldID, constants.FieldSysDashboard_Name, constants.FieldSysDashboard_Description, constants.FieldSysDashboard_Layout, constants.FieldSysDashboard_Widgets, constants.TableDashboard, constants.FieldID),
		id,
	))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return db, nil
}

// GetListViews queries list views for an object
func (r *MetadataRepository) GetListViews(ctx context.Context, objectAPIName string) ([]*models.ListView, error) {
	query := fmt.Sprintf("SELECT %s, %s, %s, %s, %s FROM %s WHERE LOWER(%s) = LOWER(?)", constants.FieldID, constants.FieldObjectAPIName, constants.FieldSysListView_Label, constants.FieldSysListView_FilterExpr, constants.FieldSysListView_Fields, constants.TableListView, constants.FieldObjectAPIName)
	rows, err := r.db.QueryContext(ctx, query, objectAPIName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	views := make([]*models.ListView, 0)
	for rows.Next() {
		view, err := r.scanListView(rows)
		if err != nil {
			log.Printf("Warning: Failed to scan list view: %v", err)
			continue
		}
		views = append(views, view)
	}
	return views, nil
}

// GetScheduledFlows returns all scheduled flows
func (r *MetadataRepository) GetScheduledFlows(ctx context.Context) ([]*models.Flow, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s 
		WHERE %s = ? AND (%s = false OR %s IS NULL)
	`, strings.Join(flowColumns, ", "), constants.TableFlow, constants.FieldSysFlow_TriggerType, constants.FieldIsDeleted, constants.FieldIsDeleted)
	// Note: The service logic used a bigger SELECT with description, schedule, etc.
	// But `scanFlow` only scans standard fields.
	// I should update `scanFlow` or use custom scan if schema differs.
	// `constants.TableFlow` has schedule columns?
	// The Service `GetScheduledFlows` selects: id, name... description, status, schedule, schedule_timezone, last_run_at, next_run_at, is_running.
	// `scanFlow` (Line 1133) scans: id, name, trigger_object, trigger_type, trigger_condition, action_type, action_config, status, flow_type, last_modified.
	// It misses schedule fields!
	// Schema change? Or standard `Flow` model has them?
	// `models.Flow` has `Schedule`, `ScheduleTimezone`, `LastRunAt`, `NextRunAt`, `IsRunning`?
	// Step 991 Line 117-158 shows it does.
	// `scanFlow` in Repo is INCOMPLETE for scheduled flows?
	// Or `scanFlow` is for "Metadata". Scheduled flows are "Runtime" or "Metadata"?
	// They are in `sys_flow` table.
	// I should probably update `scanFlow` to include all columns if they exist in DB.
	// Or create `GetScheduledFlows` with its own scan logic inside Repo.
	// I'll stick to `scanFlow` for now and assume standard metadata.
	// If `scanFlow` misses schedule info, Schedule Trigger won't work?
	// I will update `scanFlow` later. For now, I'll match scanFlow columns.

	rows, err := r.db.QueryContext(ctx, query, constants.TriggerTypeSchedule)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flows := make([]*models.Flow, 0)
	for rows.Next() {
		flow, err := r.scanFlow(rows)
		if err != nil {
			log.Printf("Warning: Failed to scan flow: %v", err)
			continue
		}
		flows = append(flows, flow)
	}
	return flows, nil
}

// CreateListView creates a new list view
func (r *MetadataRepository) CreateListView(ctx context.Context, view *models.ListView) error {
	fieldsJSON, err := r.marshalJSON(view.Fields)
	if err != nil {
		return fmt.Errorf("failed to marshal fields: %w", err)
	}

	cols := strings.Join([]string{
		constants.FieldID, constants.FieldObjectAPIName, constants.FieldSysListView_Label,
		constants.FieldSysListView_FilterExpr, constants.FieldSysListView_Fields,
	}, ", ")
	_, err = r.db.ExecContext(ctx,
		fmt.Sprintf("INSERT INTO %s (%s) VALUES (?, ?, ?, ?, ?)", constants.TableListView, cols),
		view.ID, view.ObjectAPIName, view.Label, view.FilterExpr, fieldsJSON,
	)
	return err
}

// UpdateListView updates a list view
func (r *MetadataRepository) UpdateListView(ctx context.Context, id string, updates *models.ListView) error {
	fieldsJSON, err := r.marshalJSON(updates.Fields)
	if err != nil {
		return fmt.Errorf("failed to marshal fields: %w", err)
	}

	updateFields := strings.Join([]string{
		fmt.Sprintf("%s = ?", constants.FieldSysListView_Label),
		fmt.Sprintf("%s = ?", constants.FieldSysListView_FilterExpr),
		fmt.Sprintf("%s = ?", constants.FieldSysListView_Fields),
	}, ", ")

	result, err := r.db.ExecContext(ctx,
		fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?", constants.TableListView, updateFields, constants.FieldID),
		updates.Label, updates.FilterExpr, fieldsJSON, id,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteListView deletes a list view
func (r *MetadataRepository) DeleteListView(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = ?", constants.TableListView, constants.FieldID), id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CountListViews counts list views for an object
func (r *MetadataRepository) CountListViews(ctx context.Context, objectAPIName string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ?", constants.TableListView, constants.FieldObjectAPIName), objectAPIName).Scan(&count)
	return count, err
}

// UpsertLayout updates or inserts a page layout
func (r *MetadataRepository) UpsertLayout(ctx context.Context, layout *models.PageLayout) error {
	configJSON, err := r.marshalJSON(layout)
	if err != nil {
		return fmt.Errorf("failed to marshal layout: %w", err)
	}

	cols := strings.Join([]string{
		constants.FieldID, constants.FieldObjectAPIName, constants.FieldSysLayout_Config,
		constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	updates := strings.Join([]string{
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldObjectAPIName, constants.FieldObjectAPIName),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysLayout_Config, constants.FieldSysLayout_Config),
		fmt.Sprintf("%s = NOW()", constants.FieldLastModifiedDate),
	}, ", ")
	// Used for EnsureDefaultLayout and CreateSchema
	_, err = r.db.ExecContext(ctx,
		fmt.Sprintf("INSERT INTO %s (%s) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) ON DUPLICATE KEY UPDATE %s", constants.TableLayout, cols, updates),
		layout.ID, layout.ObjectAPIName, configJSON,
	)
	return err
}

// BatchUpsertLayouts updates or inserts multiple page layouts
func (r *MetadataRepository) BatchUpsertLayouts(ctx context.Context, layouts []*models.PageLayout) error {
	for _, layout := range layouts {
		if err := r.UpsertLayout(ctx, layout); err != nil {
			return err
		}
	}
	return nil
}

// GetChildRelationships returns fields on OTHER objects that lookup to this object
func (r *MetadataRepository) GetChildRelationships(ctx context.Context, parentObjectAPIName string) ([]*models.ObjectMetadata, error) {
	// Query fields that reference this object
	query := fmt.Sprintf(`
		SELECT o.%s 
		FROM %s f
		JOIN %s o ON f.%s = o.%s
		WHERE (f.%s = ? OR f.%s LIKE ?) AND f.%s = 'Lookup'
	`,
		constants.FieldAPIName,
		constants.TableField,
		constants.TableObject, constants.FieldObjectID, constants.FieldID,
		constants.FieldReferenceTo, constants.FieldReferenceTo, constants.FieldType)

	likePattern := fmt.Sprintf("%%%s%%", parentObjectAPIName)
	rows, err := r.db.QueryContext(ctx, query, parentObjectAPIName, likePattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var children []*models.ObjectMetadata
	for rows.Next() {
		var apiName string
		if err := rows.Scan(&apiName); err != nil {
			log.Printf("Warning: Failed to scan child relationship: %v", err)
			continue
		}

		// Load full schema for child
		// Use internal method or recursive call?
		// GetSchemaByAPIName is available on r *MetadataRepository
		if schema, err := r.GetSchemaByAPIName(ctx, apiName); err == nil && schema != nil {
			children = append(children, schema)
		}
	}
	return children, nil
}

// UpsertUIComponent inserts or updates a UI component definition
func (r *MetadataRepository) UpsertUIComponent(ctx context.Context, component *models.UIComponent) error {
	// Check if exists by Name
	var existingID string
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", constants.FieldID, constants.TableUIComponent, constants.FieldSysUIComponent_Name), component.Name).Scan(&existingID)

	if err == nil {
		// Found, update it
		component.ID = existingID
		updates := strings.Join([]string{
			fmt.Sprintf("%s = ?", constants.FieldSysUIComponent_Description),
			fmt.Sprintf("%s = ?", constants.FieldSysUIComponent_Type),
			fmt.Sprintf("%s = ?", constants.FieldSysUIComponent_IsEmbeddable),
			fmt.Sprintf("%s = ?", constants.FieldSysUIComponent_ComponentPath),
			fmt.Sprintf("%s = CURRENT_TIMESTAMP", constants.FieldLastModifiedDate),
		}, ", ")
		query := fmt.Sprintf(`
			UPDATE %s SET %s
			WHERE %s = ?`, constants.TableUIComponent, updates, constants.FieldID)

		_, err = r.db.ExecContext(ctx, query, component.Description, component.Type, component.IsEmbeddable, component.ComponentPath, component.ID)
		return err
	}

	if err != sql.ErrNoRows {
		return err // Real error
	}

	// Not found, Insert
	if component.ID == "" {
		component.ID = utils.GenerateID()
	}

	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysUIComponent_Name, constants.FieldSysUIComponent_Type,
		constants.FieldSysUIComponent_IsEmbeddable, constants.FieldSysUIComponent_Description,
		constants.FieldSysUIComponent_ComponentPath, constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	query := fmt.Sprintf(`
		INSERT INTO %s (%s)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, constants.TableUIComponent, cols)

	_, err = r.db.ExecContext(ctx, query, component.ID, component.Name, component.Type, component.IsEmbeddable, component.Description, component.ComponentPath)
	return err
}

// UpsertSetupPage inserts or updates a setup page definition
func (r *MetadataRepository) UpsertSetupPage(ctx context.Context, page *models.SetupPage) error {
	// Check by component_name which is unique - not ID which may not be set on first load
	var existingID string
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", constants.FieldID, constants.TableSetupPage, constants.FieldSysSetupPage_ComponentName), page.ComponentName).Scan(&existingID)

	if err == nil {
		updates := strings.Join([]string{
			fmt.Sprintf("%s = ?", constants.FieldSysSetupPage_Label),
			fmt.Sprintf("%s = ?", constants.FieldSysSetupPage_Icon),
			fmt.Sprintf("%s = ?", constants.FieldSysSetupPage_ComponentName),
			fmt.Sprintf("%s = ?", constants.FieldSysSetupPage_Category),
			fmt.Sprintf("%s = ?", constants.FieldSysSetupPage_PageOrder),
			fmt.Sprintf("%s = ?", constants.FieldSysSetupPage_PermissionRequired),
			fmt.Sprintf("%s = ?", constants.FieldSysSetupPage_IsEnabled),
			fmt.Sprintf("%s = ?", constants.FieldSysSetupPage_Description),
			fmt.Sprintf("%s = CURRENT_TIMESTAMP", constants.FieldLastModifiedDate),
		}, ", ")
		query := fmt.Sprintf(`
			UPDATE %s SET %s
			WHERE %s = ?`, constants.TableSetupPage, updates, constants.FieldID)

		_, err = r.db.ExecContext(ctx, query, page.Label, page.Icon, page.ComponentName, page.Category, page.PageOrder, page.PermissionRequired, page.IsEnabled, page.Description, existingID)
		return err
	}

	if err != sql.ErrNoRows {
		return err
	}

	if page.ID == "" {
		page.ID = utils.GenerateID()
	}

	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysSetupPage_Label, constants.FieldSysSetupPage_Icon,
		constants.FieldSysSetupPage_ComponentName, constants.FieldSysSetupPage_Category,
		constants.FieldSysSetupPage_PageOrder, constants.FieldSysSetupPage_PermissionRequired,
		constants.FieldSysSetupPage_IsEnabled, constants.FieldSysSetupPage_Description,
		constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	query := fmt.Sprintf(`
		INSERT INTO %s (%s)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, constants.TableSetupPage, cols)

	_, err = r.db.ExecContext(ctx, query, page.ID, page.Label, page.Icon, page.ComponentName, page.Category, page.PageOrder, page.PermissionRequired, page.IsEnabled, page.Description)
	return err
}

// GetSetupPages returns all setup pages
func (r *MetadataRepository) GetSetupPages(ctx context.Context) ([]models.SetupPage, error) {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysSetupPage_Label, constants.FieldSysSetupPage_Icon,
		constants.FieldSysSetupPage_ComponentName, constants.FieldSysSetupPage_Category,
		constants.FieldSysSetupPage_PageOrder, constants.FieldSysSetupPage_PermissionRequired,
		constants.FieldSysSetupPage_IsEnabled, constants.FieldSysSetupPage_Description,
		constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s 
		ORDER BY %s ASC`, cols, constants.TableSetupPage, constants.FieldSysSetupPage_PageOrder)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []models.SetupPage
	for rows.Next() {
		var p models.SetupPage
		var createdDateVal, lastModifiedDateVal interface{}

		if err := rows.Scan(
			&p.ID, &p.Label, &p.Icon, &p.ComponentName, &p.Category,
			&p.PageOrder, &p.PermissionRequired, &p.IsEnabled, &p.Description,
			&createdDateVal, &lastModifiedDateVal,
		); err != nil {
			return nil, err
		}

		p.CreatedDate = parseTime(createdDateVal)
		p.LastModifiedDate = parseTime(lastModifiedDateVal)
		pages = append(pages, p)
	}

	return pages, rows.Err()
}

// ==================== Theme Methods ====================

func (r *MetadataRepository) scanTheme(row Scannable) (*models.Theme, error) {
	var theme models.Theme
	var colorsJSON, logoURL sql.NullString
	var createdDateVal, lastModifiedDateVal interface{}

	// Columns: id, name, is_active, colors, density, logo_url, __sys_gen_created_date, last_modified_date
	if err := row.Scan(&theme.ID, &theme.Name, &theme.IsActive, &colorsJSON, &theme.Density, &logoURL, &createdDateVal, &lastModifiedDateVal); err != nil {
		return nil, err
	}

	theme.LogoURL = models.NullStringToPtr(logoURL)
	if colorsJSON.Valid {
		if err := models.ParseJSON(colorsJSON.String, &theme.Colors); err != nil {
			log.Printf("Warning: Failed to parse theme colors: %v", err)
		}
	}
	theme.CreatedDate = parseTime(createdDateVal)
	theme.LastModifiedDate = parseTime(lastModifiedDateVal)

	return &theme, nil
}

// GetActiveTheme returns the currently active theme
func (r *MetadataRepository) GetActiveTheme(ctx context.Context) (*models.Theme, error) {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysTheme_Name, constants.FieldSysTheme_IsActive,
		constants.FieldSysTheme_Colors, constants.FieldSysTheme_Density, constants.FieldSysTheme_LogoURL,
		constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = true LIMIT 1", cols, constants.TableTheme, constants.FieldSysTheme_IsActive)
	row := r.db.QueryRowContext(ctx, query)
	theme, err := r.scanTheme(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil if no active theme found
		}
		return nil, err
	}
	return theme, nil
}

// GetThemeByName returns a theme by name
func (r *MetadataRepository) GetThemeByName(ctx context.Context, name string) (*models.Theme, error) {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysTheme_Name, constants.FieldSysTheme_IsActive,
		constants.FieldSysTheme_Colors, constants.FieldSysTheme_Density, constants.FieldSysTheme_LogoURL,
		constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", cols, constants.TableTheme, constants.FieldSysTheme_Name)
	row := r.db.QueryRowContext(ctx, query, name)
	theme, err := r.scanTheme(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return theme, nil
}

// CreateTheme creates a new theme
func (r *MetadataRepository) CreateTheme(ctx context.Context, theme *models.Theme) error {
	colorsJSON, err := r.marshalJSON(theme.Colors)
	if err != nil {
		return fmt.Errorf("failed to marshal colors: %w", err)
	}

	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysTheme_Name, constants.FieldSysTheme_IsActive,
		constants.FieldSysTheme_Colors, constants.FieldSysTheme_Density, constants.FieldSysTheme_LogoURL,
		constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", constants.TableTheme, cols)
	_, err = r.db.ExecContext(ctx, query, theme.ID, theme.Name, theme.IsActive, colorsJSON, theme.Density, theme.LogoURL, theme.CreatedDate, theme.LastModifiedDate)
	return err
}

// UpdateTheme updates an existing theme
func (r *MetadataRepository) UpdateTheme(ctx context.Context, theme *models.Theme) error {
	colorsJSON, err := r.marshalJSON(theme.Colors)
	if err != nil {
		return fmt.Errorf("failed to marshal colors: %w", err)
	}

	updates := strings.Join([]string{
		fmt.Sprintf("%s = ?", constants.FieldSysTheme_IsActive),
		fmt.Sprintf("%s = ?", constants.FieldSysTheme_Colors),
		fmt.Sprintf("%s = ?", constants.FieldSysTheme_Density),
		fmt.Sprintf("%s = ?", constants.FieldSysTheme_LogoURL),
		fmt.Sprintf("%s = ?", constants.FieldLastModifiedDate),
	}, ", ")
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?", constants.TableTheme, updates, constants.FieldID)
	_, err = r.db.ExecContext(ctx, query, theme.IsActive, colorsJSON, theme.Density, theme.LogoURL, theme.LastModifiedDate, theme.ID)
	return err
}

// ActivateTheme sets a specific theme as active and deactivates all others
func (r *MetadataRepository) ActivateTheme(ctx context.Context, themeID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Deactivate all
	_, err = tx.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET %s = false", constants.TableTheme, constants.FieldSysTheme_IsActive))
	if err != nil {
		return fmt.Errorf("failed to deactivate themes: %w", err)
	}

	// 2. Activate target
	result, err := tx.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET %s = true WHERE %s = ?", constants.TableTheme, constants.FieldSysTheme_IsActive, constants.FieldID), themeID)
	if err != nil {
		return fmt.Errorf("failed to activate theme: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("theme not found: %s", themeID)
	}

	return tx.Commit()
}

// UpsertAutoNumber inserts or updates an auto number configuration
func (r *MetadataRepository) UpsertAutoNumber(ctx context.Context, id, objectAPIName, fieldAPIName, displayFormat string, startingNumber, currentNumber int) error {
	query := fmt.Sprintf(`INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s, %s)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW()) 
		ON DUPLICATE KEY UPDATE %s = VALUES(%s), %s = NOW()`,
		constants.TableAutoNumber, constants.FieldID, constants.FieldObjectAPIName, constants.FieldSysAutoNumber_FieldAPIName,
		constants.FieldSysAutoNumber_DisplayFormat, constants.FieldSysAutoNumber_StartingNumber,
		constants.FieldSysAutoNumber_CurrentNumber, constants.FieldCreatedDate, constants.FieldLastModifiedDate,
		constants.FieldSysAutoNumber_DisplayFormat, constants.FieldSysAutoNumber_DisplayFormat, constants.FieldLastModifiedDate)
	_, err := r.db.ExecContext(ctx, query, id, objectAPIName, fieldAPIName, displayFormat, startingNumber, currentNumber)
	return err
}

// GetRelatedListConfigs queries related list configs for relationship
func (r *MetadataRepository) GetRelatedListConfigs(ctx context.Context, layoutObjectAPIName string) ([]struct {
	LookupFieldAPI    string
	ChildObjectAPI    string
	ChildPluralLabel  string
	RelatedListFields sql.NullString
}, error) {

	query := fmt.Sprintf(`
		SELECT f.%s, o.%s, o.%s, r.%s
		FROM %s f
		JOIN %s o ON f.%s = o.%s
		LEFT JOIN %s r ON r.%s = o.%s AND r.%s = f.%s
		WHERE f.%s = ? AND f.%s = 'Lookup'
	`,
		constants.FieldAPIName, constants.FieldAPIName, constants.FieldPluralLabel, constants.FieldSysRelationship_RelatedListFields,
		constants.TableField,
		constants.TableObject, constants.FieldObjectID, constants.FieldID,
		constants.TableRelationship, constants.FieldSysRelationship_ChildObjectAPIName, constants.FieldAPIName, constants.FieldSysRelationship_FieldAPIName, constants.FieldAPIName,
		constants.FieldReferenceTo, constants.FieldType)

	rows, err := r.db.QueryContext(ctx, query, layoutObjectAPIName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		LookupFieldAPI    string
		ChildObjectAPI    string
		ChildPluralLabel  string
		RelatedListFields sql.NullString
	}

	for rows.Next() {
		var res struct {
			LookupFieldAPI    string
			ChildObjectAPI    string
			ChildPluralLabel  string
			RelatedListFields sql.NullString
		}
		if err := rows.Scan(&res.LookupFieldAPI, &res.ChildObjectAPI, &res.ChildPluralLabel, &res.RelatedListFields); err != nil {
			log.Printf("Warning: Failed to scan child relationship row: %v", err)
			continue
		}
		results = append(results, res)
	}
	return results, nil
}

// =================================================================================
// Scan Helpers (Local to Package)
// =================================================================================

type Scannable interface {
	Scan(dest ...interface{}) error
}

func (r *MetadataRepository) scanObject(row Scannable) (*models.ObjectMetadata, error) {
	var obj models.ObjectMetadata
	var description, icon, pathField, listFieldsJSON, appID sql.NullString
	var isCustom bool

	err := row.Scan(
		&obj.ID, &obj.APIName, &obj.Label, &obj.PluralLabel,
		&icon, &description, &isCustom, &pathField, &listFieldsJSON,
		&appID, &obj.ThemeColor,
	)
	if err != nil {
		return nil, err
	}

	if description.Valid {
		val := description.String
		obj.Description = &val
	}
	if icon.Valid {
		obj.Icon = icon.String
	}
	if appID.Valid {
		val := appID.String
		obj.AppID = &val
	}
	if pathField.Valid {
		val := pathField.String
		obj.PathField = &val
	}
	obj.IsCustom = isCustom
	obj.IsSystem = !isCustom
	// Unmarshal ListFields
	if listFieldsJSON.Valid {
		r.unmarshalJSON(listFieldsJSON.String, &obj.ListFields)
	}
	obj.SharingModel = constants.SharingModelPrivate
	obj.Searchable = true
	obj.EnableHierarchySharing = false
	obj.Fields = make([]models.FieldMetadata, 0)

	return &obj, nil
}

func (r *MetadataRepository) scanField(row Scannable) (*models.FieldMetadata, string, error) {
	var field models.FieldMetadata
	var id, objectAPIName string
	var required, unique, isSystem, trackHistory, isNameField, isMasterDetail, isPolymorphic sql.NullBool
	var options, referenceTo, formula, returnType, defaultValue, helpText, controllingField, picklistDependency, rollupConfig, deleteRule, relationshipName, regex, regexMessage, validator, description sql.NullString
	var minValue, maxValue sql.NullFloat64
	var minLength, maxLength sql.NullInt64

	err := row.Scan(
		&id, &objectAPIName, &field.APIName, &field.Label, &field.Type,
		&required, &unique, &isSystem, &isNameField, &options,
		&referenceTo, &deleteRule, &isMasterDetail, &relationshipName,
		&formula, &returnType, &defaultValue, &isPolymorphic, &helpText, &description,
		&trackHistory, &minValue, &maxValue, &minLength, &maxLength,
		&regex, &regexMessage, &validator, &controllingField,
		&picklistDependency, &rollupConfig,
	)
	if err != nil {
		return nil, "", err
	}

	field.Required = required.Bool
	field.IsUnique = unique.Bool
	field.IsSystem = isSystem.Bool
	field.TrackHistory = trackHistory.Bool
	field.IsNameField = isNameField.Bool
	field.IsMasterDetail = isMasterDetail.Bool
	field.IsPolymorphic = isPolymorphic.Bool

	if formula.Valid {
		field.Formula = &formula.String
	}
	if defaultValue.Valid {
		field.DefaultValue = &defaultValue.String
	}
	if helpText.Valid {
		field.HelpText = &helpText.String
	}
	if controllingField.Valid {
		field.ControllingField = &controllingField.String
	}
	if relationshipName.Valid {
		field.RelationshipName = &relationshipName.String
	}
	if regex.Valid {
		field.Regex = &regex.String
	}
	if regexMessage.Valid {
		field.RegexMessage = &regexMessage.String
	}
	if validator.Valid {
		field.Validator = &validator.String
	}
	if returnType.Valid {
		rt := models.FieldType(returnType.String)
		field.ReturnType = &rt
	}
	if deleteRule.Valid {
		dr := models.DeleteRule(deleteRule.String)
		field.DeleteRule = &dr
	}

	if minValue.Valid {
		field.MinValue = &minValue.Float64
	}
	if maxValue.Valid {
		field.MaxValue = &maxValue.Float64
	}
	if minLength.Valid {
		val := int(minLength.Int64)
		field.MinLength = &val
	}
	if maxLength.Valid {
		val := int(maxLength.Int64)
		field.MaxLength = &val
	}

	// Unmarshal JSON fields
	if referenceTo.Valid {
		r.unmarshalJSON(referenceTo.String, &field.ReferenceTo)
	}
	if options.Valid {
		r.unmarshalJSON(options.String, &field.Options)
	}
	if picklistDependency.Valid {
		r.unmarshalJSON(picklistDependency.String, &field.PicklistDependency)
	}
	if rollupConfig.Valid {
		var rc models.RollupConfig
		r.unmarshalJSON(rollupConfig.String, &rc)
		field.RollupConfig = &rc
	}

	return &field, objectAPIName, nil
}

func (r *MetadataRepository) scanAction(row Scannable) (*models.ActionMetadata, error) {
	var action models.ActionMetadata
	var targetObject, configJSON sql.NullString
	if err := row.Scan(&action.ID, &action.ObjectAPIName, &action.Name, &action.Label, &action.Type, &action.Icon, &targetObject, &configJSON); err != nil {
		return nil, err
	}
	if targetObject.Valid {
		action.TargetObject = &targetObject.String
	}
	if configJSON.Valid {
		r.unmarshalJSON(configJSON.String, &action.Config)
	}
	return &action, nil
}

func (r *MetadataRepository) scanValidationRule(row Scannable) (*models.ValidationRule, error) {
	var rule models.ValidationRule
	var active int
	if err := row.Scan(&rule.ID, &rule.ObjectAPIName, &rule.Name, &active, &rule.Condition, &rule.ErrorMessage); err != nil {
		return nil, err
	}
	rule.Active = active != 0
	return &rule, nil
}

func (r *MetadataRepository) scanFlow(row Scannable) (*models.Flow, error) {
	var flow models.Flow
	var lastModifiedDateVal, lastRunAtVal, nextRunAtVal interface{}
	var schedule, scheduleTimezone, actionConfigJSON sql.NullString

	if err := row.Scan(
		&flow.ID, &flow.Name, &flow.TriggerObject, &flow.TriggerType, &flow.TriggerCondition,
		&flow.ActionType, &actionConfigJSON, &flow.Status, &flow.FlowType,
		&schedule, &scheduleTimezone, &lastRunAtVal, &nextRunAtVal, &flow.IsRunning,
		&lastModifiedDateVal,
	); err != nil {
		return nil, err
	}
	if actionConfigJSON.Valid {
		r.unmarshalJSON(actionConfigJSON.String, &flow.ActionConfig)
	}

	flow.Schedule = models.NullStringToPtr(schedule)
	flow.ScheduleTimezone = models.NullStringToPtr(scheduleTimezone)

	lastRun := parseTime(lastRunAtVal)
	if !lastRun.IsZero() {
		flow.LastRunAt = &lastRun
	}
	nextRun := parseTime(nextRunAtVal)
	if !nextRun.IsZero() {
		flow.NextRunAt = &nextRun
	}

	lastMod := parseTime(lastModifiedDateVal)
	if !lastMod.IsZero() {
		flow.LastModified = lastMod.Format(time.RFC3339)
	}

	return &flow, nil
}

func (r *MetadataRepository) scanApp(row Scannable) (*models.AppConfig, error) {
	var app models.AppConfig
	var name string
	var description, icon, color, navItems sql.NullString
	var createdDateVal, lastModifiedDateVal interface{}

	if err := row.Scan(&app.ID, &name, &app.Label, &description, &icon, &color, &navItems, &createdDateVal, &lastModifiedDateVal); err != nil {
		return nil, err
	}

	if description.Valid {
		app.Description = description.String
	}
	if icon.Valid {
		app.Icon = icon.String
	}
	if color.Valid {
		app.Color = color.String
	}

	if navItems.Valid {
		r.unmarshalJSON(navItems.String, &app.NavigationItems)
	}

	app.CreatedDate = parseTime(createdDateVal)
	app.LastModifiedDate = parseTime(lastModifiedDateVal)

	return &app, nil
}

func parseTime(val interface{}) time.Time {
	if val == nil {
		return time.Time{}
	}
	switch v := val.(type) {
	case time.Time:
		return v
	case []uint8:
		str := string(v)
		if t, err := time.Parse("2006-01-02 15:04:05", str); err == nil {
			return t
		}
		if t, err := time.Parse(time.RFC3339, str); err == nil {
			return t
		}
	}
	return time.Time{}
}

func (r *MetadataRepository) scanLayout(row Scannable) (*models.PageLayout, error) {
	var configJSON string
	var createdDateVal, lastModifiedDateVal interface{}
	if err := row.Scan(&configJSON, &createdDateVal, &lastModifiedDateVal); err != nil {
		return nil, err
	}

	var layout models.PageLayout
	if err := json.Unmarshal([]byte(configJSON), &layout); err != nil {
		return nil, err
	}
	layout.CreatedDate = parseTime(createdDateVal)
	layout.LastModifiedDate = parseTime(lastModifiedDateVal)
	return &layout, nil
}

func (r *MetadataRepository) scanDashboard(row Scannable) (*models.DashboardConfig, error) {
	var db models.DashboardConfig
	var description, widgetsJSON sql.NullString

	if err := row.Scan(&db.ID, &db.Label, &description, &db.Layout, &widgetsJSON); err != nil {
		return nil, err
	}

	if description.Valid {
		db.Description = &description.String
	}
	if widgetsJSON.Valid {
		r.unmarshalJSON(widgetsJSON.String, &db.Widgets)
	}
	return &db, nil
}

func (r *MetadataRepository) scanListView(row Scannable) (*models.ListView, error) {
	var view models.ListView
	var filterExpr, fieldsJSON sql.NullString
	if err := row.Scan(&view.ID, &view.ObjectAPIName, &view.Label, &filterExpr, &fieldsJSON); err != nil {
		return nil, err
	}
	if filterExpr.Valid {
		view.FilterExpr = filterExpr.String
	}
	if fieldsJSON.Valid {
		r.unmarshalJSON(fieldsJSON.String, &view.Fields)
	}
	return &view, nil
}
