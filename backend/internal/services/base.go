// Package services provides business logic layer for the blog application.
// This package contains all service interfaces and implementations that handle
// the core business logic, data processing, and external service interactions.
package services

import (
	"context"

	"go-next/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseService defines the interface for common service operations.
// This interface provides methods that are shared across multiple services
// for basic CRUD operations and common business logic.
type BaseService interface {
	// Generic CRUD operations - Common database operations

	// Create adds a new record to the database.
	// Validates the model and handles any pre-save operations.
	Create(model interface{}) error

	// GetByID retrieves a record by its unique identifier.
	// Returns the model or an error if not found.
	GetByID(id uuid.UUID, model interface{}) error

	// Update modifies an existing record in the database.
	// Validates the model and handles any pre-update operations.
	Update(model interface{}) error

	// Delete removes a record from the database.
	// Handles soft deletes if the model supports it.
	Delete(id uuid.UUID, model interface{}) error

	// List retrieves a paginated list of records.
	// Supports filtering, sorting, and pagination.
	List(page, perPage int, model interface{}, filters map[string]interface{}) ([]interface{}, int64, error)

	// Exists checks if a record exists by the given criteria.
	// Useful for validation and duplicate checking.
	Exists(criteria map[string]interface{}, model interface{}) (bool, error)

	// Count returns the total number of records matching the criteria.
	// Useful for pagination and statistics.
	Count(criteria map[string]interface{}, model interface{}) (int64, error)

	// Transaction management - Methods for handling database transactions

	// WithTransaction executes a function within a database transaction.
	// Automatically handles commit/rollback based on function return value.
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error

	// BeginTransaction starts a new database transaction.
	// Returns the transaction object for manual control.
	BeginTransaction() *gorm.DB

	// CommitTransaction commits the current transaction.
	// Should be called after successful operations.
	CommitTransaction(tx *gorm.DB) error

	// RollbackTransaction rolls back the current transaction.
	// Should be called on errors to prevent data corruption.
	RollbackTransaction(tx *gorm.DB) error

	// Utility methods - Helper functions for common operations

	// ValidateModel performs validation on a model.
	// Uses the model's validation tags and custom validation rules.
	ValidateModel(model interface{}) error

	// SanitizeModel cleans and prepares a model for database operations.
	// Removes sensitive data and sets default values.
	SanitizeModel(model interface{}) error

	// LogOperation records an operation for audit purposes.
	// Useful for tracking changes and debugging.
	LogOperation(operation string, model interface{}, userID uuid.UUID) error
}

// baseService implements the BaseService interface.
// This struct holds the database connection and provides the actual implementation
// of all common service operations.
type baseService struct {
	db *gorm.DB // Database connection for all data operations
}

// NewBaseService creates and returns a new instance of BaseService.
// This factory function initializes the service with the global database connection.
func NewBaseService() BaseService {
	return &baseService{db: database.DB}
}

// Create adds a new record to the database.
// Validates the model and handles any pre-save operations.
//
// Parameters:
//   - model: The model to create (must be a pointer to a struct)
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
//	err := baseService.Create(user)
//	if err != nil {
//	    // Handle error (validation, database, etc.)
//	}
func (s *baseService) Create(model interface{}) error {
	// Validate the model before creating
	if err := s.ValidateModel(model); err != nil {
		return err
	}

	// Sanitize the model
	if err := s.SanitizeModel(model); err != nil {
		return err
	}

	// Create the record
	return s.db.Create(model).Error
}

// GetByID retrieves a record by its unique identifier.
// Returns the model or an error if not found.
//
// Parameters:
//   - id: UUID of the record to retrieve
//   - model: Pointer to the model to populate with data
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	var user models.User
//	err := baseService.GetByID(userID, &user)
//	if err != nil {
//	    // Handle error (not found, database error, etc.)
//	}
func (s *baseService) GetByID(id uuid.UUID, model interface{}) error {
	return s.db.First(model, id).Error
}

// Update modifies an existing record in the database.
// Validates the model and handles any pre-update operations.
//
// Parameters:
//   - model: The model to update (must be a pointer to a struct)
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	user.Username = "updated_username"
//	err := baseService.Update(user)
//	if err != nil {
//	    // Handle error (validation, database, etc.)
//	}
func (s *baseService) Update(model interface{}) error {
	// Validate the model before updating
	if err := s.ValidateModel(model); err != nil {
		return err
	}

	// Sanitize the model
	if err := s.SanitizeModel(model); err != nil {
		return err
	}

	// Update the record
	return s.db.Save(model).Error
}

// Delete removes a record from the database.
// Handles soft deletes if the model supports it.
//
// Parameters:
//   - id: UUID of the record to delete
//   - model: The model type to delete
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := baseService.Delete(userID, &models.User{})
//	if err != nil {
//	    // Handle error (not found, database error, etc.)
//	}
func (s *baseService) Delete(id uuid.UUID, model interface{}) error {
	return s.db.Delete(model, id).Error
}

// List retrieves a paginated list of records.
// Supports filtering, sorting, and pagination.
//
// Parameters:
//   - page: Current page number (1-based)
//   - perPage: Number of records per page
//   - model: The model type to query
//   - filters: Map of field names to filter values
//
// Returns:
//   - []interface{}: List of records
//   - int64: Total count of matching records
//   - error: Any error encountered during the operation
//
// Example:
//
//	filters := map[string]interface{}{
//	    "is_active": true,
//	    "role":      "admin",
//	}
//	records, total, err := baseService.List(1, 10, &models.User{}, filters)
//	if err != nil {
//	    // Handle error
//	}
func (s *baseService) List(page, perPage int, model interface{}, filters map[string]interface{}) ([]interface{}, int64, error) {
	var total int64
	var records []interface{}

	// Build the query
	query := s.db.Model(model)

	// Apply filters
	for field, value := range filters {
		query = query.Where(field+" = ?", value)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * perPage
	if err := query.Offset(offset).Limit(perPage).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// Exists checks if a record exists by the given criteria.
// Useful for validation and duplicate checking.
//
// Parameters:
//   - criteria: Map of field names to values to check
//   - model: The model type to check
//
// Returns:
//   - bool: True if record exists, false otherwise
//   - error: Any error encountered during the operation
//
// Example:
//
//	criteria := map[string]interface{}{
//	    "email": "user@example.com",
//	}
//	exists, err := baseService.Exists(criteria, &models.User{})
//	if err != nil {
//	    // Handle error
//	}
//	if exists {
//	    // Record exists
//	}
func (s *baseService) Exists(criteria map[string]interface{}, model interface{}) (bool, error) {
	var count int64
	query := s.db.Model(model)

	// Apply criteria
	for field, value := range criteria {
		query = query.Where(field+" = ?", value)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

// Count returns the total number of records matching the criteria.
// Useful for pagination and statistics.
//
// Parameters:
//   - criteria: Map of field names to values to filter by
//   - model: The model type to count
//
// Returns:
//   - int64: Total count of matching records
//   - error: Any error encountered during the operation
//
// Example:
//
//	criteria := map[string]interface{}{
//	    "is_active": true,
//	}
//	count, err := baseService.Count(criteria, &models.User{})
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Printf("Total active users: %d\n", count)
func (s *baseService) Count(criteria map[string]interface{}, model interface{}) (int64, error) {
	var count int64
	query := s.db.Model(model)

	// Apply criteria
	for field, value := range criteria {
		query = query.Where(field+" = ?", value)
	}

	err := query.Count(&count).Error
	return count, err
}

// WithTransaction executes a function within a database transaction.
// Automatically handles commit/rollback based on function return value.
//
// Parameters:
//   - ctx: Context for the transaction
//   - fn: Function to execute within the transaction
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := baseService.WithTransaction(ctx, func(tx *gorm.DB) error {
//	    // Create user
//	    if err := tx.Create(&user).Error; err != nil {
//	        return err
//	    }
//	    // Create user profile
//	    return tx.Create(&profile).Error
//	})
//	if err != nil {
//	    // Handle error (transaction rolled back)
//	}
func (s *baseService) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return s.db.WithContext(ctx).Transaction(fn)
}

// BeginTransaction starts a new database transaction.
// Returns the transaction object for manual control.
//
// Returns:
//   - *gorm.DB: The transaction object
//
// Example:
//
//	tx := baseService.BeginTransaction()
//	defer func() {
//	    if r := recover(); r != nil {
//	        baseService.RollbackTransaction(tx)
//	    }
//	}()
//
//	// Perform operations
//	if err := tx.Create(&user).Error; err != nil {
//	    baseService.RollbackTransaction(tx)
//	    return err
//	}
//
//	return baseService.CommitTransaction(tx)
func (s *baseService) BeginTransaction() *gorm.DB {
	return s.db.Begin()
}

// CommitTransaction commits the current transaction.
// Should be called after successful operations.
//
// Parameters:
//   - tx: The transaction object to commit
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	tx := baseService.BeginTransaction()
//	// ... perform operations ...
//	err := baseService.CommitTransaction(tx)
//	if err != nil {
//	    // Handle error
//	}
func (s *baseService) CommitTransaction(tx *gorm.DB) error {
	return tx.Commit().Error
}

// RollbackTransaction rolls back the current transaction.
// Should be called on errors to prevent data corruption.
//
// Parameters:
//   - tx: The transaction object to rollback
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	tx := baseService.BeginTransaction()
//	if err := tx.Create(&user).Error; err != nil {
//	    baseService.RollbackTransaction(tx)
//	    return err
//	}
//	return baseService.CommitTransaction(tx)
func (s *baseService) RollbackTransaction(tx *gorm.DB) error {
	return tx.Rollback().Error
}

// ValidateModel performs validation on a model.
// Uses the model's validation tags and custom validation rules.
//
// Parameters:
//   - model: The model to validate
//
// Returns:
//   - error: Any validation errors encountered
//
// Example:
//
//	user := &models.User{
//	    Email: "invalid-email",
//	}
//	err := baseService.ValidateModel(user)
//	if err != nil {
//	    // Handle validation error
//	}
func (s *baseService) ValidateModel(model interface{}) error {
	// TODO: Implement validation logic
	// This would typically use a validation library like go-playground/validator
	// For now, return nil as a placeholder
	return nil
}

// SanitizeModel cleans and prepares a model for database operations.
// Removes sensitive data and sets default values.
//
// Parameters:
//   - model: The model to sanitize
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	user := &models.User{
//	    Username: "  john_doe  ",
//	    Email:    "JOHN@EXAMPLE.COM",
//	}
//	err := baseService.SanitizeModel(user)
//	if err != nil {
//	    // Handle error
//	}
//	// user.Username is now "john_doe"
//	// user.Email is now "john@example.com"
func (s *baseService) SanitizeModel(model interface{}) error {
	// TODO: Implement sanitization logic
	// This would typically:
	// 1. Trim whitespace from string fields
	// 2. Convert emails to lowercase
	// 3. Remove sensitive fields
	// 4. Set default values
	// For now, return nil as a placeholder
	return nil
}

// LogOperation records an operation for audit purposes.
// Useful for tracking changes and debugging.
//
// Parameters:
//   - operation: Description of the operation performed
//   - model: The model that was operated on
//   - userID: ID of the user who performed the operation
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := baseService.LogOperation("user_created", user, adminID)
//	if err != nil {
//	    // Handle error
//	}
func (s *baseService) LogOperation(operation string, model interface{}, userID uuid.UUID) error {
	// TODO: Implement logging logic
	// This would typically:
	// 1. Create an audit log entry
	// 2. Store operation details
	// 3. Include user ID and timestamp
	// 4. Optionally store model data
	// For now, return nil as a placeholder
	return nil
}
