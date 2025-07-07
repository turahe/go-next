package models

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// Comment represents a comment on a post with hierarchical structure
type Comment struct {
	BaseWithHierarchy
	Content string `gorm:"type:text;not null" json:"content" validate:"required,min=1,max=10000"`
	UserID  uint   `gorm:"not null;index" json:"user_id" validate:"required"`
	PostID  uint   `gorm:"not null;index" json:"post_id" validate:"required"`
	Status  string `gorm:"default:'pending';index;size:20" json:"status" validate:"oneof=pending approved rejected"`

	// Relationships
	User     User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Post     Post      `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"post,omitempty"`
	Medias   []Media   `gorm:"foreignKey:ModelID;constraint:OnDelete:CASCADE" json:"medias,omitempty"`
	Parent   *Comment  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Comment `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// TableName specifies the table name for Comment
func (Comment) TableName() string {
	return "comments"
}

// BeforeCreate sets timestamps and validates comment data
func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if err := c.BaseWithHierarchy.BeforeCreate(tx); err != nil {
		return err
	}
	// Set default status
	if c.Status == "" {
		c.Status = "pending"
	}

	// Clean content
	c.Content = strings.TrimSpace(c.Content)

	return c.validate()
}

// BeforeUpdate validates comment data before update
func (c *Comment) BeforeUpdate(tx *gorm.DB) error {
	if err := c.BaseWithHierarchy.BeforeUpdate(tx); err != nil {
		return err
	}

	// Clean content
	c.Content = strings.TrimSpace(c.Content)

	return c.validate()
}

// validate performs validation on comment fields
func (c *Comment) validate() error {
	if len(strings.TrimSpace(c.Content)) < 1 {
		return errors.New("comment content cannot be empty")
	}

	if len(c.Content) > 10000 {
		return errors.New("comment content cannot exceed 10000 characters")
	}

	if c.UserID == 0 {
		return errors.New("user is required")
	}

	if c.PostID == 0 {
		return errors.New("post is required")
	}

	validStatuses := []string{"pending", "approved", "rejected"}
	statusValid := false
	for _, status := range validStatuses {
		if c.Status == status {
			statusValid = true
			break
		}
	}
	if !statusValid {
		return errors.New("invalid status")
	}

	// Prevent circular references
	if c.ParentID != nil && *c.ParentID == int64(c.ID) {
		return errors.New("comment cannot be its own parent")
	}

	return nil
}

// IsRoot checks if this comment is a root comment (no parent)
func (c *Comment) IsRoot() bool {
	return c.ParentID == nil
}

// IsReply checks if this comment is a reply to another comment
func (c *Comment) IsReply() bool {
	return c.ParentID != nil
}

// IsApproved checks if the comment is approved
func (c *Comment) IsApproved() bool {
	return c.Status == "approved"
}

// IsPending checks if the comment is pending approval
func (c *Comment) IsPending() bool {
	return c.Status == "pending"
}

// IsRejected checks if the comment is rejected
func (c *Comment) IsRejected() bool {
	return c.Status == "rejected"
}

// Approve marks the comment as approved
func (c *Comment) Approve() {
	c.Status = "approved"
}

// Reject marks the comment as rejected
func (c *Comment) Reject() {
	c.Status = "rejected"
}

// GetDepth returns the depth of this comment in the hierarchy
func (c *Comment) GetDepth() int64 {
	if c.RecordDept != nil {
		return *c.RecordDept
	}
	return 0
}

// GetReplyCount returns the number of replies to this comment
func (c *Comment) GetReplyCount() int {
	return len(c.Children)
}

// HasReplies checks if this comment has replies
func (c *Comment) HasReplies() bool {
	return len(c.Children) > 0
}

// GetWordCount returns the number of words in the comment
func (c *Comment) GetWordCount() int {
	words := strings.Fields(c.Content)
	return len(words)
}
