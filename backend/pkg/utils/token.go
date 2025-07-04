package utils

import (
	"crypto/rand"
	"encoding/hex"
	"time"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateJWT creates a JWT for a given userID using their secret key and expiration from the DB
func GenerateJWT(userID uint) (string, error) {
	var jwtKey models.JWTKey
	err := database.DB.Where("user_id = ?", userID).First(&jwtKey).Error
	if err != nil {
		return "", err
	}
	exp := time.Now().Add(time.Duration(jwtKey.TokenExpiration) * time.Second).Unix()
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     exp,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtKey.SecretKey))
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
