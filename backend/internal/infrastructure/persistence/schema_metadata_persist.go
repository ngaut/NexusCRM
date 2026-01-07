package persistence

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/nexuscrm/backend/internal/domain/schema"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// FieldWithContext holds field metadata with context for batch operations
type FieldWithContext struct {
	Field    *models.FieldMetadata
	ObjectID string
	FieldID  string
}

// SaveObjectMetadata upserts object metadata into _System_Object
func (r *SchemaRepository) SaveObjectMetadata(obj *models.ObjectMetadata, exec Executor) error {
	if exec == nil {
		exec = r.db
	}

	// Determine Object ID if not set
	if obj.ID == "" {
		obj.ID = GenerateObjectID(obj.APIName)
	}

	values, err := r.prepareObjectDBValues(obj)
	if err != nil {
		return fmt.Errorf("failed to prepare object values for %s: %w", obj.APIName, err)
	}

	args := append([]interface{}{obj.ID}, values...)
	_, err = exec.Exec(r.getObjectInsertQuery(), args...)

	return err
}

// InsertObjectMetadata inserts object metadata into _System_Object (Strict - Fails on Unique Constraint)
func (r *SchemaRepository) InsertObjectMetadata(obj *models.ObjectMetadata, exec Executor) error {
	if exec == nil {
		exec = r.db
	}

	// Determine Object ID if not set
	if obj.ID == "" {
		obj.ID = GenerateObjectID(obj.APIName)
	}

	values, err := r.prepareObjectDBValues(obj)
	if err != nil {
		return fmt.Errorf("failed to prepare object values for %s: %w", obj.APIName, err)
	}

	args := append([]interface{}{obj.ID}, values...)
	_, err = exec.Exec(r.getObjectStrictInsertQuery(), args...)

	return err
}

// SaveFieldMetadataWithIDs upserts field metadata with explicit IDs
func (r *SchemaRepository) SaveFieldMetadataWithIDs(field *models.FieldMetadata, objectID string, fieldID string, exec Executor) error {
	if exec == nil {
		exec = r.db
	}

	values, err := r.prepareFieldDBValues(field)
	if err != nil {
		return fmt.Errorf("failed to prepare field values for %s: %w", field.APIName, err)
	}

	args := append([]interface{}{fieldID, objectID}, values...)
	_, err = exec.Exec(r.getFieldInsertQuery(), args...)

	return err
}

// BatchSaveObjectMetadata inserts multiple objects in a single statement
func (r *SchemaRepository) BatchSaveObjectMetadata(objs []*models.ObjectMetadata, exec Executor) error {
	if len(objs) == 0 {
		return nil
	}
	if exec == nil {
		exec = r.db
	}

	// Build multi-row INSERT
	var valuePlaceholders []string
	var args []interface{}

	for _, obj := range objs {
		// Ensure ID is set
		if obj.ID == "" {
			obj.ID = GenerateObjectID(obj.APIName)
		}

		values, err := r.prepareObjectDBValues(obj)
		if err != nil {
			return err
		}

		valuePlaceholders = append(valuePlaceholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())")
		args = append(args, obj.ID)
		args = append(args, values...)
	}

	query := fmt.Sprintf(`INSERT INTO %s (
		id, api_name, label, plural_label, icon, description, 
		is_custom, sharing_model, app_id, list_fields, path_field, theme_color, table_type,
		created_date, last_modified_date
	) VALUES %s
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
	`, constants.TableObject, strings.Join(valuePlaceholders, ", "))

	_, err := exec.Exec(query, args...)
	return err
}

