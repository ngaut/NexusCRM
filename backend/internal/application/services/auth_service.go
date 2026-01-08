package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	"github.com/nexuscrm/backend/pkg/auth"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/models"
)

// AuthService handles authentication, session management, and password operations
// AuthService handles authentication, session management, and password operations
// AuthService handles authentication, session management, and password operations
type AuthService struct {
	persistence    *PersistenceService
	userRepo       *persistence.UserRepository
	permissionRepo *persistence.PermissionRepository
	sessionRepo    *persistence.SessionRepository
}

// NewAuthService creates a new AuthService
func NewAuthService(
	persistSvc *PersistenceService,
	userRepo *persistence.UserRepository,
	sessionRepo *persistence.SessionRepository,
	permissionRepo *persistence.PermissionRepository,
) *AuthService {
	return &AuthService{
		persistence:    persistSvc,
		userRepo:       userRepo,
		permissionRepo: permissionRepo,
		sessionRepo:    sessionRepo,
	}
}

// LoginResult contains the result of a successful login
type LoginResult struct {
	Token     string
	User      auth.UserSession
	ExpiresAt time.Time
}

// Login authenticates a user and creates a session
func (s *AuthService) Login(ctx context.Context, email, password, ip, userAgent string) (*LoginResult, error) {
	// 1. Find user by email
	user, err := s.userRepo.FindUserByEmailWithPassword(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if user == nil {
		log.Printf("⚠️ Login failed for %s: user not found", email)
		return nil, errors.NewUnauthorizedError("Invalid email or password")
	}

	// Construct Display Name
	displayName := user.Username
	fullNameParts := []string{}
	if user.FirstName != "" {
		fullNameParts = append(fullNameParts, user.FirstName)
	}
	if user.LastName != "" {
		fullNameParts = append(fullNameParts, user.LastName)
	}
	if len(fullNameParts) > 0 {
		displayName = strings.Join(fullNameParts, " ")
	}

	// 2. Verify password
	if user.PasswordHash == "" {
		log.Printf("⚠️ Login failed for %s: Password not valid (NULL) in DB", email)
		return nil, errors.NewUnauthorizedError("Password authentication not configured for this user")
	}

	if !auth.VerifyPassword(password, user.PasswordHash) {
		log.Printf("⚠️ Login failed for %s: invalid password", email)
		return nil, errors.NewUnauthorizedError("Invalid email or password")
	}

	// 3. Create user session object
	// RoleID is populated by FindUserByEmailWithPassword (Repository Layer)

	userSession := auth.UserSession{
		ID:        user.ID,
		Name:      displayName,
		Email:     user.Email,
		ProfileId: user.ProfileID,
		RoleId:    user.RoleID,
	}

	// 4. Generate JWT token
	token, err := auth.GenerateToken(userSession)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 5. Decode token to get expiry
	claims, _ := auth.DecodeToken(token)
	expiresAt := time.Unix(claims.ExpiresAt.Unix(), 0)
	createdAt := time.Now()

	// 6. Store session in database using SessionRepository
	sessionStruct := &models.SystemSession{
		ID:           claims.RegisteredClaims.ID,
		UserID:       user.ID,
		Token:        token,
		ExpiresAt:    expiresAt,
		IPAddress:    ip,
		UserAgent:    userAgent,
		IsRevoked:    false,
		LastActivity: createdAt,
	}

	if err := s.sessionRepo.InsertSession(ctx, sessionStruct); err != nil {
		return nil, fmt.Errorf("failed to persist session: %w", err)
	}

	return &LoginResult{
		Token:     token,
		User:      userSession,
		ExpiresAt: expiresAt,
	}, nil
}

// ValidateSession checks if a session token is valid and active in the database
func (s *AuthService) ValidateSession(ctx context.Context, tokenString string) (*auth.Claims, error) {
	// 1. Verify JWT signature and claims
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// 2. Check DB for revocation using SessionRepository
	session, err := s.sessionRepo.GetSession(ctx, claims.RegisteredClaims.ID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if session == nil {
		return nil, errors.NewUnauthorizedError("Session not found")
	}

	if session.IsRevoked {
		return nil, errors.NewUnauthorizedError("Session has been revoked")
	}

	return claims, nil
}

// TouchSession updates the last activity timestamp for a session
func (s *AuthService) TouchSession(sessionID string) {
	// Fire and forget - errors are acceptable for non-critical activity timestamps
	go func() {
		_ = s.sessionRepo.UpdateLastActivity(context.Background(), sessionID)
	}()
}

// Logout Revokes a session
func (s *AuthService) Logout(ctx context.Context, tokenString string) error {
	claims, err := auth.DecodeToken(tokenString)
	if err != nil {
		return errors.NewValidationError("token", "Invalid token")
	}

	err = s.sessionRepo.RevokeSession(ctx, claims.RegisteredClaims.ID)
	if err == nil {
		log.Printf("👋 User logged out: %s (Session: %s)", claims.RegisteredClaims.Subject, claims.RegisteredClaims.ID)
	}
	return err
}

// ChangePassword updates a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// 1. Validate Strength
	if err := auth.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	// 2. Load current password
	user, err := s.userRepo.FindUserByIDWithPassword(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to retrieve user: %w", err)
	}
	if user == nil {
		// Should verify if user exists even? Assuming middleware handled it.
		return errors.NewNotFoundError("User", userID)
	}

	if user.PasswordHash == "" {
		return errors.NewValidationError("password", "Password authentication not configured for this user")
	}

	// 3. Verify
	if !auth.VerifyPassword(currentPassword, user.PasswordHash) {
		return errors.NewUnauthorizedError("Current password is incorrect")
	}

	// 4. Hash New
	newHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 5. Update
	err = s.userRepo.UpdatePassword(ctx, userID, newHash)
	if err == nil {
		log.Printf("🔐 Password changed for user: %s", userID)
	}
	return err
}

// GetUserByID retrieves a user session object by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*models.UserSession, error) {
	// Reuse Repo logic
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NewNotFoundError("User", userID)
	}

	// Fetch Role ID separately since GetUserByID might not return it if not cached/joined
	roleID, err := s.userRepo.GetUserRoleID(ctx, userID)
	if err != nil {
		log.Printf("Warning: failed to fetch role for user %s: %v", userID, err)
	}

	displayName := user.Username
	fullNameParts := []string{}
	if user.FirstName != "" {
		fullNameParts = append(fullNameParts, user.FirstName)
	}
	if user.LastName != "" {
		fullNameParts = append(fullNameParts, user.LastName)
	}
	if len(fullNameParts) > 0 {
		displayName = strings.Join(fullNameParts, " ")
	}

	return &models.UserSession{
		ID:        user.ID,
		Name:      displayName,
		Email:     &user.Email,
		ProfileID: user.ProfileID,
		RoleID:    roleID,
	}, nil
}

// User management functions are in auth_user_mgmt.go:
// - CreateUserRequest, UpdateUserRequest types
// - CreateUser, UpdateUser, DeleteUser
// - GetUsers, GetProfiles, splitName
