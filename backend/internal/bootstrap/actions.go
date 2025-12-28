package bootstrap

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/internal/domain/models"
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

	coreObjects := []string{"account", "contact", "opportunity", "lead", "project", "task"}

	for _, objName := range coreObjects {
		for _, action := range actions {
			newAction := &models.ActionMetadata{
				ObjectAPIName: objName,
				Name:          action.Name,
				Label:         action.Label,
				Type:          action.Type,
				Icon:          action.Icon,
			}

			if err := metadataService.CreateAction(newAction); err != nil {
				// Only log if it's not a duplicate/exists error
				errStr := err.Error()
				if !strings.Contains(errStr, "already exists") {
					log.Printf("   ⚠️  Failed to create action %s.%s: %v", objName, action.Name, err)
				}
			} else {
				log.Printf("   ✅ Action created: %s.%s", objName, action.Name)
			}
		}
	}

	return nil
}
