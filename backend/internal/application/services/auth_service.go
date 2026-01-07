package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	"github.com/nexuscrm/backend/pkg/auth"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// AuthService handles authentication, session management, and password operations
type AuthService struct {
	db          *database.TiDBConnection
	persistence *PersistenceService
	userRepo    *persistence.UserRepository
}

// NewAuthService creates a new AuthService
func NewAuthService(db *database.TiDBConnection, persistence *PersistenceService, userRepo *persistence.UserRepository) *AuthService {
	return &AuthService{
		db:          db,
		persistence: persistence,
		userRepo:    userRepo,
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
	var user struct {
		ID        string
		Username  string
		Email     string
		Password  sql.NullString
		ProfileId string
		RoleId    sql.NullString
		FirstName sql.NullString
		LastName  sql.NullString
	}

	query := fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE %s = ? LIMIT 1",
		constants.FieldID, constants.FieldUsername, constants.FieldEmail, constants.FieldPassword, constants.FieldProfileID, constants.FieldRoleID, constants.FieldFirstName, constants.FieldLastName,
		constants.TableUser, constants.FieldEmail)

	err := s.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.ProfileId, &user.RoleId, &user.FirstName, &user.LastName)
	if err == sql.ErrNoRows {
		log.Printf("⚠️ Login failed for %s: user not found", email)
		return nil, errors.NewUnauthorizedError("Invalid email or password")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Construct Display Name
	displayName := user.Username
	if user.FirstName.Valid || user.LastName.Valid {
		parts := []string{}
		if user.FirstName.Valid && user.FirstName.String != "" {
			parts = append(parts, user.FirstName.String)
		}
		if user.LastName.Valid && user.LastName.String != "" {
			parts = append(parts, user.LastName.String)
		}
		if len(parts) > 0 {
			// Join with space
			fullName := ""
			for i, p := range parts {
				if i > 0 {
					fullName += " "
				}
				fullName += p
			}
			displayName = fullName
		}
	}

	// 2. Verify password
	if !user.Password.Valid {
		log.Printf("⚠️ Login failed for %s: Password not valid (NULL) in DB", email)
		return nil, errors.NewUnauthorizedError("Password authentication not configured for this user")
	}
	// DEBUG: Print hash length and prefix
	if len(user.Password.String) > 10 {
		log.Printf("🔍 Debug Login: Stored hash len=%d, prefix=%s", len(user.Password.String), user.Password.String[:10])
	} else {
		log.Printf("🔍 Debug Login: Stored hash len=%d (too short!)", len(user.Password.String))
	}

	if !auth.VerifyPassword(password, user.Password.String) {
		log.Printf("⚠️ Login failed for %s: invalid password", email)
		return nil, errors.NewUnauthorizedError("Invalid email or password")
	}

	// 3. Create user session object
	var roleIdPtr *string
	if user.RoleId.Valid {
		roleIdPtr = &user.RoleId.String
	}

	userSession := auth.UserSession{
		ID:        user.ID,
		Name:      displayName,
		Email:     user.Email,
		ProfileId: user.ProfileId,
		RoleId:    roleIdPtr,
	}

	// 4. Generate JWT token
	token, err := auth.GenerateToken(userSession)
	// ... (lines omitted for brevity, keeping same logic)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 5. Decode token to get expiry
	claims, _ := auth.DecodeToken(token)
	expiresAt := time.Unix(claims.ExpiresAt.Unix(), 0)
	createdAt := time.Now()

	// 6. Store session in database using PersistenceService
	sessionStruct := models.SystemSession{
		ID:           claims.RegisteredClaims.ID,
		UserID:       user.ID,
		Token:        token,
		ExpiresAt:    expiresAt,
		IPAddress:    ip,
		UserAgent:    userAgent,
		IsRevoked:    false,
		LastActivity: createdAt,
	}

	// Create system user context for permission bypass/system operations
	systemContext := &models.UserSession{
		ID:        "system-login",
		Name:      "System Login Process",
		ProfileID: constants.ProfileSystemAdmin,
	}

	ctx = context.Background() // Use Passed context or keep system context?
	// Actually, for session creation, we can use passed context with timeout.
	// But here we want to ensure session is created even if request cancels? Maybe.
	// Let's use passed context for consistency with clean architecture.
	if _, err := s.persistence.Insert(ctx, constants.TableSession, sessionStruct.ToSObject(), systemContext); err != nil {
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

	// 2. Check DB for revocation
	var isRevoked bool
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ? LIMIT 1", constants.FieldIsRevoked, constants.TableSession, constants.FieldID)

	err = s.db.QueryRowContext(ctx, query, claims.RegisteredClaims.ID).Scan(&isRevoked)
	if err == sql.ErrNoRows {
		return nil, errors.NewUnauthorizedError("Session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if isRevoked {
		return nil, errors.NewUnauthorizedError("Session has been revoked")
	}

	return claims, nil
}

// TouchSession updates the last activity timestamp for a session
func (s *AuthService) TouchSession(sessionID string) {
	// Fire and forget - errors are acceptable for non-critical activity timestamps
	go func() {
		query := fmt.Sprintf("UPDATE %s SET %s = NOW() WHERE %s = ?", constants.TableSession, constants.FieldLastActivity, constants.FieldID)
		_, _ = s.db.Exec(query, sessionID)
	}()
}

// Logout Revokes a session
func (s *AuthService) Logout(ctx context.Context, tokenString string) error {
	claims, err := auth.DecodeToken(tokenString)
	if err != nil {
		return errors.NewValidationError("token", "Invalid token")
	}

	query := fmt.Sprintf("UPDATE %s SET %s = 1 WHERE %s = ?", constants.TableSession, constants.FieldIsRevoked, constants.FieldID)
	_, err = s.db.ExecContext(ctx, query, claims.RegisteredClaims.ID)
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
	var storedPassword sql.NullString
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ? LIMIT 1", constants.FieldPassword, constants.TableUser, constants.FieldID)
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&storedPassword)
	if err != nil {
		return fmt.Errorf("failed to retrieve user: %w", err)
	}

	if !storedPassword.Valid {
		return errors.NewValidationError("password", "Password authentication not configured for this user")
	}

	// 3. Verify
	if !auth.VerifyPassword(currentPassword, storedPassword.String) {
		return errors.NewUnauthorizedError("Current password is incorrect")
	}

	// 4. Hash New
	newHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 5. Update
	updateQuery := fmt.Sprintf("UPDATE %s SET %s = ?, %s = NOW() WHERE %s = ?", constants.TableUser, constants.FieldPassword, constants.FieldLastModifiedDate, constants.FieldID)
	_, err = s.db.ExecContext(ctx, updateQuery, newHash, userID)
	if err == nil {
		log.Printf("🔐 Password changed for user: %s", userID)
	}
	return err
}

// GetUserByID retrieves a user session object by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*models.UserSession, error) {
	// Reusing the struct from Login helper basically
	var user struct {
		ID        string
		Username  string
		Email     string
		ProfileId string
		RoleId    sql.NullString
	}

	query := fmt.Sprintf("SELECT %s, %s, %s, %s, %s FROM %s WHERE %s = ? LIMIT 1",
		constants.FieldID, constants.FieldUsername, constants.FieldEmail, constants.FieldProfileID, constants.FieldRoleID,
		constants.TableUser, constants.FieldID)

	err := s.db.QueryRowContext(ctx, query, userID).Scan(&user.ID, &user.Username, &user.Email, &user.ProfileId, &user.RoleId)
	if err != nil {
		return nil, err
	}

	var roleIdPtr *string
	if user.RoleId.Valid {
		roleIdPtr = &user.RoleId.String
	}

	return &models.UserSession{
		ID:        user.ID,
		Name:      user.Username,
		Email:     &user.Email,
		ProfileID: user.ProfileId,
		RoleID:    roleIdPtr,
	}, nil
}

// User management functions are in auth_user_mgmt.go:
// - CreateUserRequest, UpdateUserRequest types
// - CreateUser, UpdateUser, DeleteUser
// - GetUsers, GetProfiles, splitName
