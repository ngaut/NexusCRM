package rest

import (
	"github.com/nexuscrm/backend/pkg/fieldtypes"
)

// FieldTypeInfo represents field type information for API response
type FieldTypeInfo struct {
	Name         string   `json:"name"`
	Label        string   `json:"label"`
	Description  string   `json:"description"`
	Icon         string   `json:"icon"`
	SQLType      string   `json:"sqlType"`
	IsSearchable bool     `json:"isSearchable"`
	IsGroupable  bool     `json:"isGroupable"`
	IsSummable   bool     `json:"isSummable"`
	IsVirtual    bool     `json:"isVirtual"`
	Operators    []string `json:"operators"`
	IsPlugin     bool     `json:"isPlugin"`
}

// GetAllFieldTypes returns all available field types including plugins
func GetAllFieldTypes() []FieldTypeInfo {
	result := make([]FieldTypeInfo, 0)

	// Get built-in field types from JSON registry
	builtinTypes := fieldtypes.GetAllFieldTypes()
	for _, ft := range builtinTypes {
		result = append(result, FieldTypeInfo{
			Name:         ft.Name,
			Label:        ft.Label,
			Description:  ft.Description,
			Icon:         ft.Icon,
			SQLType:      ft.SQLType,
			IsSearchable: ft.IsSearchable,
			IsGroupable:  ft.IsGroupable,
			IsSummable:   ft.IsSummable,
			IsVirtual:    ft.IsVirtual,
			Operators:    ft.Operators,
			IsPlugin:     false,
		})
	}

	// Get plugin field types
	plugins := fieldtypes.GetPluginRegistry().GetAll()
	for _, plugin := range plugins {
		result = append(result, FieldTypeInfo{
			Name:         plugin.Name(),
			Label:        plugin.Label(),
			Description:  plugin.Description(),
			Icon:         plugin.Icon(),
			SQLType:      plugin.SQLType(),
			IsSearchable: plugin.IsSearchable(),
			IsGroupable:  plugin.IsGroupable(),
			IsSummable:   plugin.IsSummable(),
			IsVirtual:    plugin.IsVirtual(),
			Operators:    plugin.DefaultOperators(),
			IsPlugin:     true,
		})
	}

	return result
}
