package fieldtypes

import (
	"embed"
	"encoding/json"
	"sync"
)

//go:embed fieldTypes.json
var fieldTypesFS embed.FS

// FieldTypeDefinition represents a field type configuration
type FieldTypeDefinition struct {
	SQLType           *string  `json:"sqlType"`
	Icon              string   `json:"icon"`
	Label             string   `json:"label"`
	Description       string   `json:"description"`
	IsSearchable      bool     `json:"isSearchable"`
	IsGroupable       bool     `json:"isGroupable"`
	IsSummable        bool     `json:"isSummable"`
	IsFK              bool     `json:"isFK,omitempty"`
	IsVirtual         bool     `json:"isVirtual,omitempty"`
	IsSystemOnly      bool     `json:"isSystemOnly,omitempty"`
	ValidationPattern *string  `json:"validationPattern,omitempty"`
	ValidationMessage *string  `json:"validationMessage,omitempty"`
	Operators         []string `json:"operators"`
}

// Registry holds field type definitions
type Registry struct {
	types map[string]FieldTypeDefinition
	mu    sync.RWMutex
}

var (
	defaultRegistry *Registry
	once            sync.Once
)

// GetRegistry returns the singleton field types registry
func GetRegistry() *Registry {
	once.Do(func() {
		defaultRegistry = &Registry{
			types: make(map[string]FieldTypeDefinition),
		}
		defaultRegistry.loadFromEmbedded()
	})
	return defaultRegistry
}

// loadFromEmbedded loads field types from the embedded JSON file
func (r *Registry) loadFromEmbedded() error {
	data, err := fieldTypesFS.ReadFile("fieldTypes.json")
	if err != nil {
		return err
	}

	var types map[string]FieldTypeDefinition
	if err := json.Unmarshal(data, &types); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.types = types
	return nil
}

// Get returns a field type definition by name
func (r *Registry) Get(typeName string) (FieldTypeDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	def, ok := r.types[typeName]
	return def, ok
}

// GetSQLType returns the SQL type for a field type name
func (r *Registry) GetSQLType(typeName string) string {
	def, ok := r.Get(typeName)
	if !ok || def.SQLType == nil {
		return ""
	}
	return *def.SQLType
}

// IsSearchable returns whether a field type is searchable
func (r *Registry) IsSearchable(typeName string) bool {
	def, ok := r.Get(typeName)
	if !ok {
		return false
	}
	return def.IsSearchable
}

// IsGroupable returns whether a field type can be grouped
func (r *Registry) IsGroupable(typeName string) bool {
	def, ok := r.Get(typeName)
	if !ok {
		return false
	}
	return def.IsGroupable
}

// IsSummable returns whether a field type can be aggregated
func (r *Registry) IsSummable(typeName string) bool {
	def, ok := r.Get(typeName)
	if !ok {
		return false
	}
	return def.IsSummable
}

// IsVirtual returns whether a field type is virtual (computed, not stored)
func (r *Registry) IsVirtual(typeName string) bool {
	def, ok := r.Get(typeName)
	if !ok {
		return false
	}
	return def.IsVirtual
}

// IsFK returns whether a field type is a foreign key reference
func (r *Registry) IsFK(typeName string) bool {
	def, ok := r.Get(typeName)
	if !ok {
		return false
	}
	return def.IsFK
}

// GetOperators returns the valid filter operators for a field type
func (r *Registry) GetOperators(typeName string) []string {
	def, ok := r.Get(typeName)
	if !ok {
		return nil
	}
	return def.Operators
}

// GetValidationPattern returns the validation regex pattern and message for a field type
func (r *Registry) GetValidationPattern(typeName string) (pattern string, message string) {
	def, ok := r.Get(typeName)
	if !ok {
		return "", ""
	}
	if def.ValidationPattern != nil {
		pattern = *def.ValidationPattern
	}
	if def.ValidationMessage != nil {
		message = *def.ValidationMessage
	}
	return pattern, message
}

// GetAll returns all registered field types
func (r *Registry) GetAll() map[string]FieldTypeDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]FieldTypeDefinition, len(r.types))
	for k, v := range r.types {
		result[k] = v
	}
	return result
}

// Package-level convenience functions using the default registry

// GetSQLType returns the SQL type for a field type name
func GetSQLType(typeName string) string {
	return GetRegistry().GetSQLType(typeName)
}

// IsSearchable returns whether a field type is searchable
func IsSearchable(typeName string) bool {
	return GetRegistry().IsSearchable(typeName)
}

// IsGroupable returns whether a field type can be grouped
func IsGroupable(typeName string) bool {
	return GetRegistry().IsGroupable(typeName)
}

// IsSummable returns whether a field type can be aggregated
func IsSummable(typeName string) bool {
	return GetRegistry().IsSummable(typeName)
}

// IsVirtual returns whether a field type is virtual (computed, not stored)
func IsVirtual(typeName string) bool {
	return GetRegistry().IsVirtual(typeName)
}

// IsFK returns whether a field type is a foreign key reference
func IsFK(typeName string) bool {
	return GetRegistry().IsFK(typeName)
}

// GetOperators returns the valid filter operators for a field type
func GetOperators(typeName string) []string {
	return GetRegistry().GetOperators(typeName)
}

// GetValidationPattern returns the validation regex pattern and message for a field type
func GetValidationPattern(typeName string) (pattern string, message string) {
	return GetRegistry().GetValidationPattern(typeName)
}

// FieldTypeWithName includes the name in the field type definition
type FieldTypeWithName struct {
	Name         string   `json:"name"`
	SQLType      string   `json:"sqlType"`
	Icon         string   `json:"icon"`
	Label        string   `json:"label"`
	Description  string   `json:"description"`
	IsSearchable bool     `json:"isSearchable"`
	IsGroupable  bool     `json:"isGroupable"`
	IsSummable   bool     `json:"isSummable"`
	IsFK         bool     `json:"isFK,omitempty"`
	IsVirtual    bool     `json:"isVirtual,omitempty"`
	IsSystemOnly bool     `json:"isSystemOnly,omitempty"`
	Operators    []string `json:"operators"`
}

// GetAllFieldTypes returns all built-in field types as a slice with names
func GetAllFieldTypes() []FieldTypeWithName {
	registry := GetRegistry()
	allTypes := registry.GetAll()
	result := make([]FieldTypeWithName, 0, len(allTypes))

	for name, def := range allTypes {
		sqlType := ""
		if def.SQLType != nil {
			sqlType = *def.SQLType
		}
		result = append(result, FieldTypeWithName{
			Name:         name,
			SQLType:      sqlType,
			Icon:         def.Icon,
			Label:        def.Label,
			Description:  def.Description,
			IsSearchable: def.IsSearchable,
			IsGroupable:  def.IsGroupable,
			IsSummable:   def.IsSummable,
			IsFK:         def.IsFK,
			IsVirtual:    def.IsVirtual,
			IsSystemOnly: def.IsSystemOnly,
			Operators:    def.Operators,
		})
	}

	return result
}
