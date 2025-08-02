package requests

import (
	"github.com/google/uuid"
)

type PostCreateRequest struct {
	Title      string    `json:"title" validate:"required,min=3,max=255"`
	Content    string    `json:"content" validate:"required,min=1"`
	UserID     uuid.UUID `json:"user_id" validate:"required"`
	CategoryID uuid.UUID `json:"category_id" validate:"required"`
}

type PostUpdateRequest struct {
	Title      string    `json:"title" validate:"required,min=3,max=255"`
	Content    string    `json:"content" validate:"required,min=1"`
	UserID     uuid.UUID `json:"user_id" validate:"required"`
	CategoryID uuid.UUID `json:"category_id" validate:"required"`
}
