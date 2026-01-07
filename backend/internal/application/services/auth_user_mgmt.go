package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nexuscrm/backend/pkg/auth"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== User Management ====================

// CreateUserRequest contains the data needed to create a new user
type CreateUserRequest struct {
	Name      string
	Email     string
	Password  string
	ProfileID string
}

// CreateUser creates a new user account
func (s *AuthService) CreateUser(req CreateUserRequest) (*models.UserSession, error) {
	// 1. Validate Email
	if !auth.IsValidEmail(req.Email) {
		return nil, errors.NewValidationError(constants.FieldEmail, "Invalid email format")
	}

	// 2. Validate Password
	if err := auth.ValidatePasswordStrength(req.Password); err != nil {
		return nil, err
	}

	// 3. Check for Existing User
	exists, err := s.userRepo.CheckUserExistsByEmail(context.Background(), req.Email)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if exists {
		return nil, errors.NewConflictError(constants.TableUser, constants.FieldEmail, req.Email)
	}

	// 4. Hash Password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 5. Prepare Data
	userID := GenerateID()
	now := time.Now()
	profileID := req.ProfileID
	if profileID == "" {
		profileID = constants.ProfileStandardUser
	}

	// Split Name
	firstName, lastName := splitName(req.Name)

	// 6. Insert User using PersistenceService
	userStruct := models.SystemUser{
		ID:          userID,
		Username:    req.Email,
		Email:       req.Email,
		Password:    string(hashedPassword),
		FirstName:   firstName,
		LastName:    lastName,
		ProfileID:   profileID,
		CreatedDate: now,
		IsActive:    true,
	}

	// System Context
	systemContext := &models.UserSession{
		ID:        "system-create-user",
		Name:      "System User Creator",
		ProfileID: constants.ProfileSystemAdmin,
	}

	ctx := context.Background()
	if _, err := s.persistence.Insert(ctx, constants.TableUser, userStruct.ToSObject(), systemContext); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &models.UserSession{
		ID:        userID,
		Name:      req.Name,
		Email:     &req.Email,
		ProfileID: profileID,
	}, nil
}

// UpdateUserRequest contains the data that can be updated on a user
type UpdateUserRequest struct {
	Name      string
	Email     string
	Password  string
	ProfileID string
	IsActive  *bool
}

// UpdateUser updates an existing user's information
func (s *AuthService) UpdateUser(userID string, req UpdateUserRequest) error {
	// 1. Check Existence
	exists, err := s.userRepo.CheckUserExistsByID(context.Background(), userID)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if !exists {
		return errors.NewNotFoundError("User", userID)
	}

	// 2. Build Update Query
	updates := []string{}
	args := []interface{}{}

	if req.Name != "" {
		firstName, lastName := splitName(req.Name)
		updates = append(updates, fmt.Sprintf("%s = ?, %s = ?, %s = ?", constants.FieldUsername, constants.FieldFirstName, constants.FieldLastName))
		args = append(args, req.Name, firstName, lastName)
	}

	if req.Email != "" {
		if !auth.IsValidEmail(req.Email) {
			return errors.NewValidationError(constants.FieldEmail, "Invalid email format")
		}

		// Check for email uniqueness
		emailExists, err := s.userRepo.CheckEmailConflict(context.Background(), req.Email, userID)
		if err != nil {
			return fmt.Errorf("database error checking email: %w", err)
		}
		if emailExists {
			return errors.NewConflictError(constants.TableUser, constants.FieldEmail, req.Email)
		}

		updates = append(updates, fmt.Sprintf("%s = ?", constants.FieldEmail))
		args = append(args, req.Email)
	}

	if req.ProfileID != "" {
		updates = append(updates, fmt.Sprintf("%s = ?", constants.FieldProfileID))
		args = append(args, req.ProfileID)
	}

	if req.Password != "" {
		if err := auth.ValidatePasswordStrength(req.Password); err != nil {
			return err
		}
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		updates = append(updates, fmt.Sprintf("%s = ?", constants.FieldPassword))
		args = append(args, hash)
	}

	if req.IsActive != nil {
		updates = append(updates, fmt.Sprintf("%s = ?", constants.FieldIsActive))
		args = append(args, *req.IsActive)
	}

	if len(updates) == 0 {
		return nil // No changes
	}

	updates = append(updates, fmt.Sprintf("%s = ?", constants.FieldLastModifiedDate))
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE %s SET "+strings.Join(updates, ", ")+" WHERE %s = ?", constants.TableUser, constants.FieldID)
	args = append(args, userID)

	_, err = s.db.Exec(query, args...)
	if err == nil {
		log.Printf("📝 User updated: %s", userID)
	}
	return err
}

