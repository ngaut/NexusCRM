package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/backend/pkg/utils"
)

// UserSession represents the user session data stored in JWT
type UserSession struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	ProfileId string  `json:"profile_id"`        // Required: User's permissions profile
	RoleId    *string `json:"role_id,omitempty"` // Optional: Role for hierarchy-based data sharing (Salesforce pattern)
}

// IsSuperUser checks if the user has super user privileges
func (u UserSession) IsSuperUser() bool {
	return constants.IsSuperUser(u.ProfileId)
}

// Claims represents JWT claims
type Claims struct {
	User UserSession `json:"user"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-in-production"
	}
	return secret
}

// GenerateToken creates a JWT token for a user session
func GenerateToken(session UserSession) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	jti := utils.GenerateID()

	claims := &Claims{
		User: session,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// DecodeToken decodes a token without validation (for extracting JTI)
func DecodeToken(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// ToMap converts UserSession to a map for formula context
func (u UserSession) ToMap() map[string]interface{} {
	return map[string]interface{}{
		constants.FieldID:   u.ID,
		constants.FieldName: u.Name,
		"email":             u.Email,
		"profile_id":        u.ProfileId,
		// Handle nil pointer safely for optional fields
		"role_id": func() string {
			if u.RoleId != nil {
				return *u.RoleId
			}
			return ""
		}(),
	}
}
