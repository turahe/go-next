package services

import (
	"context"
	"go-next/internal/models"
	"go-next/pkg/database"
	"go-next/pkg/redis"

	"gorm.io/gorm"
)

type PostService interface {
	GetAllPosts() ([]models.Post, error)
	GetPostByID(id string) (*models.Post, error)
	CreatePost(post *models.Post) error
	UpdatePost(post *models.Post) error
	DeletePost(id string) error
	GetPublishedPosts(ctx context.Context) ([]*models.Post, error)
	GetPostCount(ctx context.Context) (int64, error)
	GetPublishedPostCount(ctx context.Context) (int64, error)

	// Content management methods
	AddContentToPost(postID string, contentType, content string, sortOrder int) (*models.Content, error)
	UpdatePostContent(postID, contentID string, contentType, content string, sortOrder int) error
	RemoveContentFromPost(postID, contentID string) error
	GetPostContents(postID string) ([]models.Content, error)
	GetPostContentsByType(postID, contentType string) ([]models.Content, error)
	ReorderPostContents(postID string, contentOrder []string) error
}

type postService struct {
	redisService *redis.RedisService
}

func NewPostService(redisService *redis.RedisService) PostService {
	return &postService{
		redisService: redisService,
	}
}

func (s *postService) GetAllPosts() ([]models.Post, error) {
	var posts []models.Post
	err := database.DB.Preload("User").Preload("Category").Find(&posts).Error
	return posts, err
}

func (s *postService) GetPostByID(id string) (*models.Post, error) {
	var post models.Post
	err := database.DB.Preload("User").Preload("Category").Preload("Contents", "sort_order ASC").First(&post, id).Error
	return &post, err
}

func (s *postService) CreatePost(post *models.Post) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(post).Error; err != nil {
			return err
		}
		for i := range post.Contents {
			post.Contents[i].ModelID = post.ID
			post.Contents[i].ModelType = "post"
			if err := tx.Create(&post.Contents[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *postService) UpdatePost(post *models.Post) error {
	return database.DB.Save(post).Error
}

func (s *postService) DeletePost(id string) error {
	return database.DB.Delete(&models.Post{}, id).Error
}

func (s *postService) GetPublishedPosts(ctx context.Context) ([]*models.Post, error) {
	var posts []*models.Post
	err := database.DB.Where("status = ?", "published").Find(&posts).Error
	return posts, err
}

func (s *postService) GetPostCount(ctx context.Context) (int64, error) {
	var count int64
	err := database.DB.Model(&models.Post{}).Count(&count).Error
	return count, err
}

func (s *postService) GetPublishedPostCount(ctx context.Context) (int64, error) {
	var count int64
	err := database.DB.Model(&models.Post{}).Where("status = ?", "published").Count(&count).Error
	return count, err
}

// AddContentToPost adds a new content block to a post
func (s *postService) AddContentToPost(postID string, contentType, content string, sortOrder int) (*models.Content, error) {
	var post models.Post
	if err := database.DB.First(&post, postID).Error; err != nil {
		return nil, err
	}

	newContent := &models.Content{
		ModelID:    post.ID,
		ModelType:  "post",
		Type:       contentType,
		ContentRaw: content,
		SortOrder:  sortOrder,
	}

	if err := database.DB.Create(newContent).Error; err != nil {
		return nil, err
	}

	return newContent, nil
}

// UpdatePostContent updates a content block in a post
func (s *postService) UpdatePostContent(postID, contentID string, contentType, content string, sortOrder int) error {
	var contentModel models.Content
	if err := database.DB.Where("id = ? AND model_id = ? AND model_type = ?", contentID, postID, "post").First(&contentModel).Error; err != nil {
		return err
	}

	contentModel.Type = contentType
	contentModel.ContentRaw = content
	contentModel.SortOrder = sortOrder

	return database.DB.Save(&contentModel).Error
}

// RemoveContentFromPost removes a content block from a post
func (s *postService) RemoveContentFromPost(postID, contentID string) error {
	return database.DB.Where("id = ? AND model_id = ? AND model_type = ?", contentID, postID, "post").Delete(&models.Content{}).Error
}

// GetPostContents gets all content blocks for a post
func (s *postService) GetPostContents(postID string) ([]models.Content, error) {
	var contents []models.Content
	err := database.DB.Where("model_id = ? AND model_type = ?", postID, "post").Order("sort_order ASC").Find(&contents).Error
	return contents, err
}

// GetPostContentsByType gets content blocks of a specific type for a post
func (s *postService) GetPostContentsByType(postID, contentType string) ([]models.Content, error) {
	var contents []models.Content
	err := database.DB.Where("model_id = ? AND model_type = ? AND type = ?", postID, "post", contentType).Order("sort_order ASC").Find(&contents).Error
	return contents, err
}

// ReorderPostContents reorders content blocks for a post
func (s *postService) ReorderPostContents(postID string, contentOrder []string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		for i, contentID := range contentOrder {
			if err := tx.Model(&models.Content{}).Where("id = ? AND model_id = ? AND model_type = ?", contentID, postID, "post").Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

var PostSvc PostService = &postService{}
