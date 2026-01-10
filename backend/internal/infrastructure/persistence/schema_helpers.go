package persistence

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// Helper methods to eliminate redundancy and ensuring consistency in metadata persistence

func (r *SchemaRepository) prepareObjectDBValues(obj *models.ObjectMetadata) ([]interface{}, error) {
	description := ToNullString(obj.Description)
	var icon sql.NullString
	if obj.Icon != "" {
		icon = sql.NullString{String: obj.Icon, Valid: true}
	}
	appID := ToNullString(obj.AppID)
	sharingModel := obj.SharingModel
	if sharingModel == "" {
		sharingModel = models.SharingModel(constants.SharingModelPrivate)
	}

	var listFields sql.NullString
	if len(obj.ListFields) > 0 {
		b, err := json.Marshal(obj.ListFields)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal list fields: %w", err)
		}
		listFields = sql.NullString{String: string(b), Valid: true}
	}

	pathField := ToNullString(obj.PathField)
	themeColor := ToNullString(obj.ThemeColor)

	tableType := constants.TableTypeSystemMetadata
	if obj.IsCustom {
		tableType = constants.TableTypeCustomObject
	}
	// For core objects bootstrapped via this path, we default to metadata/custom.
	// Real system core objects are often inserted via dedicated bootstrap scripts, but if using this, defaulting to metadata is safer than null.

	// Order matches ObjectInsertQuery
	return []interface{}{
		obj.APIName, obj.Label, obj.PluralLabel, icon, description,
		obj.IsCustom, sharingModel, appID, listFields, pathField, themeColor, tableType,
	}, nil
}

func (r *SchemaRepository) prepareFieldDBValues(field *models.FieldMetadata) ([]interface{}, error) {
	defaultValue := ToNullString(field.DefaultValue)
	helpText := ToNullString(field.HelpText)
	minLength := ToNullInt64(field.MinLength)
	maxLength := ToNullInt64(field.MaxLength)
	referenceTo := SliceToNullJSON(field.ReferenceTo)
	formula := ToNullString(field.Formula)

	var returnTypeVal *string
	if field.ReturnType != nil {
		s := string(*field.ReturnType)
		returnTypeVal = &s
	}
	returnType := ToNullString(returnTypeVal)

	var optionsJSON interface{}
	if field.Options != nil {
		b, err := json.Marshal(field.Options)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal options: %w", err)
		}
		optionsJSON = string(b)
	}

	var rollupConfigJSON interface{}
	if field.RollupConfig != nil {
		b, err := json.Marshal(field.RollupConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal rollup config: %w", err)
		}
		rollupConfigJSON = string(b)
	}

	var deleteRuleVal *string
	if field.DeleteRule != nil {
		s := string(*field.DeleteRule)
		deleteRuleVal = &s
	}
	deleteRule := ToNullString(deleteRuleVal)
	relationshipName := ToNullString(field.RelationshipName)

	// Convert bools to ints for compatibility
	isSystem := 0
	if field.IsSystem {
		isSystem = 1
	}
	isNameField := 0
	if field.IsNameField {
		isNameField = 1
	}
	required := 0
	if field.Required {
		required = 1
	}
	unique := 0
	if field.IsUnique {
		unique = 1
	}
	isMasterDetail := 0
	if field.IsMasterDetail {
		isMasterDetail = 1
	}
	isPolymorphic := 0
	if field.IsPolymorphic {
		isPolymorphic = 1
	}

	// Order matches FieldInsertQuery
	return []interface{}{
		field.APIName, field.Label, field.Type, required, unique,
		defaultValue, helpText, isSystem, isNameField, optionsJSON,
		minLength, maxLength, referenceTo, formula, returnType, rollupConfigJSON,
		isMasterDetail, isPolymorphic, deleteRule, relationshipName,
	}, nil
}

