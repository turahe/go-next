package database

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// Database error checking functions
func IsNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func IsDuplicateError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common duplicate error messages
	errMsg := strings.ToLower(err.Error())
	duplicateKeywords := []string{
		"duplicate",
		"already exists",
		"unique constraint",
		"duplicate key",
		"duplicate entry",
	}

	for _, keyword := range duplicateKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}

	return false
}

func IsConstraintError(err error) bool {
	if err == nil {
		return false
	}

	// Check for constraint violation errors
	errMsg := strings.ToLower(err.Error())
	constraintKeywords := []string{
		"constraint",
		"foreign key",
		"not null",
		"check constraint",
		"violation",
	}

	for _, keyword := range constraintKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}

	return false
}

func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	// Check for connection-related errors
	errMsg := strings.ToLower(err.Error())
	connectionKeywords := []string{
		"connection",
		"timeout",
		"network",
		"dial",
		"refused",
		"unreachable",
	}

	for _, keyword := range connectionKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}

	return false
}

// GetDatabaseErrorType returns the type of database error
func GetDatabaseErrorType(err error) string {
	if IsNotFoundError(err) {
		return "not_found"
	}
	if IsDuplicateError(err) {
		return "duplicate"
	}
	if IsConstraintError(err) {
		return "constraint"
	}
	if IsConnectionError(err) {
		return "connection"
	}
	return "unknown"
}
