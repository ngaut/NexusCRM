package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/pkg/auth"
	"github.com/nexuscrm/shared/pkg/constants"
)

// RequireAuth is a middleware that validates JWT tokens
func RequireAuth(authSvc *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader(constants.HeaderAuthorization)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				constants.ResponseError: "Unauthorized",
				constants.FieldMessage:  "No authorization token provided",
				"code":                  "UNAUTHORIZED",
				"data":                  nil,
			})
			c.Abort()
			return
		}

		// Extract token (format: "Bearer <token>")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				constants.ResponseError: "Unauthorized",
				constants.FieldMessage:  "Invalid authorization header format",
				"code":                  "UNAUTHORIZED",
				"data":                  nil,
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token and session via AuthService
		claims, err := authSvc.ValidateSession(c.Request.Context(), tokenString)
		if err != nil {
			// Determine status code based on error type?
			// For now, 401 is safe for all session failures
			c.JSON(http.StatusUnauthorized, gin.H{
				constants.ResponseError: "Unauthorized",
				constants.FieldMessage:  err.Error(),
				"code":                  "UNAUTHORIZED",
				"data":                  nil,
			})
			c.Abort()
			return
		}

		// Update last activity (Fire and forget)
		authSvc.TouchSession(claims.RegisteredClaims.ID)

		// Set user session in context
		c.Set(constants.ContextKeyUser, claims.User)
		c.Set("token", tokenString)

		c.Next()
	}
}

// RequireSystemAdmin checks if the user is a system administrator
func RequireSystemAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get(constants.ContextKeyUser)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				constants.ResponseError: "Unauthorized",
				constants.FieldMessage:  "User not authenticated",
				"code":                  "UNAUTHORIZED",
				"data":                  nil,
			})
			c.Abort()
			return
		}

		user := userInterface.(auth.UserSession)
		if !user.IsSuperUser() {
			c.JSON(http.StatusForbidden, gin.H{
				constants.ResponseError: "Forbidden",
				constants.FieldMessage:  "Only System Administrators can access this resource",
				"code":                  "FORBIDDEN",
				"data":                  nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
