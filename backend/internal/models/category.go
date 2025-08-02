package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Category represents a content category
type Category struct {
	BaseModelWithOrdering
	Name        string `json:"name" gorm:"not null;size:100" validate:"required,min=1,max=100"`
	Slug        string `json:"slug" gorm:"uniqueIndex;not null;size:100" validate:"required,min=1,max=100"`
	Description string `json:"description" gorm:"size:500"`
	IsActive    bool   `json:"is_active" gorm:"default:true;index"`
	SortOrder   int    `json:"sort_order" gorm:"default:0;index"`

	// Relationships
	Parent   *Category  `json:"parent,omitempty" gorm:"foreignKey:ParentID;constraint:OnDelete:SET NULL"`
	Children []Category `json:"children,omitempty" gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE"`
	Posts    []Post     `json:"posts,omitempty" gorm:"foreignKey:CategoryID;constraint:OnDelete:SET NULL"`
}

// TableName specifies the table name for Category
func (Category) TableName() string {
	return "categories"
}

// BeforeCreate hook for Category
func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate hook for Category
func (c *Category) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// IsRoot checks if the category is a root category
func (c *Category) IsRoot() bool {
	return c.ParentID == nil
}

// HasChildren checks if the category has children
func (c *Category) HasChildren() bool {
	return len(c.Children) > 0
}

// GetDepth returns the depth of the category in the hierarchy
func (c *Category) GetDepth() int {
	return c.RecordDept
}

// GetIsActive returns the active status
func (c *Category) GetIsActive() bool {
	return c.IsActive
}

// Activate activates the category
func (c *Category) Activate() {
	c.IsActive = true
}

// Deactivate deactivates the category
func (c *Category) Deactivate() {
	c.IsActive = false
}
