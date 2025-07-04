package middleware

import (
	"net/http"
	"strings"
	"wordpress-go-next/backend/pkg/database"

	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/internal/services"

	"github.com/gin-gonic/gin"
)

func CasbinMiddleware(obj string, act string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		// Get the user's role name for Casbin enforcement
		var role string
		if len(user.Roles) > 0 {
			role = user.Roles[0].Name
		} else {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User has no role assigned"})
			return
		}
		// Normalize path for policy matching
		path := obj
		if strings.Contains(obj, ":") {
			// Remove :id or :param for policy
			path = obj[:strings.Index(obj, ":")-1]
		}
		allowed, err := services.Enforcer.Enforce(role, path, act)
		if err != nil || !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		c.Next()
	}
}
