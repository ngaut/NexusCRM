package bootstrap

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/shared/pkg/models"
)

//go:embed standard_actions.json
var standardActionsJSON []byte

type StandardActionDef struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Type  string `json:"type"`
	Icon  string `json:"icon"`
}

// InitializeStandardActions ensures standard actions exist for core objects
func InitializeStandardActions(metadataService *services.MetadataService) error {
	log.Println("🔧 Initializing standard actions...")

	var actions []StandardActionDef
	if err := json.Unmarshal(standardActionsJSON, &actions); err != nil {
		return fmt.Errorf("failed to parse standard_actions.json: %w", err)
	}

	// Step 2: Ensure 'Edit' and 'Delete' exist for ALL objects in metadata
	schemas := metadataService.GetSchemas()
	log.Printf("   ...Ensuring %d standard actions for %d objects...", len(actions), len(schemas))

	for _, schema := range schemas {
		for _, action := range actions {
			newAction := &models.ActionMetadata{
				ObjectAPIName: schema.APIName,
				Name:          action.Name,
				Label:         action.Label,
				Type:          action.Type,
				Icon:          action.Icon,
			}

			if err := metadataService.CreateAction(newAction); err != nil {
				// Only log if it's not a duplicate/exists error
				errStr := err.Error()
				if !strings.Contains(errStr, "already exists") {
					log.Printf("   ⚠️  Failed to create action %s.%s: %v", schema.APIName, action.Name, err)
				}
			}
		}
	}

	return nil
}
