package services

import (
	"context"
	"go-next/internal/models"
	"go-next/pkg/database"
	"go-next/pkg/redis"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MenuService interface {
	GetAllMenus() ([]models.Menu, error)
	GetMenuByID(id uuid.UUID) (*models.Menu, error)
	CreateMenu(menu *models.Menu) error
	UpdateMenu(menu *models.Menu) error
	DeleteMenu(id uuid.UUID) error
	GetMenuTree() ([]models.Menu, error)
	GetMenuByParentID(parentID uuid.UUID) ([]models.Menu, error)
	GetAllMenusWithContext(ctx context.Context) ([]models.Menu, error)
	GetMenuCount(ctx context.Context) (int64, error)
	GetMenuDescendants(id uuid.UUID) ([]models.Menu, error)
	GetMenuAncestors(id uuid.UUID) ([]models.Menu, error)
	GetMenuSiblings(id uuid.UUID) ([]models.Menu, error)
	MoveMenu(id uuid.UUID, newParentID uuid.UUID) error

	// Role-related methods
	AssignRoleToMenu(menuID, roleID uuid.UUID) error
	RemoveRoleFromMenu(menuID, roleID uuid.UUID) error
	GetMenuRoles(menuID uuid.UUID) ([]models.Role, error)
	GetRoleMenus(roleID uuid.UUID) ([]models.Menu, error)
}

type menuService struct {
	redisService *redis.RedisService
}

func NewMenuService(redisService *redis.RedisService) MenuService {
	return &menuService{
		redisService: redisService,
	}
}

func (s *menuService) GetAllMenus() ([]models.Menu, error) {
	var menus []models.Menu
	err := database.DB.Preload("Children").Where("parent_id IS NULL").Order("record_ordering ASC").Find(&menus).Error
	return menus, err
}

func (s *menuService) GetMenuByID(id uuid.UUID) (*models.Menu, error) {
	var menu models.Menu
	err := database.DB.Preload("Children").First(&menu, id).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (s *menuService) CreateMenu(menu *models.Menu) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// If parent_id is provided, implement nested set model
		if menu.ParentID != uuid.Nil {
			var parent models.Menu
			if err := tx.First(&parent, menu.ParentID).Error; err != nil {
				return err
			}
			if parent.RecordRight == 0 {
				return gorm.ErrRecordNotFound
			}

			// Update parent's right value and all nodes to the right
			tx.Model(&models.Menu{}).
				Where("record_right >= ?", parent.RecordRight).
				Update("record_right", gorm.Expr("record_right + 2"))
			tx.Model(&models.Menu{}).
				Where("record_left > ?", parent.RecordRight).
				Update("record_left", gorm.Expr("record_left + 2"))

			// Set nested set values for the new menu
			menu.RecordLeft = parent.RecordRight
			menu.RecordRight = parent.RecordRight + 1
			menu.RecordDept = parent.RecordDept + 1
		} else {
			// Create as root - find the maximum right value
			var maxRight int
			tx.Model(&models.Menu{}).Select("COALESCE(MAX(record_right), 0)").Scan(&maxRight)

			menu.RecordLeft = maxRight + 1
			menu.RecordRight = maxRight + 2
			menu.RecordDept = 0
		}

		return tx.Create(menu).Error
	})
}

func (s *menuService) UpdateMenu(menu *models.Menu) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Get the existing menu to check if parent is changing
		var existingMenu models.Menu
		if err := tx.First(&existingMenu, menu.ID).Error; err != nil {
			return err
		}

		// If parent is changing, we need to move the subtree
		if existingMenu.ParentID != menu.ParentID {
			if err := s.moveMenuSubtree(tx, &existingMenu, menu.ParentID); err != nil {
				return err
			}
		}

		// Update the menu fields
		return tx.Save(menu).Error
	})
}

