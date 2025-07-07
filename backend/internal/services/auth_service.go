package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"
	"wordpress-go-next/backend/pkg/redis"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	GenerateToken() string
	CreateVerificationToken(ctx context.Context, userID uint64, tokenType string) (string, error)
	ValidateVerificationToken(ctx context.Context, token, tokenType string) (*models.VerificationToken, error)
	MarkEmailVerified(ctx context.Context, user *models.User) error
	MarkPhoneVerified(ctx context.Context, user *models.User) error
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) (bool, error)
	ResetUserPassword(ctx context.Context, user *models.User, newPassword string) error
	InvalidateUserAuthCache(ctx context.Context, userID uint64) error
	AuthenticateUser(ctx context.Context, username, password string) (*models.User, error)
	GenerateTokens(user *models.User) (string, string, error)
	RefreshTokens(refreshToken string) (string, string, error)
	IsRateLimited(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type authService struct {
	Redis *redis.RedisService
}

var JWTSecret []byte

func init() {
	JWTSecret = []byte(os.Getenv("JWT_SECRET"))
	if len(JWTSecret) == 0 {
		JWTSecret = []byte("your-very-secret-key") // fallback for dev
	}
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

func (s *authService) getUserAuthCacheKey(userID uint64) string {
	return fmt.Sprintf("%s%d", userAuthCacheKeyPrefix, userID)
}

func (s *authService) GenerateToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func (s *authService) CreateVerificationToken(ctx context.Context, userID uint64, tokenType string) (string, error) {
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
				err := s.Redis.Delete(ctx, cacheKey)
				if err != nil {
					return nil, err
				}
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
		err := s.Redis.Delete(ctx, cacheKey)
		if err != nil {
			return nil, err
		}
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

func (s *authService) InvalidateUserAuthCache(ctx context.Context, userID uint64) error {
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

func (s *authService) invalidateUserAuthCache(ctx context.Context, userID uint64) {
	cacheKey := s.getUserAuthCacheKey(userID)
	err := s.Redis.Delete(ctx, cacheKey)
	if err != nil {
		return
	}
}

// Add stub methods for JWT/refresh token support
func (s *authService) AuthenticateUser(ctx context.Context, username, password string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}
	ok, err := s.VerifyPassword(password, user.PasswordHash)
	if err != nil || !ok {
		return nil, fmt.Errorf("invalid credentials")
	}
	return &user, nil
}

func (s *authService) GenerateTokens(user *models.User) (string, string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(JWTSecret)
	if err != nil {
		return "", "", err
	}
	refreshToken := s.GenerateToken()
	if s.Redis != nil {
		key := "refresh:" + refreshToken
		err := s.Redis.SetWithTTL(context.Background(), key, fmt.Sprintf("%d", user.ID), 7*24*time.Hour)
		if err != nil {
			return "", "", err
		}
	}
	return accessToken, refreshToken, nil
}

func (s *authService) RefreshTokens(refreshToken string) (string, string, error) {
	key := "refresh:" + refreshToken
	userIDStr, err := s.Redis.Get(context.Background(), key)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token")
	}
	err = s.Redis.Delete(context.Background(), key)
	if err != nil {
		return "", "", err
	}
	var user models.User
	if err := database.DB.First(&user, userIDStr).Error; err != nil {
		return "", "", fmt.Errorf("user not found")
	}
	return s.GenerateTokens(&user)
}

func (s *authService) IsRateLimited(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	countStr, _ := s.Redis.Get(ctx, key)
	count := 0
	if countStr != "" {
		count, _ = strconv.Atoi(countStr)
	}
	if count >= limit {
		return true, nil
	}
	if count == 0 {
		_ = s.Redis.SetWithTTL(ctx, key, "1", window)
	} else {
		_ = s.Redis.Incr(ctx, key)
	}
	return false, nil
}

var AuthSvc AuthService = &authService{}
