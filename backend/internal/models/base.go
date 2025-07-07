package models

import (
	"time"

	"gorm.io/gorm"
)

// Base contains common fields for all models
type Base struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BaseWithUser contains common fields for models that track user actions
type BaseWithUser struct {
	Base
	CreatedBy *uint64 `gorm:"index" json:"created_by,omitempty"`
	UpdatedBy *uint64 `gorm:"index" json:"updated_by,omitempty"`
	DeletedBy *uint64 `gorm:"index" json:"deleted_by,omitempty"`
}

// BaseWithHierarchy contains common fields for hierarchical models (nested sets)
type BaseWithHierarchy struct {
	Base
	RecordLeft     *uint64 `gorm:"index" json:"record_left,omitempty"`
	RecordRight    *uint64 `gorm:"index" json:"record_right,omitempty"`
	RecordDept     *uint64 `gorm:"index" json:"record_dept,omitempty"`
	RecordOrdering *uint64 `gorm:"index" json:"record_ordering,omitempty"`
	ParentID       *uint64 `gorm:"index" json:"parent_id,omitempty"`
}

// BeforeCreate sets timestamps before creating a record
func (b *Base) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if b.CreatedAt.IsZero() {
		b.CreatedAt = now
	}
	if b.UpdatedAt.IsZero() {
		b.UpdatedAt = now
	}
	return nil
}

// BeforeUpdate sets updated timestamp before updating a record
func (b *Base) BeforeUpdate(tx *gorm.DB) error {
	b.UpdatedAt = time.Now()
	return nil
}

// IsDeleted checks if the model is soft deleted
func (b *Base) IsDeleted() bool {
	return !b.DeletedAt.Time.IsZero()
}

// GetAge returns the age of the record in seconds
func (b *Base) GetAge() time.Duration {
	return time.Since(b.CreatedAt)
}

// GetLastModified returns the time since last update
func (b *Base) GetLastModified() time.Duration {
	return time.Since(b.UpdatedAt)
}
