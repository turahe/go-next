package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wordpress-go-next/backend/internal/http/responses"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"
	"wordpress-go-next/backend/pkg/redis"
)

type RoleService interface {
	GetAllRoles(ctx context.Context) ([]models.Role, error)
	GetRolesWithPagination(ctx context.Context, page, perPage int, search string) (*responses.PaginationResponse, error)
	GetRoleByID(ctx context.Context, id string) (*models.Role, error)
	GetRoleByName(ctx context.Context, name string) (*models.Role, error)
	CreateRole(ctx context.Context, role *models.Role) error
	UpdateRole(ctx context.Context, role *models.Role) error
	DeleteRole(ctx context.Context, id string) error
	GetRolesByUser(ctx context.Context, userID uint) ([]models.Role, error)
	InvalidateRoleCache(ctx context.Context, roleID uint64) error
}

type roleService struct {
	Redis *redis.RedisService
}

func NewRoleService(redisService *redis.RedisService) RoleService {
	return &roleService{
		Redis: redisService,
	}
}

// Cache keys
const (
	roleCacheKeyPrefix     = "role:"
	roleNameCacheKeyPrefix = "role:name:"
	roleAllKeyPrefix       = "role:all:"
	roleUserKeyPrefix      = "role:user:"
)

func (s *roleService) getRoleCacheKey(id uint64) string {
	return fmt.Sprintf("%s%d", roleCacheKeyPrefix, id)
}

func (s *roleService) getRoleNameCacheKey(name string) string {
	return fmt.Sprintf("%s%s", roleNameCacheKeyPrefix, name)
}

func (s *roleService) getRoleAllCacheKey() string {
	return roleAllKeyPrefix + "list"
}

func (s *roleService) getRoleUserCacheKey(userID uint) string {
	return fmt.Sprintf("%s%d", roleUserKeyPrefix, userID)
}

func (s *roleService) GetAllRoles(ctx context.Context) ([]models.Role, error) {
	cacheKey := s.getRoleAllCacheKey()

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var roles []models.Role
		if err := json.Unmarshal([]byte(cached), &roles); err == nil {
			return roles, nil
		}
	}

	var roles []models.Role
	err := database.DB.WithContext(ctx).Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get all roles: %w", err)
	}

	// Cache the result
	if data, err := json.Marshal(roles); err == nil {
		err := s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute)
		if err != nil {
			return nil, err
		}
	}

	return roles, err
}

func (s *roleService) GetRolesWithPagination(ctx context.Context, page, perPage int, search string) (*responses.PaginationResponse, error) {
	var roles []models.Role
	params := PaginationParams{
		Page:    page,
		PerPage: perPage,
	}

	query := database.DB
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name LIKE ?", like)
	}

	result, err := (&BaseService{Redis: s.Redis}).PaginateWithCacheQuery(ctx, &models.Role{}, params, &roles, "", 0, query)
	if err != nil {
		return nil, err
	}

	result.Data = roles
	return result, nil
}

func (s *roleService) GetRoleByID(ctx context.Context, id string) (*models.Role, error) {
	// Parse ID to uint64 for cache key
	var roleID uint64
	if _, err := fmt.Sscanf(id, "%d", &roleID); err != nil {
		return nil, fmt.Errorf("invalid role ID format: %w", err)
	}

	cacheKey := s.getRoleCacheKey(roleID)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var role models.Role
		if err := json.Unmarshal([]byte(cached), &role); err == nil {
			return &role, nil
		}
	}

	var role models.Role
	err := database.DB.WithContext(ctx).First(&role, id).Error
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	// Cache the result
	if err := s.cacheRole(ctx, &role); err != nil {
		fmt.Printf("Warning: failed to cache role %d: %v\n", role.ID, err)
	}

	return &role, err
}

