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
	Excerpt     string     `json:"excerpt" gorm:"size:500"`
	Description string     `json:"description" gorm:"size:500"`
	Status      PostStatus `json:"status" gorm:"default:'draft';index"`
	PublishedAt *time.Time `json:"published_at,omitempty" gorm:"index"`
	CategoryID  *uuid.UUID `json:"category_id,omitempty" gorm:"type:uuid;index"`
	ViewCount   int        `json:"view_count" gorm:"default:0"`
	Public      bool       `json:"public" gorm:"default:true"`

	// Relationships
	User     *User     `json:"user,omitempty" gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE"`
	Category *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID;constraint:OnDelete:SET NULL"`
	Comments []Comment `json:"comments,omitempty" gorm:"foreignKey:ModelID;constraint:OnDelete:CASCADE"`
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
	return p.Status == PostStatusPublished && p.PublishedAt != nil
}

// IsPublic checks if the post is public
func (p *Post) IsPublic() bool {
	return p.Public
}

// Publish publishes the post
func (p *Post) Publish() {
	p.Status = PostStatusPublished
	now := time.Now()
	p.PublishedAt = &now
}

// Unpublish unpublishes the post
func (p *Post) Unpublish() {
	p.Status = PostStatusDraft
	p.PublishedAt = nil
}

// Archive archives the post
func (p *Post) Archive() {
	p.Status = PostStatusArchived
}

// AddContent adds a new content block to the post
func (p *Post) AddContent(contentType, content string, sortOrder int) *Content {
	newContent := &Content{
		ModelID:    p.ID,
		ModelType:  string(ModelTypePost),
		Type:       contentType,
		ContentRaw: content,
		SortOrder:  sortOrder,
	}
	p.Contents = append(p.Contents, *newContent)
	return newContent
}

// GetContentByType gets all content blocks of a specific type
func (p *Post) GetContentByType(contentType string) []Content {
	var filtered []Content
	for _, content := range p.Contents {
		if content.Type == contentType {
			filtered = append(filtered, content)
		}
	}
	return filtered
}

// GetSortedContents gets all content blocks sorted by sort order
func (p *Post) GetSortedContents() []Content {
	// This would typically be handled by the database query with ORDER BY
	// For now, we'll return the contents as they are
	// In a real implementation, you'd want to sort them by SortOrder
	return p.Contents
}

// RemoveContent removes a content block by ID
func (p *Post) RemoveContent(contentID uuid.UUID) {
	for i, content := range p.Contents {
		if content.ID == contentID {
			p.Contents = append(p.Contents[:i], p.Contents[i+1:]...)
			break
		}
	}
}

// UpdateContent updates a content block
func (p *Post) UpdateContent(contentID uuid.UUID, contentType, content string, sortOrder int) bool {
	for i, c := range p.Contents {
		if c.ID == contentID {
			p.Contents[i].Type = contentType
			p.Contents[i].ContentRaw = content
			p.Contents[i].SortOrder = sortOrder
			return true
		}
	}
	return false
}

// HasContent checks if the post has any content blocks
func (p *Post) HasContent() bool {
	return len(p.Contents) > 0
}

// GetContentCount returns the number of content blocks
func (p *Post) GetContentCount() int {
	return len(p.Contents)
}
