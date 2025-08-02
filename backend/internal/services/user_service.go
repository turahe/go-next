package services

import (
	"context"
	"go-next/internal/models"
	"go-next/pkg/database"
	"go-next/pkg/redis"
	"time"
)

type UserService interface {
	GetUserByID(id string) (*models.User, error)
	UpdateUserProfile(user *models.User, username, email, phone string, emailVerified, phoneVerified *time.Time) error
	GetActiveUsers(ctx context.Context) ([]*models.User, error)
	GetUserCount(ctx context.Context) (int64, error)
}

type userService struct {
	redisService *redis.RedisService
}

func NewUserService(redisService *redis.RedisService) UserService {
	return &userService{
		redisService: redisService,
	}
}

func (s *userService) GetUserByID(id string) (*models.User, error) {
	var user models.User
	err := database.DB.Preload("Roles").First(&user, id).Error
	return &user, err
}

func (s *userService) UpdateUserProfile(user *models.User, username, email, phone string, emailVerified, phoneVerified *time.Time) error {
	user.Username = username
	user.Email = email
	user.Phone = phone
	user.EmailVerified = emailVerified
	user.PhoneVerified = phoneVerified
	return database.DB.Save(user).Error
}

func (s *userService) GetActiveUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	err := database.DB.Where("is_active = ?", true).Find(&users).Error
	return users, err
}

func (s *userService) GetUserCount(ctx context.Context) (int64, error) {
	var count int64
	err := database.DB.Model(&models.User{}).Count(&count).Error
	return count, err
}

var UserSvc UserService = &userService{}
