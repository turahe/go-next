package models

import (
	"time"
)

// Tag represents a tag that can be associated with various entities
type Tag struct {
	ID          uint64     `json:"id" gorm:"primaryKey"`
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

// TableName specifies the table name for Tag
func (Tag) TableName() string {
	return "tags"
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
