// Package services provides business logic layer for the blog application.
// This package contains all service interfaces and implementations that handle
// the core business logic, data processing, and external service interactions.
package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go-next/internal/models"
	"go-next/pkg/redis"

	"github.com/google/uuid"
)

// TokenService defines the interface for all token-related business operations.
// This interface provides methods for managing refresh tokens, API tokens,
// and verification tokens using Redis for storage.
type TokenService interface {
	// Refresh token operations - Methods for managing refresh tokens
	StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time, ipAddress, userAgent string) error
	GetRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error
	IsRefreshTokenValid(ctx context.Context, token string) (bool, error)

	// API token operations - Methods for managing API tokens
	StoreAPIToken(ctx context.Context, userID uuid.UUID, token, tokenType string, expiresAt *time.Time, ipAddress, userAgent string) error
	GetAPIToken(ctx context.Context, token string) (*models.Token, error)
	DeleteAPIToken(ctx context.Context, token string) error
	UpdateAPITokenLastUsed(ctx context.Context, token string) error
	IsAPITokenValid(ctx context.Context, token string) (bool, error)

	// Verification token operations - Methods for managing verification tokens
	StoreVerificationToken(ctx context.Context, userID uuid.UUID, token string, tokenType models.VerificationTokenType, expiresAt time.Time, ipAddress, userAgent string) error
	GetVerificationToken(ctx context.Context, token string) (*models.VerificationToken, error)
	DeleteVerificationToken(ctx context.Context, token string) error
	MarkVerificationTokenAsUsed(ctx context.Context, token string) error
	IsVerificationTokenValid(ctx context.Context, token string) (bool, error)

	// Utility methods - Helper functions for token management
	CleanupExpiredTokens(ctx context.Context) error
	GetUserTokens(ctx context.Context, userID uuid.UUID) ([]string, error)
}

// tokenService implements the TokenService interface.
// This struct holds the Redis client and provides the actual implementation
// of all token-related business logic using Redis for storage.
type tokenService struct {
	redis *redis.RedisService // Redis client for token storage
}

// NewTokenService creates and returns a new instance of TokenService.
// This factory function initializes the service with the global Redis client.
func NewTokenService() TokenService {
	return &tokenService{redis: GlobalRedisClient}
}

// generateRedisKey creates a Redis key for token storage with proper prefixing.
// This helper function ensures consistent key naming across the application.
func (s *tokenService) generateRedisKey(tokenType, token string) string {
	return "token:" + tokenType + ":" + token
}

// generateUserKey creates a Redis key for user-specific token collections.
// This helper function ensures consistent key naming for user token management.
func (s *tokenService) generateUserKey(userID uuid.UUID, tokenType string) string {
	return "user:" + userID.String() + ":" + tokenType
}

// StoreRefreshToken stores a refresh token in Redis with expiration.
// The token is stored with metadata including user ID, expiration, and device info.
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: ID of the user the token belongs to
//   - token: The refresh token string
//   - expiresAt: When the token expires
//   - ipAddress: IP address of the device that created the token
//   - userAgent: User agent string of the device
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := tokenService.StoreRefreshToken(ctx, userID, "refresh_token_123", time.Now().AddDate(0, 0, 30), "192.168.1.1", "Mozilla/5.0...")
//	if err != nil {
//	    // Handle error
//	}
func (s *tokenService) StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time, ipAddress, userAgent string) error {
	// Create refresh token model for storage
	refreshToken := &models.RefreshToken{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		IsActive:  true,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	// Serialize the token data
	tokenData, err := json.Marshal(refreshToken)
	if err != nil {
		return err
	}

	// Calculate TTL for Redis
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return errors.New("token expiration time has passed")
	}

	// Store in Redis with expiration
	key := s.generateRedisKey("refresh", token)
	if err := s.redis.SetWithExpiration(ctx, key, string(tokenData), ttl); err != nil {
		return err
	}

	// Add to user's refresh token collection
	userKey := s.generateUserKey(userID, "refresh_tokens")
	userTokens, _ := s.redis.Get(ctx, userKey)
	var tokens []string
	if userTokens != "" {
		json.Unmarshal([]byte(userTokens), &tokens)
	}
	tokens = append(tokens, token)

	userTokensData, _ := json.Marshal(tokens)
	s.redis.SetWithExpiration(ctx, userKey, string(userTokensData), ttl)

	return nil
}

