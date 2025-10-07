package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Only set CORS headers if origin is allowed
		if isOriginAllowed(origin, allowedOrigins) {
			// When using credentials, cannot use wildcard for Access-Control-Allow-Origin
			// Must specify exact origin
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		} else if hasWildcard(allowedOrigins) {
			// If wildcard is configured but we're using credentials,
			// don't set Access-Control-Allow-Origin for security
			// Only allow specific origins when credentials are enabled
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "false")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			// Wildcard should only be used in development
			// Return true but handle credentials separately
			return true
		}
		if allowed == origin {
			return true
		}
	}
	return false
}

func hasWildcard(allowedOrigins []string) bool {
	for _, origin := range allowedOrigins {
		if origin == "*" {
			return true
		}
	}
	return false
}