// BatchSaveFieldMetadata inserts multiple fields in a single statement
func (r *SchemaRepository) BatchSaveFieldMetadata(fields []FieldWithContext, exec Executor) error {
	if len(fields) == 0 {
		return nil
	}
	if exec == nil {
		exec = r.db
	}

	// Process in batches of 50 to avoid overly long queries
	batchSize := 50
	for i := 0; i < len(fields); i += batchSize {
		end := i + batchSize
		if end > len(fields) {
			end = len(fields)
		}
		batch := fields[i:end]

		var valuePlaceholders []string
		var args []interface{}

		for _, fc := range batch {
			values, err := r.prepareFieldDBValues(fc.Field)
			if err != nil {
				return err
			}

			valuePlaceholders = append(valuePlaceholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())")
			args = append(args, fc.FieldID, fc.ObjectID)
			args = append(args, values...)
		}

		query := fmt.Sprintf(`INSERT INTO %s (
			id, object_id, api_name, label, type, required, `+"`unique`"+`, 
			default_value, help_text, is_system, is_name_field, options,
			min_length, max_length, reference_to, formula, return_type, rollup_config,
			is_master_detail, delete_rule, relationship_name,
			created_date, last_modified_date
		) VALUES %s
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
		`, constants.TableField, strings.Join(valuePlaceholders, ", "))

		if _, err := exec.Exec(query, args...); err != nil {
			return fmt.Errorf("batch field insert failed: %w", err)
		}
	}

	return nil
}

// PrepareFieldForBatch converts a column definition to FieldWithContext for batch processing
func (r *SchemaRepository) PrepareFieldForBatch(tableName string, col schema.ColumnDefinition) FieldWithContext {
	objectID := GenerateObjectID(tableName)
	fieldID := GenerateFieldID(tableName, col.Name)

	fieldType := r.mapSQLTypeToLogical(col.Type)
	if col.LogicalType != "" {
		fieldType = col.LogicalType
	}

	isNameField := strings.EqualFold(col.Name, constants.FieldName) || col.IsNameField
	isSystem := r.IsSystemColumn(col.Name)
	required := !col.Nullable && !isSystem

	label := col.Label
	if label == "" {
		// Auto-humanize: replace underscores with spaces and Title Case
		// e.g. "created_by_id" -> "Created By Id"
		// We can also strip "_id" suffix if desired, but for now simple humanization is safer
		name := strings.ReplaceAll(col.Name, "_", " ")
		label = strings.Title(name) // Standard Go strings.Title (or use a better caser if available)
	}

	f := FieldWithContext{
		ObjectID: objectID,
		FieldID:  fieldID,
		Field: &models.FieldMetadata{
			APIName:     col.Name,
			Label:       label,
			Type:        models.FieldType(fieldType),
			Required:    required,
			Unique:      col.Unique,
			IsSystem:    isSystem,
			IsNameField: isNameField,
			ReferenceTo: WrapStringToSlice(col.ReferenceTo),
		},
	}

	if col.Default != "" {
		f.Field.DefaultValue = &col.Default
	}

	return f
}

// TableRegistryItem represents a table in the registry
type TableRegistryItem struct {
	ID               string
	TableName        string
	TableType        string
	Category         string
	Description      string
	IsManaged        bool
	SchemaVersion    string
	CreatedBy        string
	CreatedDate      string
	LastModifiedDate string
}

// GetTableRegistry retrieves all registered tables
func (r *SchemaRepository) GetTableRegistry() ([]*TableRegistryItem, error) {
	query := fmt.Sprintf(`
		SELECT %s, %s, %s, %s, %s, %s, 
		       %s, %s, %s, %s
		FROM %s
		ORDER BY %s, %s
	`, constants.FieldID, constants.FieldTableName, constants.FieldTableType, constants.FieldCategory, constants.FieldDescription, constants.FieldIsManaged,
		constants.FieldSchemaVersion, constants.FieldCreatedBy, constants.FieldCreatedDate, constants.FieldLastModifiedDate,
		constants.TableTable,
		constants.FieldTableType, constants.FieldTableName)

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []*TableRegistryItem
	for rows.Next() {
		var t TableRegistryItem
		var description sql.NullString
		var category sql.NullString
		var schemaVersion sql.NullString

		err := rows.Scan(
			&t.ID, &t.TableName, &t.TableType, &category, &description,
			&t.IsManaged, &schemaVersion, &t.CreatedBy, &t.CreatedDate, &t.LastModifiedDate,
		)
		if err != nil {
			return nil, err
		}

		if description.Valid {
			t.Description = description.String
		}
		if category.Valid {
			t.Category = category.String
		}
		if schemaVersion.Valid {
			t.SchemaVersion = schemaVersion.String
		}

		tables = append(tables, &t)
	}

	return tables, nil
}
