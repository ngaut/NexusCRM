package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// Helper methods to eliminate redundancy and ensuring consistency in metadata persistence

func (sm *SchemaManager) prepareObjectDBValues(obj *models.ObjectMetadata) ([]interface{}, error) {
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

func (sm *SchemaManager) prepareFieldDBValues(field *models.FieldMetadata) ([]interface{}, error) {
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
	if field.Unique {
		unique = 1
	}
	isMasterDetail := 0
	if field.IsMasterDetail {
		isMasterDetail = 1
	}

	// Order matches FieldInsertQuery
	return []interface{}{
		field.APIName, field.Label, field.Type, required, unique,
		defaultValue, helpText, isSystem, isNameField, optionsJSON,
		minLength, maxLength, referenceTo, formula, returnType, rollupConfigJSON,
		isMasterDetail, deleteRule, relationshipName,
	}, nil
}

func (sm *SchemaManager) getObjectInsertQuery() string {
	return fmt.Sprintf(`INSERT INTO %s (
		id, api_name, label, plural_label, icon, description, 
		is_custom, sharing_model, app_id, list_fields, path_field, theme_color, table_type,
		created_date, last_modified_date
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	ON DUPLICATE KEY UPDATE
		label = VALUES(label),
		plural_label = VALUES(plural_label),
		icon = VALUES(icon),
		description = VALUES(description),
		sharing_model = VALUES(sharing_model),
		app_id = VALUES(app_id),
		list_fields = VALUES(list_fields),
		path_field = VALUES(path_field),
		theme_color = VALUES(theme_color),
		table_type = VALUES(table_type),
		last_modified_date = NOW()
	`, constants.TableObject)
}

func (sm *SchemaManager) getObjectStrictInsertQuery() string {
	return fmt.Sprintf(`INSERT INTO %s (
		id, api_name, label, plural_label, icon, description, 
		is_custom, sharing_model, app_id, list_fields, path_field, theme_color, table_type,
		created_date, last_modified_date
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`, constants.TableObject)
}

func (sm *SchemaManager) getFieldInsertQuery() string {
	return fmt.Sprintf(`INSERT INTO %s (
		id, object_id, api_name, label, type, required, `+"`unique`"+`,
		default_value, help_text, is_system, is_name_field, options,
		min_length, max_length, reference_to, formula, return_type, rollup_config,
		is_master_detail, delete_rule, relationship_name,
		created_date, last_modified_date
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	ON DUPLICATE KEY UPDATE
		label = VALUES(label),
		type = VALUES(type),
		required = VALUES(required),
		`+"`unique`"+` = VALUES(`+"`unique`"+`),
		default_value = VALUES(default_value),
		help_text = VALUES(help_text),
		is_system = VALUES(is_system),
		is_name_field = VALUES(is_name_field),
		options = VALUES(options),
		min_length = VALUES(min_length),
		max_length = VALUES(max_length),
		reference_to = VALUES(reference_to),
		formula = VALUES(formula),
		return_type = VALUES(return_type),
		rollup_config = VALUES(rollup_config),
		is_master_detail = VALUES(is_master_detail),
		delete_rule = VALUES(delete_rule),
		relationship_name = VALUES(relationship_name),
		last_modified_date = NOW()
	`, constants.TableField)
}
