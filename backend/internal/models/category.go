package models

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// Category represents a content category with hierarchical structure
type Category struct {
	BaseWithHierarchy
	Name        string `gorm:"uniqueIndex;not null;size:100" json:"name" validate:"required,min=2,max=100"`
	Description string `gorm:"type:text" json:"description,omitempty"`
	Slug        string `gorm:"uniqueIndex;not null;size:100" json:"slug"`
	IsActive    bool   `gorm:"default:true;index" json:"is_active"`

	// Relationships
	Posts    []Post     `gorm:"foreignKey:CategoryID;constraint:OnDelete:RESTRICT" json:"posts,omitempty"`
	Medias   []Media    `gorm:"foreignKey:ModelID;constraint:OnDelete:CASCADE" json:"medias,omitempty"`
	Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// TableName specifies the table name for Category
func (Category) TableName() string {
	return "categories"
}

// BeforeCreate sets timestamps and validates category data
func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if err := c.BaseWithHierarchy.BeforeCreate(tx); err != nil {
		return err
	}

	// Generate slug if not provided
	if c.Slug == "" {
		c.Slug = c.generateSlug()
	}

	// Set default values
	if c.IsActive == false {
		c.IsActive = true
	}

	return c.validate()
}

// BeforeUpdate validates category data before update
func (c *Category) BeforeUpdate(tx *gorm.DB) error {
	if err := c.BaseWithHierarchy.BeforeUpdate(tx); err != nil {
		return err
	}

	// Generate slug if name changed and slug is empty
	if c.Slug == "" {
		c.Slug = c.generateSlug()
	}

	return c.validate()
}

// validate performs validation on category fields
func (c *Category) validate() error {
	if len(c.Name) < 2 || len(c.Name) > 100 {
		return errors.New("category name must be between 2 and 100 characters")
	}

	if len(c.Slug) < 2 || len(c.Slug) > 100 {
		return errors.New("category slug must be between 2 and 100 characters")
	}

	// Prevent circular references
	if c.ParentID != nil && *c.ParentID == c.ID {
		return errors.New("category cannot be its own parent")
	}

	return nil
}

// generateSlug creates a URL-friendly slug from the name
func (c *Category) generateSlug() string {
	slug := strings.ToLower(c.Name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	// Remove special characters except hyphens
	// This is a simplified version - you might want to use a proper slug library
	return slug
}

// IsRoot checks if this category is a root category (no parent)
func (c *Category) IsRoot() bool {
	return c.ParentID == nil
}

// IsLeaf checks if this category has no children
func (c *Category) IsLeaf() bool {
	return len(c.Children) == 0
}

// GetDepth returns the depth of this category in the hierarchy
func (c *Category) GetDepth() uint64 {
	if c.RecordDept != nil {
		return *c.RecordDept
	}
	return 0
}

// GetPostCount returns the number of posts in this category
func (c *Category) GetPostCount() int {
	return len(c.Posts)
}

// GetChildCount returns the number of child categories
func (c *Category) GetChildCount() int {
	return len(c.Children)
}

// HasChildren checks if this category has child categories
func (c *Category) HasChildren() bool {
	return len(c.Children) > 0
}

// GetFullPath returns the full hierarchical path of this category
func (c *Category) GetFullPath() string {
	if c.Parent == nil {
		return c.Name
	}
	return c.Parent.GetFullPath() + " > " + c.Name
}
