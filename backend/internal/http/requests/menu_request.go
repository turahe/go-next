package requests

import (
	"github.com/google/uuid"
)

type MenuCreateRequest struct {
	Name        string     `json:"name" validate:"required,min=2,max=50"`
	Description string     `json:"description" validate:"omitempty,max=255"`
	Icon        string     `json:"icon" validate:"omitempty,max=50"`
	URL         string     `json:"url" validate:"omitempty,max=255"`
	ParentID    *uuid.UUID `json:"parent_id" validate:"omitempty"`
	Ordering    int        `json:"ordering" validate:"omitempty,min=0"`
}

type MenuUpdateRequest struct {
	Name        string     `json:"name" validate:"required,min=2,max=50"`
	Description string     `json:"description" validate:"omitempty,max=255"`
	Icon        string     `json:"icon" validate:"omitempty,max=50"`
	URL         string     `json:"url" validate:"omitempty,max=255"`
	ParentID    *uuid.UUID `json:"parent_id" validate:"omitempty"`
	Ordering    int        `json:"ordering" validate:"omitempty,min=0"`
}
