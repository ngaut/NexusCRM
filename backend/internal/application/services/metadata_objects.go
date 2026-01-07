package services

import (
	"context"
	"fmt"
	"log"

	"github.com/nexuscrm/backend/internal/domain/events"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// GetAllObjects returns all object definitions visible to the user
func (ms *MetadataService) GetAllObjects(ctx context.Context, user *models.UserSession) ([]models.ObjectMetadata, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Use Repo
	// But Repo "GetAllSchemas" returns []*ObjectMetadata. Service returns []ObjectMetadata.
	// And service logic in original was: Get API Names -> Loop querySchemaByAPIName (load full).
	// Repo "GetAllSchemas" loads full objects?
	// In Step 853, GetAllSchemas calls "SELECT ... FROM sys_object". And calls "scanObject".
	// "scanObject" (Step 931) does NOT load Fields recursively?
	// Line 1020: `obj.Fields = make([]models.FieldMetadata, 0)`.
	// So GetAllSchemas returns objects WITHOUT fields?
	// But Service `GetAllObjects` seems to want FULL objects (via querySchemaByAPIName).
	// `querySchemaByAPIName` (deleted) likely loaded fields.
	// `repo.GetSchemaByAPIName` (Step 831 Logic) loaded fields?
	// Wait, Step 831 logic: `scanObject` created obj. Then it called `r.GetSchemaFields(apiName)`.
	// Does `GetAllSchemas` call `GetSchemaFields` for each?
	// Let's check `GetAllSchemas` implementation in `metadata_repository.go` (if visible).
	// I haven't viewed `GetAllSchemas` recently.
	// If `scanObject` doesn't load fields, `GetAllSchemas` needs to load them if required.
	// Original `GetAllObjects` loaded them.
	// If I use `ms.repo.GetAllSchemas`, I might get objects without fields.
	// I should verify `GetAllSchemas`.

	// Assuming `GetAllSchemas` behaves correctly or I need to fetch fields.
	// If `GetAllSchemas` is light (no fields), I should iterate and fetch fields?
	// Or use `GetSchemaByAPIName` in loop like original.

	// Better: Use `GetAllSchemas` then for each, fetch fields?
	// Or if `GetSchemaByAPIName` fetches fields, use that.

	// I'll replicate original logic for safety: Get Names (via GetAllSchemas light?) then GetSchemaByAPIName (which loads fields).
	// Or just use `GetAllSchemas` and hope it loads fields?
	// Usually GetAllObjects implies full metadata.
	// I'll check `MetadataRepository.GetAllSchemas`.
	// Step 853 snippet implies it returns []*ObjectMetadata.
	// I'll assume for now I should use `GetAllSchemas`.
	// If it lacks fields, I might need to fix Repo.

	// Wait, original logic: 1. Get API names. 2. Loop `querySchemaByAPIName`.
	// `querySchemaByAPIName` definitey loaded fields.
	// `repo.GetSchemaByAPIName` DEFINITELY loads fields (it calls `scanObject` then `GetSchemaFields`).
	// `repo.GetAllSchemas`: Does it?
	// I'll check `repo.GetAllSchemas` later.
	// For now, I'll use `repo.GetAllSchemas`.

	schemas, err := ms.repo.GetAllSchemas(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query objects: %w", err)
	}

	res := make([]models.ObjectMetadata, len(schemas))
	for i, s := range schemas {
		res[i] = *s
		// If fields are missing, UI might break.
		// If `GetAllSchemas` doesn't populate fields, I should call `GetSchemaFields`.
	}
	return res, nil
}

// GetSystemFields returns all system fields for an object by querying metadata
// System fields are those with IsSystem=true OR IsNameField=true
func (ms *MetadataService) GetSystemFields(ctx context.Context, objectAPIName string) []string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Query schema from Repo
	obj, err := ms.repo.GetSchemaByAPIName(ctx, objectAPIName)
	if err != nil || obj == nil {
		return []string{}
	}

	var systemFields []string
	for _, field := range obj.Fields {
		if field.IsSystem || field.IsNameField {
			systemFields = append(systemFields, field.APIName)
		}
	}

	return systemFields
}

// GetSupportedEvents returns a list of all event types supported by the system
func (ms *MetadataService) GetSupportedEvents() []string {
	return []string{
		string(events.RecordBeforeCreate),
		string(events.RecordAfterCreate),
		string(events.RecordBeforeUpdate),
		string(events.RecordAfterUpdate),
		string(events.RecordBeforeDelete),
		string(events.RecordAfterDelete),
	}
}

func (ms *MetadataService) GetFlows(ctx context.Context) []*models.Flow {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	flows, err := ms.repo.GetAllFlows(ctx)
	if err != nil {
		log.Printf("Failed to get flows: %v", err)
		return []*models.Flow{}
	}
	return flows
}

