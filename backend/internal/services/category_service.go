package services

import (
	"context"
	"fmt"

	"wordpress-go-next/backend/internal/http/responses"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"
	"wordpress-go-next/backend/pkg/redis"

	"gorm.io/gorm"
)

type CategoryService interface {
	GetAllCategories(ctx context.Context) ([]models.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*models.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error)
	GetCategoriesWithPagination(ctx context.Context, page, perPage int) (*responses.PaginationResponse, error)
	GetRootCategories(ctx context.Context) ([]models.Category, error)
	GetActiveCategories(ctx context.Context) ([]models.Category, error)
	CreateCategory(ctx context.Context, category *models.Category) error
	UpdateCategory(ctx context.Context, category *models.Category) error
	DeleteCategory(ctx context.Context, id string) error
	CreateNested(ctx context.Context, category *models.Category, parentID *int64) error
	MoveNested(ctx context.Context, id uint, newParentID *int64) error
	DeleteNested(ctx context.Context, id uint) error
	GetSiblingCategory(ctx context.Context, id uint) ([]models.Category, error)
	GetParentCategory(ctx context.Context, id uint) (*models.Category, error)
	GetDescendantCategories(ctx context.Context, id uint) ([]models.Category, error)
	GetChildrenCategories(ctx context.Context, id uint) ([]models.Category, error)
	GetAncestorCategories(ctx context.Context, id uint) ([]models.Category, error)
	SearchCategories(ctx context.Context, query string) ([]models.Category, error)
	GetCategoryStats(ctx context.Context, categoryID string) (map[string]interface{}, error)
	GetCategoryCount(ctx context.Context) (int64, error)
	GetCategoryTree(ctx context.Context) ([]models.Category, error)
}

type categoryService struct {
	*BaseService
}

func NewCategoryService(redisService *redis.RedisService) CategoryService {
	return &categoryService{
		BaseService: NewBaseService(redisService),
	}
}

func (s *categoryService) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	cacheKey := s.GetListCacheKey(redis.CategoryCachePrefix)

	err := s.GetAllWithCacheAndPreload(ctx, &categories, cacheKey, redis.DefaultTTL, "Posts", "Parent", "Children")
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	var category models.Category
	cacheKey := s.GetCacheKey(redis.CategoryCachePrefix, id)

	err := s.GetByIDWithCacheAndPreload(ctx, id, &category, cacheKey, redis.DefaultTTL, "Posts", "Parent", "Children")
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (s *categoryService) GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error) {
	var category models.Category
	cacheKey := fmt.Sprintf("%sslug:%s", redis.CategoryCachePrefix, slug)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &category); err == nil {
			return &category, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("Posts").Preload("Parent").Preload("Children").
		Where("slug = ?", slug).First(&category).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		s.Redis.SetCache(ctx, cacheKey, &category, redis.DefaultTTL)
	}

	return &category, nil
}

func (s *categoryService) GetCategoriesWithPagination(ctx context.Context, page, perPage int) (*responses.PaginationResponse, error) {
	var categories []models.Category
	cacheKey := s.GetListCacheKey(redis.CategoryCachePrefix)

	params := PaginationParams{
		Page:    page,
		PerPage: perPage,
	}

	result, err := s.PaginateWithCache(ctx, &models.Category{}, params, &categories, cacheKey, redis.DefaultTTL)
	if err != nil {
		return nil, err
	}

	// Preload relationships for each category
	for i := range categories {
		if err := database.DB.Preload("Posts").Preload("Parent").Preload("Children").
			First(&categories[i], categories[i].ID).Error; err != nil {
			return nil, err
		}
	}

	result.Data = categories
	return result, nil
}

func (s *categoryService) GetRootCategories(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	cacheKey := fmt.Sprintf("%sroot", redis.CategoryCachePrefix)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &categories); err == nil {
			return categories, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("Posts").Preload("Children").
		Where("parent_id IS NULL").Find(&categories).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		s.Redis.SetCache(ctx, cacheKey, categories, redis.DefaultTTL)
	}

	return categories, nil
}

func (s *categoryService) GetActiveCategories(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	cacheKey := fmt.Sprintf("%sactive", redis.CategoryCachePrefix)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &categories); err == nil {
			return categories, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("Posts").Preload("Parent").Preload("Children").
		Where("is_active = ?", true).Find(&categories).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		s.Redis.SetCache(ctx, cacheKey, categories, redis.ShortTTL)
	}

	return categories, nil
}

func (s *categoryService) CreateCategory(ctx context.Context, category *models.Category) error {
	return s.CreateWithCache(ctx, category, redis.CategoryCachePrefix)
}

