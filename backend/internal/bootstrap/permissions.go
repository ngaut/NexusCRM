package bootstrap

import (
	_ "embed"
	"log"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

type ObjectPermission struct {
	APIName   string `json:"object_api_name"`
	Read      bool   `json:"allow_read"`
	Create    bool   `json:"allow_create"`
	Edit      bool   `json:"allow_edit"`
	Delete    bool   `json:"allow_delete"`
	ViewAll   bool   `json:"view_all"`
	ModifyAll bool   `json:"modify_all"`
}

type ProfilePermissions struct {
	ProfileID string             `json:"profile_id"`
	Objects   []ObjectPermission `json:"objects"`
}

type PermissionsData struct {
	ObjectPermissions []ProfilePermissions `json:"object_permissions"`
}

// InitializePermissions seeds default object permissions for all profiles
// It first seeds from permissions.json, then ensures ALL objects in metadata have permissions
func InitializePermissions(permSvc *services.PermissionService, metadata *services.MetadataService) error {
	log.Println("🔧 Initializing permissions...")

	// Step 1: Seed from JSON file (explicit overrides) removed - files missing
	seededObjects := make(map[string]map[string]bool) // profile -> objects already seeded

	// Step 2: Ensure ALL objects in metadata have permissions for system_admin and standard_user
	profiles := []string{constants.ProfileSystemAdmin, constants.ProfileStandardUser}
	schemas := metadata.GetSchemas()

	log.Printf("   ...Batching permission updates for %d objects across %d profiles...", len(schemas), len(profiles))

	for _, profileID := range profiles {
		if seededObjects[profileID] == nil {
			seededObjects[profileID] = make(map[string]bool)
		}

		for i, schema := range schemas {
			// Skip if already seeded from JSON
			if seededObjects[profileID][schema.APIName] {
				continue
			}

			// DEBUG LOG every 10 items
			if i%10 == 0 {
				log.Printf("   ...Processing permission for %s/%s (%d/%d)...", profileID, schema.APIName, i+1, len(schemas))
			}

			// Determine default permissions based on profile and object type
			isSystemTable := constants.IsSystemTable(schema.APIName)
			isSystemAdmin := profileID == constants.ProfileSystemAdmin

			var read, create, edit, del, viewAll, modifyAll bool
			if isSystemAdmin {
				// System admin gets full access to everything
				read, create, edit, del, viewAll, modifyAll = true, true, true, true, true, true
			} else {
				// Standard user: read access to system tables, full CRUD (except delete) for business objects
				if isSystemTable {
					read, create, edit, del, viewAll, modifyAll = true, false, false, false, false, false
				} else {
					read, create, edit, del, viewAll, modifyAll = true, true, true, false, false, false
				}
			}

			perm := models.ObjectPermission{
				ProfileID:     &profileID,
				ObjectAPIName: schema.APIName,
				AllowRead:     read,
				AllowCreate:   create,
				AllowEdit:     edit,
				AllowDelete:   del,
				ViewAll:       viewAll,
				ModifyAll:     modifyAll,
			}

			// We use UpdateObjectPermission instead of batch insert for consistency and DRY
			// Performance impact is negligible for bootstrap
			if err := permSvc.UpdateObjectPermission(perm); err != nil {
				log.Printf("   ⚠️  Failed to auto-seed permissions for %s/%s: %v", profileID, schema.APIName, err)
			}

			seededObjects[profileID][schema.APIName] = true
		}
	}

	log.Printf("   ✅ Permissions initialized for %d objects across %d profiles", len(schemas), len(profiles))
	return nil
}