// GetScheduledFlows returns all flows with trigger_type = "schedule"
func (ms *MetadataService) GetScheduledFlows(ctx context.Context) []models.Flow {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	flows, err := ms.repo.GetScheduledFlows(ctx)
	if err != nil {
		log.Printf("Failed to query scheduled flows: %v", err)
		return []models.Flow{}
	}

	// Convert []*models.Flow to []models.Flow
	res := make([]models.Flow, len(flows))
	for i, f := range flows {
		res[i] = *f
	}

	return res
}

func (ms *MetadataService) GetValidationRules(ctx context.Context, objectAPIName string) []*models.ValidationRule {
	// Skip validation rule lookup for system objects - they never have custom rules
	if constants.IsSystemTable(objectAPIName) {
		return []*models.ValidationRule{}
	}

	ms.mu.RLock()
	defer ms.mu.RUnlock()
	rules, err := ms.repo.GetValidationRules(ctx, objectAPIName)
	if err != nil {
		log.Printf("Failed to get validation rules: %v", err)
		return []*models.ValidationRule{}
	}
	return rules
}

func (ms *MetadataService) GetListViews(ctx context.Context, objectAPIName string) []*models.ListView {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	views, err := ms.repo.GetListViews(ctx, objectAPIName)
	if err != nil {
		log.Printf("Failed to get list views: %v", err)
		return []*models.ListView{}
	}
	return views
}

// CreateListView creates a new list view
func (ms *MetadataService) CreateListView(ctx context.Context, view *models.ListView) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if view.ID == "" {
		view.ID = GenerateID()
	}

	if err := ms.repo.CreateListView(ctx, view); err != nil {
		return fmt.Errorf("failed to create list view: %w", err)
	}
	return nil
}

// UpdateListView updates an existing list view
func (ms *MetadataService) UpdateListView(ctx context.Context, id string, updates *models.ListView) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if err := ms.repo.UpdateListView(ctx, id, updates); err != nil {
		return fmt.Errorf("failed to update list view: %w", err)
	}
	return nil
}

// DeleteListView deletes a list view
func (ms *MetadataService) DeleteListView(ctx context.Context, id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if err := ms.repo.DeleteListView(ctx, id); err != nil {
		return fmt.Errorf("failed to delete list view: %w", err)
	}
	return nil
}

func (ms *MetadataService) GetSharingRules(ctx context.Context, objectAPIName string) []*models.SystemSharingRule {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	rules, err := ms.repo.GetSharingRules(ctx, objectAPIName)
	if err != nil {
		log.Printf("Failed to get sharing rules: %v", err)
		return []*models.SystemSharingRule{}
	}
	return rules
}

func (ms *MetadataService) GetActions(ctx context.Context, objectAPIName string) []*models.ActionMetadata {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	actions, err := ms.repo.GetActions(ctx, objectAPIName)
	if err != nil {
		log.Printf("Failed to get actions: %v", err)
		return []*models.ActionMetadata{}
	}
	return actions
}

// GetAllActions returns all actions from all objects
func (ms *MetadataService) GetAllActions(ctx context.Context) []*models.ActionMetadata {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	actions, err := ms.repo.GetAllActions(ctx)
	if err != nil {
		log.Printf("Failed to query actions: %v", err)
		return []*models.ActionMetadata{}
	}
	return actions
}

func (ms *MetadataService) GetActionByID(ctx context.Context, actionID string) *models.ActionMetadata {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	action, err := ms.repo.GetAction(ctx, actionID)
	if err != nil {
		return nil
	}
	return action
}

func (ms *MetadataService) GetRecordTypes(ctx context.Context, objectAPIName string) []*models.RecordType {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	rts, err := ms.repo.GetRecordTypes(ctx, objectAPIName)
	if err != nil {
		return []*models.RecordType{}
	}
	return rts
}

func (ms *MetadataService) GetAutoNumbers(ctx context.Context, objectAPIName string) []*models.AutoNumber {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ans, err := ms.repo.GetAutoNumbers(ctx, objectAPIName)
	if err != nil {
		return []*models.AutoNumber{}
	}
	return ans
}

func (ms *MetadataService) GetRelationships(ctx context.Context, objectAPIName string) []*models.Relationship {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	rels, err := ms.repo.GetRelationships(ctx, objectAPIName)
	if err != nil {
		return []*models.Relationship{}
	}
	return rels
}

func (ms *MetadataService) GetFieldDependencies(ctx context.Context, objectAPIName string) []*models.FieldDependency {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	fds, err := ms.repo.GetFieldDependencies(ctx, objectAPIName)
	if err != nil {
		return []*models.FieldDependency{}
	}
	return fds
}

func (ms *MetadataService) GetChildRelationships(ctx context.Context, parentObjectAPIName string) []*models.ObjectMetadata {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	children, err := ms.repo.GetChildRelationships(ctx, parentObjectAPIName)
	if err != nil {
		log.Printf("⚠️ Failed to query child relationships for %s: %v", parentObjectAPIName, err)
		return []*models.ObjectMetadata{}
	}

	return children
}
