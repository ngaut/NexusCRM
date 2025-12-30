package services

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/nexuscrm/shared/pkg/models"
	"github.com/nexuscrm/shared/pkg/constants"
)

// scanAction scans a row into an ActionMetadata struct
func (ms *MetadataService) scanAction(row Scannable) (*models.ActionMetadata, error) {
	var action models.ActionMetadata
	var targetObject, configJSON sql.NullString
	if err := row.Scan(&action.ID, &action.ObjectAPIName, &action.Name, &action.Label, &action.Type, &action.Icon, &targetObject, &configJSON); err != nil {
		return nil, err
	}
	action.TargetObject = ScanNullString(targetObject)
	UnmarshalJSONField(configJSON, &action.Config)
	return &action, nil
}

// queryActions queries actions for an object
func (ms *MetadataService) queryActions(objectAPIName string) ([]*models.ActionMetadata, error) {
	query := fmt.Sprintf("SELECT id, object_api_name, name, label, type, icon, target_object, config FROM %s WHERE object_api_name = ?", constants.TableAction)
	rows, err := ms.db.Query(query, objectAPIName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actions := make([]*models.ActionMetadata, 0)
	for rows.Next() {
		action, err := ms.scanAction(rows)
		if err != nil {
			continue
		}
		actions = append(actions, action)
	}
	return actions, nil
}

// queryAction queries a single action by ID
func (ms *MetadataService) queryAction(id string) (*models.ActionMetadata, error) {
	query := fmt.Sprintf("SELECT id, object_api_name, name, label, type, icon, target_object, config FROM %s WHERE id = ?", constants.TableAction)

	action, err := ms.scanAction(ms.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return action, nil
}

// scanValidationRule scans a row into a ValidationRule struct
func (ms *MetadataService) scanValidationRule(row Scannable) (*models.ValidationRule, error) {
	var rule models.ValidationRule
	var active int
	if err := row.Scan(&rule.ID, &rule.ObjectAPIName, &rule.Name, &active, &rule.Condition, &rule.ErrorMessage); err != nil {
		return nil, err
	}
	rule.Active = active != 0
	return &rule, nil
}

// queryValidationRules queries validation rules for an object
func (ms *MetadataService) queryValidationRules(objectAPIName string) ([]*models.ValidationRule, error) {
	query := fmt.Sprintf("SELECT id, object_api_name, name, active, condition, error_message FROM %s WHERE object_api_name = ?", constants.TableValidation)
	rows, err := ms.db.Query(query, objectAPIName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]*models.ValidationRule, 0)
	for rows.Next() {
		rule, err := ms.scanValidationRule(rows)
		if err != nil {
			continue
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// scanFlow scans a row into a Flow struct
func (ms *MetadataService) scanFlow(row Scannable) (*models.Flow, error) {
	var flow models.Flow
	var actionConfigJSON sql.NullString
	if err := row.Scan(&flow.ID, &flow.Name, &flow.TriggerObject, &flow.TriggerType, &flow.TriggerCondition, &flow.ActionType, &actionConfigJSON, &flow.Status, &flow.FlowType, &flow.LastModified); err != nil {
		return nil, err
	}
	UnmarshalJSONField(actionConfigJSON, &flow.ActionConfig)
	return &flow, nil
}

// queryFlows queries all flows
func (ms *MetadataService) queryFlows() ([]*models.Flow, error) {
	query := fmt.Sprintf("SELECT id, name, trigger_object, trigger_type, trigger_condition, action_type, action_config, status, flow_type, last_modified_date FROM %s", constants.TableFlow)
	rows, err := ms.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flows := make([]*models.Flow, 0)
	for rows.Next() {
		flow, err := ms.scanFlow(rows)
		if err != nil {
			log.Printf("⚠️ Failed to scan flow: %v\n", err) // Using log for visibility
			continue
		}
		flows = append(flows, flow)
	}
	return flows, nil
}

// queryFlow queries a single flow
func (ms *MetadataService) queryFlow(id string) (*models.Flow, error) {
	query := fmt.Sprintf("SELECT id, name, trigger_object, trigger_type, trigger_condition, action_type, action_config, status, flow_type, last_modified_date FROM %s WHERE id = ?", constants.TableFlow)

	flow, err := ms.scanFlow(ms.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return flow, nil
}

// querySharingRules queries sharing rules for an object
func (ms *MetadataService) querySharingRules(objectAPIName string) ([]*models.SharingRule, error) {
	query := fmt.Sprintf("SELECT id, object_api_name, name, criteria, access_level, share_with_role_id FROM %s WHERE object_api_name = ?", constants.TableSharingRule)
	rows, err := ms.db.Query(query, objectAPIName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]*models.SharingRule, 0)
	for rows.Next() {
		var rule models.SharingRule
		if err := rows.Scan(&rule.ID, &rule.ObjectAPIName, &rule.Name, &rule.Criteria, &rule.AccessLevel, &rule.ShareWithRoleID); err != nil {
			continue
		}
		rules = append(rules, &rule)
	}
	return rules, nil
}
