package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"
	"wordpress-go-next/backend/pkg/redis"
)

type UserRoleService interface {
	AssignRoleToUser(ctx context.Context, user *models.User, role *models.Role) error
	RemoveRoleFromUser(ctx context.Context, user *models.User, role *models.Role) error
	ListUserRoles(ctx context.Context, user *models.User) ([]models.Role, error)
	GetUsersByRole(ctx context.Context, role *models.Role) ([]models.User, error)
	HasRole(ctx context.Context, user *models.User, roleName string) (bool, error)
	InvalidateUserRoleCache(ctx context.Context, userID uint) error
}

type userRoleService struct {
	Redis *redis.RedisService
}

func NewUserRoleService(redisService *redis.RedisService) UserRoleService {
	return &userRoleService{
		Redis: redisService,
	}
}

// Cache keys
const (
	userRoleCacheKeyPrefix = "user_role:"
	roleUserCacheKeyPrefix = "role_user:"
)

func (s *userRoleService) getUserRoleCacheKey(userID uint) string {
	return fmt.Sprintf("%s%d", userRoleCacheKeyPrefix, userID)
}

func (s *userRoleService) getRoleUserCacheKey(roleID uint) string {
	return fmt.Sprintf("%s%d", roleUserCacheKeyPrefix, roleID)
}

func (s *userRoleService) AssignRoleToUser(ctx context.Context, user *models.User, role *models.Role) error {
	if err := database.DB.WithContext(ctx).Model(user).Association("Roles").Append(role); err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	// Invalidate related caches
	s.invalidateUserRoleCaches(ctx, user.ID, role.ID)

	return nil
}

func (s *userRoleService) RemoveRoleFromUser(ctx context.Context, user *models.User, role *models.Role) error {
	if err := database.DB.WithContext(ctx).Model(user).Association("Roles").Delete(role); err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}

	// Invalidate related caches
	s.invalidateUserRoleCaches(ctx, user.ID, role.ID)

	return nil
}

func (s *userRoleService) ListUserRoles(ctx context.Context, user *models.User) ([]models.Role, error) {
	cacheKey := s.getUserRoleCacheKey(user.ID)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var roles []models.Role
		if err := json.Unmarshal([]byte(cached), &roles); err == nil {
			return roles, nil
		}
	}

	var u models.User
	err := database.DB.WithContext(ctx).Preload("Roles").First(&u, user.ID).Error
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Cache the result
	if data, err := json.Marshal(u.Roles); err == nil {
		s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute)
	}

	return u.Roles, err
}

func (s *userRoleService) GetUsersByRole(ctx context.Context, role *models.Role) ([]models.User, error) {
	cacheKey := s.getRoleUserCacheKey(role.ID)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var users []models.User
		if err := json.Unmarshal([]byte(cached), &users); err == nil {
			return users, nil
		}
	}

	var users []models.User
	err := database.DB.WithContext(ctx).
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", role.ID).
		Find(&users).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get users by role: %w", err)
	}

	// Cache the result
	if data, err := json.Marshal(users); err == nil {
		s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute)
	}

	return users, err
}

func (s *userRoleService) HasRole(ctx context.Context, user *models.User, roleName string) (bool, error) {
	roles, err := s.ListUserRoles(ctx, user)
	if err != nil {
		return false, fmt.Errorf("failed to get user roles: %w", err)
	}

	for _, role := range roles {
		if role.Name == roleName {
			return true, nil
		}
	}

	return false, nil
}

func (s *userRoleService) InvalidateUserRoleCache(ctx context.Context, userID uint) error {
	s.invalidateUserRoleCaches(ctx, userID, 0)
	return nil
}

// Helper methods
func (s *userRoleService) invalidateUserRoleCaches(ctx context.Context, userID, roleID uint) {
	cacheKeys := []string{
		s.getUserRoleCacheKey(userID),
	}

	if roleID != 0 {
		cacheKeys = append(cacheKeys, s.getRoleUserCacheKey(roleID))
	}

	for _, key := range cacheKeys {
		s.Redis.Delete(ctx, key)
	}
}

var UserRoleSvc UserRoleService = &userRoleService{}