func (r *SchemaRepository) getObjectInsertQuery() string {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysObject_APIName, constants.FieldSysObject_Label, constants.FieldSysObject_PluralLabel,
		constants.FieldSysObject_Icon, constants.FieldSysObject_Description, constants.FieldSysObject_IsCustom,
		constants.FieldSysObject_SharingModel, constants.FieldSysObject_AppID, constants.FieldSysObject_ListFields,
		constants.FieldSysObject_PathField, constants.FieldSysObject_ThemeColor, constants.FieldSysObject_TableType,
		constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")

	updates := strings.Join([]string{
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObject_Label, constants.FieldSysObject_Label),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObject_PluralLabel, constants.FieldSysObject_PluralLabel),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObject_Icon, constants.FieldSysObject_Icon),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObject_Description, constants.FieldSysObject_Description),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObject_SharingModel, constants.FieldSysObject_SharingModel),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObject_AppID, constants.FieldSysObject_AppID),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObject_ListFields, constants.FieldSysObject_ListFields),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObject_PathField, constants.FieldSysObject_PathField),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObject_ThemeColor, constants.FieldSysObject_ThemeColor),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObject_TableType, constants.FieldSysObject_TableType),
		fmt.Sprintf("%s = NOW()", constants.FieldLastModifiedDate),
	}, ", ")

	return fmt.Sprintf(`%s %s (%s) %s (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)
	%s %s`, KeywordInsertInto, constants.TableObject, cols, KeywordValues, FuncNow, FuncNow,
		KeywordOnDuplicate, updates)
}

func (r *SchemaRepository) getObjectStrictInsertQuery() string {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysObject_APIName, constants.FieldSysObject_Label, constants.FieldSysObject_PluralLabel,
		constants.FieldSysObject_Icon, constants.FieldSysObject_Description, constants.FieldSysObject_IsCustom,
		constants.FieldSysObject_SharingModel, constants.FieldSysObject_AppID, constants.FieldSysObject_ListFields,
		constants.FieldSysObject_PathField, constants.FieldSysObject_ThemeColor, constants.FieldSysObject_TableType,
		constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	return fmt.Sprintf(`%s %s (%s) %s (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`,
		KeywordInsertInto, constants.TableObject, cols, KeywordValues, FuncNow, FuncNow)
}

func (r *SchemaRepository) getFieldInsertQuery() string {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldObjectID, constants.FieldAPIName, constants.FieldSysField_Label, constants.FieldType,
		constants.FieldSysField_Required, constants.FieldSysField_IsUnique, constants.FieldSysField_DefaultValue, constants.FieldSysField_HelpText,
		constants.FieldSysField_IsSystem, constants.FieldSysField_IsNameField, constants.FieldSysField_Options,
		constants.FieldSysField_MinLength, constants.FieldSysField_MaxLength, constants.FieldReferenceTo,
		constants.FieldSysField_Formula, constants.FieldSysField_ReturnType, constants.FieldSysField_RollupConfig,
		constants.FieldSysField_IsMasterDetail, constants.FieldSysField_IsPolymorphic, constants.FieldSysField_DeleteRule,
		constants.FieldSysField_RelationshipName, constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")

	updates := strings.Join([]string{
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_Label, constants.FieldSysField_Label),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldType, constants.FieldType),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_Required, constants.FieldSysField_Required),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_IsUnique, constants.FieldSysField_IsUnique),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_DefaultValue, constants.FieldSysField_DefaultValue),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_HelpText, constants.FieldSysField_HelpText),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_IsSystem, constants.FieldSysField_IsSystem),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_IsNameField, constants.FieldSysField_IsNameField),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_Options, constants.FieldSysField_Options),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_MinLength, constants.FieldSysField_MinLength),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_MaxLength, constants.FieldSysField_MaxLength),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldReferenceTo, constants.FieldReferenceTo),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_Formula, constants.FieldSysField_Formula),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_ReturnType, constants.FieldSysField_ReturnType),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_RollupConfig, constants.FieldSysField_RollupConfig),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_IsMasterDetail, constants.FieldSysField_IsMasterDetail),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_IsPolymorphic, constants.FieldSysField_IsPolymorphic),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_DeleteRule, constants.FieldSysField_DeleteRule),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysField_RelationshipName, constants.FieldSysField_RelationshipName),
		fmt.Sprintf("%s = NOW()", constants.FieldLastModifiedDate),
	}, ", ")

	return fmt.Sprintf(`%s %s (%s) %s (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)
	%s %s`, KeywordInsertInto, constants.TableField, cols, KeywordValues, FuncNow, FuncNow,
		KeywordOnDuplicate, updates)
}
