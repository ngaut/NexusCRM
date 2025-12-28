package bootstrap

import (
	_ "embed"
	"encoding/json"
	"log"

	"github.com/nexuscrm/backend/internal/domain/schema"
)

//go:embed system_tables.json
var systemTablesJSON []byte

// TableDefinitionJSON matches the JSON structure for unmarshaling
type TableDefinitionJSON struct {
	TableName   string           `json:"tableName"`
	TableType   string           `json:"tableType"`
	Category    string           `json:"category"`
	Description string           `json:"description"`
	Columns     []ColumnJSON     `json:"columns"`
	Indices     []IndexJSON      `json:"indices,omitempty"`
	ForeignKeys []ForeignKeyJSON `json:"foreignKeys,omitempty"`
}

type ColumnJSON struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	PrimaryKey    bool   `json:"primaryKey,omitempty"`
	Nullable      bool   `json:"nullable,omitempty"`
	Unique        bool   `json:"unique,omitempty"`
	Default       string `json:"default,omitempty"`
	AutoIncrement bool   `json:"autoIncrement,omitempty"`
	LogicalType   string `json:"logicalType,omitempty"`
	ReferenceTo   string `json:"referenceTo,omitempty"`
	IsNameField   bool   `json:"isNameField,omitempty"`
}

type IndexJSON struct {
	Name    string   `json:"name,omitempty"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique,omitempty"`
}

type ForeignKeyJSON struct {
	Column     string `json:"column"`
	References string `json:"references"`
	OnDelete   string `json:"onDelete,omitempty"`
	OnUpdate   string `json:"onUpdate,omitempty"`
}

// GetSystemTableDefinitions returns definitions for all system tables
// Loaded from embedded JSON file for easy maintenance without code changes
func GetSystemTableDefinitions() []schema.TableDefinition {
	var jsonDefs []TableDefinitionJSON
	if err := json.Unmarshal(systemTablesJSON, &jsonDefs); err != nil {
		log.Fatalf("Failed to parse system_tables.json: %v", err)
	}

	definitions := make([]schema.TableDefinition, 0, len(jsonDefs))
	for _, jd := range jsonDefs {
		def := schema.TableDefinition{
			TableName:   jd.TableName,
			TableType:   jd.TableType,
			Category:    jd.Category,
			Description: jd.Description,
		}

		// Convert columns
		for _, jc := range jd.Columns {
			col := schema.ColumnDefinition{
				Name:          jc.Name,
				Type:          jc.Type,
				PrimaryKey:    jc.PrimaryKey,
				Nullable:      jc.Nullable,
				Unique:        jc.Unique,
				Default:       jc.Default,
				AutoIncrement: jc.AutoIncrement,
				LogicalType:   jc.LogicalType,
				ReferenceTo:   jc.ReferenceTo,
				IsNameField:   jc.IsNameField,
			}
			def.Columns = append(def.Columns, col)
		}

		// Convert indices
		for _, ji := range jd.Indices {
			idx := schema.IndexDefinition{
				Name:    ji.Name,
				Columns: ji.Columns,
				Unique:  ji.Unique,
			}
			def.Indices = append(def.Indices, idx)
		}

		// Convert foreign keys
		for _, jfk := range jd.ForeignKeys {
			fk := schema.ForeignKeyDefinition{
				Column:     jfk.Column,
				References: jfk.References,
				OnDelete:   jfk.OnDelete,
				OnUpdate:   jfk.OnUpdate,
			}
			def.ForeignKeys = append(def.ForeignKeys, fk)
		}

		definitions = append(definitions, def)
	}

	return definitions
}
