package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// PermissionService handles permission checks for objects and fields.
//
// This service queries the database for every permission check (per-request model).
// This ensures permission changes are immediately effective without cache invalidation.
//
// Permission evaluation order:
//  1. SuperUser check (system_admin profile bypasses all checks)
//  2. Object-level permissions from _System_ObjectPerms
//  3. Field-level permissions from _System_FieldPerms
//  4. Record-level access (ownership, sharing rules) - future enhancement
type PermissionService struct {
	db       *database.TiDBConnection
	metadata *MetadataService
	formula  *formula.Engine

	// Role hierarchy cache: maps role_id -> parent_role_id
	roleHierarchyCache map[string]*string
	roleHierarchyMu    sync.RWMutex
}

// NewPermissionService creates a new PermissionService
func NewPermissionService(db *database.TiDBConnection, metadata *MetadataService) *PermissionService {
	ps := &PermissionService{
		db:                 db,
		metadata:           metadata,
		formula:            formula.NewEngine(),
		roleHierarchyCache: make(map[string]*string),
	}
	// Pre-load role hierarchy
	ps.refreshRoleHierarchy()
	return ps
}

// ==================== Object Permission Queries ====================

// loadObjectPermission queries the database for a specific object permission
func (ps *PermissionService) loadObjectPermission(profileID, objectAPIName string) (*models.ObjectPermission, error) {
	query := fmt.Sprintf(`
		SELECT profile_id, object_api_name, allow_read, allow_create, allow_edit, allow_delete, view_all, modify_all
		FROM %s
		WHERE profile_id = ? AND object_api_name = ?
		LIMIT 1
	`, constants.TableObjectPerms)

	var p models.ObjectPermission
	err := ps.db.QueryRow(query, profileID, strings.ToLower(objectAPIName)).Scan(
		&p.ProfileID, &p.ObjectAPIName,
		&p.AllowRead, &p.AllowCreate, &p.AllowEdit, &p.AllowDelete,
		&p.ViewAll, &p.ModifyAll,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No permission record = no access
	}
	if err != nil {
		log.Printf("⚠️ PermissionService: failed to load object permission for %s/%s: %v", profileID, objectAPIName, err)
		return nil, err
	}

	return &p, nil
}

// loadFieldPermission queries the database for a specific field permission
func (ps *PermissionService) loadFieldPermission(profileID, objectAPIName, fieldAPIName string) (*models.FieldPermission, error) {
	query := fmt.Sprintf(`
		SELECT profile_id, object_api_name, field_api_name, readable, editable
		FROM %s
		WHERE profile_id = ? AND object_api_name = ? AND field_api_name = ?
		LIMIT 1
	`, constants.TableFieldPerms)

	var p models.FieldPermission
	err := ps.db.QueryRow(query, profileID, strings.ToLower(objectAPIName), strings.ToLower(fieldAPIName)).Scan(
		&p.ProfileID, &p.ObjectAPIName, &p.FieldAPIName,
		&p.Readable, &p.Editable,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No field permission = use object-level defaults
	}
	if err != nil {
		log.Printf("⚠️ PermissionService: failed to load field permission for %s/%s.%s: %v", profileID, objectAPIName, fieldAPIName, err)
		return nil, err
	}

	return &p, nil
}

// loadEffectiveObjectPermission loads permissions considering Profile AND Permission Sets
func (ps *PermissionService) loadEffectiveObjectPermission(user *models.UserSession, objectAPIName string) (*models.ObjectPermission, error) {
	query := fmt.Sprintf(`
		SELECT 
			MAX(allow_read), MAX(allow_create), MAX(allow_edit), MAX(allow_delete), MAX(view_all), MAX(modify_all)
		FROM %s
		WHERE object_api_name = ?
		AND (
			profile_id = ? 
			OR 
			permission_set_id IN (SELECT permission_set_id FROM %s WHERE assignee_id = ?)
		)
	`, constants.TableObjectPerms, constants.TablePermissionSetAssignment)

	var p models.ObjectPermission
	p.ObjectAPIName = objectAPIName
	// ProfileID and PermissionSetID are irrelevant in aggregated view, kept nil

	// We scan into NullBool to detect if logic returned rows (though aggregation always returns 1 row, if no matches, values are NULL)
	var r, c, e, d, va, ma sql.NullBool

	err := ps.db.QueryRow(query, strings.ToLower(objectAPIName), user.ProfileID, user.ID).Scan(
		&r, &c, &e, &d, &va, &ma,
	)
	if err != nil {
		return nil, err
	}

	if !r.Valid {
		// All are NULL -> No permissions defined for this object for this user
		return nil, nil // No access
	}

	p.AllowRead = r.Bool
	p.AllowCreate = c.Bool
	p.AllowEdit = e.Bool
	p.AllowDelete = d.Bool
	p.ViewAll = va.Bool
	p.ModifyAll = ma.Bool

	return &p, nil
}

