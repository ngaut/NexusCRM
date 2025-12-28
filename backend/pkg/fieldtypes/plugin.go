package fieldtypes

import (
	"fmt"
	"sync"
)

// FieldTypePlugin defines the interface for custom field type plugins
// Plugins can be registered to extend the system with new field types
type FieldTypePlugin interface {
	// Name returns the unique identifier for this field type
	Name() string

	// Label returns the human-readable label for this field type
	Label() string

	// Description returns a description of what this field type is for
	Description() string

	// Icon returns the icon name (lucide-react icon) for UI display
	Icon() string

	// SQLType returns the SQL column type for this field
	SQLType() string

	// IsSearchable returns whether this field type supports full-text search
	IsSearchable() bool

	// IsGroupable returns whether this field type can be used in GROUP BY
	IsGroupable() bool

	// IsSummable returns whether this field type supports aggregation (SUM, AVG, etc.)
	IsSummable() bool

	// IsVirtual returns whether this field type has no physical column
	IsVirtual() bool

	// DefaultOperators returns the default filter operators for this field type
	DefaultOperators() []string

	// Validate validates a value for this field type
	// Returns nil if valid, error otherwise
	Validate(value interface{}, config map[string]interface{}) error

	// Transform transforms a value before storage (optional)
	// Returns the transformed value
	Transform(value interface{}, config map[string]interface{}) (interface{}, error)

	// Format formats a value for display (optional)
	// Returns the formatted string
	Format(value interface{}, config map[string]interface{}) string
}

// BasePlugin provides default implementations for optional plugin methods
type BasePlugin struct {
	name        string
	label       string
	description string
	icon        string
	sqlType     string
	operators   []string
}

// NewBasePlugin creates a new base plugin with required fields
func NewBasePlugin(name, label, description, icon, sqlType string, operators []string) BasePlugin {
	return BasePlugin{
		name:        name,
		label:       label,
		description: description,
		icon:        icon,
		sqlType:     sqlType,
		operators:   operators,
	}
}

func (p BasePlugin) Name() string               { return p.name }
func (p BasePlugin) Label() string              { return p.label }
func (p BasePlugin) Description() string        { return p.description }
func (p BasePlugin) Icon() string               { return p.icon }
func (p BasePlugin) SQLType() string            { return p.sqlType }
func (p BasePlugin) IsSearchable() bool         { return false }
func (p BasePlugin) IsGroupable() bool          { return false }
func (p BasePlugin) IsSummable() bool           { return false }
func (p BasePlugin) IsVirtual() bool            { return false }
func (p BasePlugin) DefaultOperators() []string { return p.operators }

func (p BasePlugin) Validate(value interface{}, config map[string]interface{}) error {
	return nil // Default: no validation
}

func (p BasePlugin) Transform(value interface{}, config map[string]interface{}) (interface{}, error) {
	return value, nil // Default: no transformation
}

func (p BasePlugin) Format(value interface{}, config map[string]interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// PluginRegistry manages registered field type plugins
type PluginRegistry struct {
	plugins map[string]FieldTypePlugin
	mu      sync.RWMutex
}

var (
	pluginRegistry     *PluginRegistry
	pluginRegistryOnce sync.Once
)

// GetPluginRegistry returns the singleton plugin registry
func GetPluginRegistry() *PluginRegistry {
	pluginRegistryOnce.Do(func() {
		pluginRegistry = &PluginRegistry{
			plugins: make(map[string]FieldTypePlugin),
		}
	})
	return pluginRegistry
}

// Register adds a plugin to the registry
func (r *PluginRegistry) Register(plugin FieldTypePlugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := plugin.Name()
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("field type plugin '%s' is already registered", name)
	}

	r.plugins[name] = plugin
	return nil
}

// Get retrieves a plugin by name
func (r *PluginRegistry) Get(name string) (FieldTypePlugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	plugin, ok := r.plugins[name]
	return plugin, ok
}

// List returns all registered plugin names
func (r *PluginRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

// GetAll returns all registered plugins as a map
func (r *PluginRegistry) GetAll() map[string]FieldTypePlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]FieldTypePlugin, len(r.plugins))
	for name, plugin := range r.plugins {
		result[name] = plugin
	}
	return result
}

// Unregister removes a plugin from the registry
func (r *PluginRegistry) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; exists {
		delete(r.plugins, name)
		return true
	}
	return false
}

// Package-level convenience functions

// RegisterPlugin registers a field type plugin
func RegisterPlugin(plugin FieldTypePlugin) error {
	return GetPluginRegistry().Register(plugin)
}

// GetPlugin retrieves a field type plugin by name
func GetPlugin(name string) (FieldTypePlugin, bool) {
	return GetPluginRegistry().Get(name)
}

// ListPlugins returns all registered plugin names
func ListPlugins() []string {
	return GetPluginRegistry().List()
}
