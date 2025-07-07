package services

import (
	"context"
	"fmt"
	"time"

	"wordpress-go-next/backend/internal/http/responses"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"
	"wordpress-go-next/backend/pkg/redis"
)

type UserService interface {
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
	GetUsersWithPagination(ctx context.Context, page, perPage int) (*responses.PaginationResponse, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUserProfile(ctx context.Context, user *models.User, username, email, phone string, emailVerified, phoneVerified *time.Time) error
	DeleteUser(ctx context.Context, id string) error
	GetUserStats(ctx context.Context, userID string) (map[string]interface{}, error)
	GetActiveUsers(ctx context.Context) ([]models.User, error)
	GetUsersByRole(ctx context.Context, roleName string) ([]models.User, error)
	SearchUsers(ctx context.Context, query string) ([]models.User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
	GetUserCount(ctx context.Context) (int64, error)
}

type userService struct {
	*BaseService
}

func NewUserService(redisService *redis.RedisService) UserService {
	return &userService{
		BaseService: NewBaseService(redisService),
	}
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	cacheKey := s.GetCacheKey(redis.UserCachePrefix, id)

	err := s.GetByIDWithCacheAndPreload(ctx, id, &user, cacheKey, redis.DefaultTTL, "Roles")
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	cacheKey := fmt.Sprintf("%semail:%s", redis.UserCachePrefix, email)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &user); err == nil {
			return &user, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("Roles").Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		err := s.Redis.SetCache(ctx, cacheKey, &user, redis.DefaultTTL)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	cacheKey := fmt.Sprintf("%susername:%s", redis.UserCachePrefix, username)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &user); err == nil {
			return &user, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("Roles").Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		err := s.Redis.SetCache(ctx, cacheKey, &user, redis.DefaultTTL)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s *userService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	cacheKey := s.GetListCacheKey(redis.UserCachePrefix)

	err := s.GetAllWithCacheAndPreload(ctx, &users, cacheKey, redis.DefaultTTL, "Roles")
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *userService) GetUsersWithPagination(ctx context.Context, page, perPage int) (*responses.PaginationResponse, error) {
	var users []models.User
	cacheKey := s.GetListCacheKey(redis.UserCachePrefix)

	params := PaginationParams{
		Page:    page,
		PerPage: perPage,
	}

	result, err := s.PaginateWithCache(ctx, &models.User{}, params, &users, cacheKey, redis.DefaultTTL)
	if err != nil {
		return nil, err
	}

	// Preload roles for each user
	for i := range users {
		if err := database.DB.Preload("Roles").First(&users[i], users[i].ID).Error; err != nil {
			return nil, err
		}
	}

	result.Data = users
	return result, nil
}

func (s *userService) CreateUser(ctx context.Context, user *models.User) error {
	return s.CreateWithCache(ctx, user, redis.UserCachePrefix)
}

func (s *userService) UpdateUserProfile(ctx context.Context, user *models.User, username, email, phone string, emailVerified, phoneVerified *time.Time) error {
	user.Username = username
	user.Email = email
	if phone != "" {
		user.Phone = &phone
	}
	user.EmailVerified = emailVerified
	user.PhoneVerified = phoneVerified

	return s.UpdateWithCache(ctx, user, redis.UserCachePrefix)
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	user := &models.User{}
	if err := database.DB.First(user, id).Error; err != nil {
		return err
	}

	return s.DeleteWithCache(ctx, user, redis.UserCachePrefix)
}

func (s *userService) GetUserStats(ctx context.Context, userID string) (map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("%sstats:%s", redis.UserCachePrefix, userID)

	// Try cache first
	if s.Redis != nil {
		var cachedStats map[string]interface{}
		if err := s.Redis.GetCache(ctx, cacheKey, &cachedStats); err == nil {
			return cachedStats, nil
		}
	}

	// Calculate stats from database
	var postCount int64
	var commentCount int64
	var user models.User

	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}

	if err := database.DB.Model(&models.Post{}).Where("created_by = ?", userID).Count(&postCount).Error; err != nil {
		return nil, err
	}

	if err := database.DB.Model(&models.Comment{}).Where("user_id = ?", userID).Count(&commentCount).Error; err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"user_id":        userID,
		"post_count":     postCount,
		"comment_count":  commentCount,
		"is_active":      user.IsActive,
		"email_verified": user.IsEmailVerified(),
		"phone_verified": user.IsPhoneVerified(),
		"created_at":     user.CreatedAt,
		"last_login":     user.LastLoginAt,
	}

	// Cache the stats
	if s.Redis != nil {
		err := s.Redis.SetCache(ctx, cacheKey, stats, redis.ShortTTL)
		if err != nil {
			return nil, err
		}
	}

	return stats, nil
}

