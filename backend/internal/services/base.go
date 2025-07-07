package services

import (
	"context"
	"fmt"
	"time"

	"wordpress-go-next/backend/internal/http/responses"
	"wordpress-go-next/backend/pkg/database"
	"wordpress-go-next/backend/pkg/redis"
)

type BaseService struct {
	Redis *redis.RedisService
}

// NewBaseService creates a new base service with Redis caching
func NewBaseService(redisService *redis.RedisService) *BaseService {
	return &BaseService{
		Redis: redisService,
	}
}

// CreateWithCache creates a record and invalidates related caches
func (s *BaseService) CreateWithCache(ctx context.Context, value interface{}, cacheKey string) error {
	if err := database.DB.Create(value).Error; err != nil {
		return err
	}

	// Invalidate related caches
	if s.Redis != nil && cacheKey != "" {
		err := s.Redis.DeletePattern(ctx, fmt.Sprintf("%s*", cacheKey))
		if err != nil {
			return err
		}
	}

	return nil
}

// SaveWithCache saves a record and invalidates related caches
func (s *BaseService) SaveWithCache(ctx context.Context, value interface{}, cacheKey string) error {
	if err := database.DB.Save(value).Error; err != nil {
		return err
	}

	// Invalidate related caches
	if s.Redis != nil && cacheKey != "" {
		err := s.Redis.DeletePattern(ctx, fmt.Sprintf("%s*", cacheKey))
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateWithCache updates a record and invalidates related caches
func (s *BaseService) UpdateWithCache(ctx context.Context, value interface{}, cacheKey string) error {
	if err := database.DB.Save(value).Error; err != nil {
		return err
	}

	// Invalidate related caches
	if s.Redis != nil && cacheKey != "" {
		err := s.Redis.DeletePattern(ctx, fmt.Sprintf("%s*", cacheKey))
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteWithCache deletes a record and invalidates related caches
func (s *BaseService) DeleteWithCache(ctx context.Context, value interface{}, cacheKey string) error {
	if err := database.DB.Delete(value).Error; err != nil {
		return err
	}

	// Invalidate related caches
	if s.Redis != nil && cacheKey != "" {
		err := s.Redis.DeletePattern(ctx, fmt.Sprintf("%s*", cacheKey))
		if err != nil {
			return err
		}
	}

	return nil
}

// GetByIDWithCache retrieves a record by ID with caching
func (s *BaseService) GetByIDWithCache(ctx context.Context, id string, dest interface{}, cacheKey string, ttl time.Duration) error {
	if s.Redis != nil && cacheKey != "" {
		// Try to get from cache first
		fullKey := fmt.Sprintf("%s%s", cacheKey, id)
		if err := s.Redis.GetCache(ctx, fullKey, dest); err == nil {
			return nil // Cache hit
		}
	}

	// Cache miss or Redis not available, get from database
	if err := database.DB.First(dest, id).Error; err != nil {
		return err
	}

	// Cache the result
	if s.Redis != nil && cacheKey != "" {
		fullKey := fmt.Sprintf("%s%s", cacheKey, id)
		err := s.Redis.SetCache(ctx, fullKey, dest, ttl)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetByIDWithCacheAndPreload retrieves a record by ID with caching and preloading
func (s *BaseService) GetByIDWithCacheAndPreload(ctx context.Context, id string, dest interface{}, cacheKey string, ttl time.Duration, preloads ...string) error {
	if s.Redis != nil && cacheKey != "" {
		// Try to get from cache first
		fullKey := fmt.Sprintf("%s%s", cacheKey, id)
		if err := s.Redis.GetCache(ctx, fullKey, dest); err == nil {
			return nil // Cache hit
		}
	}

	// Cache miss or Redis not available, get from database
	db := database.DB
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	if err := db.First(dest, id).Error; err != nil {
		return err
	}

	// Cache the result
	if s.Redis != nil && cacheKey != "" {
		fullKey := fmt.Sprintf("%s%s", cacheKey, id)
		err := s.Redis.SetCache(ctx, fullKey, dest, ttl)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetAllWithCache retrieves all records with caching
func (s *BaseService) GetAllWithCache(ctx context.Context, dest interface{}, cacheKey string, ttl time.Duration) error {
	if s.Redis != nil && cacheKey != "" {
		// Try to get from cache first
		if err := s.Redis.GetCache(ctx, cacheKey, dest); err == nil {
			return nil // Cache hit
		}
	}

	// Cache miss or Redis not available, get from database
	if err := database.DB.Find(dest).Error; err != nil {
		return err
	}

	// Cache the result
	if s.Redis != nil && cacheKey != "" {
		err := s.Redis.SetCache(ctx, cacheKey, dest, ttl)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetAllWithCacheAndPreload retrieves all records with caching and preloading
func (s *BaseService) GetAllWithCacheAndPreload(ctx context.Context, dest interface{}, cacheKey string, ttl time.Duration, preloads ...string) error {
	if s.Redis != nil && cacheKey != "" {
		// Try to get from cache first
		if err := s.Redis.GetCache(ctx, cacheKey, dest); err == nil {
			return nil // Cache hit
		}
	}

	// Cache miss or Redis not available, get from database
	db := database.DB
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	if err := db.Find(dest).Error; err != nil {
		return err
	}

	// Cache the result
	if s.Redis != nil && cacheKey != "" {
		err := s.Redis.SetCache(ctx, cacheKey, dest, ttl)
		if err != nil {
			return err
		}
	}

	return nil
}

// PaginationParams holds pagination query parameters
type PaginationParams struct {
	Page    int
	PerPage int
}

// PaginateResult holds paginated data and meta info
type PaginateResult struct {
	Data         interface{}
	TotalCount   int64
	TotalPage    int64
	CurrentPage  int64
	LastPage     int64
	PerPage      int64
	NextPage     int64
	PreviousPage int64
}

// PaginateWithCache is a generic pagination method with caching
func (s *BaseService) PaginateWithCache(ctx context.Context, model interface{}, params PaginationParams, out interface{}, cacheKey string, ttl time.Duration) (*responses.PaginationResponse, error) {
	// Create cache key for this specific pagination
	paginationKey := fmt.Sprintf("%s:page:%d:per_page:%d", cacheKey, params.Page, params.PerPage)

	if s.Redis != nil && cacheKey != "" {
		// Try to get from cache first
		var cachedResult responses.PaginationResponse
		if err := s.Redis.GetCache(ctx, paginationKey, &cachedResult); err == nil {
			return &cachedResult, nil // Cache hit
		}
	}

	// Cache miss or Redis not available, get from database
	db := database.DB.Model(model)

	var totalCount int64
	if err := db.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 10
	}

	offset := (params.Page - 1) * params.PerPage

	if err := db.Limit(params.PerPage).Offset(offset).Find(out).Error; err != nil {
		return nil, err
	}

	totalPage := (totalCount + int64(params.PerPage) - 1) / int64(params.PerPage)
	lastPage := totalPage
	var nextPage, prevPage int64
	if int64(params.Page) < totalPage {
		nextPage = int64(params.Page) + 1
	}
	if params.Page > 1 {
		prevPage = int64(params.Page) - 1
	}

	result := &responses.PaginationResponse{
		Data:         out,
		TotalCount:   totalCount,
		TotalPage:    totalPage,
		CurrentPage:  int64(params.Page),
		LastPage:     lastPage,
		PerPage:      int64(params.PerPage),
		NextPage:     nextPage,
		PreviousPage: prevPage,
	}

	// Cache the result
	if s.Redis != nil && cacheKey != "" {
		err := s.Redis.SetCache(ctx, paginationKey, result, ttl)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// Paginate is a generic pagination method for all service models (legacy)
func (s *BaseService) Paginate(model interface{}, params PaginationParams, out interface{}) (*responses.PaginationResponse, error) {
	db := database.DB.Model(model)

	var totalCount int64
	if err := db.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 10
	}

	offset := (params.Page - 1) * params.PerPage

	if err := db.Limit(params.PerPage).Offset(offset).Find(out).Error; err != nil {
		return nil, err
	}

	totalPage := (totalCount + int64(params.PerPage) - 1) / int64(params.PerPage)
	lastPage := totalPage
	var nextPage, prevPage int64
	if int64(params.Page) < totalPage {
		nextPage = int64(params.Page) + 1
	}
	if params.Page > 1 {
		prevPage = int64(params.Page) - 1
	}

	return &responses.PaginationResponse{
		Data:         out,
		TotalCount:   totalCount,
		TotalPage:    totalPage,
		CurrentPage:  int64(params.Page),
		LastPage:     lastPage,
		PerPage:      int64(params.PerPage),
		NextPage:     nextPage,
		PreviousPage: prevPage,
	}, nil
}

// Legacy methods for backward compatibility
func (s *BaseService) Create(value interface{}) error {
	return database.DB.Create(value).Error
}

func (s *BaseService) Save(value interface{}) error {
	return database.DB.Save(value).Error
}

func (s *BaseService) Update(value interface{}) error {
	return database.DB.Save(value).Error
}

func (s *BaseService) Delete(value interface{}) error {
	return database.DB.Delete(value).Error
}

func (s *BaseService) FirstOrFail(out interface{}, query interface{}, args ...interface{}) error {
	db := database.DB
	if err := db.Where(query, args...).First(out).Error; err != nil {
		return err
	}
	return nil
}

// Cache utility methods
func (s *BaseService) InvalidateCache(ctx context.Context, pattern string) error {
	if s.Redis != nil {
		return s.Redis.DeletePattern(ctx, pattern)
	}
	return nil
}

func (s *BaseService) SetCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if s.Redis != nil {
		return s.Redis.SetCache(ctx, key, value, ttl)
	}
	return nil
}

func (s *BaseService) GetCache(ctx context.Context, key string, dest interface{}) error {
	if s.Redis != nil {
		return s.Redis.GetCache(ctx, key, dest)
	}
	return fmt.Errorf("redis not available")
}

// Cache key generators
func (s *BaseService) GetCacheKey(prefix, id string) string {
	return fmt.Sprintf("%s%s", prefix, id)
}

func (s *BaseService) GetListCacheKey(prefix string) string {
	return fmt.Sprintf("%s:list", prefix)
}

func (s *BaseService) GetSearchCacheKey(prefix, query string) string {
	return fmt.Sprintf("%s:search:%s", prefix, query)
}
