// Package services provides business logic layer for the blog application.
// This package contains all service interfaces and implementations that handle
// the core business logic, data processing, and external service interactions.
package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"go-next/internal/dto"
	"go-next/internal/models"
	"go-next/pkg/config"
	"go-next/pkg/database"
	"go-next/pkg/email"
	"go-next/pkg/jwt"
	"go-next/pkg/logger"
	"go-next/pkg/redis"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthService defines the interface for all authentication-related business operations.
// This interface provides methods for user registration, login, token management,
// and account verification operations.
type AuthService interface {
	// Authentication methods - Core login and registration functionality

	// Register creates a new user account with email verification.
	// Validates user data, hashes password, and sends verification email.
	Register(username, email, phone, countryCode, password string) error

	// Login authenticates a user with email and password.
	// Returns access and refresh tokens upon successful authentication.
	// Sends login success email notification with client details.
	Login(identity, password, clientIP, userAgent string) (*dto.AuthDTO, error)

	// RefreshToken generates new access token using a valid refresh token.
	// Used to maintain user sessions without requiring re-authentication.
	RefreshToken(refreshToken string) (*dto.AuthDTO, error)

	// Logout invalidates the current refresh token.
	// Ensures the user session is properly terminated.
	Logout(refreshToken string) error

	// Password management - Methods for password operations

	// RequestPasswordReset initiates the password reset process.
	// Sends a reset link to the user's email address.
	RequestPasswordReset(userEmail string) error

	// ResetPassword changes user password using a valid reset token.
	// Validates the reset token and updates the password securely.
	ResetPassword(token, newPassword string) error

	// ChangePassword allows authenticated users to change their password.
	// Requires current password verification for security.
	ChangePassword(userID uuid.UUID, currentPassword, newPassword string) error

	// Account verification - Methods for email and phone verification

	// VerifyEmail confirms a user's email address using a verification token.
	// Marks the email as verified and activates the account.
	VerifyEmail(token string) error

	// ResendVerificationEmail sends a new verification email to the user.
	// Useful when the original email expires or is lost.
	ResendVerificationEmail(email string) error

	// Token validation - Methods for token verification and management

	// ValidateToken checks if a token is valid and not expired.
	// Used for protecting routes and validating user sessions.
	ValidateToken(token string) (*models.User, error)

	// GetUserFromToken extracts user information from a valid token.
	// Returns the user associated with the token or an error.
	GetUserFromToken(token string) (*models.User, error)
}

// authService implements the AuthService interface.
// This struct holds the database connection, JWT service, and role service,
// providing the actual implementation of all authentication-related business logic.
type authService struct {
	db           *gorm.DB            // Database connection for all data operations
	jwt          *jwt.Service        // JWT service for token operations
	roleService  RoleService         // Role service for role operations
	emailService *email.EmailService // Email service for sending verification emails
	redisService *redis.RedisService // Redis service for token storage
}

// NewAuthService creates and returns a new instance of AuthService.
// This factory function initializes the service with the global database connection,
// JWT service, and role service.
func NewAuthService() AuthService {
	// Get configuration
	cfg := config.GetConfig()

	// Initialize email service
	emailService := email.NewEmailService(cfg.SMTP)

	// Initialize JWT service with config
	jwtConfig := &jwt.Config{
		SecretKey:     cfg.JwtSecret,
		AccessExpiry:  15 * time.Minute,   // 15 minutes
		RefreshExpiry: 7 * 24 * time.Hour, // 7 days
	}

	// Get Redis service from global service manager
	var redisService *redis.RedisService
	if ServiceMgr != nil {
		redisService = ServiceMgr.RedisService
	}

	return &authService{
		db:           database.DB,
		jwt:          jwt.NewService(jwtConfig),
		roleService:  NewRoleService(),
		emailService: emailService,
		redisService: redisService,
	}
}

// generateVerificationToken generates a secure random token for email verification
func (s *authService) generateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Simple in-memory token storage for development
// TODO: Replace with Redis in production
type tokenStorage struct {
	tokens map[string]string
	expiry map[string]time.Time
	mutex  sync.RWMutex
}

var tokenStore = &tokenStorage{
	tokens: make(map[string]string),
	expiry: make(map[string]time.Time),
}