func (s *roleService) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	cacheKey := s.getRoleNameCacheKey(name)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var role models.Role
		if err := json.Unmarshal([]byte(cached), &role); err == nil {
			return &role, nil
		}
	}

	var role models.Role
	err := database.DB.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	// Cache the result
	if err := s.cacheRole(ctx, &role); err != nil {
		fmt.Printf("Warning: failed to cache role %d: %v\n", role.ID, err)
	}

	return &role, err
}

func (s *roleService) CreateRole(ctx context.Context, role *models.Role) error {
	if err := database.DB.WithContext(ctx).Create(role).Error; err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	// Cache the new role
	if err := s.cacheRole(ctx, role); err != nil {
		fmt.Printf("Warning: failed to cache role %d: %v\n", role.ID, err)
	}

	// Invalidate related caches
	s.invalidateRelatedCaches(ctx)

	return nil
}

func (s *roleService) UpdateRole(ctx context.Context, role *models.Role) error {
	if err := database.DB.WithContext(ctx).Save(role).Error; err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	// Update cache
	if err := s.cacheRole(ctx, role); err != nil {
		fmt.Printf("Warning: failed to cache role %d: %v\n", role.ID, err)
	}

	// Invalidate related caches
	s.invalidateRelatedCaches(ctx)

	return nil
}

func (s *roleService) DeleteRole(ctx context.Context, id string) error {
	// Parse ID to uint64 for cache key
	var roleID uint64
	if _, err := fmt.Sscanf(id, "%d", &roleID); err != nil {
		return fmt.Errorf("invalid role ID format: %w", err)
	}
	// Get role first to invalidate related caches
	var role models.Role
	if err := database.DB.WithContext(ctx).First(&role, id).Error; err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if err := database.DB.WithContext(ctx).Delete(&models.Role{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	// Invalidate caches
	s.invalidateRoleCaches(ctx, roleID)
	s.invalidateRelatedCaches(ctx)

	return nil
}

func (s *roleService) GetRolesByUser(ctx context.Context, userID uint) ([]models.Role, error) {
	cacheKey := s.getRoleUserCacheKey(userID)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var roles []models.Role
		if err := json.Unmarshal([]byte(cached), &roles); err == nil {
			return roles, nil
		}
	}

	var user models.User
	err := database.DB.WithContext(ctx).Preload("Roles").First(&user, userID).Error
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Cache the result
	if data, err := json.Marshal(user.Roles); err == nil {
		err := s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute)
		if err != nil {
			return nil, err
		}
	}

	return user.Roles, err
}

func (s *roleService) InvalidateRoleCache(ctx context.Context, roleID uint64) error {
	s.invalidateRoleCaches(ctx, roleID)
	return nil
}

// Helper methods
func (s *roleService) cacheRole(ctx context.Context, role *models.Role) error {
	data, err := json.Marshal(role)
	if err != nil {
		return err
	}

	// Cache by ID
	cacheKey := s.getRoleCacheKey(role.ID)
	if err := s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute); err != nil {
		return err
	}

	// Cache by name
	nameCacheKey := s.getRoleNameCacheKey(role.Name)
	return s.Redis.SetWithTTL(ctx, nameCacheKey, string(data), 30*time.Minute)
}

func (s *roleService) invalidateRoleCaches(ctx context.Context, roleID uint64) {
	cacheKeys := []string{
		s.getRoleCacheKey(roleID),
	}

	for _, key := range cacheKeys {
		err := s.Redis.Delete(ctx, key)
		if err != nil {
			return
		}
	}
}

func (s *roleService) invalidateRelatedCaches(ctx context.Context) {
	// Invalidate all roles cache and user role caches
	patterns := []string{
		roleAllKeyPrefix + "*",
		roleUserKeyPrefix + "*",
	}

	for _, pattern := range patterns {
		err := s.Redis.DeletePattern(ctx, pattern)
		if err != nil {
			return
		}
	}
}

var RoleSvc RoleService = &roleService{}
