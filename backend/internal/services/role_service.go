package services

import (
	"errors"
	"go-next/internal/models"
	"go-next/pkg/database"
	"go-next/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleService defines the interface for role-related business operations
type RoleService interface {
	// GetOrCreateDefaultRole gets the default "user" role or creates it if it doesn't exist
	GetOrCreateDefaultRole() (*models.Role, error)

	// GetRoleByName gets a role by name
	GetRoleByName(name string) (*models.Role, error)

	// GetRoleByID gets a role by ID uuid
	GetRoleByID(id uuid.UUID) (*models.Role, error)

	// GetAllRoles gets all roles
	GetAllRoles() (*[]models.Role, error)

	// CreateRole creates a new role
	CreateRole(name, description string) (*models.Role, error)

	// UpdateRole updates a role
	UpdateRole(role *models.Role) error

	// DeleteRole deletes a role by ID
	DeleteRole(id uuid.UUID) error

	// AssignRoleToUser assigns a role to a user
	AssignRoleToUser(userID, roleID uuid.UUID) error

	// AssignRoleToUserWithTx assigns a role to a user within a transaction
	AssignRoleToUserWithTx(tx *gorm.DB, userID, roleID uuid.UUID) error

	// GetUserRoles gets all roles for a user
	GetUserRoles(userID uuid.UUID) ([]models.Role, error)

	// AssignMenuToRole assigns a menu to a role
	AssignMenuToRole(roleID, menuID uuid.UUID) error

	// RemoveMenuFromRole removes a menu from a role
	RemoveMenuFromRole(roleID, menuID uuid.UUID) error

	// GetRoleMenus gets all menus for a role
	GetRoleMenus(roleID uuid.UUID) ([]models.Menu, error)

	// GetMenuRoles gets all roles for a menu
	GetMenuRoles(menuID uuid.UUID) ([]models.Role, error)
}

// roleService implements the RoleService interface
type roleService struct {
	db *gorm.DB
}

// NewRoleService creates and returns a new instance of RoleService
func NewRoleService() RoleService {
	return &roleService{
		db: database.DB,
	}
}

// GetOrCreateDefaultRole gets the default "user" role or creates it if it doesn't exist
func (s *roleService) GetOrCreateDefaultRole() (*models.Role, error) {
	var role models.Role

	// Try to find the default "user" role
	err := s.db.Where("name = ?", "user").First(&role).Error
	if err == nil {
		// Role found, return it
		return &role, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Some other error occurred
		logger.Errorf("Error finding default role: %v", err)
		return nil, err
	}

	// Role not found, create it
	role = models.Role{
		Name:        "user",
		Description: "Default user role with basic permissions",
	}

	if err := s.db.Create(&role).Error; err != nil {
		logger.Errorf("Error creating default role: %v", err)
		return nil, err
	}

	logger.Infof("Created default role: %s", role.Name)
	return &role, nil
}

// GetRoleByName gets a role by name
func (s *roleService) GetRoleByName(name string) (*models.Role, error) {
	var role models.Role

	err := s.db.Where("name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}

	return &role, nil
}

// CreateRole creates a new role
func (s *roleService) CreateRole(name, description string) (*models.Role, error) {
	role := models.Role{
		Name:        name,
		Description: description,
	}

	if err := s.db.Create(&role).Error; err != nil {
		logger.Errorf("Error creating role: %v", err)
		return nil, err
	}

	return &role, nil
}

// AssignRoleToUser assigns a role to a user
func (s *roleService) AssignRoleToUser(userID, roleID uuid.UUID) error {
	return s.AssignRoleToUserWithTx(s.db, userID, roleID)
}

// AssignRoleToUserWithTx assigns a role to a user within a transaction
func (s *roleService) AssignRoleToUserWithTx(tx *gorm.DB, userID, roleID uuid.UUID) error {
	// Check if the assignment already exists
	var count int64
	err := tx.Model(&models.UserRole{}).Where("user_id = ? AND role_id = ?", userID, roleID).Count(&count).Error
	if err != nil {
		logger.Errorf("Error checking existing role assignment: %v", err)
		return err
	}

	if count > 0 {
		// Role already assigned
		return nil
	}

	// Create the user role assignment
	userRole := models.UserRole{
		UserID: userID,
		RoleID: roleID,
	}

	err = tx.Create(&userRole).Error
	if err != nil {
		logger.Errorf("Error assigning role to user: %v", err)
		return err
	}

	logger.Infof("Assigned role %s to user %s", roleID, userID)
	return nil
}

// GetUserRoles gets all roles for a user
func (s *roleService) GetUserRoles(userID uuid.UUID) ([]models.Role, error) {
	var roles []models.Role

	// Use a more robust query that works whether deleted_at column exists or not
	err := s.db.Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error

	if err != nil {
		logger.Errorf("Error getting user roles: %v", err)
		return nil, err
	}

	return roles, nil
}

// GetRoleByID gets a role by ID
func (s *roleService) GetRoleByID(id uuid.UUID) (*models.Role, error) {
	var role models.Role

	err := s.db.Where("id = ?", id).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}

	return &role, nil
}

// GetAllRoles gets all roles
func (s *roleService) GetAllRoles() (*[]models.Role, error) {
	var roles []models.Role

	err := s.db.Find(&roles).Error
	if err != nil {
		logger.Errorf("Error getting all roles: %v", err)
		return nil, err
	}

	return &roles, nil
}

// UpdateRole updates a role
func (s *roleService) UpdateRole(role *models.Role) error {
	if err := s.db.Save(role).Error; err != nil {
		logger.Errorf("Error updating role: %v", err)
		return err
	}

	return nil
}

// DeleteRole deletes a role by ID
func (s *roleService) DeleteRole(id uuid.UUID) error {
	var role models.Role

	err := s.db.Where("id = ?", id).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return err
	}

	if err := s.db.Delete(&role).Error; err != nil {
		logger.Errorf("Error deleting role: %v", err)
		return err
	}

	return nil
}

