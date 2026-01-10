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
	query := fmt.Sprintf("%s %s(%s 1 %s %s %s %s = ?)",
		KeywordSelect, "EXISTS", KeywordSelect, KeywordFrom, constants.TableUser, KeywordWhere, constants.FieldEmail)
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *UserRepository) CheckUserExistsByID(ctx context.Context, id string) (bool, error) {
	var exists bool
	query := fmt.Sprintf("%s %s(%s 1 %s %s %s %s = ?)",
		KeywordSelect, "EXISTS", KeywordSelect, KeywordFrom, constants.TableUser, KeywordWhere, constants.FieldID)
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *UserRepository) CheckEmailConflict(ctx context.Context, email, excludeID string) (bool, error) {
	var exists bool
	query := fmt.Sprintf("%s %s(%s 1 %s %s %s %s = ? %s %s != ?)",
		KeywordSelect, "EXISTS", KeywordSelect, KeywordFrom, constants.TableUser, KeywordWhere, constants.FieldEmail, KeywordAnd, constants.FieldID)
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

	query := fmt.Sprintf("%s %s %s %s %s %s = ?", KeywordUpdate, constants.TableUser, KeywordSet, strings.Join(setClauses, ", "), KeywordWhere, constants.FieldID)
	args = append(args, userID)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// DeleteUser deletes a user record
func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	query := fmt.Sprintf("%s %s %s = ?", KeywordDeleteFrom, constants.TableUser, constants.FieldID)
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// FindAll retrieves all users
func (r *UserRepository) FindAll(ctx context.Context) ([]*models.SystemUser, error) {
	query := fmt.Sprintf(`
		%s %s, %s, %s, %s, %s, %s, %s, %s, %s, %s 
		%s %s 
		%s %s %s`,
		KeywordSelect, constants.FieldID, constants.FieldUsername, constants.FieldEmail, constants.FieldProfileID, constants.FieldRoleID, constants.FieldIsActive, constants.FieldCreatedDate, constants.FieldLastLoginDate, constants.FieldFirstName, constants.FieldLastName,
		KeywordFrom, constants.TableUser,
		KeywordOrderBy, constants.FieldCreatedDate, KeywordDesc)

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
		// Parse dates
		if len(createdDateRaw) > 0 {
			if t, err := ParseDBTime(createdDateRaw); err == nil {
				u.CreatedDate = t
			}
		}
		if len(lastLoginRaw) > 0 {
			if t, err := ParseDBTime(lastLoginRaw); err == nil {
				u.LastLoginDate = &t
			}
		}

		users = append(users, &u)
	}
	return users, nil
}

// GetUserRoleID retrieves the role ID for a user
func (r *UserRepository) GetUserRoleID(ctx context.Context, userID string) (*string, error) {
	query := fmt.Sprintf("%s %s %s %s %s %s = ?", KeywordSelect, constants.FieldSysUser_RoleID, KeywordFrom, constants.TableUser, KeywordWhere, constants.FieldID)
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
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysUser_Username, constants.FieldSysUser_Email,
		constants.FieldSysUser_Password, constants.FieldSysUser_ProfileID, constants.FieldSysUser_RoleID,
		constants.FieldSysUser_FirstName, constants.FieldSysUser_LastName,
	}, ", ")

	query := fmt.Sprintf(`
		%s %s 
		%s %s 
		%s %s = ? %s 1`,
		KeywordSelect, cols, KeywordFrom, constants.TableUser, KeywordWhere, constants.FieldSysUser_Email, KeywordLimit)

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
	// Map nullable fields to struct
	// SystemUser generated struct includes these fields

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
		%s %s, %s 
		%s %s 
		%s %s = ? %s 1`,
		KeywordSelect, constants.FieldID, constants.FieldSysUser_Password, KeywordFrom, constants.TableUser, KeywordWhere, constants.FieldID, KeywordLimit)

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
	query := fmt.Sprintf("%s %s %s %s = ?, %s = %s %s %s = ?",
		KeywordUpdate, constants.TableUser, KeywordSet, constants.FieldSysUser_Password, constants.FieldLastModifiedDate, FuncNow, KeywordWhere, constants.FieldID)
	_, err := r.db.ExecContext(ctx, query, passwordHash, userID)
	return err
}

// GetUserByID fetches basic user info
func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*models.SystemUser, error) {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysUser_Username, constants.FieldSysUser_Email,
		constants.FieldSysUser_ProfileID, constants.FieldSysUser_FirstName, constants.FieldSysUser_LastName,
	}, ", ")

	query := fmt.Sprintf(`
		%s %s 
		%s %s 
		%s %s = ? %s 1`,
		KeywordSelect, cols, KeywordFrom, constants.TableUser, KeywordWhere, constants.FieldID, KeywordLimit)

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
