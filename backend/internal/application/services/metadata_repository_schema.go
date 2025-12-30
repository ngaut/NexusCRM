package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

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
	"`" + constants.FieldMetaType + "`",
	constants.FieldIsRequired,
	"`" + constants.FieldIsUnique + "`",
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
	constants.FieldCreatedDate,
	constants.FieldLastModifiedDate,
}

// scanObject scans a row into an ObjectMetadata struct
func (ms *MetadataService) scanObject(row Scannable) (*models.ObjectMetadata, error) {
	var obj models.ObjectMetadata
	var description, icon, pathField, listFieldsJSON, appID sql.NullString
	var isCustom bool

	err := row.Scan(
		&obj.ID,
		&obj.APIName,
		&obj.Label,
		&obj.PluralLabel,
		&icon,
		&description,
		&isCustom,
		&pathField,
		&listFieldsJSON,
		&appID,
		&obj.ThemeColor,
	)
	if err != nil {
		return nil, err
	}

	obj.Description = ScanNullString(description)
	obj.Icon = ScanNullStringValue(icon)
	obj.AppID = ScanNullString(appID)
	obj.PathField = ScanNullString(pathField)
	obj.IsCustom = isCustom
	obj.IsSystem = !isCustom

	UnmarshalJSONField(listFieldsJSON, &obj.ListFields)

	// Set defaults
	obj.SharingModel = constants.SharingModelPrivate
	obj.Searchable = true
	obj.EnableHierarchySharing = false
	obj.Fields = make([]models.FieldMetadata, 0)

	return &obj, nil
}

// querySchemaByAPIName queries a single schema from the database by API name
func (ms *MetadataService) querySchemaByAPIName(apiName string) (*models.ObjectMetadata, error) {
	// Query the object
	objectQuery := fmt.Sprintf("SELECT %s FROM %s WHERE api_name = ?", strings.Join(objectColumns, ", "), constants.TableObject)
	row := ms.db.QueryRow(objectQuery, apiName)

	obj, err := ms.scanObject(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan object: %w", err)
	}

	// Load fields for this object
	obj.Fields, err = ms.queryFieldsForObject(obj.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load fields: %w", err)
	}

	return obj, nil
}

// queryAllSchemas queries all schemas from the database
func (ms *MetadataService) queryAllSchemas() ([]*models.ObjectMetadata, error) {
	// Query all objects
	objectQuery := fmt.Sprintf("SELECT %s FROM %s", strings.Join(objectColumns, ", "), constants.TableObject)
	rows, err := ms.db.Query(objectQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query objects: %w", err)
	}
	defer rows.Close()

	schemas := make([]*models.ObjectMetadata, 0)
	idToSchema := make(map[string]*models.ObjectMetadata)

	for rows.Next() {
		obj, err := ms.scanObject(rows)
		if err != nil {
			return nil, err
		}

		schemas = append(schemas, obj)
		// IMPORTANT: Normalize ID to lowercase for robust map lookups.
		// Database IDs might be MixedCase (e.g., 'obj_ChildNew3') but Field 'object_id' references might be lowercase.
		idToSchema[strings.ToLower(obj.ID)] = obj
	}

	// Load all fields
	fieldQuery := fmt.Sprintf("SELECT %s FROM %s", strings.Join(fieldColumns, ", "), constants.TableField)
	fieldRows, err := ms.db.Query(fieldQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query fields: %w", err)
	}
	defer func() { _ = fieldRows.Close() }()

	for fieldRows.Next() {
		field, objectID, err := ms.scanField(fieldRows)
		if err != nil {
			log.Printf("⚠️ Failed to scan field: %v", err)
			continue
		}

		// Match field to parent object using lowercase ID to handle case sensitivity mismatch
		if obj, ok := idToSchema[strings.ToLower(objectID)]; ok {
			obj.Fields = append(obj.Fields, *field)
		}
	}

	return schemas, nil
}