// loadEffectiveFieldPermission loads field permissions considering Profile AND Permission Sets
func (ps *PermissionService) loadEffectiveFieldPermission(user *models.UserSession, objectAPIName, fieldAPIName string) (*models.FieldPermission, error) {
	query := fmt.Sprintf(`
		SELECT MAX(readable), MAX(editable)
		FROM %s
		WHERE object_api_name = ? AND field_api_name = ?
		AND (
			profile_id = ? 
			OR 
			permission_set_id IN (SELECT permission_set_id FROM %s WHERE assignee_id = ?)
		)
	`, constants.TableFieldPerms, constants.TablePermissionSetAssignment)

	var r, e sql.NullBool
	err := ps.db.QueryRow(query, strings.ToLower(objectAPIName), strings.ToLower(fieldAPIName), user.ProfileID, user.ID).Scan(&r, &e)
	if err != nil {
		return nil, err
	}

	if !r.Valid {
		return nil, nil
	}

	return &models.FieldPermission{
		ObjectAPIName: objectAPIName,
		FieldAPIName:  fieldAPIName,
		Readable:      r.Bool,
		Editable:      e.Bool,
	}, nil
}

// ==================== Core Permission Checks ====================

// CheckObjectPermissionWithUser checks if a user has permission for an operation on an object.
// Operations: "create", "read", "edit", "delete"
// Returns true if the operation is permitted.
func (ps *PermissionService) CheckObjectPermissionWithUser(objectAPIName string, operation string, user *models.UserSession) bool {
	// No user = no access
	if user == nil {
		return false
	}

	// SuperUser bypass - system_admin or IsSystemAdmin flag has full access
	if user.IsSystemAdmin || constants.IsSuperUser(user.ProfileID) {
		return true
	}

	// Query database for permission
	// Query database for effective permission (Profile OR Permission Sets)
	perm, err := ps.loadEffectiveObjectPermission(user, objectAPIName)
	if err != nil {
		return false
	}

	// No permission record = no access
	if perm == nil {
		return false
	}

	// Check specific operation
	switch strings.ToLower(operation) {
	case constants.PermRead:
		return perm.AllowRead
	case constants.PermCreate:
		return perm.AllowCreate
	case constants.PermEdit:
		return perm.AllowEdit
	case constants.PermDelete:
		return perm.AllowDelete
	default:
		return false
	}
}

// CheckPermissionOrErrorWithUser checks permission and returns a specific PermissionError if false
func (ps *PermissionService) CheckPermissionOrErrorWithUser(objectAPIName string, operation string, user *models.UserSession) error {
	if !ps.CheckObjectPermissionWithUser(objectAPIName, operation, user) {
		return errors.NewPermissionError(operation, objectAPIName)
	}
	return nil
}

// Record-level access functions are in permission_record_access.go:
// - CheckRecordAccess, checkManualShareAccess
// - checkTeamMemberAccess, accessLevelAllowsOperation

// CheckFieldEditabilityWithUser checks if a field can be edited by the current user
func (ps *PermissionService) CheckFieldEditabilityWithUser(objectAPIName, fieldAPIName string, user *models.UserSession) bool {
	// System fields are never editable
	if isFieldSystemReadOnlyByName(fieldAPIName) {
		return false
	}

	// No user = no access
	if user == nil {
		return false
	}

	// SuperUser bypass
	if user.IsSystemAdmin || constants.IsSuperUser(user.ProfileID) {
		return true
	}

	// Check field-level permission (Effective)
	perm, err := ps.loadEffectiveFieldPermission(user, objectAPIName, fieldAPIName)
	if err != nil {
		return false
	}

	if perm != nil {
		return perm.Editable
	}

	// Fallback to object permission
	return ps.CheckObjectPermissionWithUser(objectAPIName, constants.PermEdit, user)
}

