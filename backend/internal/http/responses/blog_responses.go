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

// Conversion functions

// ToBlogPostResponse converts a models.Post to BlogPostResponse
func ToBlogPostResponse(p *models.Post) *BlogPostResponse {
	var category *CategoryResponse
	if p.Category != nil && p.Category.ID != uuid.Nil {
		category = &CategoryResponse{
			ID:   p.Category.ID,
			Name: p.Category.Name,
			Slug: p.Category.Slug,
		}
	}

	comments := make([]CommentResponse, len(p.Comments))
	for i, comment := range p.Comments {
		var commentUser *UserResponse
		if comment.User != nil {
			commentUser = &UserResponse{
				ID:       comment.UserID,
				Username: comment.User.Username,
				Email:    comment.User.Email,
			}
		}
		comments[i] = CommentResponse{
			ID:        comment.ID,
			Content:   comment.Content,
			User:      commentUser,
			CreatedAt: comment.CreatedAt,
		}
	}

	media := make([]MediaResponse, len(p.Media))
	for i, m := range p.Media {
		media[i] = MediaResponse{
			ID:       m.ID,
			Filename: m.FileName,
			URL:      m.URL,
			Type:     m.MimeType,
		}
	}

	return &BlogPostResponse{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Content:     p.Content,
		Excerpt:     p.Excerpt,
		Status:      p.Status,
		Public:      p.Public,
		ViewCount:   p.ViewCount,
		CategoryID:  p.CategoryID,
		Category:    category,
		UserID:      p.CreatedBy,
		Comments:    comments,
		Media:       media,
		PublishedAt: p.PublishedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// ToBlogPostSimpleResponse converts a models.Post to BlogPostSimpleResponse
func ToBlogPostSimpleResponse(p *models.Post) *BlogPostSimpleResponse {
	var category *CategoryResponse
	if p.Category != nil && p.Category.ID != uuid.Nil {
		category = &CategoryResponse{
			ID:   p.Category.ID,
			Name: p.Category.Name,
			Slug: p.Category.Slug,
		}
	}

	return &BlogPostSimpleResponse{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Excerpt:     p.Excerpt,
		ViewCount:   p.ViewCount,
		Category:    category,
		PublishedAt: p.PublishedAt,
		CreatedAt:   p.CreatedAt,
	}
}

// ToBlogPostResponses converts a slice of models.Post to []*BlogPostResponse
func ToBlogPostResponses(posts []models.Post) []*BlogPostResponse {
	responses := make([]*BlogPostResponse, len(posts))
	for i, post := range posts {
		responses[i] = ToBlogPostResponse(&post)
	}
	return responses
}

// ToBlogPostSimpleResponses converts a slice of models.Post to []*BlogPostSimpleResponse
func ToBlogPostSimpleResponses(posts []models.Post) []*BlogPostSimpleResponse {
	responses := make([]*BlogPostSimpleResponse, len(posts))
	for i, post := range posts {
		responses[i] = ToBlogPostSimpleResponse(&post)
	}
	return responses
}

// ToBlogStatsResponse converts a BlogStats to BlogStatsResponse
func ToBlogStatsResponse(stats interface{}) *BlogStatsResponse {
	// Type assertion for stats
	if s, ok := stats.(map[string]interface{}); ok {
		return &BlogStatsResponse{
			TotalPosts:      int64(s["total_posts"].(float64)),
			PublishedPosts:  int64(s["published_posts"].(float64)),
			TotalViews:      int64(s["total_views"].(float64)),
			TotalComments:   int64(s["total_comments"].(float64)),
			TotalCategories: int64(s["total_categories"].(float64)),
			TotalTags:       int64(s["total_tags"].(float64)),
		}
	}
	return &BlogStatsResponse{}
}

// ToCategoryStatsResponse converts a CategoryStats to CategoryStatsResponse
func ToCategoryStatsResponse(stats interface{}) *CategoryStatsResponse {
	// Type assertion for stats
	if s, ok := stats.(map[string]interface{}); ok {
		return &CategoryStatsResponse{
			CategoryID:   uuid.MustParse(s["category_id"].(string)),
			CategoryName: s["category_name"].(string),
			CategorySlug: s["category_slug"].(string),
			PostCount:    int64(s["post_count"].(float64)),
			ViewCount:    int64(s["view_count"].(float64)),
		}
	}
	return &CategoryStatsResponse{}
}

// ToCategoryStatsResponses converts a slice of CategoryStats to []*CategoryStatsResponse
func ToCategoryStatsResponses(stats []interface{}) []*CategoryStatsResponse {
	responses := make([]*CategoryStatsResponse, len(stats))
	for i, stat := range stats {
		responses[i] = ToCategoryStatsResponse(stat)
	}
	return responses
}

// ToMonthlyArchiveResponse converts a MonthlyArchive to MonthlyArchiveResponse
func ToMonthlyArchiveResponse(archive interface{}) *MonthlyArchiveResponse {
	// Type assertion for archive
	if a, ok := archive.(map[string]interface{}); ok {
		return &MonthlyArchiveResponse{
			Year:  int(a["year"].(float64)),
			Month: int(a["month"].(float64)),
			Count: int64(a["count"].(float64)),
		}
	}
	return &MonthlyArchiveResponse{}
}

// ToMonthlyArchiveResponses converts a slice of MonthlyArchive to []*MonthlyArchiveResponse
func ToMonthlyArchiveResponses(archives []interface{}) []*MonthlyArchiveResponse {
	responses := make([]*MonthlyArchiveResponse, len(archives))
	for i, archive := range archives {
		responses[i] = ToMonthlyArchiveResponse(archive)
	}
	return responses
}

// ToCategoryResponse converts a models.Category to CategoryResponse
func ToCategoryResponse(category *models.Category) *CategoryResponse {
	return &CategoryResponse{
		ID:   category.ID,
		Name: category.Name,
		Slug: category.Slug,
	}
}

// ToCategoryResponses converts a slice of models.Category to []*CategoryResponse
func ToCategoryResponses(categories []models.Category) []*CategoryResponse {
	responses := make([]*CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = ToCategoryResponse(&category)
	}
	return responses
}

// ToTagResponse converts a models.Tag to TagResponse
func ToTagResponse(tag *models.Tag) *TagResponse {
	return &TagResponse{
		ID:   tag.ID,
		Name: tag.Name,
		Slug: tag.Slug,
	}
}

// ToTagResponses converts a slice of models.Tag to []*TagResponse
func ToTagResponses(tags []models.Tag) []*TagResponse {
	responses := make([]*TagResponse, len(tags))
	for i, tag := range tags {
		responses[i] = ToTagResponse(&tag)
	}
	return responses
}
