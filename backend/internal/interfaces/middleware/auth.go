package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/pkg/auth"
)

// RequireAuth is a middleware that validates JWT tokens
func RequireAuth(authSvc *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "No authorization token provided",
			})
			c.Abort()
			return
		}

		// Extract token (format: "Bearer <token>")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token and session via AuthService
		claims, err := authSvc.ValidateSession(tokenString)
		if err != nil {
			// Determine status code based on error type?
			// For now, 401 is safe for all session failures
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Update last activity (Fire and forget)
		authSvc.TouchSession(claims.RegisteredClaims.ID)

		// Set user session in context
		c.Set("user", claims.User)
		c.Set("token", tokenString)

		c.Next()
	}
}

// RequireSystemAdmin checks if the user is a system administrator
func RequireSystemAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "User not authenticated",
			})
			c.Abort()
			return
		}

		user := userInterface.(auth.UserSession)
		if !user.IsSuperUser() {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Only System Administrators can access this resource",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