// GetRefreshToken retrieves a refresh token from Redis.
// Returns the token data if found and not expired.
//
// Parameters:
//   - ctx: Context for the operation
//   - token: The refresh token string to retrieve
//
// Returns:
//   - *models.RefreshToken: The refresh token data or nil if not found
//   - error: Any error encountered during the operation
//
// Example:
//
//	refreshToken, err := tokenService.GetRefreshToken(ctx, "refresh_token_123")
//	if err != nil {
//	    // Handle error (token not found, expired, etc.)
//	}
//	if refreshToken != nil {
//	    // Use the token data
//	}
func (s *tokenService) GetRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	key := s.generateRedisKey("refresh", token)
	tokenData, err := s.redis.Get(ctx, key)
	if err != nil {
		return nil, errors.New("refresh token not found")
	}

	var refreshToken models.RefreshToken
	if err := json.Unmarshal([]byte(tokenData), &refreshToken); err != nil {
		return nil, err
	}

	// Check if token is expired
	if refreshToken.IsExpired() {
		// Clean up expired token
		s.redis.Delete(ctx, key)
		return nil, errors.New("refresh token expired")
	}

	return &refreshToken, nil
}

// DeleteRefreshToken removes a refresh token from Redis.
// This is typically called during logout or token invalidation.
//
// Parameters:
//   - ctx: Context for the operation
//   - token: The refresh token string to delete
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := tokenService.DeleteRefreshToken(ctx, "refresh_token_123")
//	if err != nil {
//	    // Handle error
//	}
func (s *tokenService) DeleteRefreshToken(ctx context.Context, token string) error {
	// Get token data first to remove from user's collection
	refreshToken, err := s.GetRefreshToken(ctx, token)
	if err == nil && refreshToken != nil {
		// Remove from user's token collection
		userKey := s.generateUserKey(refreshToken.UserID, "refresh_tokens")
		userTokens, _ := s.redis.Get(ctx, userKey)
		if userTokens != "" {
			var tokens []string
			json.Unmarshal([]byte(userTokens), &tokens)

			// Remove the specific token
			for i, t := range tokens {
				if t == token {
					tokens = append(tokens[:i], tokens[i+1:]...)
					break
				}
			}

			if len(tokens) > 0 {
				userTokensData, _ := json.Marshal(tokens)
				s.redis.Set(ctx, userKey, string(userTokensData))
			} else {
				s.redis.Delete(ctx, userKey)
			}
		}
	}

	// Delete the token
	key := s.generateRedisKey("refresh", token)
	return s.redis.Delete(ctx, key)
}

// DeleteUserRefreshTokens removes all refresh tokens for a specific user.
// This is typically called when a user changes their password or account is compromised.
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: ID of the user whose tokens should be deleted
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := tokenService.DeleteUserRefreshTokens(ctx, userID)
//	if err != nil {
//	    // Handle error
//	}
func (s *tokenService) DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	// Get all user's refresh tokens
	userKey := s.generateUserKey(userID, "refresh_tokens")
	userTokens, err := s.redis.Get(ctx, userKey)
	if err != nil {
		return nil // No tokens to delete
	}

	var tokens []string
	if err := json.Unmarshal([]byte(userTokens), &tokens); err != nil {
		return err
	}

	// Delete each token
	for _, token := range tokens {
		key := s.generateRedisKey("refresh", token)
		s.redis.Delete(ctx, key)
	}

	// Delete user's token collection
	return s.redis.Delete(ctx, userKey)
}