// CheckFieldVisibilityWithUser checks if a field is visible to the current user
func (ps *PermissionService) CheckFieldVisibilityWithUser(objectAPIName, fieldAPIName string, user *models.UserSession) bool {
	if user == nil {
		return false
	}

	// SuperUser bypass
	if user.IsSystemAdmin || constants.IsSuperUser(user.ProfileID) {
		return true
	}

	// Check field-level permission (Effective)
	perm, err := ps.loadEffectiveFieldPermission(user, objectAPIName, fieldAPIName)
	if err != nil {
		return false
	}

	// If explicit field permission exists, use it
	if perm != nil {
		return perm.Readable
	}

	// Fallback to object permission
	return ps.CheckObjectPermissionWithUser(objectAPIName, constants.PermRead, user)
}

// RefreshPermissions reloads permissions from the database
// This specifically refreshes the Role Hierarchy cache.
// Object/Field permissions are not cached (checked per-request), so they don't need refreshing.
func (ps *PermissionService) RefreshPermissions() error {
	ps.refreshRoleHierarchy()
	return nil
}

// GetEffectiveSchema returns the schema with field-level visibility applied
func (ps *PermissionService) GetEffectiveSchema(schema *models.ObjectMetadata, user *models.UserSession) *models.ObjectMetadata {
	if schema == nil {
		return nil
	}

	// Create a copy of the schema to avoid mutating the original (which might be cached)
	effectiveSchema := *schema
	effectiveSchema.Fields = make([]models.FieldMetadata, 0, len(schema.Fields))

	// Super users see all fields
	if user != nil && (user.IsSystemAdmin || constants.IsSuperUser(user.ProfileID)) {
		return schema
	}

	for _, field := range schema.Fields {
		if ps.CheckFieldVisibilityWithUser(schema.APIName, field.APIName, user) {
			effectiveSchema.Fields = append(effectiveSchema.Fields, field)
		}
	}

	return &effectiveSchema
}

