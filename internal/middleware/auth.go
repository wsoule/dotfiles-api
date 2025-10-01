package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"dotfiles-web/pkg/errors"
)

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := getSession(c)
		if session == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": errors.NewUnauthorizedError("authentication required"),
			})
			c.Abort()
			return
		}

		userID, exists := session["user_id"]
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": errors.NewUnauthorizedError("invalid session"),
			})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := getSession(c)
		if session != nil {
			if userID, exists := session["user_id"]; exists {
				c.Set("user_id", userID)
			}
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

func getSession(c *gin.Context) map[string]interface{} {
	sessionCookie, err := c.Request.Cookie("session")
	if err != nil {
		return nil
	}

	if sessionCookie.Value == "" {
		return nil
	}

	return nil
}

func extractBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}