// storeVerificationToken stores the verification token in memory with expiration
func (s *authService) storeVerificationToken(email, token string) error {
	tokenStore.mutex.Lock()
	defer tokenStore.mutex.Unlock()

	key := fmt.Sprintf("email_verification:%s", email)
	expiration := time.Now().Add(24 * time.Hour)

	tokenStore.tokens[key] = token
	tokenStore.expiry[key] = expiration

	return nil
}

// getVerificationToken retrieves the verification token from memory
func (s *authService) getVerificationToken(email string) (string, error) {
	tokenStore.mutex.RLock()
	defer tokenStore.mutex.RUnlock()

	key := fmt.Sprintf("email_verification:%s", email)

	token, exists := tokenStore.tokens[key]
	if !exists {
		return "", errors.New("verification token not found")
	}

	expiry, exists := tokenStore.expiry[key]
	if !exists || time.Now().After(expiry) {
		// Clean up expired token
		delete(tokenStore.tokens, key)
		delete(tokenStore.expiry, key)
		return "", errors.New("verification token expired")
	}

	return token, nil
}

// cleanupExpiredTokens removes expired tokens from memory
func (s *authService) cleanupExpiredTokens() {
	tokenStore.mutex.Lock()
	defer tokenStore.mutex.Unlock()

	now := time.Now()
	for key, expiry := range tokenStore.expiry {
		if now.After(expiry) {
			delete(tokenStore.tokens, key)
			delete(tokenStore.expiry, key)
		}
	}
}

// sendVerificationEmail sends a verification email to the user
func (s *authService) sendVerificationEmail(user *models.User, token string) error {
	// Generate verification URL
	verificationURL := fmt.Sprintf("http://localhost:8080/api/v1/auth/verify-email?token=%s", token)

	// Generate email content using template
	emailBody := email.EmailVerificationTemplate(user.Username, verificationURL)

	// Send email
	return s.emailService.SendEmail(user.Email, "Verify Your Email Address", emailBody)
}

// sendLoginSuccessEmail sends a login success notification email to the user
func (s *authService) sendLoginSuccessEmail(user *models.User, clientIP, userAgent string) error {
	// Format login time
	loginTime := time.Now().Format("2006-01-02 15:04:05 UTC")

	// Generate email content using template
	emailBody := email.LoginSuccessTemplate(user.Username, user.Email, loginTime, userAgent, clientIP)

	// Send email
	return s.emailService.SendEmail(user.Email, "🔐 Login Successful - Security Notification", emailBody)
}

// Register creates a new user account with email verification.
// Validates user data, hashes password, and sends verification email.
//
// Parameters:
//   - username: User's chosen username
//   - email: User's email address
//   - phone: User's phone number
//   - countryCode: User's country code
//   - password: User's password (plain text)
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := authService.Register("john_doe", "john@example.com", "+1234567890", "US", "securepassword")
//	if err != nil {
//	    // Handle error (validation, duplicate email, etc.)
//	}
func (s *authService) Register(username, email, phone, countryCode, password string) error {
	// Start a database transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Check if user already exists
	var existingUser models.User
	if err := tx.Where("email = ? OR username = ? OR phone = ?", email, username, phone).First(&existingUser).Error; err == nil {
		tx.Rollback()
		return errors.New("user already exists")
	}

	// Create new user
	user := &models.User{
		Username: username,
		Email:    email,
		Phone:    phone,
	}

	// Hash password
	if err := user.HashPassword(password); err != nil {
		tx.Rollback()
		return err
	}

	// Save user to database
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Generate verification token
	verificationToken, err := s.generateVerificationToken()
	if err != nil {
		tx.Rollback()
		return err
	}

	// Store verification token
	if err := s.storeVerificationToken(email, verificationToken); err != nil {
		tx.Rollback()
		return err
	}

	// Assign default role to user
	if err := s.roleService.AssignRoleToUser(user.ID, uuid.Nil); err != nil {
		logger.Errorf("Error assigning role in Casbin: %v", err)
		// Don't rollback the transaction since the user was created successfully
		// Just log the error and continue
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		logger.Errorf("Error committing transaction: %v", err)
		return err
	}

	logger.Infof("Successfully registered user %s with default role", user.Username)

	// Send verification email
	if err := s.sendVerificationEmail(user, verificationToken); err != nil {
		logger.Errorf("Error sending verification email: %v", err)
		// Don't fail the registration if email sending fails
		// The user can request a new verification email later
	}

	return nil
}

