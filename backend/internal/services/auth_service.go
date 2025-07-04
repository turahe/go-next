package services

import (
	"crypto/rand"
	"encoding/hex"
	"time"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	GenerateToken() string
	CreateVerificationToken(userID uint, tokenType string) (string, error)
	MarkEmailVerified(user *models.User) error
	MarkPhoneVerified(user *models.User) error
	HashPassword(password string) (string, error)
	ResetUserPassword(user *models.User, newPassword string) error
}

type authService struct{}

func (s *authService) GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *authService) CreateVerificationToken(userID uint, tokenType string) (string, error) {
	token := s.GenerateToken()
	t := models.VerificationToken{
		UserID:    userID,
		Token:     token,
		Type:      tokenType,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	if err := database.DB.Create(&t).Error; err != nil {
		return "", err
	}
	return token, nil
}

func (s *authService) MarkEmailVerified(user *models.User) error {
	now := time.Now()
	user.EmailVerified = &now
	return database.DB.Save(user).Error
}

func (s *authService) MarkPhoneVerified(user *models.User) error {
	now := time.Now()
	user.PhoneVerified = &now
	return database.DB.Save(user).Error
}

func (s *authService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (s *authService) ResetUserPassword(user *models.User, newPassword string) error {
	hash, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	return database.DB.Save(user).Error
}

var AuthSvc AuthService = &authService{}