// queryFieldsForObject queries all fields for a specific object
func (ms *MetadataService) queryFieldsForObject(objectID string) ([]models.FieldMetadata, error) {
	fieldQuery := fmt.Sprintf("SELECT %s FROM %s WHERE object_id = ?", strings.Join(fieldColumns, ", "), constants.TableField)

	rows, err := ms.db.Query(fieldQuery, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query fields: %w", err)
	}
	defer rows.Close()

	fields := make([]models.FieldMetadata, 0)
	for rows.Next() {
		field, _, err := ms.scanField(rows)
		if err != nil {
			log.Printf("⚠️ Failed to scan field: %v", err)
			continue
		}
		fields = append(fields, *field)
	}

	return fields, nil
}

// scanField scans a row into a FieldMetadata struct
func (ms *MetadataService) scanField(row Scannable) (*models.FieldMetadata, string, error) {
	var field models.FieldMetadata
	var id string
	var objectAPIName string
	var required, unique, isSystem, trackHistory, isNameField sql.NullBool
	var isMasterDetail sql.NullBool
	var options, referenceTo, formula, returnType, defaultValue, helpText, controllingField sql.NullString
	var picklistDependency, rollupConfig sql.NullString
	var deleteRule, relationshipName sql.NullString
	var minValue, maxValue sql.NullFloat64
	var minLength, maxLength sql.NullInt64
	var regex, regexMessage, validator sql.NullString
	var description sql.NullString
	var createdDate, lastModifiedDate sql.NullTime

	err := row.Scan(
		&id,                 // Column 1: id
		&objectAPIName,      // Column 2: ObjectId (this is the parent object's ApiName we need!)
		&field.APIName,      // Column 3: ApiName
		&field.Label,        // Column 4: Label
		&field.Type,         // Column 5: Type
		&required,           // Column 6: Required
		&unique,             // Column 7: Unique
		&isSystem,           // Column 8: isSystem
		&isNameField,        // Column 9: isNameField
		&options,            // Column 10: options (Picklist)
		&referenceTo,        // Column 11: referenceTo (Lookup)
		&deleteRule,         // Column 12: deleteRule (Lookup)
		&isMasterDetail,     // Column 13: isMasterDetail
		&relationshipName,   // Column 14: relationshipName
		&formula,            // Column 15: formula (Formula)
		&returnType,         // Column 16: returnType (Formula)
		&defaultValue,       // Column 17: defaultValue
		&helpText,           // Column 18: helpText
		&description,        // Column 19: Description
		&trackHistory,       // Column 20: trackHistory
		&minValue,           // Column 21: minValue
		&maxValue,           // Column 22: maxValue
		&minLength,          // Column 23: minLength
		&maxLength,          // Column 24: maxLength
		&regex,              // Column 25: regex
		&regexMessage,       // Column 26: regexMessage
		&validator,          // Column 27: validator
		&controllingField,   // Column 28: controllingField
		&picklistDependency, // Column 29: picklistDependency
		&rollupConfig,       // Column 30: rollupConfig
		&createdDate,        // Column 31: CreatedDate
		&lastModifiedDate,   // Column 32: LastModifiedDate
	)
	if err != nil {
		return nil, "", err
	}

	// Convert nullable fields
	field.Required = ScanNullBool(required)
	field.Unique = ScanNullBool(unique)
	field.IsSystem = ScanNullBool(isSystem)
	field.TrackHistory = ScanNullBool(trackHistory)
	field.IsNameField = ScanNullBool(isNameField)

	UnmarshalJSONField(options, &field.Options)

	// ReferenceTo: stored as JSON array, e.g., ["Account"] or ["Account", "Contact"]
	UnmarshalJSONField(referenceTo, &field.ReferenceTo)
	field.IsPolymorphic = len(field.ReferenceTo) > 1

	field.Formula = ScanNullString(formula)

	if returnType.Valid {
		rt := models.FieldType(returnType.String)
		field.ReturnType = &rt
	}

	field.DefaultValue = ScanNullString(defaultValue)
	field.HelpText = ScanNullString(helpText)
	field.ControllingField = ScanNullString(controllingField)

	if deleteRule.Valid {
		dr := models.DeleteRule(deleteRule.String)
		field.DeleteRule = &dr
	}

	field.IsMasterDetail = ScanNullBool(isMasterDetail)
	field.RelationshipName = ScanNullString(relationshipName)

	field.MinValue = ScanNullFloat64(minValue)
	field.MaxValue = ScanNullFloat64(maxValue)
	field.MinLength = ScanNullInt64(minLength)
	field.MaxLength = ScanNullInt64(maxLength)
	field.Regex = ScanNullString(regex)
	field.RegexMessage = ScanNullString(regexMessage)
	field.Validator = ScanNullString(validator)

	UnmarshalJSONField(picklistDependency, &field.PicklistDependency)
	UnmarshalJSONField(rollupConfig, &field.RollupConfig)

	return &field, objectAPIName, nil
}