// GetObjectPermissions retrieves all object permissions for a profile
func (ps *PermissionService) GetObjectPermissions(profileID string) ([]models.ObjectPermission, error) {
	query := fmt.Sprintf(`
		SELECT profile_id, object_api_name, allow_read, allow_create, allow_edit, allow_delete, view_all, modify_all
		FROM %s
		WHERE profile_id = ?
	`, constants.TableObjectPerms)

	rows, err := ps.db.Query(query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.ObjectPermission
	for rows.Next() {
		var p models.ObjectPermission
		if err := rows.Scan(&p.ProfileID, &p.ObjectAPIName, &p.AllowRead, &p.AllowCreate, &p.AllowEdit, &p.AllowDelete, &p.ViewAll, &p.ModifyAll); err != nil {
			continue
		}
		perms = append(perms, p)
	}

	return perms, nil
}

// UpdateObjectPermission creates or updates an object permission
func (ps *PermissionService) UpdateObjectPermission(perm models.ObjectPermission) error {
	return ps.updateObjectPermission(ps.db, perm)
}

// UpdateObjectPermissionTx creates or updates an object permission within a transaction
func (ps *PermissionService) UpdateObjectPermissionTx(tx *sql.Tx, perm models.ObjectPermission) error {
	return ps.updateObjectPermission(tx, perm)
}

// updateObjectPermission creates or updates an object permission using the provided executor
func (ps *PermissionService) updateObjectPermission(exec Executor, perm models.ObjectPermission) error {
	// Generate ID if likely new (won't hurt ON DUPLICATE constraint)
	id := GenerateID()

	query := fmt.Sprintf(`
		INSERT INTO %s (id, profile_id, permission_set_id, object_api_name, allow_read, allow_create, allow_edit, allow_delete, view_all, modify_all)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			allow_read = VALUES(allow_read),
			allow_create = VALUES(allow_create),
			allow_edit = VALUES(allow_edit),
			allow_delete = VALUES(allow_delete),
			view_all = VALUES(view_all),
			modify_all = VALUES(modify_all)
	`, constants.TableObjectPerms)

	_, err := exec.Exec(query, id, perm.ProfileID, perm.PermissionSetID, perm.ObjectAPIName, perm.AllowRead, perm.AllowCreate, perm.AllowEdit, perm.AllowDelete, perm.ViewAll, perm.ModifyAll)
	if err == nil {
		target := "Unknown"
		if perm.ProfileID != nil {
			target = *perm.ProfileID
		}
		if perm.PermissionSetID != nil {
			target = "PermSet:" + *perm.PermissionSetID
		}
		log.Printf("🛡️ Updated object permission for %s on %s", target, perm.ObjectAPIName)
	}
	return err
}

// GetFieldPermissions retrieves all field permissions for a profile
func (ps *PermissionService) GetFieldPermissions(profileID string) ([]models.FieldPermission, error) {
	query := fmt.Sprintf(`
		SELECT profile_id, object_api_name, field_api_name, readable, editable
		FROM %s
		WHERE profile_id = ?
	`, constants.TableFieldPerms)

	rows, err := ps.db.Query(query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.FieldPermission
	for rows.Next() {
		var p models.FieldPermission
		if err := rows.Scan(&p.ProfileID, &p.ObjectAPIName, &p.FieldAPIName, &p.Readable, &p.Editable); err != nil {
			continue
		}
		perms = append(perms, p)
	}

	return perms, nil
}

// UpdateFieldPermission creates or updates a field permission
func (ps *PermissionService) UpdateFieldPermission(perm models.FieldPermission) error {
	id := GenerateID()
	query := fmt.Sprintf(`
		INSERT INTO %s (id, profile_id, permission_set_id, object_api_name, field_api_name, readable, editable)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			readable = VALUES(readable),
			editable = VALUES(editable)
	`, constants.TableFieldPerms)

	_, err := ps.db.Exec(query, id, perm.ProfileID, perm.PermissionSetID, perm.ObjectAPIName, perm.FieldAPIName, perm.Readable, perm.Editable)
	if err == nil {
		target := "Unknown"
		if perm.ProfileID != nil {
			target = *perm.ProfileID
		}
		if perm.PermissionSetID != nil {
			target = "PermSet:" + *perm.PermissionSetID
		}
		log.Printf("🛡️ Updated field permission for %s on %s.%s", target, perm.ObjectAPIName, perm.FieldAPIName)
	}
	return err
}

// isFieldSystemReadOnlyByName checks if a field is a system read-only field based on its name
func isFieldSystemReadOnlyByName(fieldAPIName string) bool {
	return constants.IsSystemField(fieldAPIName)
}

// isFieldSystemReadOnly checks if a field is system read-only by looking up metadata
// This version takes metadata service and object name for richer checking
func isFieldSystemReadOnly(metadata *MetadataService, objectAPIName string, fieldAPIName string) bool {
	// First check if it's a common system field
	if isFieldSystemReadOnlyByName(fieldAPIName) {
		return true
	}

	// If we have metadata, check the field's is_system flag
	if metadata != nil {
		schema := metadata.GetSchema(objectAPIName)
		if schema != nil {
			for _, field := range schema.Fields {
				if field.APIName == fieldAPIName {
					return field.IsSystem
				}
			}
		}
	}

	return false
}

// GrantInitialPermissions grants default permissions for a new object to all profiles
func (ps *PermissionService) GrantInitialPermissions(objectAPIName string) error {
	return GrantInitialObjectPermissions(ps.db, objectAPIName, constants.TableProfile, constants.TableObjectPerms, constants.ProfileSystemAdmin)
}

// Role hierarchy functions are in permission_role_hierarchy.go:
// - refreshRoleHierarchy, getRoleAncestors, isUserAboveInHierarchy, getRecordOwnerRoleID, RefreshRoleHierarchy

// Sharing rule functions are in permission_sharing_rules.go:
// - isUserInRoleOrBelow, checkSharingRuleAccess, evaluateSharingCriteria

// Permission Set functions are in permission_perm_sets.go:
// - GetPermissionSetObjectPermissions, GetPermissionSetFieldPermissions
// - UpdatePermissionSetObjectPermission, UpdatePermissionSetFieldPermission
// - GetEffectiveObjectPermissions, GetEffectiveFieldPermissions
// - CreatePermissionSet, UpdatePermissionSet, DeletePermissionSet
