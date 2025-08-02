package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents a user role in the system
type Role struct {
	BaseModel
	Name        string `json:"name" gorm:"uniqueIndex;not null;size:50" validate:"required,min=2,max=50"`
	Description string `json:"description" gorm:"size:255"`
	IsActive    bool   `json:"is_active" gorm:"default:true;index"`
	Users       []User `json:"users,omitempty" gorm:"many2many:user_roles;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for Role
func (Role) TableName() string {
	return "roles"
}

// BeforeCreate hook for Role
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate hook for Role
func (r *Role) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// GetIsActive returns the active status
func (r *Role) GetIsActive() bool {
	return r.IsActive
}

// Activate activates the role
func (r *Role) Activate() {
	r.IsActive = true
}

// Deactivate deactivates the role
func (r *Role) Deactivate() {
	r.IsActive = false
}
