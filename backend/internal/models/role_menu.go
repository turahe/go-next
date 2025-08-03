package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleMenu represents the many-to-many relationship between roles and menus
type RoleMenu struct {
	RoleID uuid.UUID `json:"role_id" gorm:"type:uuid;primaryKey;not null"`
	MenuID uuid.UUID `json:"menu_id" gorm:"type:uuid;primaryKey;not null"`
}

// TableName specifies the table name for RoleMenu
func (RoleMenu) TableName() string {
	return "role_menus"
}

// BeforeCreate hook for RoleMenu
func (rm *RoleMenu) BeforeCreate(tx *gorm.DB) error {
	return nil
}