func (s *menuService) DeleteMenu(id uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Get the menu to be deleted
		var menu models.Menu
		if err := tx.First(&menu, id).Error; err != nil {
			return err
		}

		// Calculate the width of the subtree
		width := menu.RecordRight - menu.RecordLeft + 1

		// Delete the menu and all its descendants
		if err := tx.Where("record_left >= ? AND record_right <= ?", menu.RecordLeft, menu.RecordRight).Delete(&models.Menu{}).Error; err != nil {
			return err
		}

		// Update the left and right values of remaining nodes
		tx.Model(&models.Menu{}).
			Where("record_left > ?", menu.RecordRight).
			Update("record_left", gorm.Expr("record_left - ?", width))

		tx.Model(&models.Menu{}).
			Where("record_right > ?", menu.RecordRight).
			Update("record_right", gorm.Expr("record_right - ?", width))

		return nil
	})
}

func (s *menuService) GetMenuTree() ([]models.Menu, error) {
	var menus []models.Menu
	err := database.DB.Preload("Children").Where("parent_id IS NULL").Order("record_ordering ASC").Find(&menus).Error
	return menus, err
}

func (s *menuService) GetMenuByParentID(parentID uuid.UUID) ([]models.Menu, error) {
	var menus []models.Menu
	err := database.DB.Where("parent_id = ?", parentID).Order("record_ordering ASC").Find(&menus).Error
	return menus, err
}

func (s *menuService) GetAllMenusWithContext(ctx context.Context) ([]models.Menu, error) {
	var menus []models.Menu
	err := database.DB.WithContext(ctx).Preload("Children").Where("parent_id IS NULL").Order("record_ordering ASC").Find(&menus).Error
	return menus, err
}

func (s *menuService) GetMenuCount(ctx context.Context) (int64, error) {
	var count int64
	err := database.DB.WithContext(ctx).Model(&models.Menu{}).Count(&count).Error
	return count, err
}

// moveMenuSubtree moves a menu and its entire subtree to a new parent
func (s *menuService) moveMenuSubtree(tx *gorm.DB, menu *models.Menu, newParentID uuid.UUID) error {
	// Calculate the width of the subtree
	width := menu.RecordRight - menu.RecordLeft + 1

	// Get the new parent
	var newParent models.Menu
	if newParentID != uuid.Nil {
		if err := tx.First(&newParent, newParentID).Error; err != nil {
			return err
		}
	}

	// Calculate the new position
	var newLeft int
	if newParentID != uuid.Nil {
		newLeft = newParent.RecordRight
	} else {
		// Moving to root level
		var maxRight int
		tx.Model(&models.Menu{}).Select("COALESCE(MAX(record_right), 0)").Scan(&maxRight)
		newLeft = maxRight + 1
	}

	// Calculate the offset
	offset := newLeft - menu.RecordLeft

	// Update all nodes in the subtree
	tx.Model(&models.Menu{}).
		Where("record_left >= ? AND record_right <= ?", menu.RecordLeft, menu.RecordRight).
		Updates(map[string]interface{}{
			"record_left":  gorm.Expr("record_left + ?", offset),
			"record_right": gorm.Expr("record_right + ?", offset),
			"record_dept":  gorm.Expr("record_dept + ?", newParent.RecordDept-menu.RecordDept+1),
		})

	// Update nodes to the right of the old position
	tx.Model(&models.Menu{}).
		Where("record_left > ?", menu.RecordRight).
		Update("record_left", gorm.Expr("record_left - ?", width))

	tx.Model(&models.Menu{}).
		Where("record_right > ?", menu.RecordRight).
		Update("record_right", gorm.Expr("record_right - ?", width))

	// Update nodes to the right of the new position
	tx.Model(&models.Menu{}).
		Where("record_left >= ?", newLeft).
		Update("record_left", gorm.Expr("record_left + ?", width))

	tx.Model(&models.Menu{}).
		Where("record_right >= ?", newLeft).
		Update("record_right", gorm.Expr("record_right + ?", width))

	return nil
}

