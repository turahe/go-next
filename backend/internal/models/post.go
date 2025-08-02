package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Post represents a blog post or article
type Post struct {
	BaseModelWithUser
	Title       string     `json:"title" gorm:"not null;size:255" validate:"required,min=1,max=255"`
	Slug        string     `json:"slug" gorm:"uniqueIndex;not null;size:255" validate:"required,min=1,max=255"`
	Content     string     `json:"content" gorm:"type:text;not null"`
	Excerpt     string     `json:"excerpt" gorm:"size:500"`
	Status      string     `json:"status" gorm:"default:'draft';index" validate:"oneof=draft published archived"`
	Public      bool       `json:"public" gorm:"default:true;index"`
	PublishedAt *time.Time `json:"published_at,omitempty" gorm:"index"`
	ViewCount   int64      `json:"view_count" gorm:"default:0;index"`
	CategoryID  *uuid.UUID `json:"category_id,omitempty" gorm:"type:uuid;index"`

	// Relationships
	Category *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID;constraint:OnDelete:SET NULL"`
	Comments []Comment `json:"comments,omitempty" gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`
	Contents []Content `json:"contents,omitempty" gorm:"polymorphic:Model;polymorphicValue:post;constraint:OnDelete:CASCADE"`
	Media    []Media   `json:"media,omitempty" gorm:"many2many:mediables;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for Post
func (Post) TableName() string {
	return "posts"
}

// BeforeCreate hook for Post
func (p *Post) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate hook for Post
func (p *Post) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// IncrementViewCount increments the view count
func (p *Post) IncrementViewCount() {
	p.ViewCount++
}

// IsPublished checks if the post is published
func (p *Post) IsPublished() bool {
	return p.Status == "published" && p.PublishedAt != nil
}

// IsPublic checks if the post is public
func (p *Post) IsPublic() bool {
	return p.Public
}

// Publish publishes the post
func (p *Post) Publish() {
	p.Status = "published"
	now := time.Now()
	p.PublishedAt = &now
}

// Unpublish unpublishes the post
func (p *Post) Unpublish() {
	p.Status = "draft"
	p.PublishedAt = nil
}

// Archive archives the post
func (p *Post) Archive() {
	p.Status = "archived"
}
