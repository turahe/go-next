package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Comment represents a comment on a post
type Comment struct {
	BaseModelWithOrdering
	ModelID     uuid.UUID     `json:"model_id" gorm:"type:uuid;not null;index" validate:"required"`
	ModelType   ModelType     `json:"model_type" gorm:"type:varchar(255);not null;index" validate:"required"`
	Title       string        `json:"title" gorm:"size:255" validate:"required,min=1,max=255"`
	Description string        `json:"description" gorm:"size:500"`
	Status      CommentStatus `json:"status" gorm:"default:'pending';index"`
	ApprovedAt  *time.Time    `json:"approved_at,omitempty" gorm:"index"`

	// Relationships
	User     *User     `json:"user,omitempty" gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE"`
	Post     *Post     `json:"post,omitempty" gorm:"foreignKey:ModelID;constraint:OnDelete:CASCADE"`
	Parent   *Comment  `json:"parent,omitempty" gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE"`
	Children []Comment `json:"children,omitempty" gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for Comment
func (Comment) TableName() string {
	return "comments"
}

// BeforeCreate hook for Comment
func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate hook for Comment
func (c *Comment) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// IsApproved checks if the comment is approved
func (c *Comment) IsApproved() bool {
	return c.Status == CommentStatusApproved
}

// IsRejected checks if the comment is rejected
func (c *Comment) IsRejected() bool {
	return c.Status == CommentStatusRejected
}

// IsPending checks if the comment is pending approval
func (c *Comment) IsPending() bool {
	return c.Status == CommentStatusPending
}

// IsRoot checks if the comment is a root comment
func (c *Comment) IsRoot() bool {
	return c.ParentID == nil
}

// HasChildren checks if the comment has children
func (c *Comment) HasChildren() bool {
	return len(c.Children) > 0
}

// GetDepth returns the depth of the comment in the thread
func (c *Comment) GetDepth() int {
	return c.RecordDept
}

// Approve approves the comment
func (c *Comment) Approve() {
	c.Status = CommentStatusApproved
	now := time.Now()
	c.ApprovedAt = &now
}

// Reject rejects the comment
func (c *Comment) Reject() {
	c.Status = CommentStatusRejected
	c.ApprovedAt = nil
}
