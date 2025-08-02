package services

import (
	"go-next/internal/models"
	"go-next/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryService interface {
	GetAllCategories() ([]models.Category, error)
	GetCategoryByID(id string) (*models.Category, error)
	CreateCategory(category *models.Category) error
	UpdateCategory(category *models.Category) error
	DeleteCategory(id string) error
	CreateNested(category *models.Category, parentID *uuid.UUID) error
	MoveNested(id uuid.UUID, newParentID *uuid.UUID) error
	DeleteNested(id uuid.UUID) error
	GetSiblingCategory(id uuid.UUID) ([]models.Category, error)
	GetParentCategory(id uuid.UUID) (*models.Category, error)
	GetDescendantCategories(id uuid.UUID) ([]models.Category, error)
	GetChildrenCategories(id uuid.UUID) ([]models.Category, error)
}

type categoryService struct{}

func (s *categoryService) GetAllCategories() ([]models.Category, error) {
	var categories []models.Category
	err := database.DB.Find(&categories).Error
	return categories, err
}

func (s *categoryService) GetCategoryByID(id string) (*models.Category, error) {
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	var category models.Category
	err = database.DB.First(&category, categoryID).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *categoryService) CreateCategory(category *models.Category) error {
	return database.DB.Create(category).Error
}

func (s *categoryService) UpdateCategory(category *models.Category) error {
	return database.DB.Save(category).Error
}

func (s *categoryService) DeleteCategory(id string) error {
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return database.DB.Delete(&models.Category{}, categoryID).Error
}

func (s *categoryService) CreateNested(category *models.Category, parentID *uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if parentID != nil {
			var parent models.Category
			if err := tx.First(&parent, parentID).Error; err != nil {
				return err
			}
			if parent.RecordRight == 0 {
				return gorm.ErrRecordNotFound
			}

			// Update parent's right value
			tx.Model(&models.Category{}).
				Where("record_right >= ?", parent.RecordRight).
				Update("record_right", gorm.Expr("record_right + 2"))
			tx.Model(&models.Category{}).
				Where("record_left > ?", parent.RecordRight).
				Update("record_left", gorm.Expr("record_left + 2"))

			category.RecordLeft = parent.RecordRight
			category.RecordRight = parent.RecordRight + 1
			category.RecordDept = parent.RecordDept + 1
			category.ParentID = parentID
		} else {
			// Create as root
			var maxRight int
			tx.Model(&models.Category{}).Select("COALESCE(MAX(record_right), 0)").Scan(&maxRight)
			category.RecordLeft = maxRight + 1
			category.RecordRight = maxRight + 2
			category.RecordDept = 0
			category.ParentID = nil
		}

		return tx.Create(category).Error
	})
}

