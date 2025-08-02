package services

import (
	"fmt"

	"go-next/internal/models"
	"go-next/pkg/database"

	"github.com/google/uuid"
)

type TokenCacheService interface {
	// Token caching
	GetTokenByID(id uuid.UUID) (*models.Token, error)
	GetTokenByValue(token string) (*models.Token, error)
	GetTokensByUser(userID uuid.UUID) ([]models.Token, error)
	CacheToken(token *models.Token) error
	InvalidateToken(id uuid.UUID) error
	InvalidateUserTokens(userID uuid.UUID) error

	// JWT Key caching
	GetJWTKeyByID(id uuid.UUID) (*models.JWTKey, error)
	GetJWTKeyByKeyID(keyID string) (*models.JWTKey, error)
	GetActiveJWTKeys() ([]models.JWTKey, error)
	CacheJWTKey(jwtKey *models.JWTKey) error
	InvalidateJWTKey(id uuid.UUID) error
	InvalidateAllJWTKeys() error

	// Verification Token caching
	GetVerificationTokenByID(id uuid.UUID) (*models.VerificationToken, error)
	GetVerificationTokenByValue(token string) (*models.VerificationToken, error)
	GetVerificationTokensByUser(userID uuid.UUID, tokenType models.VerificationTokenType) ([]models.VerificationToken, error)
	CacheVerificationToken(token *models.VerificationToken) error
	InvalidateVerificationToken(id uuid.UUID) error
	InvalidateUserVerificationTokens(userID uuid.UUID, tokenType models.VerificationTokenType) error

	// Refresh Token caching
	GetRefreshTokenByID(id uuid.UUID) (*models.RefreshToken, error)
	GetRefreshTokenByValue(token string) (*models.RefreshToken, error)
	GetRefreshTokensByUser(userID uuid.UUID) ([]models.RefreshToken, error)
	CacheRefreshToken(token *models.RefreshToken) error
	InvalidateRefreshToken(id uuid.UUID) error
	InvalidateUserRefreshTokens(userID uuid.UUID) error
}

type tokenCacheService struct{}

func NewTokenCacheService() TokenCacheService {
	return &tokenCacheService{}
}

// Token caching methods
func (t *tokenCacheService) GetTokenByID(id uuid.UUID) (*models.Token, error) {
	cacheKey := fmt.Sprintf(CacheKeyToken, id.String())

	var token models.Token
	if err := CacheSvc.Get(cacheKey, &token); err == nil {
		return &token, nil
	}

	// Cache miss, get from database
	if err := database.DB.First(&token, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// Cache the result
	t.CacheToken(&token)
	return &token, nil
}

func (t *tokenCacheService) GetTokenByValue(tokenValue string) (*models.Token, error) {
	var token models.Token
	if err := database.DB.First(&token, "token = ?", tokenValue).Error; err != nil {
		return nil, err
	}

	// Cache the result
	t.CacheToken(&token)
	return &token, nil
}

func (t *tokenCacheService) GetTokensByUser(userID uuid.UUID) ([]models.Token, error) {
	cacheKey := fmt.Sprintf(CacheKeyTokens, userID.String())

	var tokens []models.Token
	if err := CacheSvc.Get(cacheKey, &tokens); err == nil {
		return tokens, nil
	}

	// Cache miss, get from database
	if err := database.DB.Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
		return nil, err
	}

	// Cache the result
	CacheSvc.Set(cacheKey, tokens, CacheDurationToken)
	return tokens, nil
}

func (t *tokenCacheService) CacheToken(token *models.Token) error {
	cacheKey := fmt.Sprintf(CacheKeyToken, token.ID.String())
	return CacheSvc.Set(cacheKey, token, CacheDurationToken)
}

func (t *tokenCacheService) InvalidateToken(id uuid.UUID) error {
	cacheKey := fmt.Sprintf(CacheKeyToken, id.String())
	return CacheSvc.Delete(cacheKey)
}

func (t *tokenCacheService) InvalidateUserTokens(userID uuid.UUID) error {
	cacheKey := fmt.Sprintf(CacheKeyTokens, userID.String())
	return CacheSvc.Delete(cacheKey)
}

