package services

import (
	"context"
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
	RoleID    string
}

// CreateUser creates a new user account
func (s *AuthService) CreateUser(ctx context.Context, req CreateUserRequest) (*models.UserSession, error) {
	// 1. Validate Email
	if !auth.IsValidEmail(req.Email) {
		return nil, errors.NewValidationError(constants.FieldEmail, "Invalid email format")
	}

	// 2. Validate Password
	if err := auth.ValidatePasswordStrength(req.Password); err != nil {
		return nil, err
	}

	// 3. Check for Existing User
	exists, err := s.userRepo.CheckUserExistsByEmail(ctx, req.Email)
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

	var roleID *string
	if req.RoleID != "" {
		roleID = &req.RoleID
	}

	// 6. Insert User using PersistenceService
	userStruct := models.SystemUser{
		ID:          userID,
		Username:    req.Email,
		Email:       req.Email,
		Password:    string(hashedPassword),
		FirstName:   firstName,
		LastName:    lastName,
		ProfileID:   profileID,
		RoleID:      roleID,
		CreatedDate: now,
		IsActive:    true,
	}

	// System Context
	systemContext := &models.UserSession{
		ID:        "system-create-user",
		Name:      "System User Creator",
		ProfileID: constants.ProfileSystemAdmin,
	}

	// Use propagated context
	// ctx := context.Background()
	userData := userStruct.ToSObject()
	if req.RoleID != "" {
		userData["role_id"] = req.RoleID
	}
	if _, err := s.persistence.Insert(ctx, constants.TableUser, userData, systemContext); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &models.UserSession{
		ID:        userID,
		Name:      req.Name,
		Email:     &req.Email,
		ProfileID: profileID,
		RoleID:    roleID,
	}, nil
}

// UpdateUserRequest contains the data that can be updated on a user
type UpdateUserRequest struct {
	Name      string
	Email     string
	Password  string
	ProfileID string
	RoleID    string
	IsActive  *bool
}

// UpdateUser updates an existing user's information
func (s *AuthService) UpdateUser(ctx context.Context, userID string, req UpdateUserRequest) error {
	// 1. Check Existence
	exists, err := s.userRepo.CheckUserExistsByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if !exists {
		return errors.NewNotFoundError("User", userID)
	}

	// 2. Prepare Updates
	updates := make(map[string]interface{})

	if req.Name != "" {
		firstName, lastName := splitName(req.Name)
		updates[constants.FieldUsername] = req.Name
		updates[constants.FieldFirstName] = firstName
		updates[constants.FieldLastName] = lastName
	}

	if req.Email != "" {
		if !auth.IsValidEmail(req.Email) {
			return errors.NewValidationError(constants.FieldEmail, "Invalid email format")
		}

		// Check for email uniqueness
		emailExists, err := s.userRepo.CheckEmailConflict(ctx, req.Email, userID)
		if err != nil {
			return fmt.Errorf("database error checking email: %w", err)
		}
		if emailExists {
			return errors.NewConflictError(constants.TableUser, constants.FieldEmail, req.Email)
		}

		updates[constants.FieldEmail] = req.Email
	}

	if req.ProfileID != "" {
		updates[constants.FieldProfileID] = req.ProfileID
	}

	if req.RoleID != "" {
		updates[constants.FieldRoleID] = req.RoleID
	}

	if req.Password != "" {
		if err := auth.ValidatePasswordStrength(req.Password); err != nil {
			return err
		}
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		updates[constants.FieldPassword] = string(hash)
	}

	if req.IsActive != nil {
		updates[constants.FieldIsActive] = *req.IsActive
	}

	if len(updates) == 0 {
		return nil // No changes
	}

	if err := s.userRepo.UpdateUser(ctx, userID, updates); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	log.Printf("📝 User updated: %s", userID)
	return nil
}

// DeleteUser removes a user from the system
func (s *AuthService) DeleteUser(ctx context.Context, userID string) error {
	// Check Existence
	exists, err := s.userRepo.CheckUserExistsByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if !exists {
		return errors.NewNotFoundError("User", userID)
	}

	if err := s.userRepo.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	log.Printf("🗑️ User deleted: %s", userID)
	return nil
}

// GetUsers retrieves all users in the system
func (s *AuthService) GetUsers(ctx context.Context) ([]map[string]interface{}, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	var result []map[string]interface{}
	for _, u := range users {
		fullName := strings.TrimSpace(u.FirstName + " " + u.LastName)
		if fullName == "" {
			fullName = u.Username // Fallback
		}

		userMap := map[string]interface{}{
			constants.FieldID:            u.ID,
			constants.FieldUsername:      u.Username,
			constants.FieldName:          fullName,
			constants.FieldEmail:         u.Email,
			constants.FieldProfileID:     u.ProfileID,
			constants.FieldRoleID:        u.RoleID,
			constants.FieldIsActive:      u.IsActive,
			constants.FieldLastLoginDate: u.LastLoginDate,
			constants.FieldCreatedDate:   u.CreatedDate,
		}
		result = append(result, userMap)
	}
	return result, nil
}

// GetProfiles retrieves all security profiles
func (s *AuthService) GetProfiles(ctx context.Context) ([]map[string]interface{}, error) {
	profiles, err := s.permissionRepo.GetAllProfiles(ctx)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, p := range profiles {
		profileMap := map[string]interface{}{
			constants.FieldID:          p.ID,
			constants.FieldName:        p.Name,
			constants.FieldDescription: p.Description,
			constants.FieldIsSystem:    p.IsSystem,
			constants.FieldIsActive:    p.IsActive,
		}
		result = append(result, profileMap)
	}
	return result, nil
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
