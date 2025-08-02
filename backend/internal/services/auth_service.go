// Package services provides business logic layer for the blog application.
// This package contains all service interfaces and implementations that handle
// the core business logic, data processing, and external service interactions.
package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"go-next/internal/models"
	"go-next/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthService defines the interface for all authentication-related business operations.
// This interface provides methods for user authentication, token management,
// password operations, and account verification.
type AuthService interface {
	// Authentication methods - Core login and registration functionality

	// Register creates a new user account with email verification.
	// Validates user data, hashes password, and sends verification email.
	Register(user *models.User) error

	// Login authenticates a user with email and password.
	// Returns access and refresh tokens upon successful authentication.
	Login(email, password string) (*models.User, string, string, error)

	// RefreshToken generates new access token using a valid refresh token.
	// Used to maintain user sessions without requiring re-authentication.
	RefreshToken(refreshToken string) (string, error)

	// Logout invalidates the current refresh token.
	// Ensures the user session is properly terminated.
	Logout(refreshToken string) error

	// Password management - Methods for password operations

	// RequestPasswordReset initiates the password reset process.
	// Sends a reset link to the user's email address.
	RequestPasswordReset(email string) error

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
// This struct holds the database connection and provides the actual implementation
// of all authentication-related business logic.
type authService struct {
	db           *gorm.DB     // Database connection for all data operations
	tokenService TokenService // Redis-based token service
}

// NewAuthService creates and returns a new instance of AuthService.
// This factory function initializes the service with the global database connection.
func NewAuthService() AuthService {
	return &authService{
		db:           database.DB,
		tokenService: NewTokenService(),
	}
}

// generateRandomToken creates a secure random hex string for token generation.
// This is a helper function to avoid import cycles with the utils package.
func (s *authService) generateRandomToken(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// generateJWT creates a simple JWT token for the given user ID.
// This is a simplified implementation to avoid import cycles.
func (s *authService) generateJWT(userID uuid.UUID) (string, error) {
	// For now, return a simple token format
	// In a real implementation, you would use proper JWT library
	token := userID.String() + "_" + time.Now().Add(time.Hour).Format("20060102150405")
	return token, nil
}

// Register creates a new user account with email verification.
// Validates user data, hashes password, and sends verification email.
//
// Parameters:
//   - user: User model with all required fields populated
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	user := &models.User{
//	    Username: "johndoe",
//	    Email:    "john@example.com",
//	}
//	err := user.HashPassword("securepassword")
//	if err != nil {
//	    // Handle error
//	}
//	err = authService.Register(user)
//	if err != nil {
//	    // Handle error (validation, database, or email error)
//	}
func (s *authService) Register(user *models.User) error {
	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		return errors.New("user already exists")
	}

	// Set default values
	user.IsActive = false
	user.EmailVerified = nil
	user.CreatedAt = time.Now()

	// Create the user
	if err := s.db.Create(user).Error; err != nil {
		return err
	}

	// Generate email verification token
	verificationToken, err := s.generateRandomToken(32)
	if err != nil {
		return err
	}

	// Store verification token in Redis
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hours expiration
	if err := s.tokenService.StoreVerificationToken(context.Background(), user.ID, verificationToken, models.EmailVerification, expiresAt, "", ""); err != nil {
		return err
	}

	// TODO: Send verification email with token
	// This would typically call an email service to send verification link

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
func (s *authService) Login(email, password string) (*models.User, string, string, error) {
	var user models.User

	// Find user by email
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", "", errors.New("invalid credentials")
		}
		return nil, "", "", err
	}

	// Verify password using the User model's method
	if !user.CheckPassword(password) {
		return nil, "", "", errors.New("invalid credentials")
	}

	// Check if account is active
	if !user.GetIsActive() {
		return nil, "", "", errors.New("account not active")
	}

	// Generate tokens using internal methods
	accessToken, err := s.generateJWT(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	// Generate refresh token (using the same method for now)
	refreshToken, err := s.generateJWT(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	// Store refresh token in Redis
	expiresAt := time.Now().AddDate(0, 0, 30) // 30 days
	if err := s.tokenService.StoreRefreshToken(context.Background(), user.ID, refreshToken, expiresAt, "", ""); err != nil {
		return nil, "", "", err
	}

	// Update last login time
	user.UpdateLastLogin()
	s.db.Save(&user)

	return &user, accessToken, refreshToken, nil
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
func (s *authService) RefreshToken(refreshToken string) (string, error) {
	// Find refresh token in Redis
	tokenModel, err := s.tokenService.GetRefreshToken(context.Background(), refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	// Generate new access token
	accessToken, err := s.generateJWT(tokenModel.UserID)
	if err != nil {
		return "", err
	}

	return accessToken, nil
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
	// Delete refresh token from Redis
	return s.tokenService.DeleteRefreshToken(context.Background(), refreshToken)
}

// RequestPasswordReset initiates the password reset process.
// Sends a reset link to the user's email address.
//
// Parameters:
//   - email: Email address of the user requesting password reset
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
func (s *authService) RequestPasswordReset(email string) error {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	// Generate reset token
	resetToken, err := s.generateRandomToken(32)
	if err != nil {
		return err
	}

	// Store reset token in Redis with expiration
	expiresAt := time.Now().Add(time.Hour) // 1 hour expiration
	if err := s.tokenService.StoreVerificationToken(context.Background(), user.ID, resetToken, models.PasswordReset, expiresAt, "", ""); err != nil {
		return err
	}

	// TODO: Send password reset email
	// This would typically call an email service to send reset link

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
	// Find valid reset token in Redis
	resetToken, err := s.tokenService.GetVerificationToken(context.Background(), token)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	// Check if token is for password reset
	if resetToken.Type != models.PasswordReset {
		return errors.New("invalid token type")
	}

	// Get user and hash new password
	var user models.User
	if err := s.db.First(&user, resetToken.UserID).Error; err != nil {
		return errors.New("user not found")
	}

	if err := user.HashPassword(newPassword); err != nil {
		return err
	}

	// Update user password
	if err := s.db.Save(&user).Error; err != nil {
		return err
	}

	// Delete used reset token from Redis
	return s.tokenService.DeleteVerificationToken(context.Background(), token)
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
//   - token: Email verification token from email
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
	// Find valid verification token in Redis
	verificationToken, err := s.tokenService.GetVerificationToken(context.Background(), token)
	if err != nil {
		return errors.New("invalid or expired verification token")
	}

	// Check if token is for email verification
	if verificationToken.Type != models.EmailVerification {
		return errors.New("invalid token type")
	}

	// Update user status and mark email as verified
	var user models.User
	if err := s.db.First(&user, verificationToken.UserID).Error; err != nil {
		return errors.New("user not found")
	}

	user.MarkEmailVerified()
	user.Activate()

	if err := s.db.Save(&user).Error; err != nil {
		return err
	}

	// Mark token as used and delete from Redis
	if err := s.tokenService.MarkVerificationTokenAsUsed(context.Background(), token); err != nil {
		return err
	}

	return s.tokenService.DeleteVerificationToken(context.Background(), token)
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
	if user.IsEmailVerified() {
		return errors.New("email already verified")
	}

	// Generate new verification token
	verificationToken, err := s.generateRandomToken(32)
	if err != nil {
		return err
	}

	// Store verification token in Redis
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hour expiration
	if err := s.tokenService.StoreVerificationToken(context.Background(), user.ID, verificationToken, models.EmailVerification, expiresAt, "", ""); err != nil {
		return err
	}

	// TODO: Send verification email
	// This would typically call an email service to send verification link

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
	// TODO: Implement proper JWT token parsing and validation
	// For now, this is a placeholder implementation
	// In a real implementation, you would:
	// 1. Parse the JWT token
	// 2. Validate the signature
	// 3. Check expiration
	// 4. Extract user ID and get user from database

	return nil, errors.New("token validation not implemented")
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
