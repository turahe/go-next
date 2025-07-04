package middleware

import (
	"net/http"
	"strings"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
			userID, ok := claims["user_id"].(float64)
			if !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			var jwtKey models.JWTKey
			if err := database.DB.Where("user_id = ?", uint(userID)).First(&jwtKey).Error; err != nil {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtKey.SecretKey), nil
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
		c.Set("user_id", uint(claims["user_id"].(float64)))
		c.Next()
	}
}
