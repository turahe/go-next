package requests

import (
	"time"

	"github.com/google/uuid"
)

// BlogPostCreateRequest represents the request structure for creating a blog post
type BlogPostCreateRequest struct {
	Title       string     `json:"title" binding:"required,min=1,max=255" validate:"required,min=1,max=255"`
	Slug        string     `json:"slug" binding:"required,min=1,max=255" validate:"required,min=1,max=255"`
	Content     string     `json:"content" binding:"required" validate:"required"`
	Excerpt     string     `json:"excerpt" binding:"max=500" validate:"max=500"`
	Status      string     `json:"status" binding:"oneof=draft published archived" validate:"oneof=draft published archived"`
	Public      bool       `json:"public" binding:"omitempty"`
	PublishedAt *time.Time `json:"published_at" binding:"omitempty"`
	CategoryID  *uuid.UUID `json:"category_id" binding:"omitempty"`
}

// BlogPostUpdateRequest represents the request structure for updating a blog post
type BlogPostUpdateRequest struct {
	Title       string     `json:"title" binding:"omitempty,min=1,max=255" validate:"omitempty,min=1,max=255"`
	Slug        string     `json:"slug" binding:"omitempty,min=1,max=255" validate:"omitempty,min=1,max=255"`
	Content     string     `json:"content" binding:"omitempty" validate:"omitempty"`
	Excerpt     string     `json:"excerpt" binding:"omitempty,max=500" validate:"omitempty,max=500"`
	Status      string     `json:"status" binding:"omitempty,oneof=draft published archived" validate:"omitempty,oneof=draft published archived"`
	Public      *bool      `json:"public" binding:"omitempty"`
	PublishedAt *time.Time `json:"published_at" binding:"omitempty"`
	CategoryID  *uuid.UUID `json:"category_id" binding:"omitempty"`
}

// BlogSearchRequest represents the request structure for searching blog posts
type BlogSearchRequest struct {
	Query   string `json:"query" binding:"required,min=1" validate:"required,min=1"`
	Page    int    `json:"page" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	PerPage int    `json:"per_page" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100"`
}

// BlogCommentCreateRequest represents the request structure for creating a blog comment
type BlogCommentCreateRequest struct {
	Content  string     `json:"content" binding:"required,min=1" validate:"required,min=1"`
	PostID   uuid.UUID  `json:"post_id" binding:"required" validate:"required"`
	ParentID *uuid.UUID `json:"parent_id" binding:"omitempty" validate:"omitempty"`
}

// BlogCommentUpdateRequest represents the request structure for updating a blog comment
type BlogCommentUpdateRequest struct {
	Content string `json:"content" binding:"required,min=1" validate:"required,min=1"`
}

// BlogCategoryCreateRequest represents the request structure for creating a blog category
type BlogCategoryCreateRequest struct {
	Name        string     `json:"name" binding:"required,min=1,max=100" validate:"required,min=1,max=100"`
	Slug        string     `json:"slug" binding:"required,min=1,max=100" validate:"required,min=1,max=100"`
	Description string     `json:"description" binding:"omitempty,max=500" validate:"omitempty,max=500"`
	ParentID    *uuid.UUID `json:"parent_id" binding:"omitempty" validate:"omitempty"`
	SortOrder   int        `json:"sort_order" binding:"omitempty" validate:"omitempty"`
}

// BlogTagCreateRequest represents the request structure for creating a blog tag
type BlogTagCreateRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=50" validate:"required,min=1,max=50"`
	Slug     string `json:"slug" binding:"required,min=1,max=50" validate:"required,min=1,max=50"`
	Color    string `json:"color" binding:"omitempty,max=7" validate:"omitempty,max=7"`
	IsActive bool   `json:"is_active" binding:"omitempty"`
}

// BlogStatsRequest represents the request structure for blog statistics
type BlogStatsRequest struct {
	StartDate  *time.Time `json:"start_date" binding:"omitempty" validate:"omitempty"`
	EndDate    *time.Time `json:"end_date" binding:"omitempty" validate:"omitempty"`
	CategoryID *uuid.UUID `json:"category_id" binding:"omitempty" validate:"omitempty"`
}

// BlogPostPublishRequest represents the request structure for publishing a blog post
type BlogPostPublishRequest struct {
	PublishedAt *time.Time `json:"published_at" binding:"omitempty" validate:"omitempty"`
}

// BlogViewCountRequest represents the request structure for tracking view counts
type BlogViewCountRequest struct {
	PostID uuid.UUID  `json:"post_id" binding:"required" validate:"required"`
	UserID *uuid.UUID `json:"user_id" binding:"omitempty" validate:"omitempty"`
	IP     string     `json:"ip" binding:"omitempty" validate:"omitempty"`
}
