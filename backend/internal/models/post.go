package models

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// Post represents a blog post or article
type Post struct {
	BaseWithUser
	Title      string `gorm:"not null;size:255;index" json:"title" validate:"required,min=3,max=255"`
	Content    string `gorm:"type:text;not null" json:"content" validate:"required,min=10"`
	Slug       string `gorm:"uniqueIndex;not null;size:255" json:"slug"`
	Excerpt    string `gorm:"size:500" json:"excerpt,omitempty"`
	Status     string `gorm:"default:'draft';index;size:20" json:"status" validate:"oneof=draft published archived"`
	CategoryID uint   `gorm:"not null;index" json:"category_id" validate:"required"`

	// Relationships
	Category Category  `gorm:"foreignKey:CategoryID;constraint:OnDelete:RESTRICT" json:"category,omitempty"`
	Comments []Comment `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
	Contents []Content `gorm:"foreignKey:ModelID;constraint:OnDelete:CASCADE" json:"contents,omitempty"`
	Medias   []Media   `gorm:"foreignKey:ModelID;constraint:OnDelete:CASCADE" json:"medias,omitempty"`
}

// TableName specifies the table name for Post
func (Post) TableName() string {
	return "posts"
}

// BeforeCreate sets timestamps and validates post data
func (p *Post) BeforeCreate(tx *gorm.DB) error {
	if err := p.BaseWithUser.BeforeCreate(tx); err != nil {
		return err
	}

	// Generate slug if not provided
	if p.Slug == "" {
		p.Slug = p.generateSlug()
	}

	// Set default status
	if p.Status == "" {
		p.Status = "draft"
	}

	// Generate excerpt if not provided
	if p.Excerpt == "" && len(p.Content) > 0 {
		p.Excerpt = p.generateExcerpt()
	}

	return p.validate()
}

// BeforeUpdate validates post data before update
func (p *Post) BeforeUpdate(tx *gorm.DB) error {
	if err := p.BaseWithUser.BeforeUpdate(tx); err != nil {
		return err
	}

	// Generate slug if title changed and slug is empty
	if p.Slug == "" {
		p.Slug = p.generateSlug()
	}

	// Generate excerpt if not provided
	if p.Excerpt == "" && len(p.Content) > 0 {
		p.Excerpt = p.generateExcerpt()
	}

	return p.validate()
}

// validate performs validation on post fields
func (p *Post) validate() error {
	if len(p.Title) < 3 || len(p.Title) > 255 {
		return errors.New("title must be between 3 and 255 characters")
	}

	if len(p.Content) < 10 {
		return errors.New("content must be at least 10 characters")
	}

	if p.CategoryID == 0 {
		return errors.New("category is required")
	}

	validStatuses := []string{"draft", "published", "archived"}
	statusValid := false
	for _, status := range validStatuses {
		if p.Status == status {
			statusValid = true
			break
		}
	}
	if !statusValid {
		return errors.New("invalid status")
	}

	return nil
}

// generateSlug creates a URL-friendly slug from the title
func (p *Post) generateSlug() string {
	slug := strings.ToLower(p.Title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	// Remove special characters except hyphens
	// This is a simplified version - you might want to use a proper slug library
	return slug
}

// generateExcerpt creates a short excerpt from the content
func (p *Post) generateExcerpt() string {
	if len(p.Content) <= 150 {
		return p.Content
	}

	// Find the first sentence or cut at 150 characters
	excerpt := p.Content[:150]
	if idx := strings.LastIndex(excerpt, "."); idx > 100 {
		excerpt = excerpt[:idx+1]
	} else if idx := strings.LastIndex(excerpt, " "); idx > 100 {
		excerpt = excerpt[:idx] + "..."
	} else {
		excerpt += "..."
	}

	return excerpt
}

// IsPublished checks if the post is published
func (p *Post) IsPublished() bool {
	return p.Status == "published"
}

// IsDraft checks if the post is a draft
func (p *Post) IsDraft() bool {
	return p.Status == "draft"
}

// IsArchived checks if the post is archived
func (p *Post) IsArchived() bool {
	return p.Status == "archived"
}

// Publish marks the post as published
func (p *Post) Publish() {
	p.Status = "published"
}

// Archive marks the post as archived
func (p *Post) Archive() {
	p.Status = "archived"
}

// GetCommentCount returns the number of comments on this post
func (p *Post) GetCommentCount() int {
	return len(p.Comments)
}
