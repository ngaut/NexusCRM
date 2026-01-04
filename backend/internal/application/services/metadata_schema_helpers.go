package services

import (
	"fmt"
	"regexp"

	domainSchema "github.com/nexuscrm/backend/internal/domain/schema"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== DRY Helpers ====================

// PrepareTableDefinition consolidates validation, enrichment, and mapping logic from ObjectMetadata to TableDefinition
// It returns the Physical TableDefinition and a slice of FieldWithContext for batch processing
func (ms *MetadataService) PrepareTableDefinition(schema *models.ObjectMetadata) (domainSchema.TableDefinition, []FieldWithContext, error) {
	// Validate API Name
	validName := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if !validName.MatchString(schema.APIName) {
		return domainSchema.TableDefinition{}, nil, fmt.Errorf("invalid API name '%s': must start with letter or underscore and contain only alphanumeric characters", schema.APIName)
	}

	// Determine defaults
	if schema.Description == nil || *schema.Description == "" {
		desc := schema.Label
		schema.Description = &desc
	}
	// In pure meta-driven architecture, all non-system objects use custom_object table type
	tableType := constants.TableTypeCustomObject
	if schema.ID == "" {
		schema.ID = GenerateObjectID(schema.APIName)
	}

	// Ensure System Fields
	standardFields := ms.schemaMgr.GetStandardFieldMetadata()
	for i, f := range standardFields {
		if f.IsNameField {
			// Custom objects often want "{Object} Name" instead of just "Name".
			standardFields[i].Label = schema.Label + " Name"
		}
	}
	EnrichWithSystemFields(schema, standardFields)

	// Build Table Definition
	def := domainSchema.TableDefinition{
		TableName:   schema.APIName,
		TableType:   string(tableType),
		Category:    "standard",
		Description: *schema.Description,
		Columns:     make([]domainSchema.ColumnDefinition, 0),
	}

	var batchFields []FieldWithContext

	// Map Fields to Columns
	for _, field := range schema.Fields {
		// Column Def
		colDef := domainSchema.ColumnDefinition{
			Name:        field.APIName,
			Type:        ms.schemaMgr.MapFieldTypeToSQL(string(field.Type)),
			LogicalType: string(field.Type),
			Nullable:    !field.Required, // If required, then NOT NULL
			Unique:      field.Unique,
		}

		// Special handling for ID -> Make it Primary Key
		if field.APIName == constants.FieldID {
			colDef.PrimaryKey = true
			colDef.Type = "VARCHAR(36)"
		}

		if field.DefaultValue != nil {
			colDef.Default = "'" + *field.DefaultValue + "'"
			if field.Type == constants.FieldTypeBoolean {
				if *field.DefaultValue == "false" || *field.DefaultValue == "0" {
					colDef.Default = "0"
				} else if *field.DefaultValue == "true" || *field.DefaultValue == "1" {
					colDef.Default = "1"
				}
			}
		}
		if len(field.ReferenceTo) == 1 {
			colDef.ReferenceTo = field.ReferenceTo[0] // Single reference -> add FK
		}
		colDef.AllReferences = field.ReferenceTo
		def.Columns = append(def.Columns, colDef)

		// If Polymorphic, add a "Type" column to store the object type of the reference
		if field.IsPolymorphic {
			typeColDef := domainSchema.ColumnDefinition{
				Name:        GetPolymorphicTypeColumnName(field.APIName),
				Type:        "VARCHAR(100)",
				LogicalType: string(constants.FieldTypeText),
				Nullable:    true, // Can be null if lookup is null
			}
			def.Columns = append(def.Columns, typeColDef)
		}

		// Field Metadata Context
		f := field // copy
		batchFields = append(batchFields, FieldWithContext{
			ObjectID: schema.ID,
			FieldID:  GenerateFieldID(schema.APIName, f.APIName),
			Field:    &f,
		})
	}

	return def, batchFields, nil
}

// GenerateDefaultLayout creates a default page layout for a schema
func (ms *MetadataService) GenerateDefaultLayout(schema *models.ObjectMetadata) models.PageLayout {
	layoutID := GenerateID()
	defaultLayout := models.PageLayout{
		ID:            layoutID,
		ObjectAPIName: schema.APIName,
		LayoutName:    "Default Layout",
		Type:          "Detail",
		CompactLayout: []string{},
		Sections: []models.PageSection{
			{
				ID:      GenerateID(),
				Label:   "Information",
				Columns: 2,
				Fields:  []string{},
			},
			{
				ID:      GenerateID(),
				Label:   "System Information",
				Columns: 2,
				Fields:  []string{constants.FieldCreatedByID, constants.FieldCreatedDate, constants.FieldLastModifiedByID, constants.FieldLastModifiedDate, constants.FieldOwnerID},
			},
		},
		RelatedLists:  []models.RelatedListConfig{},
		HeaderActions: []models.ActionConfig{},
		QuickActions:  []models.ActionConfig{},
	}

	for _, f := range schema.Fields {
		if !f.IsSystem && f.APIName != constants.FieldID {
			defaultLayout.Sections[0].Fields = append(defaultLayout.Sections[0].Fields, f.APIName)
		}
	}
	return defaultLayout
}
