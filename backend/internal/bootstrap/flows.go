package bootstrap

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/internal/domain/models"
)

//go:embed flows.json
var flowsJSON []byte

// InitializeFlows ensures standard flows exist
func InitializeFlows(metadata *services.MetadataService) error {
	log.Println("🔧 Initializing flows...")

	var flows []models.Flow
	if err := json.Unmarshal(flowsJSON, &flows); err != nil {
		return fmt.Errorf("failed to parse flows.json: %w", err)
	}

	for _, flow := range flows {
		// Check if flow exists
		existing := metadata.GetFlow(flow.ID)
		if existing != nil {
			// Update existing flow
			log.Printf("   🔄 Flow %s already exists, updating...", flow.Name)
			// Apply updates (status, logic, etc.)
			// Enforce system flow definition.
			if err := metadata.UpdateFlow(flow.ID, &flow); err != nil {
				log.Printf("   ⚠️  Failed to update flow %s: %v", flow.Name, err)
			}
		} else {
			// Create new flow
			// Ensure ID consistency from JSON definition.
			// CreateFlow should respect the provided ID to allow idempotent updates.
			if err := metadata.CreateFlow(&flow); err != nil {
				log.Printf("   ⚠️  Failed to create flow %s: %v", flow.Name, err)
			} else {
				log.Printf("   ✅ Flow %s created", flow.Name)
			}
		}
	}

	return nil
}
