package validation

import (
	"fmt"
	"reflect"
	"strings"

	"go-next/internal/models"
	"go-next/pkg/database"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// UniqueValidator provides database uniqueness validation
type UniqueValidator struct {
	db *gorm.DB
}

// NewUniqueValidator creates a new unique validator
func NewUniqueValidator() *UniqueValidator {
	return &UniqueValidator{
		db: database.GetDB(),
	}
}

// RegisterUniqueValidations registers all unique validation functions
func (uv *UniqueValidator) RegisterUniqueValidations(v *validator.Validate) {
	// Register unique validation functions
	v.RegisterValidation("unique_email", uv.validateUniqueEmail)
	v.RegisterValidation("unique_username", uv.validateUniqueUsername)
	v.RegisterValidation("unique_phone", uv.validateUniquePhone)
	v.RegisterValidation("unique_role", uv.validateUniqueRole)
	v.RegisterValidation("unique_field", uv.validateUniqueField)
}

// validateUniqueEmail validates that email is unique in the database
func (uv *UniqueValidator) validateUniqueEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	if email == "" {
		return true // Empty emails are handled by required validation
	}

	// Get the current user ID from the context if this is an update operation
	var currentUserID string
	if fl.Parent().Kind() == reflect.Struct {
		// Try to get user_id from the struct
		userIDField := fl.Parent().FieldByName("UserID")
		if userIDField.IsValid() {
			currentUserID = userIDField.String()
		}
	}

	var count int64
	query := uv.db.Model(&models.User{}).Where("email = ?", email)

	// If we have a current user ID, exclude it from the uniqueness check
	if currentUserID != "" {
		query = query.Where("id != ?", currentUserID)
	}

	query.Count(&count)

	return count == 0
}

// validateUniqueUsername validates that username is unique in the database
func (uv *UniqueValidator) validateUniqueUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if username == "" {
		return true // Empty usernames are handled by required validation
	}

	// Get the current user ID from the context if this is an update operation
	var currentUserID string
	if fl.Parent().Kind() == reflect.Struct {
		// Try to get user_id from the struct
		userIDField := fl.Parent().FieldByName("UserID")
		if userIDField.IsValid() {
			currentUserID = userIDField.String()
		}
	}

	var count int64
	query := uv.db.Model(&models.User{}).Where("username = ?", username)

	// If we have a current user ID, exclude it from the uniqueness check
	if currentUserID != "" {
		query = query.Where("id != ?", currentUserID)
	}

	query.Count(&count)

	return count == 0
}

// validateUniquePhone validates that phone number is unique in the database
func (uv *UniqueValidator) validateUniquePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true // Empty phones are handled by required validation
	}

	// Get the current user ID from the context if this is an update operation
	var currentUserID string
	if fl.Parent().Kind() == reflect.Struct {
		// Try to get user_id from the struct
		userIDField := fl.Parent().FieldByName("UserID")
		if userIDField.IsValid() {
			currentUserID = userIDField.String()
		}
	}

	var count int64
	query := uv.db.Model(&models.User{}).Where("phone = ?", phone)

	// If we have a current user ID, exclude it from the uniqueness check
	if currentUserID != "" {
		query = query.Where("id != ?", currentUserID)
	}

	query.Count(&count)

	return count == 0
}

// validateUniqueRole validates that role name is unique in the database
func (uv *UniqueValidator) validateUniqueRole(fl validator.FieldLevel) bool {
	roleName := fl.Field().String()
	if roleName == "" {
		return true // Empty role names are handled by required validation
	}

	// Get the current role ID from the context if this is an update operation
	var currentRoleID string
	if fl.Parent().Kind() == reflect.Struct {
		// Try to get role_id from the struct
		roleIDField := fl.Parent().FieldByName("ID")
		if roleIDField.IsValid() {
			currentRoleID = roleIDField.String()
		}
	}

	var count int64
	query := uv.db.Model(&models.Role{}).Where("name = ?", roleName)

	// If we have a current role ID, exclude it from the uniqueness check
	if currentRoleID != "" {
		query = query.Where("id != ?", currentRoleID)
	}

	query.Count(&count)

	return count == 0
}

// validateUniqueField validates that a field is unique in a specific table
// Usage: validate:"unique_field=users,email,id"
func (uv *UniqueValidator) validateUniqueField(fl validator.FieldLevel) bool {
	fieldValue := fl.Field().String()
	if fieldValue == "" {
		return true
	}

	// Parse the tag parameters
	params := strings.Split(fl.Param(), ",")
	if len(params) < 2 {
		return false
	}

	tableName := params[0]
	fieldName := params[1]

	// Optional: exclude current record ID
	var excludeID string
	if len(params) > 2 {
		excludeID = params[2]
	}

	// Build the query dynamically
	query := uv.db.Table(tableName).Where(fmt.Sprintf("%s = ?", fieldName), fieldValue)

	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}

	var count int64
	query.Count(&count)

	return count == 0
}

// ValidateUniqueEmail validates email uniqueness with custom error handling
func (uv *UniqueValidator) ValidateUniqueEmail(email string, excludeUserID string) error {
	if email == "" {
		return nil
	}

	var count int64
	query := uv.db.Model(&models.User{}).Where("email = ?", email)

	if excludeUserID != "" {
		query = query.Where("id != ?", excludeUserID)
	}

	query.Count(&count)

	if count > 0 {
		return fmt.Errorf("email '%s' is already taken", email)
	}

	return nil
}

// ValidateUniqueUsername validates username uniqueness with custom error handling
func (uv *UniqueValidator) ValidateUniqueUsername(username string, excludeUserID string) error {
	if username == "" {
		return nil
	}

	var count int64
	query := uv.db.Model(&models.User{}).Where("username = ?", username)

	if excludeUserID != "" {
		query = query.Where("id != ?", excludeUserID)
	}

	query.Count(&count)

	if count > 0 {
		return fmt.Errorf("username '%s' is already taken", username)
	}

	return nil
}

// ValidateUniquePhone validates phone uniqueness with custom error handling
func (uv *UniqueValidator) ValidateUniquePhone(phone string, excludeUserID string) error {
	if phone == "" {
		return nil
	}

	var count int64
	query := uv.db.Model(&models.User{}).Where("phone = ?", phone)

	if excludeUserID != "" {
		query = query.Where("id != ?", excludeUserID)
	}

	query.Count(&count)

	if count > 0 {
		return fmt.Errorf("phone number '%s' is already registered", phone)
	}

	return nil
}

// ValidateUniqueField validates any field uniqueness with custom error handling
func (uv *UniqueValidator) ValidateUniqueField(tableName, fieldName, fieldValue, excludeID string) error {
	if fieldValue == "" {
		return nil
	}

	var count int64
	query := uv.db.Table(tableName).Where(fmt.Sprintf("%s = ?", fieldName), fieldValue)

	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}

	query.Count(&count)

	if count > 0 {
		return fmt.Errorf("%s '%s' is already taken", fieldName, fieldValue)
	}

	return nil
}

// ValidateUniqueRole validates role name uniqueness with custom error handling
func (uv *UniqueValidator) ValidateUniqueRole(roleName string, excludeRoleID string) error {
	if roleName == "" {
		return nil
	}

	var count int64
	query := uv.db.Model(&models.Role{}).Where("name = ?", roleName)

	if excludeRoleID != "" {
		query = query.Where("id != ?", excludeRoleID)
	}

	query.Count(&count)

	if count > 0 {
		return fmt.Errorf("role name '%s' is already taken", roleName)
	}

	return nil
}
