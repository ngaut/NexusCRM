package persistence

import (
	"strings"

	"github.com/nexuscrm/backend/internal/domain/schema"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== System Column Definitions ====================

// GetStandardSystemColumns returns the default columns for every custom object
// This serves as the Single Source of Truth for system field definitions.
func (r *SchemaRepository) GetStandardSystemColumns() []schema.ColumnDefinition {
	return []schema.ColumnDefinition{
		{
			Name:       constants.FieldID,
			Type:       "VARCHAR(36)",
			PrimaryKey: true,
			Nullable:   false,
		},
		{
			Name:     constants.FieldName,
			Type:     "VARCHAR(255)",
			Nullable: false,
		},
		{
			Name:     constants.FieldOwnerID,
			Type:     "VARCHAR(36)",
			Nullable: true,
		},
		{
			Name:     constants.FieldCreatedByID,
			Type:     "VARCHAR(36)",
			Nullable: true,
		},
		{
			Name:     constants.FieldLastModifiedByID,
			Type:     "VARCHAR(36)",
			Nullable: true,
		},
		{
			Name: constants.FieldCreatedDate,
			Type: "DATETIME",
			// Default removed to avoid 'Invalid default value' error in strict mode.
			// PersistenceService populates this.
		},
		{
			Name: constants.FieldLastModifiedDate,
			Type: "DATETIME",
			// PersistenceService populates this.
		},
		{
			Name:     constants.FieldIsDeleted,
			Type:     "TINYINT(1)",
			Default:  "0",
			Nullable: false,
		},
	}
}

// GetStandardFieldMetadata returns field metadata for standard system columns
// This is derived from GetStandardSystemColumns() to avoid duplication of column defs,
// but adds logical metadata properties like Labels and Types.
func (r *SchemaRepository) GetStandardFieldMetadata() []models.FieldMetadata {
	columns := r.GetStandardSystemColumns()
	fields := make([]models.FieldMetadata, 0, len(columns))

	for _, col := range columns {
		field := models.FieldMetadata{
			APIName:  col.Name,
			Label:    col.Name,
			Type:     models.FieldType(r.mapSQLTypeToLogical(col.Type)),
			Required: !col.Nullable && !r.IsSystemColumn(col.Name),
			IsUnique: col.Unique,
			IsSystem: r.IsSystemColumn(col.Name),
		}

		// Set special flags and refining Types
		if strings.EqualFold(col.Name, constants.FieldName) {
			field.IsNameField = true
			field.Required = true
			field.Label = "Name"
		} else if strings.EqualFold(col.Name, constants.FieldID) {
			field.Label = "ID"
		} else if strings.EqualFold(col.Name, constants.FieldCreatedDate) {
			field.Label = "Created Date"
		} else if strings.EqualFold(col.Name, constants.FieldLastModifiedDate) {
			field.Label = "Last Modified Date"
		} else if strings.EqualFold(col.Name, constants.FieldOwnerID) {
			field.Label = "Owner"
			field.Type = constants.FieldTypeLookup
			field.ReferenceTo = []string{constants.TableUser}
		} else if strings.EqualFold(col.Name, constants.FieldCreatedByID) {
			field.Label = "Created By"
			field.Type = constants.FieldTypeLookup
			field.ReferenceTo = []string{constants.TableUser}
		} else if strings.EqualFold(col.Name, constants.FieldLastModifiedByID) {
			field.Label = "Last Modified By"
			field.Type = constants.FieldTypeLookup
			field.ReferenceTo = []string{constants.TableUser}
		}

		fields = append(fields, field)
	}

	return fields
}