func (s *userService) GetActiveUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	cacheKey := fmt.Sprintf("%sactive", redis.UserCachePrefix)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &users); err == nil {
			return users, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("Roles").Where("is_active = ?", true).Find(&users).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		err := s.Redis.SetCache(ctx, cacheKey, users, redis.ShortTTL)
		if err != nil {
			return nil, err
		}
	}

	return users, nil
}

func (s *userService) GetUsersByRole(ctx context.Context, roleName string) ([]models.User, error) {
	var users []models.User
	cacheKey := fmt.Sprintf("%srole:%s", redis.UserCachePrefix, roleName)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &users); err == nil {
			return users, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("Roles").Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("roles.name = ?", roleName).Find(&users).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		err := s.Redis.SetCache(ctx, cacheKey, users, redis.DefaultTTL)
		if err != nil {
			return nil, err
		}
	}

	return users, nil
}

func (s *userService) SearchUsers(ctx context.Context, query string) ([]models.User, error) {
	var users []models.User
	cacheKey := s.GetSearchCacheKey(redis.UserCachePrefix, query)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &users); err == nil {
			return users, nil
		}
	}

	// Search in database
	if err := database.DB.Preload("Roles").
		Where("username LIKE ? OR email LIKE ? OR phone LIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%").
		Find(&users).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		err := s.Redis.SetCache(ctx, cacheKey, users, redis.ShortTTL)
		if err != nil {
			return nil, err
		}
	}

	return users, nil
}

func (s *userService) UpdateLastLogin(ctx context.Context, userID string) error {
	user := &models.User{}
	if err := database.DB.First(user, userID).Error; err != nil {
		return err
	}

	now := time.Now()
	user.LastLoginAt = &now

	// Update database
	if err := database.DB.Save(user).Error; err != nil {
		return err
	}

	// Invalidate user cache
	if s.Redis != nil {
		err := s.Redis.InvalidateUserCache(ctx, userID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *userService) GetUserCount(ctx context.Context) (int64, error) {
	cacheKey := fmt.Sprintf("%scount", redis.UserCachePrefix)

	// Try cache first
	if s.Redis != nil {
		var count int64
		if err := s.Redis.GetCache(ctx, cacheKey, &count); err == nil {
			return count, nil
		}
	}

	// Get from database
	var count int64
	if err := database.DB.Model(&models.User{}).Count(&count).Error; err != nil {
		return 0, err
	}

	// Cache the result
	if s.Redis != nil {
		err := s.Redis.SetCache(ctx, cacheKey, count, redis.LongTTL)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

// Legacy methods for backward compatibility
func (s *userService) GetUserByIDLegacy(id string) (*models.User, error) {
	return s.GetUserByID(context.Background(), id)
}

func (s *userService) UpdateUserProfileLegacy(user *models.User, username, email, phone string, emailVerified, phoneVerified *time.Time) error {
	return s.UpdateUserProfile(context.Background(), user, username, email, phone, emailVerified, phoneVerified)
}

// Global service instance
var UserSvc UserService
