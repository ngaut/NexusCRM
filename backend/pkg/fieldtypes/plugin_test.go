package fieldtypes

import (
	"testing"
)

func TestPluginRegistry_Lifecycle(t *testing.T) {
	registry := GetPluginRegistry()
	mockName := "LifecyclePlugin"

	// 1. Register
	mock := MockPlugin{
		BasePlugin: NewBasePlugin(mockName, "Label", "Desc", "Icon", "VARCHAR", nil),
	}
	if err := registry.Register(mock); err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}
	defer registry.Unregister(mockName)

	// 2. Get
	p, ok := registry.Get(mockName)
	if !ok {
		t.Errorf("Failed to retrieve registered plugin")
	}
	if p.Name() != mockName {
		t.Errorf("Expected name %s, got %s", mockName, p.Name())
	}

	// 3. List
	plugins := registry.List()
	found := false
	for _, name := range plugins {
		if name == mockName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("List() did not contain registered plugin")
	}
}

// Mock plugin for testing custom registration
type MockPlugin struct {
	BasePlugin
}

func TestCustomRegistration(t *testing.T) {
	registry := GetPluginRegistry()
	mockName := "TestPlugin"

	// Ensure cleanup
	defer registry.Unregister(mockName)

	mock := MockPlugin{
		BasePlugin: NewBasePlugin(mockName, "Test Label", "Desc", "Icon", "VARCHAR", nil),
	}

	if err := registry.Register(mock); err != nil {
		t.Fatalf("Failed to register custom plugin: %v", err)
	}

	if _, ok := registry.Get(mockName); !ok {
		t.Errorf("Failed to retrieve custom registered plugin")
	}
}
