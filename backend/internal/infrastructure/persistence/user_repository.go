package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = ?)", constants.TableUser, constants.FieldEmail)
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *UserRepository) CheckUserExistsByID(ctx context.Context, id string) (bool, error) {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = ?)", constants.TableUser, constants.FieldID)
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *UserRepository) CheckEmailConflict(ctx context.Context, email, excludeID string) (bool, error) {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = ? AND %s != ?)", constants.TableUser, constants.FieldEmail, constants.FieldID)
	err := r.db.QueryRowContext(ctx, query, email, excludeID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// UpdateUser updates a user record
func (r *UserRepository) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setClauses := []string{}
	args := []interface{}{}

	for k, v := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}

	// Always update last_modified_date
	setClauses = append(setClauses, fmt.Sprintf("%s = ?", constants.FieldLastModifiedDate))
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?", constants.TableUser, strings.Join(setClauses, ", "), constants.FieldID)
	args = append(args, userID)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// DeleteUser deletes a user record
func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", constants.TableUser, constants.FieldID)
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// FindAll retrieves all users
func (r *UserRepository) FindAll(ctx context.Context) ([]*models.SystemUser, error) {
	query := fmt.Sprintf(`
		SELECT %s, %s, %s, %s, %s, %s, %s, %s, %s, %s 
		FROM %s 
		ORDER BY %s DESC`,
		constants.FieldID, constants.FieldUsername, constants.FieldEmail, constants.FieldProfileID, constants.FieldRoleID, constants.FieldIsActive, constants.FieldCreatedDate, constants.FieldLastLoginDate, constants.FieldFirstName, constants.FieldLastName,
		constants.TableUser,
		constants.FieldCreatedDate)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*models.SystemUser, 0)
	for rows.Next() {
		var u models.SystemUser
		var createdDateRaw, lastLoginRaw []byte
		var firstName, lastName, roleID sql.NullString

		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.ProfileID, &roleID, &u.IsActive, &createdDateRaw, &lastLoginRaw, &firstName, &lastName); err != nil {
			continue
		}

		u.FirstName = firstName.String
		u.LastName = lastName.String

		if roleID.Valid {
			rID := roleID.String
			u.RoleID = &rID
		}

		// Parse dates
		if len(createdDateRaw) > 0 {
			if t, err := time.Parse("2006-01-02 15:04:05", string(createdDateRaw)); err == nil {
				u.CreatedDate = t
			} else if t, err := time.Parse(time.RFC3339, string(createdDateRaw)); err == nil {
				u.CreatedDate = t
			}
		}
		if len(lastLoginRaw) > 0 {
			if t, err := time.Parse("2006-01-02 15:04:05", string(lastLoginRaw)); err == nil {
				u.LastLoginDate = &t
			} else if t, err := time.Parse(time.RFC3339, string(lastLoginRaw)); err == nil {
				u.LastLoginDate = &t
			}
		}

		users = append(users, &u)
	}
	return users, nil
}

// GetUserRoleID retrieves the role ID for a user
func (r *UserRepository) GetUserRoleID(ctx context.Context, userID string) (*string, error) {
	query := fmt.Sprintf("SELECT role_id FROM %s WHERE id = ?", constants.TableUser)
	var roleID sql.NullString
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&roleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if roleID.Valid {
		return &roleID.String, nil
	}
	return nil, nil
}

// UserWithPassword extends SystemUser to include password hash for auth checks
type UserWithPassword struct {
	*models.SystemUser
	PasswordHash string
}

// FindUserByEmailWithPassword retrieves a user and their password hash by email
func (r *UserRepository) FindUserByEmailWithPassword(ctx context.Context, email string) (*UserWithPassword, error) {
	// We select specific fields needed for Auth (plus Password)
	query := fmt.Sprintf(`
		SELECT id, username, email, password, profile_id, role_id, first_name, last_name 
		FROM %s 
		WHERE email = ? LIMIT 1`,
		constants.TableUser)

	var u UserWithPassword
	var sysUser models.SystemUser
	u.SystemUser = &sysUser

	var password, roleID, firstName, lastName sql.NullString

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&sysUser.ID,
		&sysUser.Username,
		&sysUser.Email,
		&password,
		&sysUser.ProfileID,
		&roleID,
		&firstName,
		&lastName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if password.Valid {
		u.PasswordHash = password.String
	}

	// Map nullable fields to struct
	// Note: SystemUser doesn't have RoleID/FirstName/LastName fields in struct usually?
	// Let's check SystemUser struct definition in generated code.
	// Assuming SystemUser has them or we return a struct that Auth expects.
	// Actually AuthService uses a custom struct inside Login.
	// Let's stick to returning what AuthService needs.
	// SystemUser has: ID, Username, Email, ProfileID, IsActive.
	// It MIGHT NOT have FirstName/LastName/RoleID if they are custom fields or extended.
	// Checking generated model...

	// Just return our extended struct with fields we fetched.
	// We'll store them in strict fields if SystemUser doesn't have them.
	// Actually, let's redefine UserWithPassword to have explicit fields to match AuthService expectations exactly.

	// IMPORTANT: For now I'll just rely on the struct I defined and mapping logic in Service.
	// But wait, SystemUser is generated.
	// Let's look at FindAll in this file (line 102).
	// It scans into u.FirstName, u.LastName. So SystemUser HAS them.

	if firstName.Valid {
		sysUser.FirstName = firstName.String
	}
	if lastName.Valid {
		sysUser.LastName = lastName.String
	}
	if roleID.Valid {
		sysUser.RoleID = &roleID.String
	}

	return &u, nil
}

// FindUserByIDWithPassword retrieves a user and their password by ID
func (r *UserRepository) FindUserByIDWithPassword(ctx context.Context, userID string) (*UserWithPassword, error) {
	query := fmt.Sprintf(`
		SELECT id, password 
		FROM %s 
		WHERE id = ? LIMIT 1`,
		constants.TableUser)

	var u UserWithPassword
	var sysUser models.SystemUser
	u.SystemUser = &sysUser
	var password sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&sysUser.ID, &password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if password.Valid {
		u.PasswordHash = password.String
	}
	return &u, nil
}

// UpdatePassword updates the user's password hash
func (r *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	query := fmt.Sprintf("UPDATE %s SET password = ?, last_modified_date = NOW() WHERE id = ?", constants.TableUser)
	_, err := r.db.ExecContext(ctx, query, passwordHash, userID)
	return err
}

// GetUserByID fetches basic user info
func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*models.SystemUser, error) {
	query := fmt.Sprintf(`
		SELECT id, username, email, profile_id, first_name, last_name 
		FROM %s 
		WHERE id = ? LIMIT 1`,
		constants.TableUser)

	var u models.SystemUser
	var firstName, lastName sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.ProfileID,
		&firstName,
		&lastName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if firstName.Valid {
		u.FirstName = firstName.String
	}
	if lastName.Valid {
		u.LastName = lastName.String
	}

	return &u, nil
}
