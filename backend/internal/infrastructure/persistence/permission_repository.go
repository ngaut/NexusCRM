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
	cols := strings.Join([]string{
		constants.FieldProfileID, constants.FieldObjectAPIName,
		constants.FieldSysObjectPerms_AllowRead, constants.FieldSysObjectPerms_AllowCreate,
		constants.FieldSysObjectPerms_AllowEdit, constants.FieldSysObjectPerms_AllowDelete,
		constants.FieldSysObjectPerms_ViewAll, constants.FieldSysObjectPerms_ModifyAll,
	}, ", ")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = ? AND %s = ?
		LIMIT 1
	`, cols, constants.TableObjectPerms, constants.FieldProfileID, constants.FieldObjectAPIName)

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
	cols := strings.Join([]string{
		constants.FieldProfileID, constants.FieldObjectAPIName, constants.FieldSysFieldPerms_FieldAPIName,
		constants.FieldSysFieldPerms_Readable, constants.FieldSysFieldPerms_Editable,
	}, ", ")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = ? AND %s = ? AND %s = ?
		LIMIT 1
	`, cols, constants.TableFieldPerms, constants.FieldProfileID, constants.FieldObjectAPIName, constants.FieldSysFieldPerms_FieldAPIName)

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
	aggCols := strings.Join([]string{
		fmt.Sprintf("MAX(%s)", constants.FieldSysObjectPerms_AllowRead),
		fmt.Sprintf("MAX(%s)", constants.FieldSysObjectPerms_AllowCreate),
		fmt.Sprintf("MAX(%s)", constants.FieldSysObjectPerms_AllowEdit),
		fmt.Sprintf("MAX(%s)", constants.FieldSysObjectPerms_AllowDelete),
		fmt.Sprintf("MAX(%s)", constants.FieldSysObjectPerms_ViewAll),
		fmt.Sprintf("MAX(%s)", constants.FieldSysObjectPerms_ModifyAll),
	}, ", ")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = ?
		AND (
			%s = ? 
			OR 
			%s IN (SELECT %s FROM %s WHERE %s = ?)
		)
	`, aggCols, constants.TableObjectPerms, constants.FieldObjectAPIName,
		constants.FieldProfileID,
		constants.FieldPermissionSetID, constants.FieldPermissionSetID, constants.TablePermissionSetAssignment,
		constants.FieldSysPermissionSetAssignment_AssigneeID)

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
	aggCols := strings.Join([]string{
		fmt.Sprintf("MAX(%s)", constants.FieldSysFieldPerms_Readable),
		fmt.Sprintf("MAX(%s)", constants.FieldSysFieldPerms_Editable),
	}, ", ")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = ? AND %s = ?
		AND (
			%s = ? 
			OR 
			%s IN (SELECT %s FROM %s WHERE %s = ?)
		)
	`, aggCols, constants.TableFieldPerms, constants.FieldObjectAPIName, constants.FieldSysFieldPerms_FieldAPIName,
		constants.FieldProfileID,
		constants.FieldPermissionSetID, constants.FieldPermissionSetID, constants.TablePermissionSetAssignment,
		constants.FieldSysPermissionSetAssignment_AssigneeID)

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
	cols := strings.Join([]string{
		constants.FieldProfileID, constants.FieldObjectAPIName,
		constants.FieldSysObjectPerms_AllowRead, constants.FieldSysObjectPerms_AllowCreate,
		constants.FieldSysObjectPerms_AllowEdit, constants.FieldSysObjectPerms_AllowDelete,
		constants.FieldSysObjectPerms_ViewAll, constants.FieldSysObjectPerms_ModifyAll,
	}, ", ")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = ?
	`, cols, constants.TableObjectPerms, constants.FieldProfileID)

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
	cols := strings.Join([]string{
		constants.FieldProfileID, constants.FieldObjectAPIName, constants.FieldSysFieldPerms_FieldAPIName,
		constants.FieldSysFieldPerms_Readable, constants.FieldSysFieldPerms_Editable,
	}, ", ")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = ?
	`, cols, constants.TableFieldPerms, constants.FieldProfileID)

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
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldProfileID, constants.FieldPermissionSetID,
		constants.FieldObjectAPIName,
		constants.FieldSysObjectPerms_AllowRead, constants.FieldSysObjectPerms_AllowCreate,
		constants.FieldSysObjectPerms_AllowEdit, constants.FieldSysObjectPerms_AllowDelete,
		constants.FieldSysObjectPerms_ViewAll, constants.FieldSysObjectPerms_ModifyAll,
		constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")

	updates := strings.Join([]string{
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObjectPerms_AllowRead, constants.FieldSysObjectPerms_AllowRead),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObjectPerms_AllowCreate, constants.FieldSysObjectPerms_AllowCreate),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObjectPerms_AllowEdit, constants.FieldSysObjectPerms_AllowEdit),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObjectPerms_AllowDelete, constants.FieldSysObjectPerms_AllowDelete),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObjectPerms_ViewAll, constants.FieldSysObjectPerms_ViewAll),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObjectPerms_ModifyAll, constants.FieldSysObjectPerms_ModifyAll),
		fmt.Sprintf("%s = NOW()", constants.FieldLastModifiedDate),
	}, ", ")

	query := fmt.Sprintf(`
		INSERT INTO %s (%s)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE %s
		`, constants.TableObjectPerms, cols, updates)

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
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldProfileID, constants.FieldPermissionSetID,
		constants.FieldObjectAPIName, constants.FieldSysFieldPerms_FieldAPIName,
		constants.FieldSysFieldPerms_Readable, constants.FieldSysFieldPerms_Editable,
		constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")

	updates := strings.Join([]string{
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysFieldPerms_Readable, constants.FieldSysFieldPerms_Readable),
		fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysFieldPerms_Editable, constants.FieldSysFieldPerms_Editable),
		fmt.Sprintf("%s = NOW()", constants.FieldLastModifiedDate),
	}, ", ")

	query := fmt.Sprintf(`
		INSERT INTO %s (%s)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE %s
	`, constants.TableFieldPerms, cols, updates)

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
		cols := strings.Join([]string{
			constants.FieldID, constants.FieldProfileID, constants.FieldObjectAPIName,
			constants.FieldSysObjectPerms_AllowRead, constants.FieldSysObjectPerms_AllowCreate,
			constants.FieldSysObjectPerms_AllowEdit, constants.FieldSysObjectPerms_AllowDelete,
			constants.FieldSysObjectPerms_ViewAll, constants.FieldSysObjectPerms_ModifyAll,
			constants.FieldCreatedDate, constants.FieldLastModifiedDate,
		}, ", ")

		updates := strings.Join([]string{
			fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObjectPerms_AllowRead, constants.FieldSysObjectPerms_AllowRead),
			fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObjectPerms_AllowCreate, constants.FieldSysObjectPerms_AllowCreate),
			fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObjectPerms_AllowEdit, constants.FieldSysObjectPerms_AllowEdit),
			fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObjectPerms_AllowDelete, constants.FieldSysObjectPerms_AllowDelete),
			fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObjectPerms_ViewAll, constants.FieldSysObjectPerms_ViewAll),
			fmt.Sprintf("%s = VALUES(%s)", constants.FieldSysObjectPerms_ModifyAll, constants.FieldSysObjectPerms_ModifyAll),
			fmt.Sprintf("%s = NOW()", constants.FieldLastModifiedDate),
		}, ", ")

		insertQuery := fmt.Sprintf(`
			INSERT INTO %s (%s)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
			ON DUPLICATE KEY UPDATE %s
		`, constants.TableObjectPerms, cols, updates)

		if _, err := r.db.ExecContext(ctx, insertQuery, id, profileID, objectAPIName, allowRead, allowCreate, allowEdit, allowDelete, viewAll, modifyAll); err != nil {
			log.Printf("⚠️ Warning: Failed to grant permission for profile %s: %v", profileID, err)
		}
	}
	return nil
}

