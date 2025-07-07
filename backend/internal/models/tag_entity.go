package models

import (
	"time"
)

// TaggedEntity represents the many-to-many relationship between tags and entities
type TaggedEntity struct {
	ID         uint64    `json:"id" gorm:"primaryKey"`
	TagID      uint64    `json:"tag_id" gorm:"not null;index"`
	EntityID   uint64    `json:"entity_id" gorm:"not null;index"`
	EntityType string    `json:"entity_type" gorm:"size:50;not null;index"` // post, user, media, etc.
	Group      string    `json:"group" gorm:"size:50;default:'default'"`    // Optional grouping
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relationships
	Tag Tag `json:"tag" gorm:"foreignKey:TagID"`
}

// TableName specifies the table name for TaggedEntity
func (TaggedEntity) TableName() string {
	return "tagged_entities"
}

// BeforeCreate hook for TaggedEntity
func (te *TaggedEntity) BeforeCreate() error {
	if te.CreatedAt.IsZero() {
		te.CreatedAt = time.Now()
	}
	if te.UpdatedAt.IsZero() {
		te.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for TaggedEntity
func (te *TaggedEntity) BeforeUpdate() error {
	te.UpdatedAt = time.Now()
	return nil
}
