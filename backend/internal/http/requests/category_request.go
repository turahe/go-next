package requests

import (
	"github.com/google/uuid"
)

type CategoryCreateRequest struct {
	Name        string     `json:"name" validate:"required,min=2,max=100"`
	Description string     `json:"description" validate:"omitempty,max=500"`
	ParentID    *uuid.UUID `json:"parent_id" validate:"omitempty"` // Optional parent category ID
}

type CategoryUpdateRequest struct {
	Name        string     `json:"name" validate:"required,min=2,max=100"`
	Description string     `json:"description" validate:"omitempty,max=500"`
	ParentID    *uuid.UUID `json:"parent_id" validate:"omitempty"` // Optional parent category ID
}
