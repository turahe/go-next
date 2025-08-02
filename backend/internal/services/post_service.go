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
	err := database.DB.Preload("User").Preload("Category").First(&post, id).Error
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

var PostSvc PostService = &postService{}
