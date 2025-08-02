package services

import (
	"crypto/rand"
	"encoding/hex"
	"go-next/internal/models"
	"go-next/pkg/database"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	GenerateToken() string
	CreateVerificationToken(userID uuid.UUID, tokenType string) (string, error)
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

func (s *authService) CreateVerificationToken(userID uuid.UUID, tokenType string) (string, error) {
	token := s.GenerateToken()
	t := models.VerificationToken{
		UserID:    userID,
		Token:     token,
		Type:      models.VerificationTokenType(tokenType),
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	if err := database.DB.Create(&t).Error; err != nil {
		return "", err
	}

	// Cache the verification token
	TokenCacheSvc.CacheVerificationToken(&t)

	// Invalidate user's verification tokens cache for this type
	TokenCacheSvc.InvalidateUserVerificationTokens(userID, t.Type)

	return token, nil
}

func (s *authService) MarkEmailVerified(user *models.User) error {
	now := time.Now()
	user.EmailVerified = &now
	if err := database.DB.Save(user).Error; err != nil {
		return err
	}

	// Invalidate any cached verification tokens for this user
	TokenCacheSvc.InvalidateUserVerificationTokens(user.ID, models.EmailVerification)

	return nil
}

func (s *authService) MarkPhoneVerified(user *models.User) error {
	now := time.Now()
	user.PhoneVerified = &now
	if err := database.DB.Save(user).Error; err != nil {
		return err
	}

	// Invalidate any cached verification tokens for this user
	TokenCacheSvc.InvalidateUserVerificationTokens(user.ID, models.PhoneVerification)

	return nil
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
	if err := database.DB.Save(user).Error; err != nil {
		return err
	}

	// Invalidate any cached password reset tokens for this user
	TokenCacheSvc.InvalidateUserVerificationTokens(user.ID, models.PasswordReset)

	return nil
}

var AuthSvc AuthService = &authService{}