func (s *categoryService) UpdateCategory(ctx context.Context, category *models.Category) error {
	return s.UpdateWithCache(ctx, category, redis.CategoryCachePrefix)
}

func (s *categoryService) DeleteCategory(ctx context.Context, id string) error {
	category := &models.Category{}
	if err := database.DB.First(category, id).Error; err != nil {
		return err
	}

	return s.DeleteWithCache(ctx, category, redis.CategoryCachePrefix)
}

func (s *categoryService) CreateNested(ctx context.Context, category *models.Category, parentID *int64) error {
	err := database.DB.Transaction(func(tx *gorm.DB) error {
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

	if err != nil {
		return err
	}

	// Invalidate caches
	if s.Redis != nil {
		s.Redis.DeletePattern(ctx, fmt.Sprintf("%s*", redis.CategoryCachePrefix))
	}

	return nil
}

func (s *categoryService) MoveNested(ctx context.Context, id uint, newParentID *int64) error {
	err := database.DB.Transaction(func(tx *gorm.DB) error {
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

	if err != nil {
		return err
	}

	// Invalidate caches
	if s.Redis != nil {
		s.Redis.DeletePattern(ctx, fmt.Sprintf("%s*", redis.CategoryCachePrefix))
	}

	return nil
}

func (s *categoryService) DeleteNested(ctx context.Context, id uint) error {
	err := database.DB.Transaction(func(tx *gorm.DB) error {
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

	if err != nil {
		return err
	}

	// Invalidate caches
	if s.Redis != nil {
		s.Redis.DeletePattern(ctx, fmt.Sprintf("%s*", redis.CategoryCachePrefix))
	}

	return nil
}

func (s *categoryService) GetSiblingCategory(ctx context.Context, id uint) ([]models.Category, error) {
	var categories []models.Category
	cacheKey := fmt.Sprintf("%ssiblings:%d", redis.CategoryCachePrefix, id)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &categories); err == nil {
			return categories, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("Posts").Preload("Parent").
		Where("parent_id = (SELECT parent_id FROM categories WHERE id = ?)", id).
		Where("id != ?", id).Find(&categories).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		s.Redis.SetCache(ctx, cacheKey, categories, redis.DefaultTTL)
	}

	return categories, nil
}

func (s *categoryService) GetParentCategory(ctx context.Context, id uint) (*models.Category, error) {
	var category models.Category
	cacheKey := fmt.Sprintf("%sparent:%d", redis.CategoryCachePrefix, id)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &category); err == nil {
			return &category, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("Posts").Preload("Parent").
		Where("id = (SELECT parent_id FROM categories WHERE id = ?)", id).
		First(&category).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		s.Redis.SetCache(ctx, cacheKey, &category, redis.DefaultTTL)
	}

	return &category, nil
}

func (s *categoryService) GetDescendantCategories(ctx context.Context, id uint) ([]models.Category, error) {
	var categories []models.Category
	cacheKey := fmt.Sprintf("%sdescendants:%d", redis.CategoryCachePrefix, id)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &categories); err == nil {
			return categories, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("Posts").Preload("Parent").
		Where("record_left > (SELECT record_left FROM categories WHERE id = ?) AND record_right < (SELECT record_right FROM categories WHERE id = ?)", id, id).
		Find(&categories).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		s.Redis.SetCache(ctx, cacheKey, categories, redis.DefaultTTL)
	}

	return categories, nil
}

func (s *categoryService) GetChildrenCategories(ctx context.Context, id uint) ([]models.Category, error) {
	var categories []models.Category
	cacheKey := fmt.Sprintf("%schildren:%d", redis.CategoryCachePrefix, id)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &categories); err == nil {
			return categories, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("Posts").Preload("Parent").
		Where("parent_id = ?", id).Find(&categories).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		s.Redis.SetCache(ctx, cacheKey, categories, redis.DefaultTTL)
	}

	return categories, nil
}

func (s *categoryService) GetAncestorCategories(ctx context.Context, id uint) ([]models.Category, error) {
	var categories []models.Category
	cacheKey := fmt.Sprintf("%sancestors:%d", redis.CategoryCachePrefix, id)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &categories); err == nil {
			return categories, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("Posts").Preload("Parent").
		Where("record_left < (SELECT record_left FROM categories WHERE id = ?) AND record_right > (SELECT record_right FROM categories WHERE id = ?)", id, id).
		Order("record_left ASC").Find(&categories).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		s.Redis.SetCache(ctx, cacheKey, categories, redis.DefaultTTL)
	}

	return categories, nil
}

func (s *categoryService) SearchCategories(ctx context.Context, query string) ([]models.Category, error) {
	var categories []models.Category
	cacheKey := s.GetSearchCacheKey(redis.CategoryCachePrefix, query)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &categories); err == nil {
			return categories, nil
		}
	}

	// Search in database
	if err := database.DB.Preload("Posts").Preload("Parent").
		Where("name LIKE ? OR description LIKE ?",
			"%"+query+"%", "%"+query+"%").
		Find(&categories).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		s.Redis.SetCache(ctx, cacheKey, categories, redis.ShortTTL)
	}

	return categories, nil
}

