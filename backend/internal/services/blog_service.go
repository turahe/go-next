package services

import (
	"errors"
	"time"

	"go-next/internal/models"
	"go-next/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BlogService interface {
	// Public blog endpoints
	GetPublicPosts(page, perPage int, search, categorySlug string) ([]models.Post, int64, error)
	GetPublicPost(slug string) (*models.Post, error)
	GetFeaturedPosts(limit int) ([]models.Post, error)
	GetRelatedPosts(postID uuid.UUID, limit int) ([]models.Post, error)
	GetPopularPosts(limit int, days int) ([]models.Post, error)
	GetPostsByCategory(categorySlug string, page, perPage int) ([]models.Post, int64, error)
	GetPostsByTag(tagSlug string, page, perPage int) ([]models.Post, int64, error)
	SearchPosts(query string, page, perPage int) ([]models.Post, int64, error)

	// Blog statistics
	GetBlogStats() (*BlogStats, error)
	GetCategoryStats() ([]CategoryStats, error)
	GetMonthlyArchive() ([]MonthlyArchive, error)

	// Post management
	CreatePost(post *models.Post) error
	UpdatePost(post *models.Post) error
	DeletePost(id string) error
	PublishPost(id string) error
	UnpublishPost(id string) error
	ArchivePost(id string) error
	IncrementViewCount(postID uuid.UUID) error

	// Category management
	GetPublicCategories() ([]models.Category, error)
	GetCategoryBySlug(slug string) (*models.Category, error)

	// Tag management
	GetPublicTags() ([]models.Tag, error)
	GetTagBySlug(slug string) (*models.Tag, error)
}

type blogService struct {
	db *gorm.DB
}

type BlogStats struct {
	TotalPosts      int64 `json:"total_posts"`
	PublishedPosts  int64 `json:"published_posts"`
	TotalViews      int64 `json:"total_views"`
	TotalComments   int64 `json:"total_comments"`
	TotalCategories int64 `json:"total_categories"`
	TotalTags       int64 `json:"total_tags"`
}

type CategoryStats struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	CategorySlug string    `json:"category_slug"`
	PostCount    int64     `json:"post_count"`
	ViewCount    int64     `json:"view_count"`
}

type MonthlyArchive struct {
	Year  int   `json:"year"`
	Month int   `json:"month"`
	Count int64 `json:"count"`
}

func NewBlogService() BlogService {
	return &blogService{db: database.DB}
}

