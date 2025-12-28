package services

import (
	"fmt"
)

// ConfigHelper provides shared utilities for extracting configuration values
// from generic map[string]interface{} maps, common in Action and Flow configs.

// GetConfigString safely extracts a string value from a config map.
// It returns an empty string if the key does not exist or the value is not a string.
func GetConfigString(config map[string]interface{}, key string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return ""
}

// GetConfigStringRequired extracts a string value and returns an error if missing or empty.
func GetConfigStringRequired(config map[string]interface{}, key string) (string, error) {
	val := GetConfigString(config, key)
	if val == "" {
		return "", fmt.Errorf("missing required config key: %s", key)
	}
	return val, nil
}

// GetConfigMap extracts a nested map[string]interface{} from a config map.
func GetConfigMap(config map[string]interface{}, key string) (map[string]interface{}, bool) {
	if val, ok := config[key].(map[string]interface{}); ok {
		return val, true
	}
	return nil, false
}
