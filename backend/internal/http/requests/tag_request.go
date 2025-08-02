package requests

import (
	"fmt"
	"strings"
	"go-next/internal/models"
)

// CreateTagRequest represents the request for creating a new tag
type CreateTagRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Slug        string `json:"slug" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	Color       string `json:"color" binding:"max=7"`
	Type        string `json:"type" binding:"required,oneof=general category feature system"`
	IsActive    bool   `json:"is_active"`
}

// UpdateTagRequest represents the request for updating an existing tag
type UpdateTagRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Slug        string `json:"slug" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	Color       string `json:"color" binding:"max=7"`
	Type        string `json:"type" binding:"required,oneof=general category feature system"`
	IsActive    bool   `json:"is_active"`
}

// TagSearchRequest represents the request for searching tags
type TagSearchRequest struct {
	Query  string `form:"query" binding:"required,min=1"`
	Type   string `form:"type" binding:"omitempty,oneof=general category feature system"`
	Limit  int    `form:"limit" binding:"min=1,max=100"`
	Offset int    `form:"offset" binding:"min=0"`
}

// TagListRequest represents the request for listing tags
type TagListRequest struct {
	Type   string `form:"type" binding:"omitempty,oneof=general category feature system"`
	Active *bool  `form:"active"`
	Limit  int    `form:"limit" binding:"min=1,max=100"`
	Offset int    `form:"offset" binding:"min=0"`
}

// AddTagToEntityRequest represents the request for adding a tag to an entity
type AddTagToEntityRequest struct {
	TagID      uint64 `json:"tag_id" binding:"required"`
	EntityID   uint64 `json:"entity_id" binding:"required"`
	EntityType string `json:"entity_type" binding:"required,oneof=post user media category comment"`
	Group      string `json:"group" binding:"max=50"`
}

// RemoveTagFromEntityRequest represents the request for removing a tag from an entity
type RemoveTagFromEntityRequest struct {
	TagID      uint64 `json:"tag_id" binding:"required"`
	EntityID   uint64 `json:"entity_id" binding:"required"`
	EntityType string `json:"entity_type" binding:"required,oneof=post user media category comment"`
}

// GetEntitiesByTagRequest represents the request for getting entities by tag
type GetEntitiesByTagRequest struct {
	TagID      uint64 `form:"tag_id" binding:"required"`
	EntityType string `form:"entity_type" binding:"required,oneof=post user media category comment"`
	Limit      int    `form:"limit" binding:"min=1,max=100"`
	Offset     int    `form:"offset" binding:"min=0"`
}

// BulkTagRequest represents the request for bulk tagging operations
type BulkTagRequest struct {
	TagIDs     []uint64 `json:"tag_ids" binding:"required,min=1"`
	EntityID   uint64   `json:"entity_id" binding:"required"`
	EntityType string   `json:"entity_type" binding:"required,oneof=post user media category comment"`
	Group      string   `json:"group" binding:"max=50"`
}

// BulkUntagRequest represents the request for bulk untagging operations
type BulkUntagRequest struct {
	TagIDs     []uint64 `json:"tag_ids" binding:"required,min=1"`
	EntityID   uint64   `json:"entity_id" binding:"required"`
	EntityType string   `json:"entity_type" binding:"required,oneof=post user media category comment"`
}

// TagStatisticsRequest represents the request for tag statistics
type TagStatisticsRequest struct {
	TagID      uint64 `form:"tag_id" binding:"required"`
	EntityType string `form:"entity_type" binding:"omitempty,oneof=post user media category comment"`
}

// Validate validates the CreateTagRequest
func (r *CreateTagRequest) Validate() error {
	// Validate color format if provided
	if r.Color != "" && !isValidHexColor(r.Color) {
		return fmt.Errorf("invalid hex color format")
	}

	// Validate slug format
	if !isValidSlug(r.Slug) {
		return fmt.Errorf("invalid slug format")
	}

	return nil
}

// Validate validates the UpdateTagRequest
func (r *UpdateTagRequest) Validate() error {
	// Validate color format if provided
	if r.Color != "" && !isValidHexColor(r.Color) {
		return fmt.Errorf("invalid hex color format")
	}

	// Validate slug format
	if !isValidSlug(r.Slug) {
		return fmt.Errorf("invalid slug format")
	}

	return nil
}

// ToTag converts CreateTagRequest to Tag model
func (r *CreateTagRequest) ToTag() *models.Tag {
	tag := &models.Tag{
		Name:        strings.TrimSpace(r.Name),
		Slug:        strings.TrimSpace(r.Slug),
		Description: strings.TrimSpace(r.Description),
		Color:       strings.TrimSpace(r.Color),
		Type:        r.Type,
		IsActive:    r.IsActive,
	}

	// Set default color if not provided
	if tag.Color == "" {
		tag.SetDefaultColor()
	}

	return tag
}

// ToTag converts UpdateTagRequest to Tag model
func (r *UpdateTagRequest) ToTag() *models.Tag {
	tag := &models.Tag{
		Name:        strings.TrimSpace(r.Name),
		Slug:        strings.TrimSpace(r.Slug),
		Description: strings.TrimSpace(r.Description),
		Color:       strings.TrimSpace(r.Color),
		Type:        r.Type,
		IsActive:    r.IsActive,
	}

	// Set default color if not provided
	if tag.Color == "" {
		tag.SetDefaultColor()
	}

	return tag
}

// ToTaggedEntity converts AddTagToEntityRequest to TaggedEntity model
func (r *AddTagToEntityRequest) ToTaggedEntity() *models.TaggedEntity {
	group := strings.TrimSpace(r.Group)
	if group == "" {
		group = "default"
	}

	return &models.TaggedEntity{
		TagID:      r.TagID,
		EntityID:   r.EntityID,
		EntityType: r.EntityType,
		Group:      group,
	}
}

// Helper functions

// isValidHexColor checks if the color is a valid hex color
func isValidHexColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}
	for i := 1; i < 7; i++ {
		c := color[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// isValidSlug checks if the slug is valid
func isValidSlug(slug string) bool {
	if len(slug) == 0 {
		return false
	}

	// Check for valid characters (alphanumeric, hyphens, underscores)
	for _, char := range slug {
		if !((char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}

	return true
}