// GetPublicPosts retrieves published posts for public viewing
func (s *blogService) GetPublicPosts(page, perPage int, search, categorySlug string) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	query := s.db.Model(&models.Post{}).
		Where("status = ? AND public = ?", "published", true).
		Where("published_at IS NOT NULL").
		Preload("Category").
		Preload("User").
		Preload("Media")

	// Apply search filter
	if search != "" {
		query = query.Where("title ILIKE ? OR content ILIKE ? OR excerpt ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Apply category filter
	if categorySlug != "" {
		query = query.Joins("JOIN categories ON posts.category_id = categories.id").
			Where("categories.slug = ?", categorySlug)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated posts
	offset := (page - 1) * perPage
	if err := query.Offset(offset).Limit(perPage).
		Order("published_at DESC").
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetPublicPost retrieves a single published post by slug
func (s *blogService) GetPublicPost(slug string) (*models.Post, error) {
	var post models.Post

	err := s.db.Where("slug = ? AND status = ? AND public = ? AND published_at IS NOT NULL",
		slug, "published", true).
		Preload("Category").
		Preload("User").
		Preload("Media").
		Preload("Comments", "status = ?", "approved").
		Preload("Comments.User").
		First(&post).Error

	if err != nil {
		return nil, err
	}

	return &post, nil
}

// GetFeaturedPosts retrieves featured posts
func (s *blogService) GetFeaturedPosts(limit int) ([]models.Post, error) {
	var posts []models.Post

	err := s.db.Where("status = ? AND public = ? AND published_at IS NOT NULL",
		"published", true).
		Preload("Category").
		Preload("User").
		Order("view_count DESC, published_at DESC").
		Limit(limit).
		Find(&posts).Error

	return posts, err
}

// GetRelatedPosts retrieves related posts based on category
func (s *blogService) GetRelatedPosts(postID uuid.UUID, limit int) ([]models.Post, error) {
	var post models.Post
	if err := s.db.Select("category_id").First(&post, postID).Error; err != nil {
		return nil, err
	}

	var posts []models.Post
	err := s.db.Where("id != ? AND category_id = ? AND status = ? AND public = ? AND published_at IS NOT NULL",
		postID, post.CategoryID, "published", true).
		Preload("Category").
		Preload("User").
		Order("published_at DESC").
		Limit(limit).
		Find(&posts).Error

	return posts, err
}

// GetPopularPosts retrieves popular posts based on view count
func (s *blogService) GetPopularPosts(limit int, days int) ([]models.Post, error) {
	var posts []models.Post

	query := s.db.Where("status = ? AND public = ? AND published_at IS NOT NULL",
		"published", true)

	if days > 0 {
		cutoffDate := time.Now().AddDate(0, 0, -days)
		query = query.Where("published_at >= ?", cutoffDate)
	}

	err := query.Preload("Category").
		Preload("User").
		Order("view_count DESC, published_at DESC").
		Limit(limit).
		Find(&posts).Error

	return posts, err
}

// GetPostsByCategory retrieves posts by category slug
func (s *blogService) GetPostsByCategory(categorySlug string, page, perPage int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	query := s.db.Model(&models.Post{}).
		Joins("JOIN categories ON posts.category_id = categories.id").
		Where("categories.slug = ? AND posts.status = ? AND posts.public = ? AND posts.published_at IS NOT NULL",
			categorySlug, "published", true).
		Preload("Category").
		Preload("User")

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated posts
	offset := (page - 1) * perPage
	if err := query.Offset(offset).Limit(perPage).
		Order("posts.published_at DESC").
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetPostsByTag retrieves posts by tag slug
func (s *blogService) GetPostsByTag(tagSlug string, page, perPage int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	query := s.db.Model(&models.Post{}).
		Joins("JOIN post_tags ON posts.id = post_tags.post_id").
		Joins("JOIN tags ON post_tags.tag_id = tags.id").
		Where("tags.slug = ? AND posts.status = ? AND posts.public = ? AND posts.published_at IS NOT NULL",
			tagSlug, "published", true).
		Preload("Category").
		Preload("User")

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated posts
	offset := (page - 1) * perPage
	if err := query.Offset(offset).Limit(perPage).
		Order("posts.published_at DESC").
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// SearchPosts performs full-text search on posts
func (s *blogService) SearchPosts(query string, page, perPage int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	searchQuery := s.db.Model(&models.Post{}).
		Where("status = ? AND public = ? AND published_at IS NOT NULL", "published", true).
		Where("title ILIKE ? OR content ILIKE ? OR excerpt ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%").
		Preload("Category").
		Preload("User")

	// Get total count
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated posts
	offset := (page - 1) * perPage
	if err := searchQuery.Offset(offset).Limit(perPage).
		Order("published_at DESC").
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetBlogStats retrieves blog statistics
func (s *blogService) GetBlogStats() (*BlogStats, error) {
	var stats BlogStats

	// Count total posts
	if err := s.db.Model(&models.Post{}).Count(&stats.TotalPosts).Error; err != nil {
		return nil, err
	}

	// Count published posts
	if err := s.db.Model(&models.Post{}).
		Where("status = ? AND public = ?", "published", true).
		Count(&stats.PublishedPosts).Error; err != nil {
		return nil, err
	}

	// Sum total views
	if err := s.db.Model(&models.Post{}).
		Select("COALESCE(SUM(view_count), 0)").
		Scan(&stats.TotalViews).Error; err != nil {
		return nil, err
	}

	// Count total comments
	if err := s.db.Model(&models.Comment{}).
		Where("status = ?", "approved").
		Count(&stats.TotalComments).Error; err != nil {
		return nil, err
	}

	// Count categories
	if err := s.db.Model(&models.Category{}).
		Where("is_active = ?", true).
		Count(&stats.TotalCategories).Error; err != nil {
		return nil, err
	}

	// Count tags
	if err := s.db.Model(&models.Tag{}).
		Where("is_active = ?", true).
		Count(&stats.TotalTags).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetCategoryStats retrieves statistics for each category
func (s *blogService) GetCategoryStats() ([]CategoryStats, error) {
	var stats []CategoryStats

	err := s.db.Model(&models.Category{}).
		Select(`
			categories.id as category_id,
			categories.name as category_name,
			categories.slug as category_slug,
			COUNT(posts.id) as post_count,
			COALESCE(SUM(posts.view_count), 0) as view_count
		`).
		Joins("LEFT JOIN posts ON categories.id = posts.category_id AND posts.status = ? AND posts.public = ?", "published", true).
		Where("categories.is_active = ?", true).
		Group("categories.id, categories.name, categories.slug").
		Order("post_count DESC").
		Find(&stats).Error

	return stats, err
}

// GetMonthlyArchive retrieves monthly post counts
func (s *blogService) GetMonthlyArchive() ([]MonthlyArchive, error) {
	var archives []MonthlyArchive

	err := s.db.Model(&models.Post{}).
		Select(`
			EXTRACT(YEAR FROM published_at) as year,
			EXTRACT(MONTH FROM published_at) as month,
			COUNT(*) as count
		`).
		Where("status = ? AND public = ? AND published_at IS NOT NULL", "published", true).
		Group("year, month").
		Order("year DESC, month DESC").
		Find(&archives).Error

	return archives, err
}

// CreatePost creates a new post
func (s *blogService) CreatePost(post *models.Post) error {
	return s.db.Create(post).Error
}

// UpdatePost updates an existing post
func (s *blogService) UpdatePost(post *models.Post) error {
	return s.db.Save(post).Error
}

// DeletePost deletes a post
func (s *blogService) DeletePost(id string) error {
	postID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid post ID")
	}
	return s.db.Delete(&models.Post{}, postID).Error
}

// PublishPost publishes a post
func (s *blogService) PublishPost(id string) error {
	postID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid post ID")
	}

	var post models.Post
	if err := s.db.First(&post, postID).Error; err != nil {
		return err
	}

	post.Publish()
	return s.db.Save(&post).Error
}

// UnpublishPost unpublishes a post
func (s *blogService) UnpublishPost(id string) error {
	postID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid post ID")
	}

	var post models.Post
	if err := s.db.First(&post, postID).Error; err != nil {
		return err
	}

	post.Unpublish()
	return s.db.Save(&post).Error
}

// ArchivePost archives a post
func (s *blogService) ArchivePost(id string) error {
	postID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid post ID")
	}

	var post models.Post
	if err := s.db.First(&post, postID).Error; err != nil {
		return err
	}

	post.Archive()
	return s.db.Save(&post).Error
}

// IncrementViewCount increments the view count for a post
func (s *blogService) IncrementViewCount(postID uuid.UUID) error {
	return s.db.Model(&models.Post{}).
		Where("id = ?", postID).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).
		Error
}

// GetPublicCategories retrieves active categories
func (s *blogService) GetPublicCategories() ([]models.Category, error) {
	var categories []models.Category

	err := s.db.Where("is_active = ?", true).
		Preload("Children", "is_active = ?", true).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error

	return categories, err
}

// GetCategoryBySlug retrieves a category by slug
func (s *blogService) GetCategoryBySlug(slug string) (*models.Category, error) {
	var category models.Category

	err := s.db.Where("slug = ? AND is_active = ?", slug, true).
		Preload("Children", "is_active = ?", true).
		First(&category).Error

	if err != nil {
		return nil, err
	}

	return &category, nil
}

// GetPublicTags retrieves active tags
func (s *blogService) GetPublicTags() ([]models.Tag, error) {
	var tags []models.Tag

	err := s.db.Where("is_active = ?", true).
		Order("name ASC").
		Find(&tags).Error

	return tags, err
}

// GetTagBySlug retrieves a tag by slug
func (s *blogService) GetTagBySlug(slug string) (*models.Tag, error) {
	var tag models.Tag

	err := s.db.Where("slug = ? AND is_active = ?", slug, true).
		First(&tag).Error

	if err != nil {
		return nil, err
	}

	return &tag, nil
}
