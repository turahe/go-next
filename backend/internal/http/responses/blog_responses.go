package responses

import (
	"time"

	"go-next/internal/models"

	"github.com/google/uuid"
)

// BlogPostResponse represents a detailed blog post for public viewing
type BlogPostResponse struct {
	ID          uuid.UUID         `json:"id"`
	Title       string            `json:"title"`
	Slug        string            `json:"slug"`
	Content     string            `json:"content"`
	Excerpt     string            `json:"excerpt,omitempty"`
	Status      string            `json:"status"`
	Public      bool              `json:"public"`
	ViewCount   int64             `json:"view_count"`
	CategoryID  *uuid.UUID        `json:"category_id,omitempty"`
	Category    *CategoryResponse `json:"category,omitempty"`
	UserID      *uuid.UUID        `json:"user_id,omitempty"`
	Comments    []CommentResponse `json:"comments,omitempty"`
	Media       []MediaResponse   `json:"media,omitempty"`
	PublishedAt *time.Time        `json:"published_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// BlogPostSimpleResponse represents a simplified blog post for lists
type BlogPostSimpleResponse struct {
	ID          uuid.UUID         `json:"id"`
	Title       string            `json:"title"`
	Slug        string            `json:"slug"`
	Excerpt     string            `json:"excerpt,omitempty"`
	ViewCount   int64             `json:"view_count"`
	Category    *CategoryResponse `json:"category,omitempty"`
	PublishedAt *time.Time        `json:"published_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// CategoryResponse represents a category for blog responses
type CategoryResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// UserResponse represents a user for blog responses
type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

// CommentResponse represents a comment for blog responses
type CommentResponse struct {
	ID        uuid.UUID     `json:"id"`
	Content   string        `json:"content"`
	User      *UserResponse `json:"user,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

// MediaResponse represents a media item for blog responses
type MediaResponse struct {
	ID       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	URL      string    `json:"url"`
	Type     string    `json:"type"`
}

// BlogStatsResponse represents blog statistics
type BlogStatsResponse struct {
	TotalPosts      int64 `json:"total_posts"`
	PublishedPosts  int64 `json:"published_posts"`
	TotalViews      int64 `json:"total_views"`
	TotalComments   int64 `json:"total_comments"`
	TotalCategories int64 `json:"total_categories"`
	TotalTags       int64 `json:"total_tags"`
}

// CategoryStatsResponse represents category statistics
type CategoryStatsResponse struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	CategorySlug string    `json:"category_slug"`
	PostCount    int64     `json:"post_count"`
	ViewCount    int64     `json:"view_count"`
}

// MonthlyArchiveResponse represents monthly archive data
type MonthlyArchiveResponse struct {
	Year  int   `json:"year"`
	Month int   `json:"month"`
	Count int64 `json:"count"`
}

type TagResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ToBlogPostResponse converts a Post model to BlogPostResponse
func ToBlogPostResponse(post *models.Post) *BlogPostResponse {
	if post == nil {
		return nil
	}

	response := &BlogPostResponse{
		ID:        post.ID,
		Title:     post.Title,
		Slug:      post.Slug,
		Content:   post.Description, // Post uses Description instead of Content
		Excerpt:   post.Excerpt,
		Status:    string(post.Status), // Convert PostStatus to string
		Public:    post.Public,
		ViewCount: int64(post.ViewCount), // Convert int to int64
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}

	if post.PublishedAt != nil {
		response.PublishedAt = post.PublishedAt
	}

	if post.CategoryID != nil {
		response.CategoryID = post.CategoryID
		if post.Category != nil {
			response.Category = &CategoryResponse{
				ID:   post.Category.ID,
				Name: post.Category.Name,
				Slug: post.Category.Slug,
			}
		}
	}

	// Post doesn't have UserID field, it's inherited from BaseModelWithUser
	// We can get it from the CreatedBy field if needed
	if post.CreatedBy != nil {
		response.UserID = post.CreatedBy
	}

	// Convert comments if available
	if len(post.Comments) > 0 {
		response.Comments = make([]CommentResponse, len(post.Comments))
		for i, comment := range post.Comments {
			response.Comments[i] = CommentResponse{
				ID:        comment.ID,
				Content:   comment.Description, // Comment uses Description instead of Content
				CreatedAt: comment.CreatedAt,
			}
			if comment.User != nil {
				response.Comments[i].User = &UserResponse{
					ID:       comment.User.ID,
					Username: comment.User.Username,
					Email:    comment.User.Email,
				}
			}
		}
	}

	// Convert media if available
	if len(post.Media) > 0 {
		response.Media = make([]MediaResponse, len(post.Media))
		for i, media := range post.Media {
			response.Media[i] = MediaResponse{
				ID:       media.ID,
				Filename: media.FileName,
				URL:      media.URL,
				Type:     media.MimeType, // Media uses MimeType instead of FileType
			}
		}
	}

	return response
}

// ToBlogPostSimpleResponses converts a slice of Post models to BlogPostSimpleResponse
func ToBlogPostSimpleResponses(posts []models.Post) []*BlogPostSimpleResponse {
	responses := make([]*BlogPostSimpleResponse, len(posts))
	for i, post := range posts {
		responses[i] = ToBlogPostSimpleResponse(&post)
	}
	return responses
}

// ToBlogPostSimpleResponse converts a Post model to BlogPostSimpleResponse
func ToBlogPostSimpleResponse(post *models.Post) *BlogPostSimpleResponse {
	if post == nil {
		return nil
	}

	response := &BlogPostSimpleResponse{
		ID:        post.ID,
		Title:     post.Title,
		Slug:      post.Slug,
		Excerpt:   post.Excerpt,
		ViewCount: int64(post.ViewCount), // Convert int to int64
		CreatedAt: post.CreatedAt,
	}

	if post.PublishedAt != nil {
		response.PublishedAt = post.PublishedAt
	}

	if post.Category != nil {
		response.Category = &CategoryResponse{
			ID:   post.Category.ID,
			Name: post.Category.Name,
			Slug: post.Category.Slug,
		}
	}

	return response
}