// queryRecordTypes queries record types for an object
func (ms *MetadataService) queryRecordTypes(objectAPIName string) ([]*models.RecordType, error) {
	query := fmt.Sprintf(`
		SELECT id, object_api_name, name, label, description, is_active, is_default,
		       business_process_id, created_date, last_modified_date
		FROM %s
		WHERE object_api_name = ?
	`, constants.TableRecordType)
	rows, err := ms.db.Query(query, objectAPIName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	types := make([]*models.RecordType, 0)
	for rows.Next() {
		var rt models.RecordType
		var description, businessProcessID sql.NullString
		var isActive, isDefault sql.NullBool

		if err := rows.Scan(
			&rt.ID, &rt.ObjectAPIName, &rt.Name, &rt.Label,
			&description, &isActive, &isDefault, &businessProcessID,
			&rt.CreatedDate, &rt.LastModifiedDate,
		); err != nil {
			continue
		}

		rt.Description = ScanNullString(description)
		rt.BusinessProcessID = ScanNullString(businessProcessID)
		rt.IsActive = ScanNullBool(isActive)
		rt.IsDefault = ScanNullBool(isDefault)

		types = append(types, &rt)
	}
	return types, nil
}

// queryAutoNumbers queries auto numbers for an object
func (ms *MetadataService) queryAutoNumbers(objectAPIName string) ([]*models.AutoNumber, error) {
	query := fmt.Sprintf(`
		SELECT id, object_api_name, field_api_name, display_format, starting_number,
		       current_number, created_date, last_modified_date
		FROM %s
		WHERE object_api_name = ?
	`, constants.TableAutoNumber)
	rows, err := ms.db.Query(query, objectAPIName)
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
			continue
		}
		anList = append(anList, &an)
	}
	return anList, nil
}

// queryRelationships queries relationships for a child object
func (ms *MetadataService) queryRelationships(childObjectAPIName string) ([]*models.Relationship, error) {
	query := fmt.Sprintf(`
		SELECT id, child_object_api_name, parent_object_api_name, field_api_name,
		       relationship_name, relationship_type, cascade_delete, restricted_delete,
		       related_list_label, related_list_fields, created_date, last_modified_date
		FROM %s
		WHERE child_object_api_name = ?
	`, constants.TableRelationship)
	rows, err := ms.db.Query(query, childObjectAPIName)
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
			continue
		}

		rel.RelatedListLabel = ScanNullString(relatedListLabel)
		rel.RelatedListFields = ScanNullString(relatedListFields)
		rel.CascadeDelete = ScanNullBool(cascadeDelete)
		rel.RestrictedDelete = ScanNullBool(restrictedDelete)

		rels = append(rels, &rel)
	}
	return rels, nil
}

// queryFieldDependencies queries field dependencies for an object
func (ms *MetadataService) queryFieldDependencies(objectAPIName string) ([]*models.FieldDependency, error) {
	query := fmt.Sprintf(`
		SELECT id, object_api_name, controlling_field, dependent_field, controlling_value,
		       action, is_active, created_date, last_modified_date
		FROM %s
		WHERE object_api_name = ?
	`, constants.TableFieldDependency)
	rows, err := ms.db.Query(query, objectAPIName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deps := make([]*models.FieldDependency, 0)
	for rows.Next() {
		var dep models.FieldDependency
		var controllingValue, action sql.NullString
		var isActive sql.NullBool

		if err := rows.Scan(
			&dep.ID, &dep.ObjectAPIName, &dep.ControllingField, &dep.DependentField,
			&controllingValue, &action, &isActive, &dep.CreatedDate, &dep.LastModifiedDate,
		); err != nil {
			continue
		}

		if controllingValue.Valid {
			dep.ControllingValue = controllingValue.String
		}
		if action.Valid {
			dep.Action = action.String
		}
		dep.IsActive = isActive.Bool

		deps = append(deps, &dep)
	}
	return deps, nil
}
