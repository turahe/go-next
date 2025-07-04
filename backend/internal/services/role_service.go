package services

import (
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"
)

type RoleService interface {
	GetAllRoles() ([]models.Role, error)
	GetRoleByID(id string) (*models.Role, error)
	CreateRole(role *models.Role) error
	UpdateRole(role *models.Role) error
	DeleteRole(id string) error
}

type roleService struct{}

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

var RoleSvc RoleService = &roleService{}
