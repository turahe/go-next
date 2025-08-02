package dto

import (
	"go-next/internal/models"
	"time"
)

type TagDTO struct {
	ID                  uint64    `json:"id"`
	Name                string    `json:"name"`
	Slug                string    `json:"slug"`
	Description         string    `json:"description"`
	Color               string    `json:"color"`
	Type                string    `json:"type"`
	IsActive            bool      `json:"isActive"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
	TaggedEntitiesCount int64     `json:"taggedEntitiesCount"`
}

func ToTagDTO(t *models.Tag) *TagDTO {
	return &TagDTO{
		ID:                  t.ID,
		Name:                t.Name,
		Slug:                t.Slug,
		Description:         t.Description,
		Color:               t.Color,
		Type:                t.Type,
		IsActive:            t.IsActive,
		CreatedAt:           t.CreatedAt,
		UpdatedAt:           t.UpdatedAt,
		TaggedEntitiesCount: int64(len(t.TaggedEntities)),
	}
}

func ToTagDTOs(tags []models.Tag) []*TagDTO {
	dtos := make([]*TagDTO, len(tags))
	for i, t := range tags {
		dtos[i] = ToTagDTO(&t)
	}
	return dtos
}
