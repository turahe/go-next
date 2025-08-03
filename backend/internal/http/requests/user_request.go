package requests

import "time"

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	UserName string `json:"name" validate:"required,min=2,max=50,unique_username,username"`
	Email    string `json:"email" validate:"required,email,unique_email"`
	Phone    string `json:"phone" validate:"omitempty,e164,unique_phone"`
	Password string `json:"password" validate:"required,min=8,max=255"`
	Role     string `json:"role" validate:"required,oneof=admin editor moderator user guest"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	UserName string `json:"name" validate:"omitempty,min=2,max=50,unique_username,username"`
	Email    string `json:"email" validate:"omitempty,email,unique_email"`
	Phone    string `json:"phone" validate:"omitempty,e164,unique_phone"`
	Password string `json:"password" validate:"omitempty,min=8,max=255"`
	Role     string `json:"role" validate:"omitempty,oneof=admin editor moderator user guest"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Identity string `json:"identity" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	UserName        string `json:"username" validate:"required,min=2,max=50,unique_username,username"`
	Email           string `json:"email" validate:"required,email,unique_email"`
	CountryCode     string `json:"country_code" validate:"omitempty,country_code"`
	Phone           string `json:"phone" validate:"omitempty,unique_phone"`
	Password        string `json:"password" validate:"required,min=8,max=255"`
	PasswordConfirm string `json:"password_confirm" validate:"required,eqfield=Password"`
	Terms           bool   `json:"terms" validate:"required"`
}

// CreatePostRequest represents a request to create a post
type CreatePostRequest struct {
	Title       string     `json:"title" validate:"required,min=5,max=255"`
	Content     string     `json:"content" validate:"required,min=10"`
	CategoryID  string     `json:"category_id" validate:"required,uuid"`
	Tags        []string   `json:"tags" validate:"omitempty,max=10"`
	PublishedAt *time.Time `json:"published_at" validate:"omitempty,datetime"`
	MetaData    string     `json:"meta_data" validate:"omitempty,json"`
}

// UpdatePostRequest represents a request to update a post
type UpdatePostRequest struct {
	Title       string     `json:"title" validate:"omitempty,min=5,max=255"`
	Content     string     `json:"content" validate:"omitempty,min=10"`
	CategoryID  string     `json:"category_id" validate:"omitempty,uuid"`
	Tags        []string   `json:"tags" validate:"omitempty,max=10"`
	PublishedAt *time.Time `json:"published_at" validate:"omitempty,datetime"`
	MetaData    string     `json:"meta_data" validate:"omitempty,json"`
}

// CreateCommentRequest represents a request to create a comment
type CreateCommentRequest struct {
	Content   string  `json:"content" validate:"required,min=1,max=1000"`
	PostID    string  `json:"post_id" validate:"required,uuid"`
	ParentID  *string `json:"parent_id" validate:"omitempty,uuid"`
	Anonymous bool    `json:"anonymous" validate:"omitempty"`
}

// UpdateCommentRequest represents a request to update a comment
type UpdateCommentRequest struct {
	Content   string `json:"content" validate:"required,min=1,max=1000"`
	Anonymous bool   `json:"anonymous" validate:"omitempty"`
}

// CreateCategoryRequest represents a request to create a category
type CreateCategoryRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=50"`
	Description string  `json:"description" validate:"omitempty,max=255"`
	Slug        string  `json:"slug" validate:"required,min=2,max=50"`
	ParentID    *string `json:"parent_id" validate:"omitempty,uuid"`
}

// UpdateCategoryRequest represents a request to update a category
type UpdateCategoryRequest struct {
	Name        string  `json:"name" validate:"omitempty,min=2,max=50"`
	Description string  `json:"description" validate:"omitempty,max=255"`
	Slug        string  `json:"slug" validate:"omitempty,min=2,max=50"`
	ParentID    *string `json:"parent_id" validate:"omitempty,uuid"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query   string `json:"query" validate:"required,min=1,max=100"`
	Page    int    `json:"page" validate:"omitempty,min=1"`
	PerPage int    `json:"per_page" validate:"omitempty,min=1,max=100"`
	SortBy  string `json:"sort_by" validate:"omitempty,oneof=created_at updated_at title name"`
	Order   string `json:"order" validate:"omitempty,oneof=asc desc"`
}

// FilterRequest represents a filter request
type FilterRequest struct {
	CategoryID string   `json:"category_id" validate:"omitempty,uuid"`
	Tags       []string `json:"tags" validate:"omitempty,max=10"`
	DateFrom   string   `json:"date_from" validate:"omitempty,datetime"`
	DateTo     string   `json:"date_to" validate:"omitempty,datetime"`
	AuthorID   string   `json:"author_id" validate:"omitempty,uuid"`
	Status     string   `json:"status" validate:"omitempty,oneof=draft published archived"`
}

// BulkActionRequest represents a bulk action request
type BulkActionRequest struct {
	IDs    []string `json:"ids" validate:"required,min=1,max=100"`
	Action string   `json:"action" validate:"required,oneof=delete publish archive restore"`
}

// FileUploadRequest represents a file upload request
type FileUploadRequest struct {
	File         interface{} `json:"file" validate:"required"`
	Type         string      `json:"type" validate:"required,oneof=image document video"`
	MaxSize      int64       `json:"max_size" validate:"omitempty,min=1,max=10485760"` // 10MB
	AllowedMimes []string    `json:"allowed_mimes" validate:"omitempty,max=20"`
}

// NotificationRequest represents a notification request
type NotificationRequest struct {
	Title    string                 `json:"title" validate:"required,min=1,max=255"`
	Message  string                 `json:"message" validate:"required,min=1,max=1000"`
	Type     string                 `json:"type" validate:"required,oneof=info success warning error"`
	UserIDs  []string               `json:"user_ids" validate:"required,min=1"`
	Data     map[string]interface{} `json:"data" validate:"omitempty,json"`
	Priority int                    `json:"priority" validate:"omitempty,min=1,max=5"`
}

// APIKeyRequest represents an API key request
type APIKeyRequest struct {
	Name        string     `json:"name" validate:"required,min=2,max=50"`
	Description string     `json:"description" validate:"omitempty,max=255"`
	Permissions []string   `json:"permissions" validate:"omitempty,max=20"`
	ExpiresAt   *time.Time `json:"expires_at" validate:"omitempty,datetime"`
}

// WebhookRequest represents a webhook request
type WebhookRequest struct {
	URL     string            `json:"url" validate:"required,url"`
	Events  []string          `json:"events" validate:"required,min=1,max=10"`
	Secret  string            `json:"secret" validate:"required,min=16,max=255"`
	Active  bool              `json:"active" validate:"omitempty"`
	Headers map[string]string `json:"headers" validate:"omitempty,json"`
}
