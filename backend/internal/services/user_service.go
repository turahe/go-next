package services

import (
	"time"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"
)

type UserService interface {
	GetUserByID(id string) (*models.User, error)
	UpdateUserProfile(user *models.User, username, email, phone string, emailVerified, phoneVerified *time.Time) error
}

type userService struct{}

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

var UserSvc UserService = &userService{}
