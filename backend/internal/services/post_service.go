package services

import (
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"

	"gorm.io/gorm"
)

type PostService interface {
	GetAllPosts() ([]models.Post, error)
	GetPostByID(id string) (*models.Post, error)
	CreatePost(post *models.Post) error
	UpdatePost(post *models.Post) error
	DeletePost(id string) error
}

type postService struct{}

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
			post.Contents[i].ModelId = post.ID
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

var PostSvc PostService = &postService{}