func (s *categoryService) GetCategoryStats(ctx context.Context, categoryID string) (map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("%sstats:%s", redis.CategoryCachePrefix, categoryID)

	// Try cache first
	if s.Redis != nil {
		var cachedStats map[string]interface{}
		if err := s.Redis.GetCache(ctx, cacheKey, &cachedStats); err == nil {
			return cachedStats, nil
		}
	}

	// Calculate stats from database
	var postCount int64
	var childCount int64
	var category models.Category

	if err := database.DB.First(&category, categoryID).Error; err != nil {
		return nil, err
	}

	if err := database.DB.Model(&models.Post{}).Where("category_id = ?", categoryID).Count(&postCount).Error; err != nil {
		return nil, err
	}

	if err := database.DB.Model(&models.Category{}).Where("parent_id = ?", categoryID).Count(&childCount).Error; err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"category_id": categoryID,
		"post_count":  postCount,
		"child_count": childCount,
		"is_active":   category.IsActive,
		"is_root":     category.IsRoot(),
		"depth":       category.GetDepth(),
		"created_at":  category.CreatedAt,
		"updated_at":  category.UpdatedAt,
	}

	// Cache the stats
	if s.Redis != nil {
		s.Redis.SetCache(ctx, cacheKey, stats, redis.ShortTTL)
	}

	return stats, nil
}

func (s *categoryService) GetCategoryCount(ctx context.Context) (int64, error) {
	cacheKey := fmt.Sprintf("%scount", redis.CategoryCachePrefix)

	// Try cache first
	if s.Redis != nil {
		var count int64
		if err := s.Redis.GetCache(ctx, cacheKey, &count); err == nil {
			return count, nil
		}
	}

	// Get from database
	var count int64
	if err := database.DB.Model(&models.Category{}).Count(&count).Error; err != nil {
		return 0, err
	}

	// Cache the result
	if s.Redis != nil {
		s.Redis.SetCache(ctx, cacheKey, count, redis.LongTTL)
	}

	return count, nil
}

func (s *categoryService) GetCategoryTree(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	cacheKey := fmt.Sprintf("%stree", redis.CategoryCachePrefix)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &categories); err == nil {
			return categories, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("Posts").Preload("Parent").Preload("Children").
		Order("record_left ASC").Find(&categories).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		s.Redis.SetCache(ctx, cacheKey, categories, redis.DefaultTTL)
	}

	return categories, nil
}

// Legacy methods for backward compatibility
func (s *categoryService) GetAllCategories() ([]models.Category, error) {
	return s.GetAllCategories(context.Background())
}

func (s *categoryService) GetCategoryByID(id string) (*models.Category, error) {
	return s.GetCategoryByID(context.Background(), id)
}

func (s *categoryService) CreateCategory(category *models.Category) error {
	return s.CreateCategory(context.Background(), category)
}

func (s *categoryService) UpdateCategory(category *models.Category) error {
	return s.UpdateCategory(context.Background(), category)
}

func (s *categoryService) DeleteCategory(id string) error {
	return s.DeleteCategory(context.Background(), id)
}

func (s *categoryService) CreateNested(category *models.Category, parentID *int64) error {
	return s.CreateNested(context.Background(), category, parentID)
}

func (s *categoryService) MoveNested(id uint, newParentID *int64) error {
	return s.MoveNested(context.Background(), id, newParentID)
}

func (s *categoryService) DeleteNested(id uint) error {
	return s.DeleteNested(context.Background(), id)
}

func (s *categoryService) GetSiblingCategory(id uint) ([]models.Category, error) {
	return s.GetSiblingCategory(context.Background(), id)
}

func (s *categoryService) GetParentCategory(id uint) (*models.Category, error) {
	return s.GetParentCategory(context.Background(), id)
}

func (s *categoryService) GetDescendantCategories(id uint) ([]models.Category, error) {
	return s.GetDescendantCategories(context.Background(), id)
}

func (s *categoryService) GetChildrenCategories(id uint) ([]models.Category, error) {
	return s.GetChildrenCategories(context.Background(), id)
}

// Global service instance
var CategorySvc CategoryService