func (s *menuService) GetMenuDescendants(id uuid.UUID) ([]models.Menu, error) {
	var menu models.Menu
	if err := database.DB.First(&menu, id).Error; err != nil {
		return nil, err
	}

	var descendants []models.Menu
	err := database.DB.Where("record_left > ? AND record_right < ?", menu.RecordLeft, menu.RecordRight).
		Order("record_left ASC").Find(&descendants).Error
	return descendants, err
}

func (s *menuService) GetMenuAncestors(id uuid.UUID) ([]models.Menu, error) {
	var menu models.Menu
	if err := database.DB.First(&menu, id).Error; err != nil {
		return nil, err
	}

	var ancestors []models.Menu
	err := database.DB.Where("record_left < ? AND record_right > ?", menu.RecordLeft, menu.RecordRight).
		Order("record_left ASC").Find(&ancestors).Error
	return ancestors, err
}

func (s *menuService) GetMenuSiblings(id uuid.UUID) ([]models.Menu, error) {
	var menu models.Menu
	if err := database.DB.First(&menu, id).Error; err != nil {
		return nil, err
	}

	var siblings []models.Menu
	err := database.DB.Where("record_left > ? AND record_right < ? AND record_dept = ?",
		menu.RecordLeft, menu.RecordRight, menu.RecordDept).
		Order("record_left ASC").Find(&siblings).Error
	return siblings, err
}

func (s *menuService) MoveMenu(id uuid.UUID, newParentID uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var menu models.Menu
		if err := tx.First(&menu, id).Error; err != nil {
			return err
		}

		return s.moveMenuSubtree(tx, &menu, newParentID)
	})
}

// AssignRoleToMenu assigns a role to a menu
func (s *menuService) AssignRoleToMenu(menuID, roleID uuid.UUID) error {
	// Check if the menu exists
	var menu models.Menu
	if err := database.DB.First(&menu, menuID).Error; err != nil {
		return err
	}

	// Check if the role exists
	var role models.Role
	if err := database.DB.First(&role, roleID).Error; err != nil {
		return err
	}

	// Check if the assignment already exists
	var count int64
	err := database.DB.Model(&models.Menu{}).
		Joins("JOIN role_menus ON menus.id = role_menus.menu_id").
		Where("menus.id = ? AND role_menus.role_id = ?", menuID, roleID).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		// Role already assigned to menu
		return nil
	}

	// Assign role to menu using GORM's Association
	err = database.DB.Model(&menu).Association("Roles").Append(&role)
	if err != nil {
		return err
	}

	return nil
}

// RemoveRoleFromMenu removes a role from a menu
func (s *menuService) RemoveRoleFromMenu(menuID, roleID uuid.UUID) error {
	// Check if the menu exists
	var menu models.Menu
	if err := database.DB.First(&menu, menuID).Error; err != nil {
		return err
	}

	// Check if the role exists
	var role models.Role
	if err := database.DB.First(&role, roleID).Error; err != nil {
		return err
	}

	// Remove role from menu using GORM's Association
	err := database.DB.Model(&menu).Association("Roles").Delete(&role)
	if err != nil {
		return err
	}

	return nil
}

// GetMenuRoles gets all roles for a menu
func (s *menuService) GetMenuRoles(menuID uuid.UUID) ([]models.Role, error) {
	var menu models.Menu
	var roles []models.Role

	// Check if the menu exists
	if err := database.DB.First(&menu, menuID).Error; err != nil {
		return nil, err
	}

	// Get roles for the menu
	err := database.DB.Model(&menu).Association("Roles").Find(&roles)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// GetRoleMenus gets all menus for a role
func (s *menuService) GetRoleMenus(roleID uuid.UUID) ([]models.Menu, error) {
	var role models.Role
	var menus []models.Menu

	// Check if the role exists
	if err := database.DB.First(&role, roleID).Error; err != nil {
		return nil, err
	}

	// Get menus for the role
	err := database.DB.Model(&role).Association("Menus").Find(&menus)
	if err != nil {
		return nil, err
	}

	return menus, nil
}
