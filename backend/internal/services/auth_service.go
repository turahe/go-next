package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"
	"wordpress-go-next/backend/pkg/redis"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	GenerateToken() string
	CreateVerificationToken(ctx context.Context, userID uint, tokenType string) (string, error)
	ValidateVerificationToken(ctx context.Context, token, tokenType string) (*models.VerificationToken, error)
	MarkEmailVerified(ctx context.Context, user *models.User) error
	MarkPhoneVerified(ctx context.Context, user *models.User) error
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) (bool, error)
	ResetUserPassword(ctx context.Context, user *models.User, newPassword string) error
	InvalidateUserAuthCache(ctx context.Context, userID uint) error
}

type authService struct {
	Redis *redis.RedisService
}

func NewAuthService(redisService *redis.RedisService) AuthService {
	return &authService{
		Redis: redisService,
	}
}

// Cache keys
const (
	verificationTokenCacheKeyPrefix = "verification_token:"
	userAuthCacheKeyPrefix          = "user_auth:"
)

func (s *authService) getVerificationTokenCacheKey(token string) string {
	return fmt.Sprintf("%s%s", verificationTokenCacheKeyPrefix, token)
}

func (s *authService) getUserAuthCacheKey(userID uint) string {
	return fmt.Sprintf("%s%d", userAuthCacheKeyPrefix, userID)
}

func (s *authService) GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *authService) CreateVerificationToken(ctx context.Context, userID uint, tokenType string) (string, error) {
	token := s.GenerateToken()
	t := models.VerificationToken{
		UserID:    userID,
		Token:     token,
		Type:      tokenType,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	if err := database.DB.WithContext(ctx).Create(&t).Error; err != nil {
		return "", fmt.Errorf("failed to create verification token: %w", err)
	}

	// Cache the token
	if err := s.cacheVerificationToken(ctx, &t); err != nil {
		fmt.Printf("Warning: failed to cache verification token: %v\n", err)
	}

	return token, nil
}

func (s *authService) ValidateVerificationToken(ctx context.Context, token, tokenType string) (*models.VerificationToken, error) {
	cacheKey := s.getVerificationTokenCacheKey(token)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var vt models.VerificationToken
		if err := json.Unmarshal([]byte(cached), &vt); err == nil {
			// Check if token is expired
			if time.Now().After(vt.ExpiresAt) {
				s.Redis.Delete(ctx, cacheKey)
				return nil, fmt.Errorf("verification token expired")
			}
			return &vt, nil
		}
	}

	var vt models.VerificationToken
	err := database.DB.WithContext(ctx).Where("token = ? AND type = ?", token, tokenType).First(&vt).Error
	if err != nil {
		return nil, fmt.Errorf("invalid verification token: %w", err)
	}

	// Check if token is expired
	if time.Now().After(vt.ExpiresAt) {
		// Delete expired token from database
		database.DB.WithContext(ctx).Delete(&vt)
		s.Redis.Delete(ctx, cacheKey)
		return nil, fmt.Errorf("verification token expired")
	}

	// Cache the token
	if err := s.cacheVerificationToken(ctx, &vt); err != nil {
		fmt.Printf("Warning: failed to cache verification token: %v\n", err)
	}

	return &vt, nil
}

func (s *authService) MarkEmailVerified(ctx context.Context, user *models.User) error {
	now := time.Now()
	user.EmailVerified = &now
	if err := database.DB.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("failed to mark email verified: %w", err)
	}

	// Invalidate user auth cache
	s.invalidateUserAuthCache(ctx, user.ID)

	return nil
}

func (s *authService) MarkPhoneVerified(ctx context.Context, user *models.User) error {
	now := time.Now()
	user.PhoneVerified = &now
	if err := database.DB.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("failed to mark phone verified: %w", err)
	}

	// Invalidate user auth cache
	s.invalidateUserAuthCache(ctx, user.ID)

	return nil
}

func (s *authService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), err
}

func (s *authService) VerifyPassword(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, fmt.Errorf("failed to verify password: %w", err)
	}
	return true, nil
}

func (s *authService) ResetUserPassword(ctx context.Context, user *models.User, newPassword string) error {
	hash, err := s.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}
	user.PasswordHash = hash
	if err := database.DB.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("failed to reset user password: %w", err)
	}

	// Invalidate user auth cache
	s.invalidateUserAuthCache(ctx, user.ID)

	return nil
}

func (s *authService) InvalidateUserAuthCache(ctx context.Context, userID uint) error {
	s.invalidateUserAuthCache(ctx, userID)
	return nil
}

// Helper methods
func (s *authService) cacheVerificationToken(ctx context.Context, token *models.VerificationToken) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	cacheKey := s.getVerificationTokenCacheKey(token.Token)
	ttl := time.Until(token.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("token already expired")
	}

	return s.Redis.SetWithTTL(ctx, cacheKey, string(data), ttl)
}

func (s *authService) invalidateUserAuthCache(ctx context.Context, userID uint) {
	cacheKey := s.getUserAuthCacheKey(userID)
	s.Redis.Delete(ctx, cacheKey)
}

var AuthSvc AuthService = &authService{}
