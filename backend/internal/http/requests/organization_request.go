package requests

import (
	"go-next/internal/models"

	"github.com/google/uuid"
)

// OrganizationRequest represents the request structure for organization operations
type OrganizationRequest struct {
	Name        string                  `json:"name" validate:"required,min=1,max=100"`
	Slug        string                  `json:"slug" validate:"required,min=1,max=100"`
	Description string                  `json:"description" validate:"max=500"`
	Code        string                  `json:"code" validate:"required,min=1,max=100"`
	Type        models.OrganizationType `json:"type" validate:"required"`
	ParentID    uuid.UUID               `json:"parent_id"`
}
