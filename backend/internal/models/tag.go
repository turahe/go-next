package models

import (
	"time"
	"wordpress-go-next/backend/pkg/database"
)

// Tag represents a tag that can be associated with various entities
type Tag struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Slug        string     `json:"slug" gorm:"size:100;not null;uniqueIndex"`
	Description string     `json:"description" gorm:"size:500"`
	Color       string     `json:"color" gorm:"size:7;default:'#007bff'"`
	Type        string     `json:"type" gorm:"size:50;default:'general';index"` // general, category, feature, system
	IsActive    bool       `json:"is_active" gorm:"default:true;index"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	TaggedEntities []TaggedEntity `json:"tagged_entities,omitempty" gorm:"foreignKey:TagID"`
}

// TaggedEntity represents the many-to-many relationship between tags and entities
type TaggedEntity struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	TagID      uint      `json:"tag_id" gorm:"not null;index"`
	EntityID   uint      `json:"entity_id" gorm:"not null;index"`
	EntityType string    `json:"entity_type" gorm:"size:50;not null;index"` // post, user, media, etc.
	Group      string    `json:"group" gorm:"size:50;default:'default'"`    // Optional grouping
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relationships
	Tag Tag `json:"tag" gorm:"foreignKey:TagID"`
}

// TableName specifies the table name for Tag
func (Tag) TableName() string {
	return "tags"
}

// TableName specifies the table name for TaggedEntity
func (TaggedEntity) TableName() string {
	return "tagged_entities"
}

// BeforeCreate hook for Tag
func (t *Tag) BeforeCreate() error {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate hook for Tag
func (t *Tag) BeforeUpdate() error {
	t.UpdatedAt = time.Now()
	return nil
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

// IsValidType checks if the tag type is valid
func (t *Tag) IsValidType() bool {
	validTypes := []string{"general", "category", "feature", "system"}
	for _, validType := range validTypes {
		if t.Type == validType {
			return true
		}
	}
	return false
}

// IsValidColor checks if the color is a valid hex color
func (t *Tag) IsValidColor() bool {
	if len(t.Color) != 7 || t.Color[0] != '#' {
		return false
	}
	for i := 1; i < 7; i++ {
		c := t.Color[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// GetDefaultColor returns the default color for a tag type
func (t *Tag) GetDefaultColor() string {
	switch t.Type {
	case "feature":
		return "#FFD700" // Gold
	case "category":
		return "#007BFF" // Blue
	case "system":
		return "#FF6B6B" // Red
	default:
		return "#6C757D" // Gray
	}
}

// SetDefaultColor sets the default color if none is provided
func (t *Tag) SetDefaultColor() {
	if t.Color == "" {
		t.Color = t.GetDefaultColor()
	}
}

// GetTaggedEntitiesCount returns the count of entities tagged with this tag
func (t *Tag) GetTaggedEntitiesCount() int64 {
	var count int64
	database.DB.Model(&TaggedEntity{}).Where("tag_id = ?", t.ID).Count(&count)
	return count
}

// GetTaggedEntitiesByType returns the count of entities of a specific type tagged with this tag
func (t *Tag) GetTaggedEntitiesByType(entityType string) int64 {
	var count int64
	database.DB.Model(&TaggedEntity{}).Where("tag_id = ? AND entity_type = ?", t.ID, entityType).Count(&count)
	return count
}

// TagType constants
const (
	TagTypeGeneral  = "general"
	TagTypeCategory = "category"
	TagTypeFeature  = "feature"
	TagTypeSystem   = "system"
)

// EntityType constants
const (
	EntityTypePost     = "post"
	EntityTypeUser     = "user"
	EntityTypeMedia    = "media"
	EntityTypeCategory = "category"
	EntityTypeComment  = "comment"
)
