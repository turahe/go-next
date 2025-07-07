package models

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// Content represents additional content associated with models (polymorphic)
type Content struct {
	Base
	ModelID   uint   `gorm:"not null;index" json:"model_id" validate:"required"`
	ModelType string `gorm:"not null;size:50;index" json:"model_type" validate:"required"`
	Content   string `gorm:"type:text;not null" json:"content" validate:"required,min=1"`
	Type      string `gorm:"size:50;index" json:"type,omitempty"` // content type: text, html, markdown, etc.

	// Relationships
	Post  *Post  `gorm:"foreignKey:ModelID" json:"post,omitempty"`
	User  *User  `gorm:"foreignKey:ModelID" json:"user,omitempty"`
	Media *Media `gorm:"foreignKey:ModelID" json:"media,omitempty"`
}

// TableName specifies the table name for Content
func (Content) TableName() string {
	return "contents"
}

// BeforeCreate sets timestamps and validates content data
func (c *Content) BeforeCreate(tx *gorm.DB) error {
	if err := c.Base.BeforeCreate(tx); err != nil {
		return err
	}

	// Clean content
	c.Content = strings.TrimSpace(c.Content)
	c.ModelType = strings.ToLower(strings.TrimSpace(c.ModelType))

	// Set default type
	if c.Type == "" {
		c.Type = "text"
	}

	return c.validate()
}

// BeforeUpdate validates content data before update
func (c *Content) BeforeUpdate(tx *gorm.DB) error {
	if err := c.Base.BeforeUpdate(tx); err != nil {
		return err
	}

	// Clean content
	c.Content = strings.TrimSpace(c.Content)
	c.ModelType = strings.ToLower(strings.TrimSpace(c.ModelType))

	return c.validate()
}

// validate performs validation on content fields
func (c *Content) validate() error {
	if c.ModelID == 0 {
		return errors.New("model ID is required")
	}

	if c.ModelType == "" {
		return errors.New("model type is required")
	}

	if len(strings.TrimSpace(c.Content)) < 1 {
		return errors.New("content cannot be empty")
	}

	validTypes := []string{"text", "html", "markdown", "json", "xml"}
	typeValid := false
	for _, contentType := range validTypes {
		if c.Type == contentType {
			typeValid = true
			break
		}
	}
	if !typeValid {
		return errors.New("invalid content type")
	}

	validModelTypes := []string{"post", "user", "media", "category", "comment"}
	modelTypeValid := false
	for _, modelType := range validModelTypes {
		if c.ModelType == modelType {
			modelTypeValid = true
			break
		}
	}
	if !modelTypeValid {
		return errors.New("invalid model type")
	}

	return nil
}

// IsText checks if the content is plain text
func (c *Content) IsText() bool {
	return c.Type == "text"
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

// IsXML checks if the content is XML
func (c *Content) IsXML() bool {
	return c.Type == "xml"
}

// GetWordCount returns the number of words in the content
func (c *Content) GetWordCount() int {
	words := strings.Fields(c.Content)
	return len(words)
}

// GetCharacterCount returns the number of characters in the content
func (c *Content) GetCharacterCount() int {
	return len(c.Content)
}

// GetLineCount returns the number of lines in the content
func (c *Content) GetLineCount() int {
	lines := strings.Split(c.Content, "\n")
	return len(lines)
}