// ==================== Permission Set Extensions ====================

// ListPermissionSetObjectPermissions retrieves all object permissions for a permission set
func (r *PermissionRepository) ListPermissionSetObjectPermissions(ctx context.Context, permissionSetID string) ([]models.SystemObjectPerms, error) {
	cols := strings.Join([]string{
		constants.FieldPermissionSetID, constants.FieldObjectAPIName,
		constants.FieldSysObjectPerms_AllowRead, constants.FieldSysObjectPerms_AllowCreate,
		constants.FieldSysObjectPerms_AllowEdit, constants.FieldSysObjectPerms_AllowDelete,
		constants.FieldSysObjectPerms_ViewAll, constants.FieldSysObjectPerms_ModifyAll,
	}, ", ")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = ?
	`, cols, constants.TableObjectPerms, constants.FieldPermissionSetID)

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
	cols := strings.Join([]string{
		constants.FieldPermissionSetID, constants.FieldObjectAPIName,
		constants.FieldSysFieldPerms_FieldAPIName, constants.FieldSysFieldPerms_Readable,
		constants.FieldSysFieldPerms_Editable,
	}, ", ")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = ?
	`, cols, constants.TableFieldPerms, constants.FieldPermissionSetID)

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
			%s,
			MAX(%s), MAX(%s), MAX(%s), MAX(%s), MAX(%s), MAX(%s)
		FROM %s
		WHERE 
			%s = ? 
			OR 
			%s IN (SELECT %s FROM %s WHERE %s = ?)
		GROUP BY %s
	`, constants.FieldObjectAPIName,
		constants.FieldSysObjectPerms_AllowRead, constants.FieldSysObjectPerms_AllowCreate,
		constants.FieldSysObjectPerms_AllowEdit, constants.FieldSysObjectPerms_AllowDelete,
		constants.FieldSysObjectPerms_ViewAll, constants.FieldSysObjectPerms_ModifyAll,
		constants.TableObjectPerms, constants.FieldProfileID, constants.FieldPermissionSetID, constants.FieldPermissionSetID, constants.TablePermissionSetAssignment, constants.FieldSysPermissionSetAssignment_AssigneeID, constants.FieldObjectAPIName)

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
			%s, %s,
			MAX(%s), MAX(%s)
		FROM %s
		WHERE 
			%s = ? 
			OR 
			%s IN (SELECT %s FROM %s WHERE %s = ?)
		GROUP BY %s, %s
	`, constants.FieldObjectAPIName, constants.FieldSysFieldPerms_FieldAPIName,
		constants.FieldSysFieldPerms_Readable, constants.FieldSysFieldPerms_Editable,
		constants.TableFieldPerms, constants.FieldProfileID, constants.FieldPermissionSetID, constants.FieldPermissionSetID, constants.TablePermissionSetAssignment, constants.FieldSysPermissionSetAssignment_AssigneeID, constants.FieldObjectAPIName, constants.FieldSysFieldPerms_FieldAPIName)

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
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysPermissionSet_Name, constants.FieldSysPermissionSet_Label,
		constants.FieldSysPermissionSet_Description, constants.FieldSysPermissionSet_IsActive,
		constants.FieldCreatedDate,
	}, ", ")

	query := fmt.Sprintf(`
		INSERT INTO %s (%s)
		VALUES (?, ?, ?, ?, true, NOW())
	`, constants.TablePermissionSet, cols)

	_, err := r.db.ExecContext(ctx, query, id, name, label, description)
	if err != nil {
		return "", err
	}
	return id, nil
}

