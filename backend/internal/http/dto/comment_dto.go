package dto

import (
	"time"
	"github.com/google/uuid"
	"go-next/internal/models"
)

type CommentDTO struct {
	ID         uuid.UUID     `json:"id"`
	UserID     uuid.UUID     `json:"userId"`
	PostID     uuid.UUID     `json:"postId"`
	Status     string        `json:"status"`
	ParentID   *uuid.UUID    `json:"parentId,omitempty"`
	Content    *string       `json:"content,omitempty"`
	CreatedAt  time.Time     `json:"createdAt"`
	UpdatedAt  time.Time     `json:"updatedAt"`
	ReplyCount int           `json:"replyCount"`
	User       *UserDTO      `json:"user,omitempty"`
	Children   []*CommentDTO `json:"children,omitempty"`
}

func ToCommentDTO(c *models.Comment) *CommentDTO {
	var content *string
	if c.Content != "" {
		content = &c.Content
	}
	var user *UserDTO
	if c.User != nil && c.User.ID != uuid.Nil {
		user = ToUserDTO(c.User)
	}
	children := make([]*CommentDTO, len(c.Children))
	for i, child := range c.Children {
		children[i] = ToCommentDTO(&child)
	}
	return &CommentDTO{
		ID:         c.ID,
		UserID:     c.UserID,
		PostID:     c.PostID,
		Status:     c.Status,
		ParentID:   c.ParentID,
		Content:    content,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
		ReplyCount: len(c.Children),
		User:       user,
		Children:   children,
	}
}

// ToCommentDTOs maps a slice of models.Comment to a slice of *CommentDTO
func ToCommentDTOs(comments []models.Comment) []*CommentDTO {
	dtos := make([]*CommentDTO, len(comments))
	for i, comment := range comments {
		dtos[i] = ToCommentDTO(&comment)
	}
	return dtos
}
