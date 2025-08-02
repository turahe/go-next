package dto

import (
	"time"
	"wordpress-go-next/backend/internal/models"
)

type CommentDTO struct {
	ID         uint64        `json:"id"`
	UserID     uint64        `json:"userId"`
	PostID     uint64        `json:"postId"`
	Status     string        `json:"status"`
	ParentID   *uint64       `json:"parentId,omitempty"`
	Content    *string       `json:"content,omitempty"`
	CreatedAt  time.Time     `json:"createdAt"`
	UpdatedAt  time.Time     `json:"updatedAt"`
	ReplyCount int           `json:"replyCount"`
	User       *UserDTO      `json:"user,omitempty"`
	Children   []*CommentDTO `json:"children,omitempty"`
}

func ToCommentDTO(c *models.Comment) *CommentDTO {
	var content *string
	if c.Content != nil {
		content = &c.Content.Content
	}
	var user *UserDTO
	if c.User.ID != 0 {
		user = ToUserDTO(&c.User)
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
