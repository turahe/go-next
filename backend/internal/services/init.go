package services

import (
	"context"
	"log"
	"time"

	"go-next/pkg/database"
	"go-next/pkg/logger"
	"go-next/pkg/redis"
	"go-next/pkg/storage"
)

// ServiceManager manages all service instances
type ServiceManager struct {
	RedisService    *redis.RedisService
	StorageService  storage.StorageService
	Logger          *logger.ServiceLogger
	UserService     UserService
	PostService     PostService
	CategoryService CategoryService
	CommentService  CommentService
	MediaService    MediaService
	RoleService     RoleService
	UserRoleService UserRoleService
	AuthService     AuthService
	TagService      TagService
	MenuService     MenuService
}

// NewServiceManager creates a new service manager with all services initialized
func NewServiceManager(redisService *redis.RedisService, storageService storage.StorageService, logLevel logger.LogLevel) *ServiceManager {
	// Initialize logger
	logger := logger.NewServiceLogger(logLevel, "ServiceManager")

	manager := &ServiceManager{
		RedisService:   redisService,
		StorageService: storageService,
		Logger:         logger,
	}

	// Initialize all services
	manager.UserService = NewUserService()
	manager.PostService = NewPostService(redisService)
	manager.CategoryService = NewCategoryService(redisService)
	manager.CommentService = NewCommentService(redisService)
	manager.MediaService = NewMediaService()
	manager.RoleService = NewRoleService()
	manager.UserRoleService = NewUserRoleService(redisService)
	manager.AuthService = NewAuthService()
	manager.TagService = NewTagService(redisService)
	manager.MenuService = NewMenuService(redisService)

	// Log service initialization
	logger.Info("NewServiceManager: All services initialized successfully", "services_count", 9, "redis_available", redisService != nil, "storage_available", storageService != nil)

	return manager
}

// InitializeServices initializes all global service instances
func InitializeServices(redisService *redis.RedisService, storageService storage.StorageService, logLevel logger.LogLevel) {
	manager := NewServiceManager(redisService, storageService, logLevel)

	// Set global service instances
	UserSvc = manager.UserService
	PostSvc = manager.PostService
	CategorySvc = manager.CategoryService
	CommentSvc = manager.CommentService
	RoleSvc = manager.RoleService
	UserRoleSvc = manager.UserRoleService
	AuthSvc = manager.AuthService
	TagSvc = manager.TagService
	MediaSvc = manager.MediaService
	MenuSvc = manager.MenuService

	// Set global service manager
	ServiceMgr = manager

	// Initialize global logger
	logger.InitializeLogger(logLevel, "GlobalServices")

	// Warm up caches
	go func() {
		ctx := context.Background()
		if err := manager.WarmCache(ctx); err != nil {
			log.Printf("Warning: failed to warm cache: %v", err)
		}
	}()

	log.Println("All services initialized with Redis caching and logging")
}

// WarmCache warms up the cache with frequently accessed data
func (sm *ServiceManager) WarmCache(ctx context.Context) error {
	log.Println("Warming up cache...")

	// Warm user cache
	if err := sm.warmUserCache(ctx); err != nil {
		log.Printf("Error warming user cache: %v", err)
	}

	// Warm post cache
	if err := sm.warmPostCache(ctx); err != nil {
		log.Printf("Error warming post cache: %v", err)
	}

	// Warm category cache
	if err := sm.warmCategoryCache(ctx); err != nil {
		log.Printf("Error warming category cache: %v", err)
	}

	// Warm role cache
	if err := sm.warmRoleCache(ctx); err != nil {
		log.Printf("Error warming role cache: %v", err)
	}

	log.Println("Cache warming completed")
	return nil
}

// warmUserCache warms up user-related caches
func (sm *ServiceManager) warmUserCache(ctx context.Context) error {
	// TODO: Implement user cache warming when methods are available
	// Warm active users cache
	// if _, err := sm.UserService.GetActiveUsers(ctx); err != nil {
	// 	return err
	// }

	// Warm user count cache
	// if _, err := sm.UserService.GetUserCount(ctx); err != nil {
	// 	return err
	// }

	return nil
}

// warmPostCache warms up post-related caches
func (sm *ServiceManager) warmPostCache(ctx context.Context) error {
	// Warm published posts cache
	if _, err := sm.PostService.GetPublishedPosts(ctx); err != nil {
		return err
	}

	// Warm post count cache
	if _, err := sm.PostService.GetPostCount(ctx); err != nil {
		return err
	}

	// Warm published post count cache
	if _, err := sm.PostService.GetPublishedPostCount(ctx); err != nil {
		return err
	}

	return nil
}

