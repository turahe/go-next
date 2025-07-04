package services

import (
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"
)

type UserRoleService interface {
	AssignRoleToUser(user *models.User, role *models.Role) error
	RemoveRoleFromUser(user *models.User, role *models.Role) error
	ListUserRoles(user *models.User) ([]models.Role, error)
}

type userRoleService struct{}

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
