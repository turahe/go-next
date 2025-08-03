package rules

import (
	"go-next/internal/models"
	"go-next/pkg/database"
	"strings"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

// uniqueUsername validates that the username is unique in the users table
func uniqueUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if username == "" {
		return true // Let other validators handle empty values
	}

	var count int64
	database.DB.Model(&models.User{}).Where("LOWER(username) = ?", strings.ToLower(username)).Count(&count)
	return count == 0
}

// uniqueEmail validates that the email is unique in the users table
func uniqueEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	if email == "" {
		return true // Let other validators handle empty values
	}

	var count int64
	database.DB.Model(&models.User{}).Where("LOWER(email) = ?", strings.ToLower(email)).Count(&count)
	return count == 0
}

// uniquePhone validates that the phone is unique in the users table
func uniquePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true // Let other validators handle empty values
	}

	var count int64

	database.DB.Model(&models.User{}).Where("LOWER(phone) = ?", strings.ToLower(phone)).Count(&count)
	return count == 0
}

// RegisterCustomValidators registers all custom validation functions
func RegisterCustomValidators() {
	Validate.RegisterValidation("unique_username", uniqueUsername)
	Validate.RegisterValidation("unique_email", uniqueEmail)
	Validate.RegisterValidation("unique_phone", uniquePhone)
}
