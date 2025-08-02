package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Content represents polymorphic content for various models
type Content struct {
	BaseModel
	ModelID   uuid.UUID `json:"model_id" gorm:"type:uuid;not null;index" validate:"required"`
	ModelType string    `json:"model_type" gorm:"not null;size:50;index" validate:"required,min=1,max=50"`
	Type      string    `json:"type" gorm:"not null;size:20" validate:"required,oneof=html markdown json text"`
	Content   string    `json:"content" gorm:"type:text;not null" validate:"required,min=1"`
	SortOrder int       `json:"sort_order" gorm:"default:0;index"`
}

// TableName specifies the table name for Content
func (Content) TableName() string {
	return "contents"
}

// BeforeCreate hook for Content
func (c *Content) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate hook for Content
func (c *Content) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// GetModelType returns the model type
func (c *Content) GetModelType() string {
	return c.ModelType
}

// IsHTML checks if the content is HTML
func (c *Content) IsHTML() bool {
	return c.Type == "html"
}

// IsMarkdown checks if the content is Markdown
func (c *Content) IsMarkdown() bool {
	return c.Type == "markdown"
}

// IsJSON checks if the content is JSON
func (c *Content) IsJSON() bool {
	return c.Type == "json"
}

// IsText checks if the content is plain text
func (c *Content) IsText() bool {
	return c.Type == "text"
}
