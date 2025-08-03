package middleware

import (
	"go-next/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CasbinMiddleware(obj string, act string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Convert userID to UUID
		userUUID, ok := userID.(uuid.UUID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			return
		}

		// Normalize path for policy matching
		path := obj
		if strings.Contains(obj, ":") {
			// Remove :id or :param for policy matching
			path = obj[:strings.Index(obj, ":")]
		}

		// Use Casbin service to check permissions
		casbinService := services.NewCasbinService()
		allowed, err := casbinService.Enforce(userUUID, path, act)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Authorization check failed"})
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		c.Next()
	}
}

// CasbinMiddlewareWithDomain checks permissions with domain context
func CasbinMiddlewareWithDomain(obj string, act string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Convert userID to UUID
		userUUID, ok := userID.(uuid.UUID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			return
		}

		// Get domain from header or query parameter
		domain := c.GetHeader("X-Domain")
		if domain == "" {
			domain = c.Query("domain")
		}
		if domain == "" {
			// Fallback to global permissions
			CasbinMiddleware(obj, act)(c)
			return
		}

		// Normalize path for policy matching
		path := obj
		if strings.Contains(obj, ":") {
			// Remove :id or :param for policy matching
			path = obj[:strings.Index(obj, ":")]
		}

		// Use Casbin service to check permissions with domain context
		casbinService := services.NewCasbinService()
		allowed, err := casbinService.EnforceWithDomain(userUUID, path, act, domain)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Authorization check failed"})
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		c.Next()
	}
}
