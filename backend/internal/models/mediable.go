package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Mediable represents a polymorphic relationship between Media and other models
type Mediable struct {
	MediaID      uuid.UUID `json:"media_id" gorm:"type:uuid;primaryKey" validate:"required"`
	MediableID   uuid.UUID `json:"mediable_id" gorm:"type:uuid;primaryKey" validate:"required"`
	MediableType string    `json:"mediable_type" gorm:"size:50;primaryKey;index" validate:"required,min=1,max=50"`
	Group        string    `json:"group" gorm:"size:50;default:'default';index" validate:"omitempty,min=1,max=50"`
	SortOrder    int       `json:"sort_order" gorm:"default:0;index"`
}

// TableName specifies the table name for Mediable
func (Mediable) TableName() string {
	return "mediables"
}

// BeforeCreate hook for Mediable
func (m *Mediable) BeforeCreate(tx *gorm.DB) error {
	if m.Group == "" {
		m.Group = "default"
	}
	return nil
}

// BeforeUpdate hook for Mediable
func (m *Mediable) BeforeUpdate(tx *gorm.DB) error {
	if m.Group == "" {
		m.Group = "default"
	}
	return nil
}

// IsDefaultGroup checks if this is the default group
func (m *Mediable) IsDefaultGroup() bool {
	return m.Group == "default"
}

// GetGroup returns the group name
func (m *Mediable) GetGroup() string {
	if m.Group == "" {
		return "default"
	}
	return m.Group
}
