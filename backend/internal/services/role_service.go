package services

import (
	"context"
	"go-next/internal/models"
	"go-next/pkg/database"
	"go-next/pkg/redis"
)

type RoleService interface {
	GetAllRoles() ([]models.Role, error)
	GetRoleByID(id string) (*models.Role, error)
	CreateRole(role *models.Role) error
	UpdateRole(role *models.Role) error
	DeleteRole(id string) error
	GetAllRolesWithContext(ctx context.Context) ([]models.Role, error)
}

type roleService struct {
	redisService *redis.RedisService
}

func NewRoleService(redisService *redis.RedisService) RoleService {
	return &roleService{
		redisService: redisService,
	}
}

func (s *roleService) GetAllRoles() ([]models.Role, error) {
	var roles []models.Role
	err := database.DB.Find(&roles).Error
	return roles, err
}

func (s *roleService) GetRoleByID(id string) (*models.Role, error) {
	var role models.Role
	err := database.DB.First(&role, id).Error
	return &role, err
}

func (s *roleService) CreateRole(role *models.Role) error {
	return database.DB.Create(role).Error
}

func (s *roleService) UpdateRole(role *models.Role) error {
	return database.DB.Save(role).Error
}

func (s *roleService) DeleteRole(id string) error {
	return database.DB.Delete(&models.Role{}, id).Error
}

func (s *roleService) GetAllRolesWithContext(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	err := database.DB.Find(&roles).Error
	return roles, err
}

var RoleSvc RoleService = &roleService{}
