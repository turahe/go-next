// Package services provides business logic layer for the blog application.
// This package contains all service interfaces and implementations that handle
// the core business logic, data processing, and external service interactions.
package services

import (
	"errors"
	"time"

	"go-next/internal/models"
	"go-next/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BlogService defines the interface for all blog-related business operations.
// This interface provides methods for public blog access, post management,
// statistics generation, and content organization.
type BlogService interface {
	// Public blog endpoints - These methods handle public-facing blog functionality

	// GetPublicPosts retrieves a paginated list of published posts for public viewing.
	// Supports filtering by search query and category slug.
	// Returns posts, total count, and any error encountered.
	GetPublicPosts(page, perPage int, search, categorySlug string) ([]models.Post, int64, error)

	// GetPublicPost retrieves a single published post by its slug for public viewing.
	// Returns the post with all related data (category, user, media) or an error if not found.
	GetPublicPost(slug string) (*models.Post, error)

	// GetFeaturedPosts retrieves a limited number of featured posts for display.
	// Featured posts are typically displayed on the homepage or special sections.
	GetFeaturedPosts(limit int) ([]models.Post, error)

	// GetRelatedPosts finds posts related to a given post based on category and tags.
	// Used for suggesting similar content to readers.
	GetRelatedPosts(postID uuid.UUID, limit int) ([]models.Post, error)

	// GetPopularPosts retrieves the most viewed posts within a specified time period.
	// Useful for displaying trending content or popular articles.
	GetPopularPosts(limit int, days int) ([]models.Post, error)

	// GetPostsByCategory retrieves all posts belonging to a specific category.
	// Supports pagination for large category collections.
	GetPostsByCategory(categorySlug string, page, perPage int) ([]models.Post, int64, error)

	// GetPostsByTag retrieves all posts tagged with a specific tag.
	// Supports pagination for large tag collections.
	GetPostsByTag(tagSlug string, page, perPage int) ([]models.Post, int64, error)

	// SearchPosts performs full-text search across post titles, content, and excerpts.
	// Returns paginated results with total count.
	SearchPosts(query string, page, perPage int) ([]models.Post, int64, error)

	// Blog statistics - Methods for generating analytics and insights

	// GetBlogStats returns comprehensive statistics about the blog including
	// total posts, published posts, views, comments, categories, and tags.
	GetBlogStats() (*BlogStats, error)

	// GetCategoryStats returns statistics for each category including
	// post count and view count for analytics purposes.
	GetCategoryStats() ([]CategoryStats, error)

	// GetMonthlyArchive returns post counts grouped by year and month
	// for creating archive navigation and analytics.
	GetMonthlyArchive() ([]MonthlyArchive, error)

	// Post management - Administrative methods for post lifecycle management

	// CreatePost creates a new post in the database.
	// The post is initially created as a draft and must be published separately.
	CreatePost(post *models.Post) error

	// UpdatePost updates an existing post with new data.
	// Only the post owner or admin can update posts.
	UpdatePost(post *models.Post) error

	// DeletePost permanently removes a post from the database.
	// This action cannot be undone and should be used with caution.
	DeletePost(id string) error

	// PublishPost changes a post's status to published and sets the published_at timestamp.
	// Published posts become visible to the public.
	PublishPost(id string) error

	// UnpublishPost changes a post's status back to draft.
	// Unpublished posts are not visible to the public.
	UnpublishPost(id string) error

	// ArchivePost moves a post to archived status.
	// Archived posts are typically hidden from public view but preserved.
	ArchivePost(id string) error

	// IncrementViewCount increases the view count for a specific post.
	// Called automatically when a post is viewed.
	IncrementViewCount(postID uuid.UUID) error

	// Category management - Methods for organizing content by categories

	// GetPublicCategories retrieves all active categories for public display.
	// Used for category navigation and filtering.
	GetPublicCategories() ([]models.Category, error)

	// GetCategoryBySlug retrieves a specific category by its slug.
	// Returns the category with its metadata or an error if not found.
	GetCategoryBySlug(slug string) (*models.Category, error)

	// Tag management - Methods for organizing content by tags

	// GetPublicTags retrieves all active tags for public display.
	// Used for tag clouds and tag-based navigation.
	GetPublicTags() ([]models.Tag, error)

	// GetTagBySlug retrieves a specific tag by its slug.
	// Returns the tag with its metadata or an error if not found.
	GetTagBySlug(slug string) (*models.Tag, error)
}

// blogService implements the BlogService interface.
// This struct holds the database connection and provides the actual implementation
// of all blog-related business logic.
type blogService struct {
	db *gorm.DB // Database connection for all data operations
}

// BlogStats represents comprehensive statistics about the blog.
// This struct is used for analytics, dashboard displays, and API responses.
type BlogStats struct {
	TotalPosts      int64 `json:"total_posts"`      // Total number of posts (all statuses)
	PublishedPosts  int64 `json:"published_posts"`  // Number of published posts
	TotalViews      int64 `json:"total_views"`      // Total view count across all posts
	TotalComments   int64 `json:"total_comments"`   // Total comment count across all posts
	TotalCategories int64 `json:"total_categories"` // Number of active categories
	TotalTags       int64 `json:"total_tags"`       // Number of active tags
}

// CategoryStats represents statistics for a specific category.
// Used for category analytics and performance tracking.
type CategoryStats struct {
	CategoryID   uuid.UUID `json:"category_id"`   // Unique identifier for the category
	CategoryName string    `json:"category_name"` // Human-readable category name
	CategorySlug string    `json:"category_slug"` // URL-friendly category identifier
	PostCount    int64     `json:"post_count"`    // Number of posts in this category
	ViewCount    int64     `json:"view_count"`    // Total views for posts in this category
}

// MonthlyArchive represents post count for a specific year and month.
// Used for creating archive navigation and historical analytics.
type MonthlyArchive struct {
	Year  int   `json:"year"`  // Year (e.g., 2024)
	Month int   `json:"month"` // Month (1-12)
	Count int64 `json:"count"` // Number of posts published in this month
}

// NewBlogService creates and returns a new instance of BlogService.
// This factory function initializes the service with the global database connection.
func NewBlogService() BlogService {
	return &blogService{db: database.DB}
}

// GetPublicPosts retrieves published posts for public viewing with pagination and filtering.
//
// Parameters:
//   - page: Current page number (1-based)
//   - perPage: Number of posts per page
//   - search: Optional search query to filter posts by title, content, or excerpt
//   - categorySlug: Optional category slug to filter posts by category
//
// Returns:
//   - []models.Post: List of posts with related data (category, user, media)
//   - int64: Total count of matching posts for pagination
//   - error: Any error encountered during the operation
//
// Example:
//
//	posts, total, err := blogService.GetPublicPosts(1, 10, "golang", "programming")
//	if err != nil {
//	    // Handle error
//	}
func (s *blogService) GetPublicPosts(page, perPage int, search, categorySlug string) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	// Build the base query for published posts only
	query := s.db.Model(&models.Post{}).
		Where("status = ? AND public = ?", "published", true).
		Where("published_at IS NOT NULL").
		Preload("Category"). // Include category data
		Preload("User").     // Include author data
		Preload("Media")     // Include media attachments

	// Apply search filter if provided
	if search != "" {
		query = query.Where("title ILIKE ? OR content ILIKE ? OR excerpt ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Apply category filter if provided
	if categorySlug != "" {
		query = query.Joins("JOIN categories ON posts.category_id = categories.id").
			Where("categories.slug = ?", categorySlug)
	}

	// Count total matching records for pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	offset := (page - 1) * perPage
	if err := query.Order("published_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetPublicPost retrieves a single published post by its slug for public viewing.
//
// Parameters:
//   - slug: URL-friendly identifier for the post
//
// Returns:
//   - *models.Post: The post with all related data or nil if not found
//   - error: Any error encountered during the operation
//
// Example:
//
//	post, err := blogService.GetPublicPost("my-awesome-post")
//	if err != nil {
//	    // Handle error (post not found or database error)
//	}
func (s *blogService) GetPublicPost(slug string) (*models.Post, error) {
	var post models.Post

	// Query for published post with all related data
	err := s.db.Where("slug = ? AND status = ? AND public = ? AND published_at IS NOT NULL",
		slug, "published", true).
		Preload("Category").
		Preload("User").
		Preload("Media").
		Preload("Comments").
		Preload("Comments.User").
		First(&post).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}

	return &post, nil
}

// GetFeaturedPosts retrieves a limited number of featured posts for display.
// Featured posts are typically displayed on the homepage or special sections.
//
// Parameters:
//   - limit: Maximum number of featured posts to return
//
// Returns:
//   - []models.Post: List of featured posts with related data
//   - error: Any error encountered during the operation
//
// Example:
//
//	featuredPosts, err := blogService.GetFeaturedPosts(5)
//	if err != nil {
//	    // Handle error
//	}
func (s *blogService) GetFeaturedPosts(limit int) ([]models.Post, error) {
	var posts []models.Post

	// Query for featured, published posts ordered by featured date
	err := s.db.Where("featured = ? AND status = ? AND public = ? AND published_at IS NOT NULL",
		true, "published", true).
		Preload("Category").
		Preload("User").
		Preload("Media").
		Order("featured_at DESC").
		Limit(limit).
		Find(&posts).Error

	return posts, err
}

// GetRelatedPosts finds posts related to a given post based on category and tags.
// Used for suggesting similar content to readers.
//
// Parameters:
//   - postID: UUID of the reference post
//   - limit: Maximum number of related posts to return
//
// Returns:
//   - []models.Post: List of related posts with basic data
//   - error: Any error encountered during the operation
//
// Example:
//
//	relatedPosts, err := blogService.GetRelatedPosts(postID, 3)
//	if err != nil {
//	    // Handle error
//	}
func (s *blogService) GetRelatedPosts(postID uuid.UUID, limit int) ([]models.Post, error) {
	// First, get the reference post to find its category and tags
	var post models.Post
	if err := s.db.Preload("Category").Preload("Tags").First(&post, postID).Error; err != nil {
		return nil, err
	}

	var posts []models.Post

	// Query for related posts based on category and tags
	query := s.db.Where("id != ? AND status = ? AND public = ? AND published_at IS NOT NULL",
		postID, "published", true).
		Preload("Category").
		Preload("User").
		Preload("Media")

	// If post has a category, prioritize posts in the same category
	if post.CategoryID != nil {
		query = query.Where("category_id = ?", post.CategoryID)
	}

	// Order by relevance (same category first, then by publish date)
	err := query.Order("published_at DESC").
		Limit(limit).
		Find(&posts).Error

	return posts, err
}

// GetPopularPosts retrieves the most viewed posts within a specified time period.
// Useful for displaying trending content or popular articles.
//
// Parameters:
//   - limit: Maximum number of popular posts to return
//   - days: Number of days to look back for popularity calculation
//
// Returns:
//   - []models.Post: List of popular posts with related data
//   - error: Any error encountered during the operation
//
// Example:
//
//	popularPosts, err := blogService.GetPopularPosts(10, 30) // Last 30 days
//	if err != nil {
//	    // Handle error
//	}
func (s *blogService) GetPopularPosts(limit int, days int) ([]models.Post, error) {
	var posts []models.Post

	// Calculate the date threshold for popularity calculation
	threshold := time.Now().AddDate(0, 0, -days)

	// Query for popular posts within the specified time period
	err := s.db.Where("status = ? AND public = ? AND published_at IS NOT NULL AND published_at >= ?",
		"published", true, threshold).
		Preload("Category").
		Preload("User").
		Preload("Media").
		Order("view_count DESC").
		Limit(limit).
		Find(&posts).Error

	return posts, err
}

// GetPostsByCategory retrieves all posts belonging to a specific category.
// Supports pagination for large category collections.
//
// Parameters:
//   - categorySlug: URL-friendly identifier for the category
//   - page: Current page number (1-based)
//   - perPage: Number of posts per page
//
// Returns:
//   - []models.Post: List of posts in the category with related data
//   - int64: Total count of posts in the category for pagination
//   - error: Any error encountered during the operation
//
// Example:
//
//	posts, total, err := blogService.GetPostsByCategory("programming", 1, 10)
//	if err != nil {
//	    // Handle error
//	}
func (s *blogService) GetPostsByCategory(categorySlug string, page, perPage int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	// Build query with category join
	query := s.db.Model(&models.Post{}).
		Joins("JOIN categories ON posts.category_id = categories.id").
		Where("categories.slug = ? AND posts.status = ? AND posts.public = ? AND posts.published_at IS NOT NULL",
			categorySlug, "published", true).
		Preload("Category").
		Preload("User").
		Preload("Media")

	// Count total posts in category
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	offset := (page - 1) * perPage
	if err := query.Order("posts.published_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetPostsByTag retrieves all posts tagged with a specific tag.
// Supports pagination for large tag collections.
//
// Parameters:
//   - tagSlug: URL-friendly identifier for the tag
//   - page: Current page number (1-based)
//   - perPage: Number of posts per page
//
// Returns:
//   - []models.Post: List of posts with the tag and related data
//   - int64: Total count of posts with the tag for pagination
//   - error: Any error encountered during the operation
//
// Example:
//
//	posts, total, err := blogService.GetPostsByTag("golang", 1, 10)
//	if err != nil {
//	    // Handle error
//	}
func (s *blogService) GetPostsByTag(tagSlug string, page, perPage int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	// Build query with tag join through post_tags table
	query := s.db.Model(&models.Post{}).
		Joins("JOIN post_tags ON posts.id = post_tags.post_id").
		Joins("JOIN tags ON post_tags.tag_id = tags.id").
		Where("tags.slug = ? AND posts.status = ? AND posts.public = ? AND posts.published_at IS NOT NULL",
			tagSlug, "published", true).
		Preload("Category").
		Preload("User").
		Preload("Media")

	// Count total posts with this tag
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	offset := (page - 1) * perPage
	if err := query.Order("posts.published_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// SearchPosts performs full-text search across post titles, content, and excerpts.
// Returns paginated results with total count.
//
// Parameters:
//   - query: Search term to look for in post content
//   - page: Current page number (1-based)
//   - perPage: Number of posts per page
//
// Returns:
//   - []models.Post: List of matching posts with related data
//   - int64: Total count of matching posts for pagination
//   - error: Any error encountered during the operation
//
// Example:
//
//	posts, total, err := blogService.SearchPosts("docker kubernetes", 1, 10)
//	if err != nil {
//	    // Handle error
//	}
func (s *blogService) SearchPosts(query string, page, perPage int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	// Build search query across multiple fields
	searchQuery := s.db.Model(&models.Post{}).
		Where("status = ? AND public = ? AND published_at IS NOT NULL", "published", true).
		Where("title ILIKE ? OR content ILIKE ? OR excerpt ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%").
		Preload("Category").
		Preload("User").
		Preload("Media")

	// Count total matching posts
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	offset := (page - 1) * perPage
	if err := searchQuery.Order("published_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetBlogStats returns comprehensive statistics about the blog including
// total posts, published posts, views, comments, categories, and tags.
//
// Returns:
//   - *BlogStats: Comprehensive blog statistics
//   - error: Any error encountered during the operation
//
// Example:
//
//	stats, err := blogService.GetBlogStats()
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Printf("Total posts: %d, Published: %d, Views: %d\n",
//	    stats.TotalPosts, stats.PublishedPosts, stats.TotalViews)
func (s *blogService) GetBlogStats() (*BlogStats, error) {
	var stats BlogStats

	// Count total posts
	if err := s.db.Model(&models.Post{}).Count(&stats.TotalPosts).Error; err != nil {
		return nil, err
	}

	// Count published posts
	if err := s.db.Model(&models.Post{}).
		Where("status = ? AND public = ? AND published_at IS NOT NULL", "published", true).
		Count(&stats.PublishedPosts).Error; err != nil {
		return nil, err
	}

	// Sum total views
	if err := s.db.Model(&models.Post{}).
		Select("COALESCE(SUM(view_count), 0)").
		Where("status = ? AND public = ? AND published_at IS NOT NULL", "published", true).
		Scan(&stats.TotalViews).Error; err != nil {
		return nil, err
	}

	// Count total comments
	if err := s.db.Model(&models.Comment{}).
		Where("status = ?", "approved").
		Count(&stats.TotalComments).Error; err != nil {
		return nil, err
	}

	// Count active categories
	if err := s.db.Model(&models.Category{}).
		Where("active = ?", true).
		Count(&stats.TotalCategories).Error; err != nil {
		return nil, err
	}

	// Count active tags
	if err := s.db.Model(&models.Tag{}).
		Where("active = ?", true).
		Count(&stats.TotalTags).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetCategoryStats returns statistics for each category including
// post count and view count for analytics purposes.
//
// Returns:
//   - []CategoryStats: List of category statistics
//   - error: Any error encountered during the operation
//
// Example:
//
//	categoryStats, err := blogService.GetCategoryStats()
//	if err != nil {
//	    // Handle error
//	}
//	for _, stat := range categoryStats {
//	    fmt.Printf("Category: %s, Posts: %d, Views: %d\n",
//	        stat.CategoryName, stat.PostCount, stat.ViewCount)
//	}
func (s *blogService) GetCategoryStats() ([]CategoryStats, error) {
	var stats []CategoryStats

	// Query for category statistics with post counts and view counts
	err := s.db.Table("categories").
		Select("categories.id as category_id, categories.name as category_name, "+
			"categories.slug as category_slug, "+
			"COUNT(posts.id) as post_count, "+
			"COALESCE(SUM(posts.view_count), 0) as view_count").
		Joins("LEFT JOIN posts ON categories.id = posts.category_id "+
			"AND posts.status = ? AND posts.public = ? AND posts.published_at IS NOT NULL",
			"published", true).
		Where("categories.active = ?", true).
		Group("categories.id, categories.name, categories.slug").
		Order("post_count DESC").
		Scan(&stats).Error

	return stats, err
}

// GetMonthlyArchive returns post counts grouped by year and month
// for creating archive navigation and analytics.
//
// Returns:
//   - []MonthlyArchive: List of monthly post counts
//   - error: Any error encountered during the operation
//
// Example:
//
//	archives, err := blogService.GetMonthlyArchive()
//	if err != nil {
//	    // Handle error
//	}
//	for _, archive := range archives {
//	    fmt.Printf("%d-%02d: %d posts\n", archive.Year, archive.Month, archive.Count)
//	}
func (s *blogService) GetMonthlyArchive() ([]MonthlyArchive, error) {
	var archives []MonthlyArchive

	// Query for monthly post counts
	err := s.db.Table("posts").
		Select("EXTRACT(YEAR FROM published_at) as year, "+
			"EXTRACT(MONTH FROM published_at) as month, "+
			"COUNT(*) as count").
		Where("status = ? AND public = ? AND published_at IS NOT NULL", "published", true).
		Group("year, month").
		Order("year DESC, month DESC").
		Scan(&archives).Error

	return archives, err
}

// CreatePost creates a new post in the database.
// The post is initially created as a draft and must be published separately.
//
// Parameters:
//   - post: Post model with all required fields populated
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	post := &models.Post{
//	    Title: "My New Post",
//	    Content: "Post content...",
//	    UserID: userID,
//	    CategoryID: &categoryID,
//	}
//	err := blogService.CreatePost(post)
//	if err != nil {
//	    // Handle error
//	}
func (s *blogService) CreatePost(post *models.Post) error {
	return s.db.Create(post).Error
}

// UpdatePost updates an existing post with new data.
// Only the post owner or admin can update posts.
//
// Parameters:
//   - post: Post model with updated fields
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	post.Title = "Updated Title"
//	post.Content = "Updated content..."
//	err := blogService.UpdatePost(post)
//	if err != nil {
//	    // Handle error
//	}
func (s *blogService) UpdatePost(post *models.Post) error {
	return s.db.Save(post).Error
}

// DeletePost permanently removes a post from the database.
// This action cannot be undone and should be used with caution.
//
// Parameters:
//   - id: String representation of the post UUID
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := blogService.DeletePost("550e8400-e29b-41d4-a716-446655440000")
//	if err != nil {
//	    // Handle error
//	}
func (s *blogService) DeletePost(id string) error {
	postID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid post ID format")
	}

	// Delete the post and all related data
	return s.db.Delete(&models.Post{}, postID).Error
}

// PublishPost changes a post's status to published and sets the published_at timestamp.
// Published posts become visible to the public.
//
// Parameters:
//   - id: String representation of the post UUID
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := blogService.PublishPost("550e8400-e29b-41d4-a716-446655440000")
//	if err != nil {
//	    // Handle error
//	}
func (s *blogService) PublishPost(id string) error {
	postID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid post ID format")
	}

	// Update post status to published and set published timestamp
	return s.db.Model(&models.Post{}).
		Where("id = ?", postID).
		Updates(map[string]interface{}{
			"status":       "published",
			"published_at": time.Now(),
		}).Error
}

// UnpublishPost changes a post's status back to draft.
// Unpublished posts are not visible to the public.
//
// Parameters:
//   - id: String representation of the post UUID
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := blogService.UnpublishPost("550e8400-e29b-41d4-a716-446655440000")
//	if err != nil {
//	    // Handle error
//	}
func (s *blogService) UnpublishPost(id string) error {
	postID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid post ID format")
	}

	// Update post status to draft and clear published timestamp
	return s.db.Model(&models.Post{}).
		Where("id = ?", postID).
		Updates(map[string]interface{}{
			"status":       "draft",
			"published_at": nil,
		}).Error
}

// ArchivePost moves a post to archived status.
// Archived posts are typically hidden from public view but preserved.
//
// Parameters:
//   - id: String representation of the post UUID
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := blogService.ArchivePost("550e8400-e29b-41d4-a716-446655440000")
//	if err != nil {
//	    // Handle error
//	}
func (s *blogService) ArchivePost(id string) error {
	postID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid post ID format")
	}

	// Update post status to archived
	return s.db.Model(&models.Post{}).
		Where("id = ?", postID).
		Update("status", "archived").Error
}

// IncrementViewCount increases the view count for a specific post.
// Called automatically when a post is viewed.
//
// Parameters:
//   - postID: UUID of the post to increment view count for
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := blogService.IncrementViewCount(postUUID)
//	if err != nil {
//	    // Handle error
//	}
func (s *blogService) IncrementViewCount(postID uuid.UUID) error {
	return s.db.Model(&models.Post{}).
		Where("id = ?", postID).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// GetPublicCategories retrieves all active categories for public display.
// Used for category navigation and filtering.
//
// Returns:
//   - []models.Category: List of active categories
//   - error: Any error encountered during the operation
//
// Example:
//
//	categories, err := blogService.GetPublicCategories()
//	if err != nil {
//	    // Handle error
//	}
//	for _, category := range categories {
//	    fmt.Printf("Category: %s (%s)\n", category.Name, category.Slug)
//	}
func (s *blogService) GetPublicCategories() ([]models.Category, error) {
	var categories []models.Category

	err := s.db.Where("active = ?", true).
		Order("name ASC").
		Find(&categories).Error

	return categories, err
}

// GetCategoryBySlug retrieves a specific category by its slug.
// Returns the category with its metadata or an error if not found.
//
// Parameters:
//   - slug: URL-friendly identifier for the category
//
// Returns:
//   - *models.Category: The category or nil if not found
//   - error: Any error encountered during the operation
//
// Example:
//
//	category, err := blogService.GetCategoryBySlug("programming")
//	if err != nil {
//	    // Handle error (category not found or database error)
//	}
func (s *blogService) GetCategoryBySlug(slug string) (*models.Category, error) {
	var category models.Category

	err := s.db.Where("slug = ? AND active = ?", slug, true).
		First(&category).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	return &category, nil
}

// GetPublicTags retrieves all active tags for public display.
// Used for tag clouds and tag-based navigation.
//
// Returns:
//   - []models.Tag: List of active tags
//   - error: Any error encountered during the operation
//
// Example:
//
//	tags, err := blogService.GetPublicTags()
//	if err != nil {
//	    // Handle error
//	}
//	for _, tag := range tags {
//	    fmt.Printf("Tag: %s (%s)\n", tag.Name, tag.Slug)
//	}
func (s *blogService) GetPublicTags() ([]models.Tag, error) {
	var tags []models.Tag

	err := s.db.Where("active = ?", true).
		Order("name ASC").
		Find(&tags).Error

	return tags, err
}

// GetTagBySlug retrieves a specific tag by its slug.
// Returns the tag with its metadata or an error if not found.
//
// Parameters:
//   - slug: URL-friendly identifier for the tag
//
// Returns:
//   - *models.Tag: The tag or nil if not found
//   - error: Any error encountered during the operation
//
// Example:
//
//	tag, err := blogService.GetTagBySlug("golang")
//	if err != nil {
//	    // Handle error (tag not found or database error)
//	}
func (s *blogService) GetTagBySlug(slug string) (*models.Tag, error) {
	var tag models.Tag

	err := s.db.Where("slug = ? AND active = ?", slug, true).
		First(&tag).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tag not found")
		}
		return nil, err
	}

	return &tag, nil
}
