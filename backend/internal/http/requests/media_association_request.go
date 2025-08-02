package requests

import (
	"github.com/google/uuid"
)

type MediaAssociationInput struct {
	MediableID   uuid.UUID `json:"mediable_id" binding:"required"`
	MediableType string    `json:"mediable_type" binding:"required"`
	Group        string    `json:"group" binding:"required"`
}

type MediaCreateRequest struct {
	Name   string    `json:"name" validate:"required,min=1,max=255"`
	URL    string    `json:"url" validate:"required,url"`
	Type   string    `json:"type" validate:"required"`
	Size   int64     `json:"size" validate:"required,gt=0"`
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

type MediaUpdateRequest struct {
	Name   string    `json:"name" validate:"required,min=1,max=255"`
	URL    string    `json:"url" validate:"required,url"`
	Type   string    `json:"type" validate:"required"`
	Size   int64     `json:"size" validate:"required,gt=0"`
	UserID uuid.UUID `json:"user_id" validate:"required"`
}
