package services

import (
	"go-next/internal/models"
	"go-next/pkg/database"
	"go-next/pkg/redis"
)

type UserRoleService interface {
	AssignRoleToUser(user *models.User, role *models.Role) error
	RemoveRoleFromUser(user *models.User, role *models.Role) error
	ListUserRoles(user *models.User) ([]models.Role, error)
}

type userRoleService struct {
	redisService *redis.RedisService
}

func NewUserRoleService(redisService *redis.RedisService) UserRoleService {
	return &userRoleService{
		redisService: redisService,
	}
}

func (s *userRoleService) AssignRoleToUser(user *models.User, role *models.Role) error {
	return database.DB.Model(user).Association("Roles").Append(role)
}
func (s *userRoleService) RemoveRoleFromUser(user *models.User, role *models.Role) error {
	return database.DB.Model(user).Association("Roles").Delete(role)
}
func (s *userRoleService) ListUserRoles(user *models.User) ([]models.Role, error) {
	var u models.User
	err := database.DB.Preload("Roles").First(&u, user.ID).Error
	return u.Roles, err
}

var UserRoleSvc UserRoleService = &userRoleService{}
