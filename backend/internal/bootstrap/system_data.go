package bootstrap

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/shared/pkg/models"
)

//go:embed system_data.json
var systemDataJSON []byte

type SystemData struct {
	Profiles []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"profiles"`
	Users []struct {
		ID        string `json:"id,omitempty"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		ProfileID string `json:"profile_id"`
	} `json:"users"`
}

// InitializeSystemData ensures required system data exists
// This should be called during server startup BEFORE accepting requests
func InitializeSystemData(sys *services.SystemManager) error {
	log.Println("🔧 Initializing system data...")

	var data SystemData
	if err := json.Unmarshal(systemDataJSON, &data); err != nil {
		return fmt.Errorf("failed to parse system_data.json: %w", err)
	}

	// 1. Process Profiles (Batch)
	profiles := make([]models.Profile, len(data.Profiles))
	for i, p := range data.Profiles {
		profiles[i] = models.Profile{
			ID:          p.ID,
			Name:        p.Name,
			Description: &p.Description,
			IsActive:    true,
			IsSystem:    true,
		}
	}
	if err := sys.BatchUpsertProfiles(profiles); err != nil {
		return fmt.Errorf("failed to batch upsert profiles: %w", err)
	}
	log.Printf("   ✅ Ensure %d system profiles (batch)", len(profiles))

	// 2. Process Users (Batch)
	users := make([]models.SystemUser, len(data.Users))
	for i, u := range data.Users {
		id := u.ID
		if id == "" {
			id = services.GenerateID()
		}

		// Split Name
		parts := strings.SplitN(u.Name, " ", 2)
		firstName := parts[0]
		lastName := ""
		if len(parts) > 1 {
			lastName = parts[1]
		}

		users[i] = models.SystemUser{
			ID:        id,
			Username:  u.Email,
			Email:     u.Email,
			Password:  u.Password,
			ProfileID: u.ProfileID,
			FirstName: firstName,
			LastName:  lastName,
			IsActive:  true,
		}
	}

	if err := sys.BatchUpsertUsers(users); err != nil {
		return fmt.Errorf("failed to batch upsert users: %w", err)
	}
	log.Printf("   ✅ Ensure %d system users (batch)", len(users))

	return nil
}

//go:embed apps.json
var appsJSON []byte

type AppsData struct {
	Apps []models.AppConfig `json:"apps"`
}

// InitializeApps ensures standard apps exist
func InitializeApps(sm *services.ServiceManager) error {
	log.Println("🔧 Initializing apps...")

	var data AppsData
	if err := json.Unmarshal(appsJSON, &data); err != nil {
		return fmt.Errorf("failed to parse apps.json: %w", err)
	}

	// Process Apps
	for _, app := range data.Apps {
		if err := sm.UIMetadata.CreateApp(&app); err != nil {
			if err.Error() != "app with ID '"+app.ID+"' already exists" {
				log.Printf("   ⚠️  Failed to ensure app %s: %v", app.ID, err)
			} else {
				log.Printf("   🔄 %s app already exists, updating configuration...", app.ID)
				if err := sm.UIMetadata.UpdateApp(app.ID, &app); err != nil {
					log.Printf("   ⚠️  Failed to update app %s: %v", app.ID, err)
				} else {
					log.Printf("   ✅ %s app updated", app.ID)
				}
			}
		} else {
			log.Printf("   ✅ %s app created", app.ID)
		}
	}

	return nil
}

//go:embed themes.json
var themesJSON []byte

type ThemesData struct {
	Themes []models.Theme `json:"themes"`
}

// InitializeThemes ensures standard themes exist
func InitializeThemes(sm *services.ServiceManager) error {
	log.Println("🔧 Initializing themes...")

	var data ThemesData
	if err := json.Unmarshal(themesJSON, &data); err != nil {
		return fmt.Errorf("failed to parse themes.json: %w", err)
	}

	for _, theme := range data.Themes {
		if err := sm.Metadata.UpsertTheme(&theme); err != nil {
			log.Printf("   ⚠️  Failed to ensure theme %s: %v", theme.Name, err)
		} else {
			log.Printf("   ✅ %s theme ensured", theme.Name)
		}
	}

	return nil
}

//go:embed ui_components.json
var uiComponentsJSON []byte

type UIComponentsData struct {
	Components []models.UIComponent `json:"components"`
}

// InitializeUIComponents ensures standard UI components exist (from ui_components.json)
func InitializeUIComponents(sm *services.ServiceManager) error {
	log.Println("🔧 Initializing UI components...")

	var data UIComponentsData
	if err := json.Unmarshal(uiComponentsJSON, &data); err != nil {
		return fmt.Errorf("failed to parse ui_components.json: %w", err)
	}

	for _, comp := range data.Components {
		// Clean description if nil
		if comp.Description != nil && *comp.Description == "" {
			comp.Description = nil
		}

		if err := sm.Metadata.UpsertUIComponent(&comp); err != nil {
			log.Printf("   ⚠️  Failed to ensure component %s: %v", comp.Name, err)
		}
	}
	log.Printf("   ✅ Ensure %d UI components", len(data.Components))

	return nil
}

// InitializeSetupPages loads setup page metadata from JSON
func InitializeSetupPages(sm *services.ServiceManager) error {
	log.Println("🔧 Initializing setup pages...")

	// Read setup_pages.json
	filePath := "internal/bootstrap/setup_pages.json"

	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("   ⚠️  Setup pages seed file not found at %s (skipping)", filePath)
		return nil
	}

	var pages []models.SetupPage
	if err := json.Unmarshal(content, &pages); err != nil {
		return fmt.Errorf("failed to parse setup_pages.json: %w", err)
	}

	for _, page := range pages {
		if err := sm.Metadata.UpsertSetupPage(&page); err != nil {
			log.Printf("   ⚠️  Failed to ensure setup page %s: %v", page.Label, err)
		}
	}
	log.Printf("   ✅ Ensure %d setup pages", len(pages))

	return nil
}
