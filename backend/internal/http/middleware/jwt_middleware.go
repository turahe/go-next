package middleware

import (
	"go-next/internal/models"
	"go-next/internal/services"
	"go-next/pkg/database"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			return
		}
		tokenStr := header[7:]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || claims["user_id"] == nil {
				return nil, jwt.ErrSignatureInvalid
			}
			if _, ok := claims["user_id"].(string); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			// Try to get JWT key from cache first
			jwtKeys, err := services.TokenCacheSvc.GetActiveJWTKeys()
			if err != nil || len(jwtKeys) == 0 {
				// Cache miss, get from database
				var jwtKey models.JWTKey
				if err := database.DB.Where("is_active = ?", true).First(&jwtKey).Error; err != nil {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtKey.Key), nil
			}
			return []byte(jwtKeys[0].Key), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["user_id"] == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			return
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}
