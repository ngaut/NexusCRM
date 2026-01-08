package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/nexuscrm/backend/pkg/utils"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// PermissionRepository handles database operations for permissions
type PermissionRepository struct {
	db *sql.DB
}

// NewPermissionRepository creates a new PermissionRepository
func NewPermissionRepository(db *sql.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// LoadObjectPermission queries the database for a specific object permission
func (r *PermissionRepository) LoadObjectPermission(ctx context.Context, profileID, objectAPIName string) (*models.SystemObjectPerms, error) {
	query := fmt.Sprintf(`
		SELECT profile_id, object_api_name, allow_read, allow_create, allow_edit, allow_delete, view_all, modify_all
		FROM %s
		WHERE profile_id = ? AND object_api_name = ?
		LIMIT 1
	`, constants.TableObjectPerms)

	var p models.SystemObjectPerms
	err := r.db.QueryRowContext(ctx, query, profileID, strings.ToLower(objectAPIName)).Scan(
		&p.ProfileID, &p.ObjectAPIName,
		&p.AllowRead, &p.AllowCreate, &p.AllowEdit, &p.AllowDelete,
		&p.ViewAll, &p.ModifyAll,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No permission record = no access
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load object permission: %w", err)
	}

	return &p, nil
}

// LoadFieldPermission queries the database for a specific field permission
func (r *PermissionRepository) LoadFieldPermission(ctx context.Context, profileID, objectAPIName, fieldAPIName string) (*models.SystemFieldPerms, error) {
	query := fmt.Sprintf(`
		SELECT profile_id, object_api_name, field_api_name, readable, editable
		FROM %s
		WHERE profile_id = ? AND object_api_name = ? AND field_api_name = ?
		LIMIT 1
	`, constants.TableFieldPerms)

	var p models.SystemFieldPerms
	err := r.db.QueryRowContext(ctx, query, profileID, strings.ToLower(objectAPIName), strings.ToLower(fieldAPIName)).Scan(
		&p.ProfileID, &p.ObjectAPIName, &p.FieldAPIName,
		&p.Readable, &p.Editable,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No field permission = use object-level defaults
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load field permission: %w", err)
	}

	return &p, nil
}

// LoadEffectiveObjectPermission loads permissions considering Profile AND Permission Sets
func (r *PermissionRepository) LoadEffectiveObjectPermission(ctx context.Context, user *models.UserSession, objectAPIName string) (*models.SystemObjectPerms, error) {
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

	var p models.SystemObjectPerms
	p.ObjectAPIName = objectAPIName

	var rr, c, e, d, va, ma sql.NullBool

	err := r.db.QueryRowContext(ctx, query, strings.ToLower(objectAPIName), user.ProfileID, user.ID).Scan(
		&rr, &c, &e, &d, &va, &ma,
	)
	if err != nil {
		return nil, err
	}

	if !rr.Valid {
		return nil, nil // No access
	}

	p.AllowRead = rr.Bool
	p.AllowCreate = c.Bool
	p.AllowEdit = e.Bool
	p.AllowDelete = d.Bool
	p.ViewAll = va.Bool
	p.ModifyAll = ma.Bool

	return &p, nil
}

// LoadEffectiveFieldPermission loads field permissions considering Profile AND Permission Sets
func (r *PermissionRepository) LoadEffectiveFieldPermission(ctx context.Context, user *models.UserSession, objectAPIName, fieldAPIName string) (*models.SystemFieldPerms, error) {
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

	var readable, editable sql.NullBool
	err := r.db.QueryRowContext(ctx, query, strings.ToLower(objectAPIName), strings.ToLower(fieldAPIName), user.ProfileID, user.ID).Scan(&readable, &editable)
	if err != nil {
		return nil, err
	}

	if !readable.Valid {
		return nil, nil
	}

	return &models.SystemFieldPerms{
		ObjectAPIName: objectAPIName,
		FieldAPIName:  fieldAPIName,
		Readable:      readable.Bool,
		Editable:      editable.Bool,
	}, nil
}

// ListObjectPermissions retrieves all object permissions for a profile
func (r *PermissionRepository) ListObjectPermissions(ctx context.Context, profileID string) ([]models.SystemObjectPerms, error) {
	query := fmt.Sprintf(`
		SELECT profile_id, object_api_name, allow_read, allow_create, allow_edit, allow_delete, view_all, modify_all
		FROM %s
		WHERE profile_id = ?
	`, constants.TableObjectPerms)

	rows, err := r.db.QueryContext(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.SystemObjectPerms
	for rows.Next() {
		var p models.SystemObjectPerms
		if err := rows.Scan(&p.ProfileID, &p.ObjectAPIName, &p.AllowRead, &p.AllowCreate, &p.AllowEdit, &p.AllowDelete, &p.ViewAll, &p.ModifyAll); err != nil {
			continue
		}
		perms = append(perms, p)
	}

	return perms, nil
}

// ListFieldPermissions retrieves all field permissions for a profile
func (r *PermissionRepository) ListFieldPermissions(ctx context.Context, profileID string) ([]models.SystemFieldPerms, error) {
	query := fmt.Sprintf(`
		SELECT profile_id, object_api_name, field_api_name, readable, editable
		FROM %s
		WHERE profile_id = ?
	`, constants.TableFieldPerms)

	rows, err := r.db.QueryContext(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.SystemFieldPerms
	for rows.Next() {
		var p models.SystemFieldPerms
		if err := rows.Scan(&p.ProfileID, &p.ObjectAPIName, &p.FieldAPIName, &p.Readable, &p.Editable); err != nil {
			continue
		}
		perms = append(perms, p)
	}

	return perms, nil
}

// UpsertObjectPermission creates or updates an object permission
func (r *PermissionRepository) UpsertObjectPermission(ctx context.Context, perm models.SystemObjectPerms) error {
	return r.upsertObjectPermission(ctx, r.db, perm)
}

// UpsertObjectPermissionTx creates or updates an object permission within a transaction
func (r *PermissionRepository) UpsertObjectPermissionTx(ctx context.Context, tx *sql.Tx, perm models.SystemObjectPerms) error {
	return r.upsertObjectPermission(ctx, tx, perm)
}

func (r *PermissionRepository) upsertObjectPermission(ctx context.Context, exec Executor, perm models.SystemObjectPerms) error {
	id := utils.GenerateID()
	query := fmt.Sprintf(`
		INSERT INTO %s (id, profile_id, permission_set_id, object_api_name, allow_read, allow_create, allow_edit, allow_delete, view_all, modify_all, created_date, last_modified_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			allow_read = VALUES(allow_read),
			allow_create = VALUES(allow_create),
			allow_edit = VALUES(allow_edit),
			allow_delete = VALUES(allow_delete),
			view_all = VALUES(view_all),
			modify_all = VALUES(modify_all),
			last_modified_date = NOW()
	`, constants.TableObjectPerms)

	_, err := exec.ExecContext(ctx, query, id, perm.ProfileID, perm.PermissionSetID, perm.ObjectAPIName, perm.AllowRead, perm.AllowCreate, perm.AllowEdit, perm.AllowDelete, perm.ViewAll, perm.ModifyAll)
	return err
}

// UpsertFieldPermission creates or updates a field permission
func (r *PermissionRepository) UpsertFieldPermission(ctx context.Context, perm models.SystemFieldPerms) error {
	return r.upsertFieldPermission(ctx, r.db, perm)
}

// UpsertFieldPermissionTx creates or updates a field permission within a transaction
func (r *PermissionRepository) UpsertFieldPermissionTx(ctx context.Context, tx *sql.Tx, perm models.SystemFieldPerms) error {
	return r.upsertFieldPermission(ctx, tx, perm)
}

func (r *PermissionRepository) upsertFieldPermission(ctx context.Context, exec Executor, perm models.SystemFieldPerms) error {
	id := utils.GenerateID()
	query := fmt.Sprintf(`
		INSERT INTO %s (id, profile_id, permission_set_id, object_api_name, field_api_name, readable, editable, created_date, last_modified_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			readable = VALUES(readable),
			editable = VALUES(editable),
			last_modified_date = NOW()
	`, constants.TableFieldPerms)

	_, err := exec.ExecContext(ctx, query, id, perm.ProfileID, perm.PermissionSetID, perm.ObjectAPIName, perm.FieldAPIName, perm.Readable, perm.Editable)
	return err
}

// GrantInitialPermissions grants default permissions for a new object to all profiles
func (r *PermissionRepository) GrantInitialPermissions(ctx context.Context, objectAPIName string) error {
	// Get all Profiles
	query := fmt.Sprintf("SELECT id FROM %s", constants.TableProfile)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to fetch profiles: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var profileID string
		if err := rows.Scan(&profileID); err != nil {
			continue
		}

		// Check if System Admin for elevated permissions
		isSystemAdmin := profileID == constants.ProfileSystemAdmin

		// Default permissions: CRUD for everyone, ModifyAll/ViewAll for Admin
		allowRead := true
		allowCreate := true
		allowEdit := true
		allowDelete := true
		viewAll := false
		modifyAll := false

		if isSystemAdmin {
			viewAll = true
			modifyAll = true
		}

		id := utils.GenerateID()
		insertQuery := fmt.Sprintf(`
			INSERT INTO %s (id, profile_id, object_api_name, allow_read, allow_create, allow_edit, allow_delete, view_all, modify_all, created_date, last_modified_date)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
			ON DUPLICATE KEY UPDATE
				allow_read = VALUES(allow_read),
				allow_create = VALUES(allow_create),
				allow_edit = VALUES(allow_edit),
				allow_delete = VALUES(allow_delete),
				view_all = VALUES(view_all),
				modify_all = VALUES(modify_all),
				last_modified_date = NOW()
		`, constants.TableObjectPerms)

		if _, err := r.db.ExecContext(ctx, insertQuery, id, profileID, objectAPIName, allowRead, allowCreate, allowEdit, allowDelete, viewAll, modifyAll); err != nil {
			log.Printf("⚠️ Warning: Failed to grant permission for profile %s: %v", profileID, err)
		}
	}
	return nil
}

// ==================== Permission Set Extensions ====================

// ListPermissionSetObjectPermissions retrieves all object permissions for a permission set
func (r *PermissionRepository) ListPermissionSetObjectPermissions(ctx context.Context, permissionSetID string) ([]models.SystemObjectPerms, error) {
	query := fmt.Sprintf(`
		SELECT permission_set_id, object_api_name, allow_read, allow_create, allow_edit, allow_delete, view_all, modify_all
		FROM %s
		WHERE permission_set_id = ?
	`, constants.TableObjectPerms)

	rows, err := r.db.QueryContext(ctx, query, permissionSetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.SystemObjectPerms
	for rows.Next() {
		var p models.SystemObjectPerms
		if err := rows.Scan(&p.PermissionSetID, &p.ObjectAPIName, &p.AllowRead, &p.AllowCreate, &p.AllowEdit, &p.AllowDelete, &p.ViewAll, &p.ModifyAll); err != nil {
			continue
		}
		perms = append(perms, p)
	}
	return perms, nil
}

// ListPermissionSetFieldPermissions retrieves all field permissions for a permission set
func (r *PermissionRepository) ListPermissionSetFieldPermissions(ctx context.Context, permissionSetID string) ([]models.SystemFieldPerms, error) {
	query := fmt.Sprintf(`
		SELECT permission_set_id, object_api_name, field_api_name, readable, editable
		FROM %s
		WHERE permission_set_id = ?
	`, constants.TableFieldPerms)

	rows, err := r.db.QueryContext(ctx, query, permissionSetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.SystemFieldPerms
	for rows.Next() {
		var p models.SystemFieldPerms
		if err := rows.Scan(&p.PermissionSetID, &p.ObjectAPIName, &p.FieldAPIName, &p.Readable, &p.Editable); err != nil {
			continue
		}
		perms = append(perms, p)
	}
	return perms, nil
}

// ListEffectiveObjectPermissionsForUser returns aggregated permissions for a user
func (r *PermissionRepository) ListEffectiveObjectPermissionsForUser(ctx context.Context, userID, profileID string) ([]models.SystemObjectPerms, error) {
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

	rows, err := r.db.QueryContext(ctx, query, profileID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.SystemObjectPerms
	for rows.Next() {
		var p models.SystemObjectPerms
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

// ListEffectiveFieldPermissionsForUser returns aggregated field permissions for a user
func (r *PermissionRepository) ListEffectiveFieldPermissionsForUser(ctx context.Context, userID, profileID string) ([]models.SystemFieldPerms, error) {
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

	rows, err := r.db.QueryContext(ctx, query, profileID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.SystemFieldPerms
	for rows.Next() {
		var p models.SystemFieldPerms
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

// CreatePermissionSet creates a new permission set
func (r *PermissionRepository) CreatePermissionSet(ctx context.Context, name, label, description string) (string, error) {
	id := utils.GenerateID()
	query := fmt.Sprintf(`
		INSERT INTO %s (id, name, label, description, is_active, created_date)
		VALUES (?, ?, ?, ?, true, NOW())
	`, constants.TablePermissionSet)

	_, err := r.db.ExecContext(ctx, query, id, name, label, description)
	if err != nil {
		return "", err
	}
	return id, nil
}

// UpdatePermissionSet updates a permission set
func (r *PermissionRepository) UpdatePermissionSet(ctx context.Context, id, name, label, description string, isActive bool) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET name = ?, label = ?, description = ?, is_active = ?
		WHERE id = ?
	`, constants.TablePermissionSet)

	_, err := r.db.ExecContext(ctx, query, name, label, description, isActive, id)
	return err
}

// DeletePermissionSet deletes a permission set
func (r *PermissionRepository) DeletePermissionSet(ctx context.Context, id string) error {
	// First delete assignments
	_, err := r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE permission_set_id = ?", constants.TablePermissionSetAssignment), id)
	if err != nil {
		return err
	}
	// Delete permissions
	_, err = r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE permission_set_id = ?", constants.TableObjectPerms), id)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE permission_set_id = ?", constants.TableFieldPerms), id)
	if err != nil {
		return err
	}
	// Delete the set itself
	_, err = r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE id = ?", constants.TablePermissionSet), id)
	return err
}

// GetUserProfileID fetches the profile ID for a given user
func (r *PermissionRepository) GetUserProfileID(ctx context.Context, userID string) (string, error) {
	query := fmt.Sprintf("SELECT profile_id FROM %s WHERE id = ?", constants.TableUser)
	var profileID string
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&profileID)
	if err != nil {
		return "", fmt.Errorf("failed to get user profile: %w", err)
	}
	return profileID, nil
}

// GetAllRoles retrieves all system roles
func (r *PermissionRepository) GetAllRoles(ctx context.Context) ([]*models.SystemRole, error) {
	query := fmt.Sprintf("SELECT id, name, description, parent_role_id FROM %s WHERE is_deleted = 0", constants.TableRole)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	var roles []*models.SystemRole
	for rows.Next() {
		var role models.SystemRole
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.ParentRoleID); err != nil {
			continue
		}
		roles = append(roles, &role)
	}
	return roles, nil
}

// CreateRole creates a new role
func (r *PermissionRepository) CreateRole(ctx context.Context, name, description string, parentRoleID *string) (string, error) {
	id := utils.GenerateID()
	query := fmt.Sprintf(`
		INSERT INTO %s (id, name, description, parent_role_id, created_date, last_modified_date, is_deleted)
		VALUES (?, ?, ?, ?, NOW(), NOW(), 0)
	`, constants.TableRole)

	_, err := r.db.ExecContext(ctx, query, id, name, description, parentRoleID)
	if err != nil {
		return "", fmt.Errorf("failed to create role: %w", err)
	}
	return id, nil
}

// GetRole retrieves a role by ID
func (r *PermissionRepository) GetRole(ctx context.Context, id string) (*models.SystemRole, error) {
	query := fmt.Sprintf("SELECT id, name, description, parent_role_id FROM %s WHERE id = ? AND is_deleted = 0", constants.TableRole)
	var role models.SystemRole
	err := r.db.QueryRowContext(ctx, query, id).Scan(&role.ID, &role.Name, &role.Description, &role.ParentRoleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return &role, nil
}

// UpdateRole updates an existing role
func (r *PermissionRepository) UpdateRole(ctx context.Context, id, name, description string, parentRoleID *string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET name = ?, description = ?, parent_role_id = ?, last_modified_date = NOW()
		WHERE id = ? AND is_deleted = 0
	`, constants.TableRole)

	_, err := r.db.ExecContext(ctx, query, name, description, parentRoleID, id)
	return err
}

// DeleteRole soft-deletes a role
func (r *PermissionRepository) DeleteRole(ctx context.Context, id string) error {
	query := fmt.Sprintf("UPDATE %s SET is_deleted = 1, last_modified_date = NOW() WHERE id = ?", constants.TableRole)
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// IsUserInGroup checks if a user is a member of a group
func (r *PermissionRepository) IsUserInGroup(ctx context.Context, groupID, userID string) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE group_id = ? AND user_id = ?", constants.TableGroupMember)
	var count int
	err := r.db.QueryRowContext(ctx, query, groupID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check group membership: %w", err)
	}
	return count > 0, nil
}

// GetManualShareAccessLevels retrieves access levels granted via manual sharing rules
func (r *PermissionRepository) GetManualShareAccessLevels(ctx context.Context, objectAPIName, recordID, userID string) ([]string, error) {
	query := fmt.Sprintf(`
		SELECT access_level FROM %s 
		WHERE object_api_name = ? AND record_id = ? AND is_deleted = 0
		AND (share_with_user_id = ? OR share_with_group_id IN (
			SELECT group_id FROM %s WHERE user_id = ?
		))
	`, constants.TableRecordShare, constants.TableGroupMember)

	rows, err := r.db.QueryContext(ctx, query, objectAPIName, recordID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query manual shares: %w", err)
	}
	defer rows.Close()

	var levels []string
	for rows.Next() {
		var level string
		if err := rows.Scan(&level); err != nil {
			continue
		}
		levels = append(levels, level)
	}
	return levels, nil
}

// GetTeamMemberAccessLevel retrieves the access level for a user in a record team
func (r *PermissionRepository) GetTeamMemberAccessLevel(ctx context.Context, objectAPIName, recordID, userID string) (*string, error) {
	query := fmt.Sprintf(`
		SELECT access_level FROM %s 
		WHERE object_api_name = ? AND record_id = ? AND user_id = ? AND is_deleted = 0
	`, constants.TableTeamMember)

	var accessLevel string
	err := r.db.QueryRowContext(ctx, query, objectAPIName, recordID, userID).Scan(&accessLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query team member access: %w", err)
	}
	return &accessLevel, nil
}

// GetAllProfiles retrieves all system profiles
func (r *PermissionRepository) GetAllProfiles(ctx context.Context) ([]*models.SystemProfile, error) {
	query := fmt.Sprintf("SELECT id, name, description, is_active, is_system FROM %s ORDER BY name ASC", constants.TableProfile)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query profiles: %w", err)
	}
	defer rows.Close()

	var profiles []*models.SystemProfile
	for rows.Next() {
		var p models.SystemProfile
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.IsActive, &p.IsSystem); err != nil {
			return nil, fmt.Errorf("failed to scan profile: %w", err)
		}
		profiles = append(profiles, &p)
	}
	return profiles, nil
}
