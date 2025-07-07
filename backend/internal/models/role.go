package models

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// Role represents a user role in the system
type Role struct {
	Base
	Name        string `gorm:"uniqueIndex;not null;size:50" json:"name" validate:"required,min=2,max=50"`
	Description string `gorm:"size:255" json:"description,omitempty"`
	IsActive    bool   `gorm:"default:true;index" json:"is_active"`

	// Relationships
	Users []User `gorm:"many2many:user_roles;constraint:OnDelete:CASCADE" json:"users,omitempty"`
}

// TableName specifies the table name for Role
func (Role) TableName() string {
	return "roles"
}

// BeforeCreate validates role data before creation
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if err := r.Base.BeforeCreate(tx); err != nil {
		return err
	}

	// Normalize name
	r.Name = strings.ToLower(strings.TrimSpace(r.Name))

	// Set default values
	if r.IsActive == false {
		r.IsActive = true
	}

	return r.validate()
}

// BeforeUpdate validates role data before update
func (r *Role) BeforeUpdate(tx *gorm.DB) error {
	if err := r.Base.BeforeUpdate(tx); err != nil {
		return err
	}

	// Normalize name
	r.Name = strings.ToLower(strings.TrimSpace(r.Name))

	return r.validate()
}

// validate performs validation on role fields
func (r *Role) validate() error {
	if len(r.Name) < 2 || len(r.Name) > 50 {
		return errors.New("role name must be between 2 and 50 characters")
	}

	if len(r.Description) > 255 {
		return errors.New("role description cannot exceed 255 characters")
	}

	return nil
}

// IsSystemRole checks if this is a system-defined role
func (r *Role) IsSystemRole() bool {
	systemRoles := []string{"admin", "user", "moderator", "guest"}
	for _, systemRole := range systemRoles {
		if r.Name == systemRole {
			return true
		}
	}
	return false
}
