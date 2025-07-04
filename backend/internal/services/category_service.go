package services

import (
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"

	"gorm.io/gorm"
)

type CategoryService interface {
	GetAllCategories() ([]models.Category, error)
	GetCategoryByID(id string) (*models.Category, error)
	CreateCategory(category *models.Category) error
	UpdateCategory(category *models.Category) error
	DeleteCategory(id string) error
	CreateNested(category *models.Category, parentID *int64) error
	MoveNested(id uint, newParentID *int64) error
	DeleteNested(id uint) error
	GetSiblingCategory(id uint) ([]models.Category, error)
	GetParentCategory(id uint) (*models.Category, error)
	GetDescendantCategories(id uint) ([]models.Category, error)
	GetChildrenCategories(id uint) ([]models.Category, error)
}

type categoryService struct{}

func (s *categoryService) GetAllCategories() ([]models.Category, error) {
	var categories []models.Category
	err := database.DB.Find(&categories).Error
	return categories, err
}

func (s *categoryService) GetCategoryByID(id string) (*models.Category, error) {
	var category models.Category
	err := database.DB.First(&category, id).Error
	return &category, err
}

func (s *categoryService) CreateCategory(category *models.Category) error {
	return database.DB.Create(category).Error
}

func (s *categoryService) UpdateCategory(category *models.Category) error {
	return database.DB.Model(&models.Category{}).Where("id = ?", category.ID).Updates(map[string]interface{}{
		"name":        category.Name,
		"description": category.Description,
	}).Error
}

func (s *categoryService) DeleteCategory(id string) error {
	return database.DB.Delete(&models.Category{}, id).Error
}

func (s *categoryService) CreateNested(category *models.Category, parentID *int64) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var left int64
		var depth int64 = 0
		if parentID != nil {
			var parent models.Category
			if err := tx.First(&parent, *parentID).Error; err != nil {
				return err
			}
			if parent.RecordRight == nil {
				return gorm.ErrRecordNotFound
			}
			left = *parent.RecordRight
			depth = 0
			if parent.RecordDept != nil {
				depth = *parent.RecordDept + 1
			}
			tx.Model(&models.Category{}).Where("record_right >= ?", left).Update("record_right", gorm.Expr("record_right + 2"))
			tx.Model(&models.Category{}).Where("record_left > ?", left-1).Update("record_left", gorm.Expr("record_left + 2"))
		} else {
			tx.Model(&models.Category{}).Select("COALESCE(MAX(record_right), 0)").Scan(&left)
			left++
		}
		right := left + 1
		category.RecordLeft = &left
		category.RecordRight = &right
		category.RecordDept = &depth
		category.ParentID = parentID
		return tx.Create(category).Error
	})
}

func (s *categoryService) MoveNested(id uint, newParentID *int64) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var node models.Category
		if err := tx.First(&node, id).Error; err != nil {
			return err
		}
		if node.RecordLeft == nil || node.RecordRight == nil {
			return gorm.ErrRecordNotFound
		}
		left := *node.RecordLeft
		right := *node.RecordRight
		width := right - left + 1

		var newParentRight int64
		var newDepth int64 = 0
		if newParentID != nil {
			var newParent models.Category
			if err := tx.First(&newParent, *newParentID).Error; err != nil {
				return err
			}
			if newParent.RecordRight == nil {
				return gorm.ErrRecordNotFound
			}
			newParentRight = *newParent.RecordRight
			if newParent.RecordDept != nil {
				newDepth = *newParent.RecordDept + 1
			}
		} else {
			tx.Model(&models.Category{}).Select("COALESCE(MAX(record_right), 0)").Scan(&newParentRight)
			newParentRight++
		}

		if newParentRight >= left && newParentRight <= right {
			return gorm.ErrInvalidData
		}

		var offset int64
		if newParentRight > right {
			offset = newParentRight - right - 1
		} else {
			offset = newParentRight - left
		}

		temp := int64(-1)
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
			depthDiff := newDepth
			if node.RecordDept != nil {
				depthDiff = newDepth - *node.RecordDept
			}
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
			depthDiff := newDepth
			if node.RecordDept != nil {
				depthDiff = newDepth - *node.RecordDept
			}
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

func (s *categoryService) DeleteNested(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var node models.Category
		if err := tx.First(&node, id).Error; err != nil {
			return err
		}
		if node.RecordLeft == nil || node.RecordRight == nil {
			return gorm.ErrRecordNotFound
		}
		left := *node.RecordLeft
		right := *node.RecordRight
		width := right - left + 1
		tx.Where("record_left >= ? AND record_right <= ?", left, right).Delete(&models.Category{})
		tx.Model(&models.Category{}).Where("record_left > ?", right).Update("record_left", gorm.Expr("record_left - ?", width))
		tx.Model(&models.Category{}).Where("record_right > ?", right).Update("record_right", gorm.Expr("record_right - ?", width))
		return nil
	})
}

func (s *categoryService) GetSiblingCategory(id uint) ([]models.Category, error) {
	var node models.Category
	if err := database.DB.First(&node, id).Error; err != nil {
		return nil, err
	}
	var siblings []models.Category
	if node.ParentID != nil {
		database.DB.Where("parent_id = ? AND id != ?", *node.ParentID, id).Find(&siblings)
	} else {
		database.DB.Where("parent_id IS NULL AND id != ?", id).Find(&siblings)
	}
	return siblings, nil
}

func (s *categoryService) GetParentCategory(id uint) (*models.Category, error) {
	var node models.Category
	if err := database.DB.First(&node, id).Error; err != nil {
		return nil, err
	}
	if node.ParentID == nil {
		return nil, nil
	}
	var parent models.Category
	if err := database.DB.First(&parent, *node.ParentID).Error; err != nil {
		return nil, err
	}
	return &parent, nil
}

func (s *categoryService) GetDescendantCategories(id uint) ([]models.Category, error) {
	var node models.Category
	if err := database.DB.First(&node, id).Error; err != nil {
		return nil, err
	}
	if node.RecordLeft == nil || node.RecordRight == nil {
		return nil, gorm.ErrRecordNotFound
	}
	var descendants []models.Category
	database.DB.Where("record_left > ? AND record_right < ?", *node.RecordLeft, *node.RecordRight).Find(&descendants)
	return descendants, nil
}

func (s *categoryService) GetChildrenCategories(id uint) ([]models.Category, error) {
	var children []models.Category
	if err := database.DB.Where("parent_id = ?", id).Find(&children).Error; err != nil {
		return nil, err
	}
	return children, nil
}

var CategorySvc CategoryService = &categoryService{}
