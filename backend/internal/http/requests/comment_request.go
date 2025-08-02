package requests

import (
	"github.com/google/uuid"
)

type CommentCreateRequest struct {
	Content string    `json:"content" validate:"required,min=1,max=1000"`
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	PostID  uuid.UUID `json:"post_id" validate:"required"`
}

type CommentUpdateRequest struct {
	Content string    `json:"content" validate:"required,min=1,max=1000"`
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	PostID  uuid.UUID `json:"post_id" validate:"required"`
}
