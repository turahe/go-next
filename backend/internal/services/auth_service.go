// Package services provides business logic layer for the blog application.
// This package contains all service interfaces and implementations that handle
// the core business logic, data processing, and external service interactions.
package services

import (
	"errors"
	"go-next/internal/dto"
	"go-next/internal/models"
	"go-next/pkg/database"
	"go-next/pkg/jwt"
	"go-next/pkg/logger"

	"github.com/google/uuid"
	"github.com/nyaruka/phonenumbers"
	"gorm.io/gorm"
)

// AuthService defines the interface for all authentication-related business operations.
// This interface provides methods for user authentication, token management,
// password operations, and account verification.
type AuthService interface {
	// Authentication methods - Core login and registration functionality

	// Register creates a new user account with email verification.
	// Validates user data, hashes password, and sends verification email.
	Register(username, email, phone, countryCode, password string) error

	// Login authenticates a user with email and password.
	// Returns access and refresh tokens upon successful authentication.
	Login(identity, password string) (*dto.AuthDTO, error)

	// RefreshToken generates new access token using a valid refresh token.
	// Used to maintain user sessions without requiring re-authentication.
	RefreshToken(refreshToken string) (*dto.AuthDTO, error)

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
	db          *gorm.DB     // Database connection for all data operations
	jwt         *jwt.Service // JWT service for token operations
	roleService RoleService  // Role service for role operations
}

// NewAuthService creates and returns a new instance of AuthService.
// This factory function initializes the service with the global database connection.
func NewAuthService() AuthService {
	return &authService{
		db:          database.DB,
		jwt:         jwt.NewService(nil), // Use default config
		roleService: NewRoleService(),
	}
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
func (s *authService) Register(username, email, phone, countryCode, password string) error {
	// Start a database transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		logger.Errorf("Error starting transaction: %v", tx.Error)
		return tx.Error
	}

	// Defer a function to handle rollback in case of error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create user
	var user models.User
	user.Username = username
	user.Email = email
	if phone != "" {
		phoneNumber, err := phonenumbers.Parse(phone, countryCode)
		if err != nil {
			logger.Errorf("Error parsing phone number: %v", err)
			tx.Rollback()
			return err
		}
		// Use E.164 format which is compact and suitable for database storage
		user.Phone = phonenumbers.Format(phoneNumber, phonenumbers.E164)
	}

	// Hash password
	if err := user.HashPassword(password); err != nil {
		logger.Errorf("Error hashing password: %v", err)
		tx.Rollback()
		return err
	}

	// Create the user within the transaction
	if err := tx.Create(&user).Error; err != nil {
		logger.Errorf("Error creating user: %v", err)
		tx.Rollback()
		return err
	}

	// Get or create the default "user" role
	defaultRole, err := s.roleService.GetOrCreateDefaultRole()
	if err != nil {
		logger.Errorf("Error getting default role: %v", err)
		tx.Rollback()
		return err
	}

	// Assign the default role to the user within the transaction
	if err := s.roleService.AssignRoleToUserWithTx(tx, user.ID, defaultRole.ID); err != nil {
		logger.Errorf("Error assigning default role to user: %v", err)
		tx.Rollback()
		return err
	}

	// Create default user settings
	settings := []models.Setting{
		{
			EntityType: "user",
			EntityID:   user.ID.String(),
			Key:        "language",
			Value:      "en",
			BaseModelWithUser: models.BaseModelWithUser{
				CreatedBy: &user.ID,
				UpdatedBy: &user.ID,
			},
		},
		{
			EntityType: "user",
			EntityID:   user.ID.String(),
			Key:        "timezone",
			Value:      "UTC",
			BaseModelWithUser: models.BaseModelWithUser{
				CreatedBy: &user.ID,
				UpdatedBy: &user.ID,
			},
		},
		{
			EntityType: "user",
			EntityID:   user.ID.String(),
			Key:        "currency",
			Value:      "USD",
			BaseModelWithUser: models.BaseModelWithUser{
				CreatedBy: &user.ID,
				UpdatedBy: &user.ID,
			},
		},
	}

	if err := tx.Create(&settings).Error; err != nil {
		logger.Errorf("Error creating setting: %v", err)
		tx.Rollback()
		return err
	}

	// Assign role in Casbin (outside transaction since Casbin has its own storage)
	casbinService := NewCasbinService()
	if err := casbinService.AddRoleForUser(user.ID, defaultRole.Name, "*"); err != nil {
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
func (s *authService) Login(identity, password string) (*dto.AuthDTO, error) {
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
	// Delete refresh token from Redis
	// TODO: Implement logout logic
	return errors.New("logout functionality not implemented")
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
	// TODO: Implement password reset logic
	return errors.New("password reset functionality not implemented")
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
	// TODO: Implement email verification logic
	return errors.New("email verification functionality not implemented")
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
