package dto

import (
	"time"

	"go-next/internal/models"
	"go-next/internal/services"

	"github.com/google/uuid"
)

// BlogPostDTO represents a detailed blog post for public viewing
type BlogPostDTO struct {
	ID          uuid.UUID          `json:"id"`
	Title       string             `json:"title"`
	Slug        string             `json:"slug"`
	Content     string             `json:"content"`
	Excerpt     string             `json:"excerpt,omitempty"`
	Status      string             `json:"status"`
	Public      bool               `json:"public"`
	ViewCount   int64              `json:"view_count"`
	CategoryID  *uuid.UUID         `json:"category_id,omitempty"`
	Category    *CategorySimpleDTO `json:"category,omitempty"`
	UserID      *uuid.UUID         `json:"user_id,omitempty"`
	Comments    []CommentSimpleDTO `json:"comments,omitempty"`
	Media       []MediaSimpleDTO   `json:"media,omitempty"`
	PublishedAt *time.Time         `json:"published_at,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// BlogPostSimpleDTO represents a simplified blog post for lists
type BlogPostSimpleDTO struct {
	ID          uuid.UUID          `json:"id"`
	Title       string             `json:"title"`
	Slug        string             `json:"slug"`
	Excerpt     string             `json:"excerpt,omitempty"`
	ViewCount   int64              `json:"view_count"`
	Category    *CategorySimpleDTO `json:"category,omitempty"`
	PublishedAt *time.Time         `json:"published_at,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
}

// UserSimpleDTO represents a simplified user for blog posts
type UserSimpleDTO struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

// CommentSimpleDTO represents a simplified comment
type CommentSimpleDTO struct {
	ID        uuid.UUID      `json:"id"`
	Content   string         `json:"content"`
	User      *UserSimpleDTO `json:"user,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// TagSimpleDTO represents a simplified tag
type TagSimpleDTO struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// MediaSimpleDTO represents a simplified media item
type MediaSimpleDTO struct {
	ID       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	URL      string    `json:"url"`
	Type     string    `json:"type"`
}

// BlogStatsDTO represents blog statistics
type BlogStatsDTO struct {
	TotalPosts      int64 `json:"total_posts"`
	PublishedPosts  int64 `json:"published_posts"`
	TotalViews      int64 `json:"total_views"`
	TotalComments   int64 `json:"total_comments"`
	TotalCategories int64 `json:"total_categories"`
	TotalTags       int64 `json:"total_tags"`
}

// CategoryStatsDTO represents category statistics
type CategoryStatsDTO struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	CategorySlug string    `json:"category_slug"`
	PostCount    int64     `json:"post_count"`
	ViewCount    int64     `json:"view_count"`
}

// MonthlyArchiveDTO represents monthly archive data
type MonthlyArchiveDTO struct {
	Year  int   `json:"year"`
	Month int   `json:"month"`
	Count int64 `json:"count"`
}

// Conversion functions

// ToBlogPostDTO converts a models.Post to BlogPostDTO
func ToBlogPostDTO(p *models.Post) *BlogPostDTO {
	var category *CategorySimpleDTO
	if p.Category != nil && p.Category.ID != uuid.Nil {
		category = &CategorySimpleDTO{
			ID:   p.Category.ID,
			Name: p.Category.Name,
		}
	}

	comments := make([]CommentSimpleDTO, len(p.Comments))
	for i, comment := range p.Comments {
		var commentUser *UserSimpleDTO
		if comment.User != nil {
			commentUser = &UserSimpleDTO{
				ID:       comment.UserID,
				Username: comment.User.Username,
				Email:    comment.User.Email,
			}
		}
		comments[i] = CommentSimpleDTO{
			ID:        comment.ID,
			Content:   comment.Content,
			User:      commentUser,
			CreatedAt: comment.CreatedAt,
		}
	}

	media := make([]MediaSimpleDTO, len(p.Media))
	for i, m := range p.Media {
		media[i] = MediaSimpleDTO{
			ID:       m.ID,
			Filename: m.FileName,
			URL:      m.URL,
			Type:     m.MimeType,
		}
	}

	return &BlogPostDTO{
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

// ToBlogPostSimpleDTO converts a models.Post to BlogPostSimpleDTO
func ToBlogPostSimpleDTO(p *models.Post) *BlogPostSimpleDTO {
	var category *CategorySimpleDTO
	if p.Category != nil && p.Category.ID != uuid.Nil {
		category = &CategorySimpleDTO{
			ID:   p.Category.ID,
			Name: p.Category.Name,
		}
	}

	return &BlogPostSimpleDTO{
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

// ToBlogPostDTOs converts a slice of models.Post to []*BlogPostDTO
func ToBlogPostDTOs(posts []models.Post) []*BlogPostDTO {
	dtos := make([]*BlogPostDTO, len(posts))
	for i, post := range posts {
		dtos[i] = ToBlogPostDTO(&post)
	}
	return dtos
}

// ToBlogPostSimpleDTOs converts a slice of models.Post to []*BlogPostSimpleDTO
func ToBlogPostSimpleDTOs(posts []models.Post) []*BlogPostSimpleDTO {
	dtos := make([]*BlogPostSimpleDTO, len(posts))
	for i, post := range posts {
		dtos[i] = ToBlogPostSimpleDTO(&post)
	}
	return dtos
}

// ToBlogStatsDTO converts services.BlogStats to BlogStatsDTO
func ToBlogStatsDTO(stats *services.BlogStats) *BlogStatsDTO {
	return &BlogStatsDTO{
		TotalPosts:      stats.TotalPosts,
		PublishedPosts:  stats.PublishedPosts,
		TotalViews:      stats.TotalViews,
		TotalComments:   stats.TotalComments,
		TotalCategories: stats.TotalCategories,
		TotalTags:       stats.TotalTags,
	}
}

// ToCategoryStatsDTOs converts services.CategoryStats to []*CategoryStatsDTO
func ToCategoryStatsDTOs(stats []services.CategoryStats) []*CategoryStatsDTO {
	dtos := make([]*CategoryStatsDTO, len(stats))
	for i, stat := range stats {
		dtos[i] = &CategoryStatsDTO{
			CategoryID:   stat.CategoryID,
			CategoryName: stat.CategoryName,
			CategorySlug: stat.CategorySlug,
			PostCount:    stat.PostCount,
			ViewCount:    stat.ViewCount,
		}
	}
	return dtos
}

// ToMonthlyArchiveDTOs converts services.MonthlyArchive to []*MonthlyArchiveDTO
func ToMonthlyArchiveDTOs(archives []services.MonthlyArchive) []*MonthlyArchiveDTO {
	dtos := make([]*MonthlyArchiveDTO, len(archives))
	for i, archive := range archives {
		dtos[i] = &MonthlyArchiveDTO{
			Year:  archive.Year,
			Month: archive.Month,
			Count: archive.Count,
		}
	}
	return dtos
}