// JWT Key caching methods
func (t *tokenCacheService) GetJWTKeyByID(id uuid.UUID) (*models.JWTKey, error) {
	cacheKey := fmt.Sprintf(CacheKeyJWTKey, id.String())

	var jwtKey models.JWTKey
	if err := CacheSvc.Get(cacheKey, &jwtKey); err == nil {
		return &jwtKey, nil
	}

	// Cache miss, get from database
	if err := database.DB.First(&jwtKey, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// Cache the result
	t.CacheJWTKey(&jwtKey)
	return &jwtKey, nil
}

func (t *tokenCacheService) GetJWTKeyByKeyID(keyID string) (*models.JWTKey, error) {
	var jwtKey models.JWTKey
	if err := database.DB.First(&jwtKey, "key_id = ?", keyID).Error; err != nil {
		return nil, err
	}

	// Cache the result
	t.CacheJWTKey(&jwtKey)
	return &jwtKey, nil
}

func (t *tokenCacheService) GetActiveJWTKeys() ([]models.JWTKey, error) {
	cacheKey := CacheKeyJWTKeys

	var jwtKeys []models.JWTKey
	if err := CacheSvc.Get(cacheKey, &jwtKeys); err == nil {
		return jwtKeys, nil
	}

	// Cache miss, get from database
	if err := database.DB.Where("is_active = ?", true).Find(&jwtKeys).Error; err != nil {
		return nil, err
	}

	// Cache the result
	CacheSvc.Set(cacheKey, jwtKeys, CacheDurationJWTKey)
	return jwtKeys, nil
}

func (t *tokenCacheService) CacheJWTKey(jwtKey *models.JWTKey) error {
	cacheKey := fmt.Sprintf(CacheKeyJWTKey, jwtKey.ID.String())
	return CacheSvc.Set(cacheKey, jwtKey, CacheDurationJWTKey)
}

func (t *tokenCacheService) InvalidateJWTKey(id uuid.UUID) error {
	cacheKey := fmt.Sprintf(CacheKeyJWTKey, id.String())
	CacheSvc.Delete(cacheKey)
	return CacheSvc.Delete(CacheKeyJWTKeys) // Also invalidate the list cache
}

func (t *tokenCacheService) InvalidateAllJWTKeys() error {
	return CacheSvc.DeletePattern("jwt_key:*")
}

// Verification Token caching methods
func (t *tokenCacheService) GetVerificationTokenByID(id uuid.UUID) (*models.VerificationToken, error) {
	cacheKey := fmt.Sprintf(CacheKeyVerificationToken, id.String())

	var token models.VerificationToken
	if err := CacheSvc.Get(cacheKey, &token); err == nil {
		return &token, nil
	}

	// Cache miss, get from database
	if err := database.DB.First(&token, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// Cache the result
	t.CacheVerificationToken(&token)
	return &token, nil
}

func (t *tokenCacheService) GetVerificationTokenByValue(tokenValue string) (*models.VerificationToken, error) {
	var token models.VerificationToken
	if err := database.DB.First(&token, "token = ?", tokenValue).Error; err != nil {
		return nil, err
	}

	// Cache the result
	t.CacheVerificationToken(&token)
	return &token, nil
}

func (t *tokenCacheService) GetVerificationTokensByUser(userID uuid.UUID, tokenType models.VerificationTokenType) ([]models.VerificationToken, error) {
	cacheKey := fmt.Sprintf(CacheKeyVerificationTokens, userID.String(), string(tokenType))

	var tokens []models.VerificationToken
	if err := CacheSvc.Get(cacheKey, &tokens); err == nil {
		return tokens, nil
	}

	// Cache miss, get from database
	if err := database.DB.Where("user_id = ? AND type = ?", userID, tokenType).Find(&tokens).Error; err != nil {
		return nil, err
	}

	// Cache the result
	CacheSvc.Set(cacheKey, tokens, CacheDurationToken)
	return tokens, nil
}

func (t *tokenCacheService) CacheVerificationToken(token *models.VerificationToken) error {
	cacheKey := fmt.Sprintf(CacheKeyVerificationToken, token.ID.String())
	return CacheSvc.Set(cacheKey, token, CacheDurationToken)
}

func (t *tokenCacheService) InvalidateVerificationToken(id uuid.UUID) error {
	cacheKey := fmt.Sprintf(CacheKeyVerificationToken, id.String())
	return CacheSvc.Delete(cacheKey)
}

func (t *tokenCacheService) InvalidateUserVerificationTokens(userID uuid.UUID, tokenType models.VerificationTokenType) error {
	cacheKey := fmt.Sprintf(CacheKeyVerificationTokens, userID.String(), string(tokenType))
	return CacheSvc.Delete(cacheKey)
}

// Refresh Token caching methods
func (t *tokenCacheService) GetRefreshTokenByID(id uuid.UUID) (*models.RefreshToken, error) {
	cacheKey := fmt.Sprintf(CacheKeyRefreshToken, id.String())

	var token models.RefreshToken
	if err := CacheSvc.Get(cacheKey, &token); err == nil {
		return &token, nil
	}

	// Cache miss, get from database
	if err := database.DB.First(&token, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// Cache the result
	t.CacheRefreshToken(&token)
	return &token, nil
}

func (t *tokenCacheService) GetRefreshTokenByValue(tokenValue string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	if err := database.DB.First(&token, "token = ?", tokenValue).Error; err != nil {
		return nil, err
	}

	// Cache the result
	t.CacheRefreshToken(&token)
	return &token, nil
}

func (t *tokenCacheService) GetRefreshTokensByUser(userID uuid.UUID) ([]models.RefreshToken, error) {
	cacheKey := fmt.Sprintf(CacheKeyRefreshTokens, userID.String())

	var tokens []models.RefreshToken
	if err := CacheSvc.Get(cacheKey, &tokens); err == nil {
		return tokens, nil
	}

	// Cache miss, get from database
	if err := database.DB.Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
		return nil, err
	}

	// Cache the result
	CacheSvc.Set(cacheKey, tokens, CacheDurationToken)
	return tokens, nil
}

func (t *tokenCacheService) CacheRefreshToken(token *models.RefreshToken) error {
	cacheKey := fmt.Sprintf(CacheKeyRefreshToken, token.ID.String())
	return CacheSvc.Set(cacheKey, token, CacheDurationToken)
}

func (t *tokenCacheService) InvalidateRefreshToken(id uuid.UUID) error {
	cacheKey := fmt.Sprintf(CacheKeyRefreshToken, id.String())
	return CacheSvc.Delete(cacheKey)
}

func (t *tokenCacheService) InvalidateUserRefreshTokens(userID uuid.UUID) error {
	cacheKey := fmt.Sprintf(CacheKeyRefreshTokens, userID.String())
	return CacheSvc.Delete(cacheKey)
}

var TokenCacheSvc TokenCacheService = NewTokenCacheService()