// DeleteUser removes a user from the system
func (s *AuthService) DeleteUser(userID string) error {
	// Check Existence
	exists, err := s.userRepo.CheckUserExistsByID(context.Background(), userID)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if !exists {
		return errors.NewNotFoundError("User", userID)
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", constants.TableUser, constants.FieldID)
	_, err = s.db.Exec(query, userID)
	if err == nil {
		log.Printf("🗑️ User deleted: %s", userID)
	}
	return err
}

// GetUsers retrieves all users in the system
func (s *AuthService) GetUsers() ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
		SELECT %s, %s, %s, %s, %s, %s, %s, %s, %s 
		FROM %s 
		ORDER BY %s DESC`,
		constants.FieldID, constants.FieldUsername, constants.FieldEmail, constants.FieldProfileID, constants.FieldIsActive, constants.FieldCreatedDate, constants.FieldLastLoginDate, constants.FieldFirstName, constants.FieldLastName,
		constants.TableUser,
		constants.FieldCreatedDate)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var users []map[string]interface{}
	for rows.Next() {
		var id, username, email, profileID, firstName, lastName string
		var isActive bool
		var createdDate, lastLogin sql.NullTime

		if err := rows.Scan(&id, &username, &email, &profileID, &isActive, &createdDate, &lastLogin, &firstName, &lastName); err != nil {
			continue
		}

		fullName := firstName
		if lastName != "" && lastName != firstName {
			fullName = firstName + " " + lastName
		} else if lastName != "" {
			// Handle case where splitName duplicated it? Or just show one?
			// If firstName == lastName, usually single word.
			fullName = firstName
		}
		// Actually, simple concat is safer:
		fullName = strings.TrimSpace(firstName + " " + lastName)
		if fullName == "" {
			fullName = username // Fallback
		}

		user := map[string]interface{}{
			constants.FieldID:            id,
			constants.FieldUsername:      username,
			constants.FieldName:          fullName,
			constants.FieldEmail:         email,
			constants.FieldProfileID:     profileID,
			constants.FieldIsActive:      isActive,
			constants.FieldLastLoginDate: nil,
		}
		if createdDate.Valid {
			user[constants.FieldCreatedDate] = createdDate.Time
		}
		if lastLogin.Valid {
			user[constants.FieldLastLoginDate] = lastLogin.Time
		}
		users = append(users, user)
	}
	return users, nil
}

// GetProfiles retrieves all security profiles
func (s *AuthService) GetProfiles() ([]map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT %s, %s, %s FROM %s ORDER BY %s ASC",
		constants.FieldID, constants.FieldName, constants.FieldDescription,
		constants.TableProfile, constants.FieldName)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query profiles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var profiles []map[string]interface{}
	for rows.Next() {
		var id, name string
		var description sql.NullString

		if err := rows.Scan(&id, &name, &description); err != nil {
			continue
		}

		profile := map[string]interface{}{
			constants.FieldID:          id,
			constants.FieldName:        name,
			constants.FieldDescription: "",
			constants.FieldIsSystem:    true,
		}
		if description.Valid {
			profile[constants.FieldDescription] = description.String
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

// splitName splits a full name into first and last name
func splitName(fullName string) (firstName, lastName string) {
	parts := strings.SplitN(fullName, " ", 2)
	firstName = parts[0]
	if len(parts) > 1 {
		lastName = parts[1]
	} else {
		// If no last name is provided, use the first name as the last name
		// This is a common fallacy in systems that require both names
		lastName = firstName
	}
	return
}
