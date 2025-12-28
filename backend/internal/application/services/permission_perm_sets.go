package services

import (
	"database/sql"
	"fmt"

	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/pkg/constants"
)

// ==================== Permission Set Permissions ====================

// GetPermissionSetObjectPermissions retrieves all object permissions for a permission set
func (ps *PermissionService) GetPermissionSetObjectPermissions(permissionSetID string) ([]models.ObjectPermission, error) {
	query := fmt.Sprintf(`
		SELECT permission_set_id, object_api_name, allow_read, allow_create, allow_edit, allow_delete, view_all, modify_all
		FROM %s
		WHERE permission_set_id = ?
	`, constants.TableObjectPerms)

	rows, err := ps.db.Query(query, permissionSetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.ObjectPermission
	for rows.Next() {
		var p models.ObjectPermission
		if err := rows.Scan(&p.PermissionSetID, &p.ObjectAPIName, &p.AllowRead, &p.AllowCreate, &p.AllowEdit, &p.AllowDelete, &p.ViewAll, &p.ModifyAll); err != nil {
			continue
		}
		perms = append(perms, p)
	}

	return perms, nil
}

// GetPermissionSetFieldPermissions retrieves all field permissions for a permission set
func (ps *PermissionService) GetPermissionSetFieldPermissions(permissionSetID string) ([]models.FieldPermission, error) {
	query := fmt.Sprintf(`
		SELECT permission_set_id, object_api_name, field_api_name, readable, editable
		FROM %s
		WHERE permission_set_id = ?
	`, constants.TableFieldPerms)

	rows, err := ps.db.Query(query, permissionSetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.FieldPermission
	for rows.Next() {
		var p models.FieldPermission
		if err := rows.Scan(&p.PermissionSetID, &p.ObjectAPIName, &p.FieldAPIName, &p.Readable, &p.Editable); err != nil {
			continue
		}
		perms = append(perms, p)
	}

	return perms, nil
}

// UpdatePermissionSetObjectPermission creates or updates an object permission for a permission set
func (ps *PermissionService) UpdatePermissionSetObjectPermission(perm models.ObjectPermission) error {
	if perm.PermissionSetID == nil || *perm.PermissionSetID == "" {
		return fmt.Errorf("permission_set_id is required")
	}
	return ps.updateObjectPermission(ps.db, perm)
}

// UpdatePermissionSetFieldPermission creates or updates a field permission for a permission set
func (ps *PermissionService) UpdatePermissionSetFieldPermission(perm models.FieldPermission) error {
	if perm.PermissionSetID == nil || *perm.PermissionSetID == "" {
		return fmt.Errorf("permission_set_id is required")
	}
	return ps.UpdateFieldPermission(perm)
}

// ==================== Effective Permissions (Admin View) ====================

// GetEffectiveObjectPermissions returns the merged object permissions for a specific user
func (ps *PermissionService) GetEffectiveObjectPermissions(userID string) ([]models.ObjectPermission, error) {
	// 1. Get user's profile ID
	queryProfile := fmt.Sprintf("SELECT profile_id FROM %s WHERE id = ?", constants.TableUser)
	var profileID string
	err := ps.db.QueryRow(queryProfile, userID).Scan(&profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// 2. Query aggregated permissions
	query := fmt.Sprintf(`
		SELECT 
			object_api_name,
			MAX(allow_read), MAX(allow_create), MAX(allow_edit), MAX(allow_delete), MAX(view_all), MAX(modify_all)
		FROM %s
		WHERE 
			profile_id = ? 
			OR 
			permission_set_id IN (SELECT permission_set_id FROM %s WHERE assignee_id = ?)
		GROUP BY object_api_name
	`, constants.TableObjectPerms, constants.TablePermissionSetAssignment)

	rows, err := ps.db.Query(query, profileID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.ObjectPermission
	for rows.Next() {
		var p models.ObjectPermission
		var r, c, e, d, va, ma sql.NullBool
		if err := rows.Scan(&p.ObjectAPIName, &r, &c, &e, &d, &va, &ma); err != nil {
			continue
		}
		p.AllowRead = r.Valid && r.Bool
		p.AllowCreate = c.Valid && c.Bool
		p.AllowEdit = e.Valid && e.Bool
		p.AllowDelete = d.Valid && d.Bool
		p.ViewAll = va.Valid && va.Bool
		p.ModifyAll = ma.Valid && ma.Bool
		perms = append(perms, p)
	}

	return perms, nil
}

// GetEffectiveFieldPermissions returns the merged field permissions for a specific user
func (ps *PermissionService) GetEffectiveFieldPermissions(userID string) ([]models.FieldPermission, error) {
	// 1. Get user's profile ID
	queryProfile := fmt.Sprintf("SELECT profile_id FROM %s WHERE id = ?", constants.TableUser)
	var profileID string
	err := ps.db.QueryRow(queryProfile, userID).Scan(&profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// 2. Query aggregated field permissions
	query := fmt.Sprintf(`
		SELECT 
			object_api_name, field_api_name,
			MAX(readable), MAX(editable)
		FROM %s
		WHERE 
			profile_id = ? 
			OR 
			permission_set_id IN (SELECT permission_set_id FROM %s WHERE assignee_id = ?)
		GROUP BY object_api_name, field_api_name
	`, constants.TableFieldPerms, constants.TablePermissionSetAssignment)

	rows, err := ps.db.Query(query, profileID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.FieldPermission
	for rows.Next() {
		var p models.FieldPermission
		var r, e sql.NullBool
		if err := rows.Scan(&p.ObjectAPIName, &p.FieldAPIName, &r, &e); err != nil {
			continue
		}
		p.Readable = r.Valid && r.Bool
		p.Editable = e.Valid && e.Bool
		perms = append(perms, p)
	}

	return perms, nil
}

// ==================== Permission Set Management ====================

// CreatePermissionSet creates a new permission set
func (ps *PermissionService) CreatePermissionSet(name, label, description string) (string, error) {
	id := GenerateID()
	query := fmt.Sprintf(`
		INSERT INTO %s (id, name, label, description, is_active, created_date)
		VALUES (?, ?, ?, ?, true, NOW())
	`, constants.TablePermissionSet)

	_, err := ps.db.Exec(query, id, name, label, description)
	if err != nil {
		return "", err
	}
	return id, nil
}

// UpdatePermissionSet updates a permission set
func (ps *PermissionService) UpdatePermissionSet(id string, name, label, description string, isActive bool) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET name = ?, label = ?, description = ?, is_active = ?
		WHERE id = ?
	`, constants.TablePermissionSet)

	_, err := ps.db.Exec(query, name, label, description, isActive, id)
	return err
}

// DeletePermissionSet deletes a permission set
func (ps *PermissionService) DeletePermissionSet(id string) error {
	// First delete assignments
	_, err := ps.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE permission_set_id = ?", constants.TablePermissionSetAssignment), id)
	if err != nil {
		return err
	}
	// Delete permissions
	_, err = ps.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE permission_set_id = ?", constants.TableObjectPerms), id)
	if err != nil {
		return err
	}
	_, err = ps.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE permission_set_id = ?", constants.TableFieldPerms), id)
	if err != nil {
		return err
	}
	// Delete the set itself
	_, err = ps.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", constants.TablePermissionSet), id)
	return err
}