// UpdatePermissionSet updates a permission set
func (r *PermissionRepository) UpdatePermissionSet(ctx context.Context, id, name, label, description string, isActive bool) error {
	updates := strings.Join([]string{
		fmt.Sprintf("%s = ?", constants.FieldSysPermissionSet_Name),
		fmt.Sprintf("%s = ?", constants.FieldSysPermissionSet_Label),
		fmt.Sprintf("%s = ?", constants.FieldSysPermissionSet_Description),
		fmt.Sprintf("%s = ?", constants.FieldSysPermissionSet_IsActive),
	}, ", ")

	query := fmt.Sprintf(`
		UPDATE %s
		SET %s
		WHERE %s = ?
	`, constants.TablePermissionSet, updates, constants.FieldID)

	_, err := r.db.ExecContext(ctx, query, name, label, description, isActive, id)
	return err
}

// DeletePermissionSet deletes a permission set
func (r *PermissionRepository) DeletePermissionSet(ctx context.Context, id string) error {
	// First delete assignments
	_, err := r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = ?", constants.TablePermissionSetAssignment, constants.FieldPermissionSetID), id)
	if err != nil {
		return err
	}
	// Delete permissions
	_, err = r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = ?", constants.TableObjectPerms, constants.FieldPermissionSetID), id)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = ?", constants.TableFieldPerms, constants.FieldPermissionSetID), id)
	if err != nil {
		return err
	}
	// Delete the set itself
	_, err = r.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = ?", constants.TablePermissionSet, constants.FieldID), id)
	return err
}

