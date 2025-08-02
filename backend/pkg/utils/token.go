package utils

import (
	"crypto/rand"
	"encoding/hex"
	"go-next/internal/models"
	"go-next/internal/services"
	"go-next/pkg/database"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// GenerateJWT creates a JWT for a given userID using their secret key and expiration from the DB
func GenerateJWT(userID uuid.UUID) (string, error) {
	// Try to get JWT key from cache first
	jwtKeys, err := services.TokenCacheSvc.GetActiveJWTKeys()
	if err != nil || len(jwtKeys) == 0 {
		// Cache miss, get from database
		var jwtKey models.JWTKey
		err := database.DB.Where("is_active = ?", true).First(&jwtKey).Error
		if err != nil {
			return "", err
		}
		// Cache the JWT key
		services.TokenCacheSvc.CacheJWTKey(&jwtKey)

		// Default expiration of 1 hour
		exp := time.Now().Add(time.Hour).Unix()
		claims := jwt.MapClaims{
			"user_id": userID.String(),
			"exp":     exp,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return token.SignedString([]byte(jwtKey.Key))
	}

	// Use cached JWT key
	jwtKey := jwtKeys[0]
	// Default expiration of 1 hour
	exp := time.Now().Add(time.Hour).Unix()
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     exp,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtKey.Key))
}

// GenerateRandomKey creates a secure random hex string of the given length
func GenerateRandomKey(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