// warmCategoryCache warms up category-related caches
func (sm *ServiceManager) warmCategoryCache(ctx context.Context) error {
	// Warm all categories cache
	if _, err := sm.CategoryService.GetAllCategoriesWithContext(ctx); err != nil {
		return err
	}

	// Warm category count cache
	if _, err := sm.CategoryService.GetCategoryCount(ctx); err != nil {
		return err
	}

	return nil
}

// warmRoleCache warms up role-related caches
func (sm *ServiceManager) warmRoleCache(ctx context.Context) error {
	// TODO: Implement role cache warming when methods are available
	// For now, we'll skip role cache warming since the current RoleService
	// doesn't have GetAllRolesWithContext method
	// if _, err := sm.RoleService.GetAllRolesWithContext(ctx); err != nil {
	// 	return err
	// }

	return nil
}

// ClearAllCaches clears all caches
func (sm *ServiceManager) ClearAllCaches(ctx context.Context) error {
	if sm.RedisService == nil {
		return nil
	}

	patterns := []string{
		"user:*",
		"post:*",
		"category:*",
		"comment:*",
		"media:*",
		"role:*",
		"user_role:*",
		"role_user:*",
		"verification_token:*",
		"user_auth:*",
		"search:*",
		"stats:*",
	}

	for _, pattern := range patterns {
		if err := sm.RedisService.DeletePattern(ctx, pattern); err != nil {
			log.Printf("Error clearing cache pattern %s: %v", pattern, err)
		}
	}

	log.Println("All caches cleared")
	return nil
}

// GetCacheStats returns cache statistics
func (sm *ServiceManager) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	if sm.RedisService == nil {
		return map[string]interface{}{"error": "Redis not available"}, nil
	}

	stats := make(map[string]interface{})

	// Get cache keys count for each prefix
	prefixes := []string{
		"user:", "post:", "category:", "comment:", "media:", "role:", "user_role:", "role_user:",
		"verification_token:", "user_auth:", "search:", "stats:",
	}

	for _, prefix := range prefixes {
		keys, err := sm.RedisService.GetKeysByPattern(ctx, prefix+"*")
		if err != nil {
			log.Printf("Error getting keys for pattern %s: %v", prefix, err)
			continue
		}
		stats[prefix] = len(keys)
	}

	// Add Redis info
	stats["timestamp"] = time.Now()
	stats["redis_available"] = true

	return stats, nil
}

// HealthCheck performs a health check on all services
func (sm *ServiceManager) HealthCheck(ctx context.Context) (map[string]interface{}, error) {
	health := make(map[string]interface{})

	// Check Redis connection
	if sm.RedisService != nil {
		if err := sm.RedisService.Ping(ctx); err != nil {
			health["redis"] = map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
			}
		} else {
			health["redis"] = map[string]interface{}{
				"status": "healthy",
			}
		}
	} else {
		health["redis"] = map[string]interface{}{
			"status":  "unavailable",
			"message": "Redis service not initialized",
		}
	}

	// Check database connection
	if err := database.DB.Raw("SELECT 1").Error; err != nil {
		health["database"] = map[string]interface{}{
			"status":  "error",
			"message": err.Error(),
		}
	} else {
		health["database"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Check services
	health["services"] = map[string]interface{}{
		"user_service":      sm.UserService != nil,
		"post_service":      sm.PostService != nil,
		"category_service":  sm.CategoryService != nil,
		"comment_service":   sm.CommentService != nil,
		"media_service":     sm.MediaService != nil,
		"role_service":      sm.RoleService != nil,
		"user_role_service": sm.UserRoleService != nil,
		"auth_service":      sm.AuthService != nil,
		"menu_service":      sm.MenuService != nil,
	}

	health["timestamp"] = time.Now()

	return health, nil
}

// Global service manager instance
var (
	UserSvc    UserService
	AuthSvc    AuthService
	MediaSvc   MediaService
	RoleSvc    RoleService
	MenuSvc    MenuService
	ServiceMgr *ServiceManager
)

// Initialize global services
func init() {
	// This will be called when the package is imported
	// The actual initialization should be done in the main application
	// with proper Redis configuration
}