// GetUserProfileID fetches the profile ID for a given user
func (r *PermissionRepository) GetUserProfileID(ctx context.Context, userID string) (string, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", constants.FieldProfileID, constants.TableUser, constants.FieldID)
	var profileID string
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&profileID)
	if err != nil {
		return "", fmt.Errorf("failed to get user profile: %w", err)
	}
	return profileID, nil
}

// GetAllRoles retrieves all system roles
func (r *PermissionRepository) GetAllRoles(ctx context.Context) ([]*models.SystemRole, error) {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysRole_Name,
		constants.FieldSysRole_Description, constants.FieldSysRole_ParentRoleID,
	}, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = 0", cols, constants.TableRole, constants.FieldIsDeleted)
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
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysRole_Name, constants.FieldSysRole_Description,
		constants.FieldSysRole_ParentRoleID, constants.FieldCreatedDate,
		constants.FieldLastModifiedDate, constants.FieldIsDeleted,
	}, ", ")
	query := fmt.Sprintf(`
		INSERT INTO %s (%s)
		VALUES (?, ?, ?, ?, NOW(), NOW(), 0)
	`, constants.TableRole, cols)

	_, err := r.db.ExecContext(ctx, query, id, name, description, parentRoleID)
	if err != nil {
		return "", fmt.Errorf("failed to create role: %w", err)
	}
	return id, nil
}

// GetRole retrieves a role by ID
func (r *PermissionRepository) GetRole(ctx context.Context, id string) (*models.SystemRole, error) {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysRole_Name,
		constants.FieldSysRole_Description, constants.FieldSysRole_ParentRoleID,
	}, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ? AND %s = 0", cols, constants.TableRole, constants.FieldID, constants.FieldIsDeleted)
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
	updates := strings.Join([]string{
		fmt.Sprintf("%s = ?", constants.FieldSysRole_Name),
		fmt.Sprintf("%s = ?", constants.FieldSysRole_Description),
		fmt.Sprintf("%s = ?", constants.FieldSysRole_ParentRoleID),
		fmt.Sprintf("%s = NOW()", constants.FieldLastModifiedDate),
	}, ", ")
	query := fmt.Sprintf(`
		UPDATE %s
		SET %s
		WHERE %s = ? AND %s = 0
	`, constants.TableRole, updates, constants.FieldID, constants.FieldIsDeleted)

	_, err := r.db.ExecContext(ctx, query, name, description, parentRoleID, id)
	return err
}

// DeleteRole soft-deletes a role
func (r *PermissionRepository) DeleteRole(ctx context.Context, id string) error {
	query := fmt.Sprintf("UPDATE %s SET %s = 1, %s = NOW() WHERE %s = ?", constants.TableRole, constants.FieldIsDeleted, constants.FieldLastModifiedDate, constants.FieldID)
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// IsUserInGroup checks if a user is a member of a group
func (r *PermissionRepository) IsUserInGroup(ctx context.Context, groupID, userID string) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ? AND %s = ?", constants.TableGroupMember, constants.FieldSysGroupMember_GroupID, constants.FieldSysGroupMember_UserID)
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
		SELECT %s FROM %s 
		WHERE %s = ? AND %s = ? AND %s = 0
		AND (%s = ? OR %s IN (
			SELECT %s FROM %s WHERE %s = ?
		))
	`, constants.FieldSysRecordShare_AccessLevel, constants.TableRecordShare,
		constants.FieldObjectAPIName, constants.FieldRecordID, constants.FieldIsDeleted,
		constants.FieldSysRecordShare_ShareWithUserID, constants.FieldSysRecordShare_ShareWithGroupID,
		constants.FieldSysGroupMember_GroupID, constants.TableGroupMember, constants.FieldSysGroupMember_UserID)

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
		SELECT %s FROM %s 
		WHERE %s = ? AND %s = ? AND %s = ? AND %s = 0
	`, constants.FieldSysTeamMember_AccessLevel, constants.TableTeamMember,
		constants.FieldObjectAPIName, constants.FieldRecordID, constants.FieldUserID, constants.FieldIsDeleted)

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
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysProfile_Name, constants.FieldSysProfile_Description,
		constants.FieldSysProfile_IsActive, constants.FieldSysProfile_IsSystem,
	}, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY %s ASC", cols, constants.TableProfile, constants.FieldSysProfile_Name)
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