// AssignMenuToRole assigns a menu to a role
func (s *roleService) AssignMenuToRole(roleID, menuID uuid.UUID) error {
	// Check if the role exists
	var role models.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return err
	}

	// Check if the menu exists
	var menu models.Menu
	if err := s.db.First(&menu, menuID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("menu not found")
		}
		return err
	}

	// Check if the assignment already exists
	var count int64
	err := s.db.Model(&models.Role{}).
		Joins("JOIN role_menus ON roles.id = role_menus.role_id").
		Where("roles.id = ? AND role_menus.menu_id = ?", roleID, menuID).
		Count(&count).Error

	if err != nil {
		logger.Errorf("Error checking existing menu assignment: %v", err)
		return err
	}

	if count > 0 {
		// Menu already assigned to role
		return nil
	}

	// Assign menu to role using GORM's Association
	err = s.db.Model(&role).Association("Menus").Append(&menu)
	if err != nil {
		logger.Errorf("Error assigning menu to role: %v", err)
		return err
	}

	logger.Infof("Assigned menu %s to role %s", menuID, roleID)
	return nil
}

// RemoveMenuFromRole removes a menu from a role
func (s *roleService) RemoveMenuFromRole(roleID, menuID uuid.UUID) error {
	// Check if the role exists
	var role models.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return err
	}

	// Check if the menu exists
	var menu models.Menu
	if err := s.db.First(&menu, menuID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("menu not found")
		}
		return err
	}

	// Remove menu from role using GORM's Association
	err := s.db.Model(&role).Association("Menus").Delete(&menu)
	if err != nil {
		logger.Errorf("Error removing menu from role: %v", err)
		return err
	}

	logger.Infof("Removed menu %s from role %s", menuID, roleID)
	return nil
}

// GetRoleMenus gets all menus for a role
func (s *roleService) GetRoleMenus(roleID uuid.UUID) ([]models.Menu, error) {
	var role models.Role
	var menus []models.Menu

	// Check if the role exists
	if err := s.db.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}

	// Get menus for the role
	err := s.db.Model(&role).Association("Menus").Find(&menus)
	if err != nil {
		logger.Errorf("Error getting role menus: %v", err)
		return nil, err
	}

	return menus, nil
}

// GetMenuRoles gets all roles for a menu
func (s *roleService) GetMenuRoles(menuID uuid.UUID) ([]models.Role, error) {
	var menu models.Menu
	var roles []models.Role

	// Check if the menu exists
	if err := s.db.First(&menu, menuID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("menu not found")
		}
		return nil, err
	}

	// Get roles for the menu
	err := s.db.Model(&menu).Association("Roles").Find(&roles)
	if err != nil {
		logger.Errorf("Error getting menu roles: %v", err)
		return nil, err
	}

	return roles, nil
}
