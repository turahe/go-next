package dto

import (
	"go-next/internal/models"
)

// TagDTO represents a tag in API responses
type TagDTO struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Type        string `json:"type"`
	IsActive    bool   `json:"is_active"`
}

// ToTagDTO converts a Tag model to TagDTO
func ToTagDTO(tag *models.Tag) *TagDTO {
	return &TagDTO{
		ID:          tag.ID,
		Name:        tag.Name,
		Slug:        tag.Slug,
		Description: tag.Description,
		Color:       tag.Color,
		Type:        tag.Type,
		IsActive:    tag.IsActive,
	}
}

// ToTagDTOs converts a slice of Tag models to TagDTOs
func ToTagDTOs(tags []models.Tag) []*TagDTO {
	result := make([]*TagDTO, len(tags))
	for i, tag := range tags {
		result[i] = ToTagDTO(&tag)
	}
	return result
}
