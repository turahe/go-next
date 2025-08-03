package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	RoleID    uuid.UUID      `json:"role_id" gorm:"type:uuid;not null;index"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName specifies the table name for UserRole
func (UserRole) TableName() string {
	return "user_roles"
}

// BeforeCreate hook for UserRole
func (ur *UserRole) BeforeCreate(tx *gorm.DB) error {
	if ur.ID == uuid.Nil {
		ur.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate hook for UserRole
func (ur *UserRole) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// GetIsActive returns the active status
func (ur *UserRole) GetIsActive() bool {
	return ur.DeletedAt.Time.IsZero()
}

// Activate activates the user role
func (ur *UserRole) Activate() {
	ur.DeletedAt = gorm.DeletedAt{Time: time.Time{}, Valid: false}
}

// Deactivate deactivates the user role
func (ur *UserRole) Deactivate() {
	ur.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
}