// Login authenticates a user with email and password.
// Returns access and refresh tokens upon successful authentication.
//
// Parameters:
//   - email: User's email address
//   - password: User's password (plain text)
//
// Returns:
//   - *models.User: The authenticated user with basic information
//   - string: Access token for API authentication
//   - string: Refresh token for token renewal
//   - error: Any error encountered during the operation
//
// Example:
//
//	user, accessToken, refreshToken, err := authService.Login("user@example.com", "password")
//	if err != nil {
//	    // Handle error (invalid credentials, account not verified, etc.)
//	}
//	// Store tokens securely and use accessToken for API calls
func (s *authService) Login(identity, password, clientIP, userAgent string) (*dto.AuthDTO, error) {
	var user models.User

	// Find user by identity
	if err := s.db.Where("email = ? OR username = ? OR phone = ?", identity, identity, identity).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	// Verify password using the User model's method
	if !user.CheckPassword(password) {
		return nil, errors.New("invalid credentials")
	}

	// Generate tokens using JWT service
	accessToken, err := s.jwt.GenerateAccessToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Send login success email (non-blocking)
	go func() {
		if err := s.sendLoginSuccessEmail(&user, clientIP, userAgent); err != nil {
			logger.Errorf("Error sending login success email: %v", err)
		}
	}()

	return &dto.AuthDTO{
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken generates new access token using a valid refresh token.
// Used to maintain user sessions without requiring re-authentication.
//
// Parameters:
//   - refreshToken: Valid refresh token string
//
// Returns:
//   - string: New access token
//   - error: Any error encountered during the operation
//
// Example:
//
//	newAccessToken, err := authService.RefreshToken(refreshToken)
//	if err != nil {
//	    // Handle error (invalid token, expired, etc.)
//	}
//	// Use newAccessToken for API calls
func (s *authService) RefreshToken(refreshToken string) (*dto.AuthDTO, error) {
	// Validate refresh token
	claims, err := s.jwt.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Check if the refresh token is blacklisted in Redis
	if s.redisService != nil {
		ctx := context.Background()
		blacklistKey := fmt.Sprintf("refresh_token_blacklist:%s", refreshToken)

		// Check if token exists in blacklist
		exists, err := s.redisService.Exists(ctx, blacklistKey)
		if err != nil {
			logger.Errorf("Error checking refresh token blacklist: %v", err)
			// Continue with refresh if Redis check fails
		} else if exists > 0 {
			return nil, errors.New("refresh token has been revoked")
		}
	}

	// Get user from database
	var user models.User
	if err := s.db.First(&user, claims.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Generate new access token
	accessToken, err := s.jwt.GenerateAccessToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	return &dto.AuthDTO{
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Logout invalidates the current refresh token.
// Ensures the user session is properly terminated.
//
// Parameters:
//   - refreshToken: Refresh token to invalidate
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := authService.Logout(refreshToken)
//	if err != nil {
//	    // Handle error
//	}
func (s *authService) Logout(refreshToken string) error {
	// Validate the refresh token first
	claims, err := s.jwt.ValidateToken(refreshToken)
	if err != nil {
		return errors.New("invalid refresh token")
	}

	// If Redis is available, store the token in a blacklist
	if s.redisService != nil {
		ctx := context.Background()

		// Create a blacklist key for the refresh token
		blacklistKey := fmt.Sprintf("refresh_token_blacklist:%s", refreshToken)

		// Store the token in blacklist with expiration (same as refresh token expiry)
		// This ensures the blacklist entry expires when the token would have expired
		expiration := time.Duration(7*24) * time.Hour // 7 days, matching refresh token expiry

		err := s.redisService.SetWithExpiration(ctx, blacklistKey, "revoked", expiration)
		if err != nil {
			logger.Errorf("Error adding refresh token to blacklist: %v", err)
			// Don't fail logout if Redis is unavailable
			// The token will expire naturally anyway
		}

		// Also store user-specific blacklist for potential future use
		userBlacklistKey := fmt.Sprintf("user_refresh_tokens:%s", claims.UserID.String())
		err = s.redisService.SetWithExpiration(ctx, userBlacklistKey, refreshToken, expiration)
		if err != nil {
			logger.Errorf("Error storing user refresh token blacklist: %v", err)
		}

		logger.Infof("Refresh token blacklisted for user %s", claims.UserID.String())
	} else {
		logger.Warnf("Redis service not available, logout will rely on token expiration")
	}

	return nil
}

// RequestPasswordReset initiates the password reset process.
// Sends a reset link to the user's email address.
//
// Parameters:
//   - userEmail: Email address of the user requesting password reset
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := authService.RequestPasswordReset("user@example.com")
//	if err != nil {
//	    // Handle error (user not found, email error, etc.)
//	}
func (s *authService) RequestPasswordReset(userEmail string) error {
	// Find user by email
	var user models.User
	if err := s.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Generate password reset token
	resetToken, err := s.generateVerificationToken()
	if err != nil {
		return err
	}

	// Store reset token in Redis with 1-hour expiration
	if s.redisService != nil {
		ctx := context.Background()
		resetKey := fmt.Sprintf("password_reset:%s", userEmail)
		expiration := time.Hour // 1 hour expiration

		err := s.redisService.SetWithExpiration(ctx, resetKey, resetToken, expiration)
		if err != nil {
			logger.Errorf("Error storing password reset token: %v", err)
			return errors.New("failed to generate reset token")
		}

		logger.Infof("Password reset token generated for user %s", user.Username)
	} else {
		// Fallback to in-memory storage if Redis is not available
		resetKey := fmt.Sprintf("password_reset:%s", userEmail)
		expiration := time.Now().Add(time.Hour)

		tokenStore.mutex.Lock()
		tokenStore.tokens[resetKey] = resetToken
		tokenStore.expiry[resetKey] = expiration
		tokenStore.mutex.Unlock()
	}

	// Generate reset URL
	resetURL := fmt.Sprintf("http://localhost:8080/api/v1/auth/reset-password?token=%s", resetToken)

	// Generate email content using template
	emailBody := email.PasswordResetTemplate(user.Username, resetURL)

	// Send password reset email
	if err := s.emailService.SendEmail(user.Email, "🔑 Password Reset Request", emailBody); err != nil {
		logger.Errorf("Error sending password reset email: %v", err)
		return errors.New("failed to send reset email")
	}

	logger.Infof("Password reset email sent to %s", userEmail)
	return nil
}

// ResetPassword changes user password using a valid reset token.
// Validates the reset token and updates the password securely.
//
// Parameters:
//   - token: Password reset token from email
//   - newPassword: New password to set
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := authService.ResetPassword(resetToken, "newSecurePassword")
//	if err != nil {
//	    // Handle error (invalid token, expired, weak password, etc.)
//	}
func (s *authService) ResetPassword(token, newPassword string) error {
	// Find the email associated with this token
	var userEmail string

	if s.redisService != nil {
		// Check Redis for the token
		ctx := context.Background()
		keys, err := s.redisService.GetKeysByPattern(ctx, "password_reset:*")
		if err != nil {
			logger.Errorf("Error searching for password reset token: %v", err)
			return errors.New("failed to validate reset token")
		}

		for _, key := range keys {
			storedToken, err := s.redisService.Get(ctx, key)
			if err != nil {
				continue
			}
			if storedToken == token {
				// Extract email from key (format: "password_reset:email@example.com")
				if len(key) > 15 { // "password_reset:" is 15 characters
					userEmail = key[15:]
				}
				break
			}
		}
	} else {
		// Check in-memory storage
		tokenStore.mutex.RLock()
		for key, storedToken := range tokenStore.tokens {
			if strings.HasPrefix(key, "password_reset:") && storedToken == token {
				// Check if token is expired
				if expiry, exists := tokenStore.expiry[key]; exists && time.Now().Before(expiry) {
					// Extract email from key
					if len(key) > 15 {
						userEmail = key[15:]
					}
				}
				break
			}
		}
		tokenStore.mutex.RUnlock()
	}

	if userEmail == "" {
		return errors.New("invalid or expired reset token")
	}

	// Find user by email
	var user models.User
	if err := s.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Hash new password
	if err := user.HashPassword(newPassword); err != nil {
		return err
	}

	// Update password in database
	if err := s.db.Save(&user).Error; err != nil {
		return err
	}

	// Remove the reset token from storage
	if s.redisService != nil {
		ctx := context.Background()
		resetKey := fmt.Sprintf("password_reset:%s", userEmail)
		s.redisService.Delete(ctx, resetKey)
	} else {
		// Remove from in-memory storage
		tokenStore.mutex.Lock()
		resetKey := fmt.Sprintf("password_reset:%s", userEmail)
		delete(tokenStore.tokens, resetKey)
		delete(tokenStore.expiry, resetKey)
		tokenStore.mutex.Unlock()
	}

	logger.Infof("Password reset successful for user %s", user.Username)
	return nil
}

// ChangePassword allows authenticated users to change their password.
// Requires current password verification for security.
//
// Parameters:
//   - userID: ID of the user changing password
//   - currentPassword: Current password for verification
//   - newPassword: New password to set
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := authService.ChangePassword(userID, "currentPass", "newSecurePass")
//	if err != nil {
//	    // Handle error (wrong current password, weak new password, etc.)
//	}
func (s *authService) ChangePassword(userID uuid.UUID, currentPassword, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	// Verify current password
	if !user.CheckPassword(currentPassword) {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	if err := user.HashPassword(newPassword); err != nil {
		return err
	}

	// Update password
	return s.db.Save(&user).Error
}

// VerifyEmail confirms a user's email address using a verification token.
// Marks the email as verified and activates the account.
//
// Parameters:
//   - token: Verification token sent to user's email
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := authService.VerifyEmail(verificationToken)
//	if err != nil {
//	    // Handle error (invalid token, expired, etc.)
//	}
func (s *authService) VerifyEmail(token string) error {
	// Clean up expired tokens first
	s.cleanupExpiredTokens()

	// Find the email associated with this token
	var userEmail string
	tokenStore.mutex.RLock()
	for key, storedToken := range tokenStore.tokens {
		if storedToken == token {
			// Extract email from key (format: "email_verification:email@example.com")
			if len(key) > 18 { // "email_verification:" is 18 characters
				userEmail = key[18:]
			}
			break
		}
	}
	tokenStore.mutex.RUnlock()

	if userEmail == "" {
		return errors.New("invalid verification token")
	}

	// Find user by email
	var user models.User
	if err := s.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Check if email is already verified
	if user.EmailVerified != nil {
		return errors.New("email is already verified")
	}

	// Mark email as verified
	now := time.Now()
	user.EmailVerified = &now

	if err := s.db.Save(&user).Error; err != nil {
		return err
	}

	// Remove the token from storage
	tokenStore.mutex.Lock()
	key := fmt.Sprintf("email_verification:%s", userEmail)
	delete(tokenStore.tokens, key)
	delete(tokenStore.expiry, key)
	tokenStore.mutex.Unlock()

	logger.Infof("Email verified for user %s", user.Username)
	return nil
}

// ResendVerificationEmail sends a new verification email to the user.
// Useful when the original email expires or is lost.
//
// Parameters:
//   - email: Email address to send verification to
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := authService.ResendVerificationEmail("user@example.com")
//	if err != nil {
//	    // Handle error (user not found, email error, etc.)
//	}
func (s *authService) ResendVerificationEmail(email string) error {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	// Check if email is already verified
	if user.EmailVerified != nil {
		return errors.New("email is already verified")
	}

	// Generate new verification token
	verificationToken, err := s.generateVerificationToken()
	if err != nil {
		return err
	}

	// Store new verification token
	if err := s.storeVerificationToken(email, verificationToken); err != nil {
		return err
	}

	// Send new verification email
	if err := s.sendVerificationEmail(&user, verificationToken); err != nil {
		return err
	}

	logger.Infof("Verification email resent to %s", email)
	return nil
}

// ValidateToken checks if a token is valid and not expired.
// Used for protecting routes and validating user sessions.
//
// Parameters:
//   - token: Access token to validate
//
// Returns:
//   - *models.User: The user associated with the token or nil if invalid
//   - error: Any error encountered during the operation
//
// Example:
//
//	user, err := authService.ValidateToken(accessToken)
//	if err != nil {
//	    // Handle error (invalid token, expired, etc.)
//	}
//	// Use user for authorization
func (s *authService) ValidateToken(token string) (*models.User, error) {
	// Validate JWT token
	claims, err := s.jwt.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	// Get user from database
	var user models.User
	if err := s.db.First(&user, claims.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// GetUserFromToken extracts user information from a valid token.
// Returns the user associated with the token or an error.
//
// Parameters:
//   - token: Access token to extract user from
//
// Returns:
//   - *models.User: The user associated with the token
//   - error: Any error encountered during the operation
//
// Example:
//
//	user, err := authService.GetUserFromToken(accessToken)
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Printf("User: %s\n", user.Username)
func (s *authService) GetUserFromToken(token string) (*models.User, error) {
	// This is essentially the same as ValidateToken but with different error handling
	return s.ValidateToken(token)
}
