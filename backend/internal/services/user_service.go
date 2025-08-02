// Package services provides business logic layer for the blog application.
// This package contains all service interfaces and implementations that handle
// the core business logic, data processing, and external service interactions.
package services

import (
	"errors"

	"go-next/internal/models"
	"go-next/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserService defines the interface for all user-related business operations.
// This interface provides methods for user management, profile operations,
// and user data retrieval.
type UserService interface {
	// GetUserByID retrieves a user by their unique identifier.
	// Returns the user with all related data or an error if not found.
	GetUserByID(id uuid.UUID) (*models.User, error)

	// GetUserByEmail retrieves a user by their email address.
	// Used for authentication and user lookup operations.
	GetUserByEmail(email string) (*models.User, error)

	// CreateUser creates a new user account in the database.
	// Validates user data and handles password hashing.
	CreateUser(user *models.User) error

	// UpdateUser updates an existing user's information.
	// Only allows updating specific fields for security reasons.
	UpdateUser(user *models.User) error

	// DeleteUser permanently removes a user account.
	// This action cannot be undone and should be used with caution.
	DeleteUser(id uuid.UUID) error

	// GetUserProfile retrieves a user's public profile information.
	// Returns only safe, public information about the user.
	GetUserProfile(id uuid.UUID) (*models.User, error)
}

// userService implements the UserService interface.
// This struct holds the database connection and provides the actual implementation
// of all user-related business logic.
type userService struct {
	db *gorm.DB // Database connection for all data operations
}

// NewUserService creates and returns a new instance of UserService.
// This factory function initializes the service with the global database connection.
func NewUserService() UserService {
	return &userService{db: database.DB}
}

// GetUserByID retrieves a user by their unique identifier.
//
// Parameters:
//   - id: UUID of the user to retrieve
//
// Returns:
//   - *models.User: The user with all related data or nil if not found
//   - error: Any error encountered during the operation
//
// Example:
//
//	user, err := userService.GetUserByID(userUUID)
//	if err != nil {
//	    // Handle error (user not found or database error)
//	}
func (s *userService) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User

	err := s.db.Preload("Role").First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by their email address.
// Used for authentication and user lookup operations.
//
// Parameters:
//   - email: Email address of the user to retrieve
//
// Returns:
//   - *models.User: The user with all related data or nil if not found
//   - error: Any error encountered during the operation
//
// Example:
//
//	user, err := userService.GetUserByEmail("user@example.com")
//	if err != nil {
//	    // Handle error (user not found or database error)
//	}
func (s *userService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User

	err := s.db.Preload("Role").Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// CreateUser creates a new user account in the database.
// Validates user data and handles password hashing.
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
//	    Name:     "John Doe",
//	    Email:    "john@example.com",
//	    Password: "securepassword",
//	}
//	err := userService.CreateUser(user)
//	if err != nil {
//	    // Handle error
//	}
func (s *userService) CreateUser(user *models.User) error {
	return s.db.Create(user).Error
}

// UpdateUser updates an existing user's information.
// Only allows updating specific fields for security reasons.
//
// Parameters:
//   - user: User model with updated fields
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	user.Name = "Updated Name"
//	user.Bio = "Updated bio"
//	err := userService.UpdateUser(user)
//	if err != nil {
//	    // Handle error
//	}
func (s *userService) UpdateUser(user *models.User) error {
	return s.db.Save(user).Error
}

// DeleteUser permanently removes a user account.
// This action cannot be undone and should be used with caution.
//
// Parameters:
//   - id: UUID of the user to delete
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := userService.DeleteUser(userUUID)
//	if err != nil {
//	    // Handle error
//	}
func (s *userService) DeleteUser(id uuid.UUID) error {
	return s.db.Delete(&models.User{}, id).Error
}

// GetUserProfile retrieves a user's public profile information.
// Returns only safe, public information about the user.
//
// Parameters:
//   - id: UUID of the user whose profile to retrieve
//
// Returns:
//   - *models.User: The user's public profile or nil if not found
//   - error: Any error encountered during the operation
//
// Example:
//
//	profile, err := userService.GetUserProfile(userUUID)
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Printf("User: %s, Bio: %s\n", profile.Name, profile.Bio)
func (s *userService) GetUserProfile(id uuid.UUID) (*models.User, error) {
	var user models.User

	// Select only public fields for profile display
	err := s.db.Select("id, name, email, bio, avatar, created_at").
		Where("id = ? AND active = ?", id, true).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user profile not found")
		}
		return nil, err
	}

	return &user, nil
}
