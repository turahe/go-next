package dto

import (
	"time"
	"wordpress-go-next/backend/internal/models"
)

type CategoryDTO struct {
	ID          uint64             `json:"id"`
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description,omitempty"`
	IsActive    bool               `json:"isActive"`
	ParentID    *uint64            `json:"parentId,omitempty"`
	Parent      *CategorySimpleDTO `json:"parent,omitempty"`
	ChildCount  int                `json:"childCount"`
	PostCount   int                `json:"postCount"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}

func ToCategoryDTO(c *models.Category) *CategoryDTO {
	var parent *CategorySimpleDTO
	if c.Parent != nil {
		parent = &CategorySimpleDTO{ID: c.Parent.ID, Name: c.Parent.Name}
	}
	return &CategoryDTO{
		ID:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		IsActive:    c.IsActive,
		ParentID:    c.ParentID,
		Parent:      parent,
		ChildCount:  len(c.Children),
		PostCount:   len(c.Posts),
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// ToCategoryDTOs maps a slice of models.Category to a slice of *CategoryDTO
func ToCategoryDTOs(categories []models.Category) []*CategoryDTO {
	dtos := make([]*CategoryDTO, len(categories))
	for i, cat := range categories {
		dtos[i] = ToCategoryDTO(&cat)
	}
	return dtos
}
