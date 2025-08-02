package dto

import (
	"time"
	"wordpress-go-next/backend/internal/models"
)

type PostDTO struct {
	ID           uint64             `json:"id"`
	Title        string             `json:"title"`
	Slug         string             `json:"slug"`
	Excerpt      string             `json:"excerpt,omitempty"`
	Status       string             `json:"status"`
	CategoryID   uint64             `json:"categoryId"`
	Category     *CategorySimpleDTO `json:"category,omitempty"`
	CommentCount int                `json:"commentCount"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
}

type CategorySimpleDTO struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

func ToPostDTO(p *models.Post) *PostDTO {
	var cat *CategorySimpleDTO
	if p.Category.ID != 0 {
		cat = &CategorySimpleDTO{ID: p.Category.ID, Name: p.Category.Name}
	}
	return &PostDTO{
		ID:           p.ID,
		Title:        p.Title,
		Slug:         p.Slug,
		Excerpt:      p.Excerpt,
		Status:       p.Status,
		CategoryID:   uint64(p.CategoryID),
		Category:     cat,
		CommentCount: len(p.Comments),
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

// ToPostDTOs maps a slice of models.Post to a slice of *PostDTO
func ToPostDTOs(posts []models.Post) []*PostDTO {
	dtos := make([]*PostDTO, len(posts))
	for i, post := range posts {
		dtos[i] = ToPostDTO(&post)
	}
	return dtos
}
