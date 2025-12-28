package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nexuscrm/backend/internal/domain/events"
	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/pkg/constants"
)

// GetAllObjects returns all object definitions visible to the user
func (ms *MetadataService) GetAllObjects(ctx context.Context, user *models.UserSession) ([]models.ObjectMetadata, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// 1. Get all object API names
	rows, err := ms.db.Query(fmt.Sprintf("SELECT api_name FROM %s", constants.TableObject))
	if err != nil {
		return nil, fmt.Errorf("failed to query objects: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var apiNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		apiNames = append(apiNames, name)
	}

	// 2. Load schema for each (in parallel or loop)
	// Optimization: This could be heavy. For now, loop.
	var objects []models.ObjectMetadata
	for _, name := range apiNames {
		// Use existing load logic which includes fields
		obj, err := ms.querySchemaByAPIName(name)
		if err != nil {
			log.Printf("Warning: Failed to load schema for %s: %v", name, err)
			continue
		}

		// 3. Filter by permissions (Phase 2)
		// if !ms.permissionService.CanReadObject(user, name) { continue }

		objects = append(objects, *obj)
	}

	return objects, nil
}

// GetSystemFields returns all system fields for an object by querying metadata
// System fields are those with IsSystem=true OR IsNameField=true
func (ms *MetadataService) GetSystemFields(objectAPIName string) []string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Query schema from DB
	obj, err := ms.querySchemaByAPIName(objectAPIName)
	if err != nil || obj == nil {
		return []string{}
	}

	systemFields := make([]string, 0)
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

func (ms *MetadataService) GetFlows() []*models.Flow {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	flows, err := ms.queryFlows()
	if err != nil {
		log.Printf("Failed to get flows: %v", err)
		return []*models.Flow{}
	}
	return flows
}

func (ms *MetadataService) GetValidationRules(objectAPIName string) []*models.ValidationRule {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	rules, err := ms.queryValidationRules(objectAPIName)
	if err != nil {
		log.Printf("Failed to get validation rules: %v", err)
		return []*models.ValidationRule{}
	}
	return rules
}

func (ms *MetadataService) GetListViews(objectAPIName string) []*models.ListView {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	views, err := ms.queryListViews(objectAPIName)
	if err != nil {
		log.Printf("Failed to get list views: %v", err)
		return []*models.ListView{}
	}
	return views
}

// CreateListView creates a new list view
func (ms *MetadataService) CreateListView(view *models.ListView) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if view.ID == "" {
		view.ID = GenerateID()
	}

	filtersJSON, _ := json.Marshal(view.Filters)
	fieldsJSON, _ := json.Marshal(view.Fields)

	_, err := ms.db.Exec(
		fmt.Sprintf("INSERT INTO %s (id, object_api_name, label, filters, fields) VALUES (?, ?, ?, ?, ?)", constants.TableListView),
		view.ID, view.ObjectAPIName, view.Label, string(filtersJSON), string(fieldsJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to create list view: %w", err)
	}
	return nil
}

// UpdateListView updates an existing list view
func (ms *MetadataService) UpdateListView(id string, updates *models.ListView) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	filtersJSON, _ := json.Marshal(updates.Filters)
	fieldsJSON, _ := json.Marshal(updates.Fields)

	result, err := ms.db.Exec(
		fmt.Sprintf("UPDATE %s SET label = ?, filters = ?, fields = ? WHERE id = ?", constants.TableListView),
		updates.Label, string(filtersJSON), string(fieldsJSON), id,
	)
	if err != nil {
		return fmt.Errorf("failed to update list view: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("list view not found: %s", id)
	}
	return nil
}

// DeleteListView deletes a list view
func (ms *MetadataService) DeleteListView(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	result, err := ms.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", constants.TableListView), id)
	if err != nil {
		return fmt.Errorf("failed to delete list view: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("list view not found: %s", id)
	}
	return nil
}

func (ms *MetadataService) GetSharingRules(objectAPIName string) []*models.SharingRule {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	rules, err := ms.querySharingRules(objectAPIName)
	if err != nil {
		log.Printf("Failed to get sharing rules: %v", err)
		return []*models.SharingRule{}
	}
	return rules
}

func (ms *MetadataService) GetActions(objectAPIName string) []*models.ActionMetadata {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	actions, err := ms.queryActions(objectAPIName)
	if err != nil {
		log.Printf("Failed to get actions: %v", err)
		return []*models.ActionMetadata{}
	}
	return actions
}

// GetAllActions returns all actions from all objects
func (ms *MetadataService) GetAllActions() []*models.ActionMetadata {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	// No helper for all actions yet, standard query
	// Implementing query here similar to loadActions but returning slice
	rows, err := ms.db.Query(fmt.Sprintf("SELECT id, object_api_name, name, label, type, icon, target_object, config FROM %s", constants.TableAction))
	if err != nil {
		return []*models.ActionMetadata{}
	}
	defer func() { _ = rows.Close() }()

	var actions []*models.ActionMetadata
	for rows.Next() {
		var action models.ActionMetadata
		var targetObject, configJSON sql.NullString
		if err := rows.Scan(&action.ID, &action.ObjectAPIName, &action.Name, &action.Label, &action.Type, &action.Icon, &targetObject, &configJSON); err != nil {
			continue
		}
		if targetObject.Valid {
			action.TargetObject = &targetObject.String
		}
		if configJSON.Valid && configJSON.String != "" {
			var config map[string]interface{}
			if err := json.Unmarshal([]byte(configJSON.String), &config); err == nil {
				action.Config = config
			}
		}
		actions = append(actions, &action)
	}
	return actions
}

func (ms *MetadataService) GetActionByID(actionID string) *models.ActionMetadata {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	action, err := ms.queryAction(actionID)
	if err != nil {
		return nil
	}
	return action
}

func (ms *MetadataService) GetRecordTypes(objectAPIName string) []*models.RecordType {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	rts, err := ms.queryRecordTypes(objectAPIName)
	if err != nil {
		return []*models.RecordType{}
	}
	return rts
}

func (ms *MetadataService) GetAutoNumbers(objectAPIName string) []*models.AutoNumber {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ans, err := ms.queryAutoNumbers(objectAPIName)
	if err != nil {
		return []*models.AutoNumber{}
	}
	return ans
}

func (ms *MetadataService) GetRelationships(objectAPIName string) []*models.Relationship {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	rels, err := ms.queryRelationships(objectAPIName)
	if err != nil {
		return []*models.Relationship{}
	}
	return rels
}

func (ms *MetadataService) GetFieldDependencies(objectAPIName string) []*models.FieldDependency {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	fds, err := ms.queryFieldDependencies(objectAPIName)
	if err != nil {
		return []*models.FieldDependency{}
	}
	return fds
}

func (ms *MetadataService) GetChildRelationships(parentObjectAPIName string) []*models.ObjectMetadata {
	// Query fields that reference this object
	query := fmt.Sprintf(`
		SELECT o.api_name 
		FROM %s f
		JOIN %s o ON f.object_id = o.id
		WHERE f.reference_to = ? AND f.type = 'Lookup'
	`, constants.TableField, constants.TableObject)

	rows, err := ms.db.Query(query, parentObjectAPIName)
	if err != nil {
		log.Printf("⚠️ Failed to query child relationships for %s: %v", parentObjectAPIName, err)
		return []*models.ObjectMetadata{}
	}
	defer rows.Close()

	var children []*models.ObjectMetadata
	for rows.Next() {
		var apiName string
		if err := rows.Scan(&apiName); err != nil {
			continue
		}

		// Load full schema for child
		if schema, err := ms.querySchemaByAPIName(apiName); err == nil && schema != nil {
			children = append(children, schema)
		}
	}

	return children
}