func (s *categoryService) MoveNested(id uuid.UUID, newParentID *uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var node models.Category
		if err := tx.First(&node, id).Error; err != nil {
			return err
		}
		if node.RecordLeft == 0 || node.RecordRight == 0 {
			return gorm.ErrRecordNotFound
		}
		left := node.RecordLeft
		right := node.RecordRight
		width := right - left + 1

		var newParentRight int
		var newDepth int = 0
		if newParentID != nil {
			var newParent models.Category
			if err := tx.First(&newParent, newParentID).Error; err != nil {
				return err
			}
			if newParent.RecordRight == 0 {
				return gorm.ErrRecordNotFound
			}
			newParentRight = newParent.RecordRight
			newDepth = newParent.RecordDept + 1
		} else {
			tx.Model(&models.Category{}).Select("COALESCE(MAX(record_right), 0)").Scan(&newParentRight)
			newParentRight++
		}

		if newParentRight >= left && newParentRight <= right {
			return gorm.ErrInvalidData
		}

		var offset int
		if newParentRight > right {
			offset = newParentRight - right - 1
		} else {
			offset = newParentRight - left
		}

		temp := -1
		tx.Model(&models.Category{}).
			Where("record_left >= ? AND record_right <= ?", left, right).
			Update("record_left", gorm.Expr("record_left * ?", temp))
		tx.Model(&models.Category{}).
			Where("record_left >= ? AND record_right <= ?", left, right).
			Update("record_right", gorm.Expr("record_right * ?", temp))

		tx.Model(&models.Category{}).
			Where("record_left > ?", right).
			Update("record_left", gorm.Expr("record_left - ?", width))
		tx.Model(&models.Category{}).
			Where("record_right > ?", right).
			Update("record_right", gorm.Expr("record_right - ?", width))

		if newParentRight > right {
			tx.Model(&models.Category{}).
				Where("record_left >= ?", newParentRight-width).
				Update("record_left", gorm.Expr("record_left + ?", width))
			tx.Model(&models.Category{}).
				Where("record_right >= ?", newParentRight-width).
				Update("record_right", gorm.Expr("record_right + ?", width))
			depthDiff := newDepth - node.RecordDept
			tx.Model(&models.Category{}).
				Where("record_left <= ? AND record_right >= ?", temp*left, temp*right).
				Updates(map[string]interface{}{
					"record_left":  gorm.Expr("record_left * -1 + ?", offset),
					"record_right": gorm.Expr("record_right * -1 + ?", offset),
					"record_dept":  gorm.Expr("record_dept + ?", depthDiff),
				})
		} else {
			tx.Model(&models.Category{}).
				Where("record_left >= ?", newParentRight).
				Update("record_left", gorm.Expr("record_left + ?", width))
			tx.Model(&models.Category{}).
				Where("record_right >= ?", newParentRight).
				Update("record_right", gorm.Expr("record_right + ?", width))
			depthDiff := newDepth - node.RecordDept
			tx.Model(&models.Category{}).
				Where("record_left <= ? AND record_right >= ?", temp*left, temp*right).
				Updates(map[string]interface{}{
					"record_left":  gorm.Expr("record_left * -1 + ?", offset),
					"record_right": gorm.Expr("record_right * -1 + ?", offset),
					"record_dept":  gorm.Expr("record_dept + ?", depthDiff),
				})
		}

		tx.Model(&models.Category{}).Where("id = ?", id).Update("parent_id", newParentID)

		return nil
	})
}

func (s *categoryService) DeleteNested(id uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var category models.Category
		if err := tx.First(&category, id).Error; err != nil {
			return err
		}
		if category.RecordLeft == 0 || category.RecordRight == 0 {
			return gorm.ErrRecordNotFound
		}

		left := category.RecordLeft
		right := category.RecordRight
		width := right - left + 1

		// Delete the category and all its descendants
		tx.Where("record_left >= ? AND record_right <= ?", left, right).Delete(&models.Category{})

		// Update the remaining nodes
		tx.Model(&models.Category{}).
			Where("record_left > ?", right).
			Update("record_left", gorm.Expr("record_left - ?", width))
		tx.Model(&models.Category{}).
			Where("record_right > ?", right).
			Update("record_right", gorm.Expr("record_right - ?", width))

		return nil
	})
}

func (s *categoryService) GetSiblingCategory(id uuid.UUID) ([]models.Category, error) {
	var category models.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		return nil, err
	}

	var siblings []models.Category
	err := database.DB.Where("parent_id = ? AND id != ?", category.ParentID, id).Find(&siblings).Error
	return siblings, err
}

func (s *categoryService) GetParentCategory(id uuid.UUID) (*models.Category, error) {
	var category models.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		return nil, err
	}

	if category.ParentID == nil {
		return nil, nil // No parent
	}

	var parent models.Category
	err := database.DB.First(&parent, category.ParentID).Error
	if err != nil {
		return nil, err
	}
	return &parent, nil
}

func (s *categoryService) GetDescendantCategories(id uuid.UUID) ([]models.Category, error) {
	var category models.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		return nil, err
	}

	var descendants []models.Category
	err := database.DB.Where("record_left > ? AND record_right < ?", category.RecordLeft, category.RecordRight).Find(&descendants).Error
	return descendants, err
}

func (s *categoryService) GetChildrenCategories(id uuid.UUID) ([]models.Category, error) {
	var children []models.Category
	err := database.DB.Where("parent_id = ?", id).Find(&children).Error
	return children, err
}

var CategorySvc CategoryService = &categoryService{}
