package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"dotfiles-api/pkg/errors"
	"dotfiles-api/internal/auth"
)

// AuthMiddleware holds the session manager
type AuthMiddleware struct {
	sessionManager *auth.SessionManager
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(sessionManager *auth.SessionManager) *AuthMiddleware {
	return &AuthMiddleware{
		sessionManager: sessionManager,
	}
}

// RequireAuth middleware that requires authentication
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, exists := am.sessionManager.GetSessionFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": errors.NewUnauthorizedError("authentication required"),
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", session.UserID)
		c.Set("username", session.Username)
		c.Set("email", session.Email)
		c.Set("session", session)
		c.Next()
	}
}

// OptionalAuth middleware that optionally sets user info if authenticated
func (am *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, exists := am.sessionManager.GetSessionFromContext(c)
		if exists {
			c.Set("user_id", session.UserID)
			c.Set("username", session.Username)
			c.Set("email", session.Email)
			c.Set("session", session)
		}
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": errors.NewForbiddenError("role information not available"),
			})
			c.Abort()
			return
		}

		if userRole != role {
			c.JSON(http.StatusForbidden, gin.H{
				"error": errors.NewForbiddenError("insufficient permissions"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireOrganizationMember(organizationID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": errors.NewUnauthorizedError("authentication required"),
			})
			c.Abort()
			return
		}

		c.Set("organization_id", organizationID)
		c.Next()
	}
}

func RequireOrganizationAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": errors.NewForbiddenError("role information not available"),
			})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok || (role != "owner" && role != "admin") {
			c.JSON(http.StatusForbidden, gin.H{
				"error": errors.NewForbiddenError("admin privileges required"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireOrganizationOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": errors.NewForbiddenError("role information not available"),
			})
			c.Abort()
			return
		}

		if userRole != "owner" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": errors.NewForbiddenError("owner privileges required"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}