// IsRefreshTokenValid checks if a refresh token is valid and not expired.
// This is a convenience method that combines retrieval and validation.
//
// Parameters:
//   - ctx: Context for the operation
//   - token: The refresh token string to validate
//
// Returns:
//   - bool: True if token is valid, false otherwise
//   - error: Any error encountered during the operation
//
// Example:
//
//	isValid, err := tokenService.IsRefreshTokenValid(ctx, "refresh_token_123")
//	if err != nil {
//	    // Handle error
//	}
//	if isValid {
//	    // Token is valid, proceed with authentication
//	}
func (s *tokenService) IsRefreshTokenValid(ctx context.Context, token string) (bool, error) {
	refreshToken, err := s.GetRefreshToken(ctx, token)
	if err != nil {
		return false, err
	}
	return refreshToken.IsValid(), nil
}

// StoreAPIToken stores an API token in Redis with expiration.
// The token is stored with metadata including user ID, type, and device info.
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: ID of the user the token belongs to
//   - token: The API token string
//   - tokenType: Type of token (access, refresh, etc.)
//   - expiresAt: When the token expires (can be nil for non-expiring tokens)
//   - ipAddress: IP address of the device that created the token
//   - userAgent: User agent string of the device
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := tokenService.StoreAPIToken(ctx, userID, "api_token_123", "access", &expiresAt, "192.168.1.1", "Mozilla/5.0...")
//	if err != nil {
//	    // Handle error
//	}
func (s *tokenService) StoreAPIToken(ctx context.Context, userID uuid.UUID, token, tokenType string, expiresAt *time.Time, ipAddress, userAgent string) error {
	// Create API token model for storage
	apiToken := &models.Token{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserID:    userID,
		Token:     token,
		Type:      tokenType,
		ExpiredAt: expiresAt,
		IsActive:  true,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	// Serialize the token data
	tokenData, err := json.Marshal(apiToken)
	if err != nil {
		return err
	}

	// Calculate TTL for Redis
	var ttl time.Duration
	if expiresAt != nil {
		ttl = time.Until(*expiresAt)
		if ttl <= 0 {
			return errors.New("token expiration time has passed")
		}
	}

	// Store in Redis
	key := s.generateRedisKey("api", token)
	if ttl > 0 {
		if err := s.redis.SetWithExpiration(ctx, key, string(tokenData), ttl); err != nil {
			return err
		}
	} else {
		if err := s.redis.Set(ctx, key, string(tokenData)); err != nil {
			return err
		}
	}

	// Add to user's API token collection
	userKey := s.generateUserKey(userID, "api_tokens")
	userTokens, _ := s.redis.Get(ctx, userKey)
	var tokens []string
	if userTokens != "" {
		json.Unmarshal([]byte(userTokens), &tokens)
	}
	tokens = append(tokens, token)

	userTokensData, _ := json.Marshal(tokens)
	if ttl > 0 {
		s.redis.SetWithExpiration(ctx, userKey, string(userTokensData), ttl)
	} else {
		s.redis.Set(ctx, userKey, string(userTokensData))
	}

	return nil
}

// GetAPIToken retrieves an API token from Redis.
// Returns the token data if found and not expired.
//
// Parameters:
//   - ctx: Context for the operation
//   - token: The API token string to retrieve
//
// Returns:
//   - *models.Token: The API token data or nil if not found
//   - error: Any error encountered during the operation
//
// Example:
//
//	apiToken, err := tokenService.GetAPIToken(ctx, "api_token_123")
//	if err != nil {
//	    // Handle error (token not found, expired, etc.)
//	}
//	if apiToken != nil {
//	    // Use the token data
//	}
func (s *tokenService) GetAPIToken(ctx context.Context, token string) (*models.Token, error) {
	key := s.generateRedisKey("api", token)
	tokenData, err := s.redis.Get(ctx, key)
	if err != nil {
		return nil, errors.New("API token not found")
	}

	var apiToken models.Token
	if err := json.Unmarshal([]byte(tokenData), &apiToken); err != nil {
		return nil, err
	}

	// Check if token is expired
	if apiToken.IsExpired() {
		// Clean up expired token
		s.redis.Delete(ctx, key)
		return nil, errors.New("API token expired")
	}

	return &apiToken, nil
}

// DeleteAPIToken removes an API token from Redis.
// This is typically called during logout or token invalidation.
//
// Parameters:
//   - ctx: Context for the operation
//   - token: The API token string to delete
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := tokenService.DeleteAPIToken(ctx, "api_token_123")
//	if err != nil {
//	    // Handle error
//	}
func (s *tokenService) DeleteAPIToken(ctx context.Context, token string) error {
	// Get token data first to remove from user's collection
	apiToken, err := s.GetAPIToken(ctx, token)
	if err == nil && apiToken != nil {
		// Remove from user's token collection
		userKey := s.generateUserKey(apiToken.UserID, "api_tokens")
		userTokens, _ := s.redis.Get(ctx, userKey)
		if userTokens != "" {
			var tokens []string
			json.Unmarshal([]byte(userTokens), &tokens)

			// Remove the specific token
			for i, t := range tokens {
				if t == token {
					tokens = append(tokens[:i], tokens[i+1:]...)
					break
				}
			}

			if len(tokens) > 0 {
				userTokensData, _ := json.Marshal(tokens)
				s.redis.Set(ctx, userKey, string(userTokensData))
			} else {
				s.redis.Delete(ctx, userKey)
			}
		}
	}

	// Delete the token
	key := s.generateRedisKey("api", token)
	return s.redis.Delete(ctx, key)
}

// UpdateAPITokenLastUsed updates the last used timestamp for an API token.
// This is typically called when a token is used for authentication.
//
// Parameters:
//   - ctx: Context for the operation
//   - token: The API token string to update
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := tokenService.UpdateAPITokenLastUsed(ctx, "api_token_123")
//	if err != nil {
//	    // Handle error
//	}
func (s *tokenService) UpdateAPITokenLastUsed(ctx context.Context, token string) error {
	apiToken, err := s.GetAPIToken(ctx, token)
	if err != nil {
		return err
	}

	// Update last used timestamp
	now := time.Now()
	apiToken.LastUsedAt = &now
	apiToken.UpdatedAt = now

	// Serialize and store back
	tokenData, err := json.Marshal(apiToken)
	if err != nil {
		return err
	}

	key := s.generateRedisKey("api", token)
	return s.redis.Set(ctx, key, string(tokenData))
}

// IsAPITokenValid checks if an API token is valid and not expired.
// This is a convenience method that combines retrieval and validation.
//
// Parameters:
//   - ctx: Context for the operation
//   - token: The API token string to validate
//
// Returns:
//   - bool: True if token is valid, false otherwise
//   - error: Any error encountered during the operation
//
// Example:
//
//	isValid, err := tokenService.IsAPITokenValid(ctx, "api_token_123")
//	if err != nil {
//	    // Handle error
//	}
//	if isValid {
//	    // Token is valid, proceed with authentication
//	}
func (s *tokenService) IsAPITokenValid(ctx context.Context, token string) (bool, error) {
	apiToken, err := s.GetAPIToken(ctx, token)
	if err != nil {
		return false, err
	}
	return apiToken.IsValid(), nil
}

// StoreVerificationToken stores a verification token in Redis with expiration.
// The token is stored with metadata including user ID, type, and device info.
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: ID of the user the token belongs to
//   - token: The verification token string
//   - tokenType: Type of verification token (email, phone, password reset, etc.)
//   - expiresAt: When the token expires
//   - ipAddress: IP address of the device that created the token
//   - userAgent: User agent string of the device
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := tokenService.StoreVerificationToken(ctx, userID, "verify_token_123", models.EmailVerification, time.Now().Add(time.Hour), "192.168.1.1", "Mozilla/5.0...")
//	if err != nil {
//	    // Handle error
//	}
func (s *tokenService) StoreVerificationToken(ctx context.Context, userID uuid.UUID, token string, tokenType models.VerificationTokenType, expiresAt time.Time, ipAddress, userAgent string) error {
	// Create verification token model for storage
	verificationToken := &models.VerificationToken{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserID:    userID,
		Token:     token,
		Type:      tokenType,
		ExpiresAt: expiresAt,
		Used:      false,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	// Serialize the token data
	tokenData, err := json.Marshal(verificationToken)
	if err != nil {
		return err
	}

	// Calculate TTL for Redis
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return errors.New("token expiration time has passed")
	}

	// Store in Redis with expiration
	key := s.generateRedisKey("verification", token)
	if err := s.redis.SetWithExpiration(ctx, key, string(tokenData), ttl); err != nil {
		return err
	}

	// Add to user's verification token collection
	userKey := s.generateUserKey(userID, "verification_tokens")
	userTokens, _ := s.redis.Get(ctx, userKey)
	var tokens []string
	if userTokens != "" {
		json.Unmarshal([]byte(userTokens), &tokens)
	}
	tokens = append(tokens, token)

	userTokensData, _ := json.Marshal(tokens)
	s.redis.SetWithExpiration(ctx, userKey, string(userTokensData), ttl)

	return nil
}

// GetVerificationToken retrieves a verification token from Redis.
// Returns the token data if found and not expired.
//
// Parameters:
//   - ctx: Context for the operation
//   - token: The verification token string to retrieve
//
// Returns:
//   - *models.VerificationToken: The verification token data or nil if not found
//   - error: Any error encountered during the operation
//
// Example:
//
//	verificationToken, err := tokenService.GetVerificationToken(ctx, "verify_token_123")
//	if err != nil {
//	    // Handle error (token not found, expired, etc.)
//	}
//	if verificationToken != nil {
//	    // Use the token data
//	}
func (s *tokenService) GetVerificationToken(ctx context.Context, token string) (*models.VerificationToken, error) {
	key := s.generateRedisKey("verification", token)
	tokenData, err := s.redis.Get(ctx, key)
	if err != nil {
		return nil, errors.New("verification token not found")
	}

	var verificationToken models.VerificationToken
	if err := json.Unmarshal([]byte(tokenData), &verificationToken); err != nil {
		return nil, err
	}

	// Check if token is expired
	if verificationToken.IsExpired() {
		// Clean up expired token
		s.redis.Delete(ctx, key)
		return nil, errors.New("verification token expired")
	}

	return &verificationToken, nil
}

// DeleteVerificationToken removes a verification token from Redis.
// This is typically called after successful verification or token cleanup.
//
// Parameters:
//   - ctx: Context for the operation
//   - token: The verification token string to delete
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := tokenService.DeleteVerificationToken(ctx, "verify_token_123")
//	if err != nil {
//	    // Handle error
//	}
func (s *tokenService) DeleteVerificationToken(ctx context.Context, token string) error {
	// Get token data first to remove from user's collection
	verificationToken, err := s.GetVerificationToken(ctx, token)
	if err == nil && verificationToken != nil {
		// Remove from user's token collection
		userKey := s.generateUserKey(verificationToken.UserID, "verification_tokens")
		userTokens, _ := s.redis.Get(ctx, userKey)
		if userTokens != "" {
			var tokens []string
			json.Unmarshal([]byte(userTokens), &tokens)

			// Remove the specific token
			for i, t := range tokens {
				if t == token {
					tokens = append(tokens[:i], tokens[i+1:]...)
					break
				}
			}

			if len(tokens) > 0 {
				userTokensData, _ := json.Marshal(tokens)
				s.redis.Set(ctx, userKey, string(userTokensData))
			} else {
				s.redis.Delete(ctx, userKey)
			}
		}
	}

	// Delete the token
	key := s.generateRedisKey("verification", token)
	return s.redis.Delete(ctx, key)
}

// MarkVerificationTokenAsUsed marks a verification token as used.
// This prevents the token from being used multiple times.
//
// Parameters:
//   - ctx: Context for the operation
//   - token: The verification token string to mark as used
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := tokenService.MarkVerificationTokenAsUsed(ctx, "verify_token_123")
//	if err != nil {
//	    // Handle error
//	}
func (s *tokenService) MarkVerificationTokenAsUsed(ctx context.Context, token string) error {
	verificationToken, err := s.GetVerificationToken(ctx, token)
	if err != nil {
		return err
	}

	// Mark as used
	verificationToken.Used = true
	verificationToken.UpdatedAt = time.Now()

	// Serialize and store back
	tokenData, err := json.Marshal(verificationToken)
	if err != nil {
		return err
	}

	key := s.generateRedisKey("verification", token)
	return s.redis.Set(ctx, key, string(tokenData))
}

// IsVerificationTokenValid checks if a verification token is valid and not expired.
// This is a convenience method that combines retrieval and validation.
//
// Parameters:
//   - ctx: Context for the operation
//   - token: The verification token string to validate
//
// Returns:
//   - bool: True if token is valid, false otherwise
//   - error: Any error encountered during the operation
//
// Example:
//
//	isValid, err := tokenService.IsVerificationTokenValid(ctx, "verify_token_123")
//	if err != nil {
//	    // Handle error
//	}
//	if isValid {
//	    // Token is valid, proceed with verification
//	}
func (s *tokenService) IsVerificationTokenValid(ctx context.Context, token string) (bool, error) {
	verificationToken, err := s.GetVerificationToken(ctx, token)
	if err != nil {
		return false, err
	}
	return verificationToken.IsValid(), nil
}

// CleanupExpiredTokens removes all expired tokens from Redis.
// This is a maintenance function that should be run periodically.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := tokenService.CleanupExpiredTokens(ctx)
//	if err != nil {
//	    // Handle error
//	}
func (s *tokenService) CleanupExpiredTokens(ctx context.Context) error {
	// This is a simplified cleanup - in a real implementation,
	// you might want to scan all keys and check expiration
	// For now, Redis will automatically expire keys based on TTL
	return nil
}

// GetUserTokens retrieves all tokens for a specific user.
// This is useful for user account management and security audits.
//
// Parameters:
//   - ctx: Context for the operation
//   - userID: ID of the user whose tokens to retrieve
//
// Returns:
//   - []string: List of token strings for the user
//   - error: Any error encountered during the operation
//
// Example:
//
//	tokens, err := tokenService.GetUserTokens(ctx, userID)
//	if err != nil {
//	    // Handle error
//	}
//	for _, token := range tokens {
//	    // Process each token
//	}
func (s *tokenService) GetUserTokens(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// Get refresh tokens
	refreshKey := s.generateUserKey(userID, "refresh_tokens")
	refreshTokens, _ := s.redis.Get(ctx, refreshKey)

	// Get API tokens
	apiKey := s.generateUserKey(userID, "api_tokens")
	apiTokens, _ := s.redis.Get(ctx, apiKey)

	// Get verification tokens
	verificationKey := s.generateUserKey(userID, "verification_tokens")
	verificationTokens, _ := s.redis.Get(ctx, verificationKey)

	var allTokens []string

	// Parse refresh tokens
	if refreshTokens != "" {
		var tokens []string
		json.Unmarshal([]byte(refreshTokens), &tokens)
		allTokens = append(allTokens, tokens...)
	}

	// Parse API tokens
	if apiTokens != "" {
		var tokens []string
		json.Unmarshal([]byte(apiTokens), &tokens)
		allTokens = append(allTokens, tokens...)
	}

	// Parse verification tokens
	if verificationTokens != "" {
		var tokens []string
		json.Unmarshal([]byte(verificationTokens), &tokens)
		allTokens = append(allTokens, tokens...)
	}

	return allTokens, nil
}
