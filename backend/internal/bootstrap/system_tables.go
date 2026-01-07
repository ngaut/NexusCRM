package bootstrap

import (
	_ "embed"
	"encoding/json"
	"log"

	"github.com/nexuscrm/backend/internal/domain/schema"
)

//go:embed system_tables.json
var systemTablesJSON []byte

// GetSystemTableDefinitions returns definitions for all system tables
// Loaded from embedded JSON file for easy maintenance without code changes
func GetSystemTableDefinitions() []schema.TableDefinition {
	var definitions []schema.TableDefinition
	if err := json.Unmarshal(systemTablesJSON, &definitions); err != nil {
		log.Fatalf("Failed to parse system_tables.json: %v", err)
	}
	return definitions
}
