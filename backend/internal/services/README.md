# Service Layer Optimizations with Redis Caching

This document outlines the comprehensive optimizations made to all services in the WordPress Go Next backend, including Redis caching integration, context support, and improved error handling.

## Overview

All services have been optimized with the following improvements:

- **Redis Caching**: Intelligent caching with TTL and cache invalidation
- **Context Support**: All methods now support context for better request handling
- **Error Handling**: Improved error messages with proper wrapping
- **Performance**: Reduced database queries through strategic caching
- **Scalability**: Better handling of concurrent requests

## Services Optimized

### 1. UserService (`user_service.go`)

**Key Features:**
- User CRUD operations with caching
- User search and filtering
- User statistics and counts
- Email and username lookups
- Active user tracking

**Cache Keys:**
- `user:{id}` - Individual user cache
- `user:email:{email}` - User by email
- `user:username:{username}` - User by username
- `user:active` - Active users list
- `user:count` - Total user count
- `user:search:{query}` - Search results

**Methods:**
```go
GetUserByID(ctx context.Context, id uint) (*models.User, error)
GetUserByEmail(ctx context.Context, email string) (*models.User, error)
GetUserByUsername(ctx context.Context, username string) (*models.User, error)
CreateUser(ctx context.Context, user *models.User) error
UpdateUser(ctx context.Context, user *models.User) error
DeleteUser(ctx context.Context, id uint) error
GetActiveUsers(ctx context.Context) ([]models.User, error)
GetUserCount(ctx context.Context) (int64, error)
SearchUsers(ctx context.Context, query string, limit, offset int) ([]models.User, int64, error)
```

### 2. PostService (`post_service.go`)

**Key Features:**
- Post CRUD operations with caching
- Published posts management
- Post search and filtering
- Post statistics
- Category-based post retrieval

**Cache Keys:**
- `post:{id}` - Individual post cache
- `post:slug:{slug}` - Post by slug
- `post:published` - Published posts list
- `post:count` - Total post count
- `post:published:count` - Published post count
- `post:category:{categoryID}` - Posts by category
- `post:search:{query}` - Search results

**Methods:**
```go
GetPostByID(ctx context.Context, id uint) (*models.Post, error)
GetPostBySlug(ctx context.Context, slug string) (*models.Post, error)
CreatePost(ctx context.Context, post *models.Post) error
UpdatePost(ctx context.Context, post *models.Post) error
DeletePost(ctx context.Context, id uint) error
GetPublishedPosts(ctx context.Context) ([]models.Post, error)
GetPostCount(ctx context.Context) (int64, error)
GetPublishedPostCount(ctx context.Context) (int64, error)
GetPostsByCategory(ctx context.Context, categoryID uint, limit, offset int) ([]models.Post, int64, error)
SearchPosts(ctx context.Context, query string, limit, offset int) ([]models.Post, int64, error)
```

### 3. CategoryService (`category_service.go`)

**Key Features:**
- Category CRUD operations with caching
- Hierarchical category management
- Category statistics
- Nested category operations

**Cache Keys:**
- `category:{id}` - Individual category cache
- `category:slug:{slug}` - Category by slug
- `category:all` - All categories list
- `category:count` - Total category count
- `category:children:{parentID}` - Children categories
- `category:descendants:{parentID}` - Descendant categories

**Methods:**
```go
GetCategoryByID(ctx context.Context, id uint) (*models.Category, error)
GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error)
CreateCategory(ctx context.Context, category *models.Category) error
UpdateCategory(ctx context.Context, category *models.Category) error
DeleteCategory(ctx context.Context, id uint) error
GetAllCategories(ctx context.Context) ([]models.Category, error)
GetCategoryCount(ctx context.Context) (int64, error)
GetChildrenCategories(ctx context.Context, parentID uint) ([]models.Category, error)
GetDescendantCategories(ctx context.Context, parentID uint) ([]models.Category, error)
```

### 4. MediaService (`media_service.go`)

**Key Features:**
- Media file upload and management
- Media association with other entities
- Media retrieval by various criteria
- File storage integration

**Cache Keys:**
- `media:{id}` - Individual media cache
- `media:uuid:{uuid}` - Media by UUID
- `media:mediable:{mediableID}:{mediableType}` - Media by mediable
- `media:user:{userID}:{limit}:{offset}` - User media with pagination
- `media:all:{limit}:{offset}` - All media with pagination

**Methods:**
```go
UploadAndSaveMedia(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, createdBy *int64) (*models.Media, error)
AssociateMedia(ctx context.Context, mediaID, mediableID uint, mediableType, group string) error
GetMediaByID(ctx context.Context, id uint) (*models.Media, error)
GetMediaByUUID(ctx context.Context, uuid string) (*models.Media, error)
GetMediaByMediable(ctx context.Context, mediableID uint, mediableType string) ([]models.Media, error)
UpdateMedia(ctx context.Context, media *models.Media) error
DeleteMedia(ctx context.Context, id uint) error
GetAllMedia(ctx context.Context, limit, offset int) ([]models.Media, int64, error)
GetMediaByUser(ctx context.Context, userID uint, limit, offset int) ([]models.Media, int64, error)
```

### 5. CommentService (`comment_service.go`)

**Key Features:**
- Comment CRUD operations with caching
- Hierarchical comment system (nested comments)
- Comment moderation
- Post-based comment retrieval

**Cache Keys:**
- `comment:{id}` - Individual comment cache
- `comment:post:{postID}` - Comments by post
- `comment:siblings:{id}` - Sibling comments
- `comment:parent:{id}` - Parent comment
- `comment:children:{id}` - Children comments
- `comment:descendants:{id}` - Descendant comments
- `comment:user:{userID}:{limit}:{offset}` - User comments with pagination
- `comment:all:{limit}:{offset}` - All comments with pagination

**Methods:**
```go
GetCommentsByPost(ctx context.Context, postID string) ([]models.Comment, error)
GetCommentByID(ctx context.Context, id string) (*models.Comment, error)
CreateComment(ctx context.Context, comment *models.Comment) error
UpdateComment(ctx context.Context, comment *models.Comment) error
DeleteComment(ctx context.Context, id string) error
CreateNested(ctx context.Context, comment *models.Comment, parentID *int64) error
GetSiblingComments(ctx context.Context, id uint) ([]models.Comment, error)
GetParentComment(ctx context.Context, id uint) (*models.Comment, error)
GetDescendantComments(ctx context.Context, id uint) ([]models.Comment, error)
GetChildrenComments(ctx context.Context, id uint) ([]models.Comment, error)
GetCommentsByUser(ctx context.Context, userID uint, limit, offset int) ([]models.Comment, int64, error)
GetAllComments(ctx context.Context, limit, offset int) ([]models.Comment, int64, error)
```

### 6. RoleService (`role_service.go`)

**Key Features:**
- Role CRUD operations with caching
- Role-based access control
- User role management

**Cache Keys:**
- `role:{id}` - Individual role cache
- `role:name:{name}` - Role by name
- `role:all:list` - All roles list
- `role:user:{userID}` - User roles

**Methods:**
```go
GetAllRoles(ctx context.Context) ([]models.Role, error)
GetRoleByID(ctx context.Context, id string) (*models.Role, error)
GetRoleByName(ctx context.Context, name string) (*models.Role, error)
CreateRole(ctx context.Context, role *models.Role) error
UpdateRole(ctx context.Context, role *models.Role) error
DeleteRole(ctx context.Context, id string) error
GetRolesByUser(ctx context.Context, userID uint) ([]models.Role, error)
```

### 7. UserRoleService (`user_role_service.go`)

**Key Features:**
- User-role association management
- Role assignment and removal
- User role queries

**Cache Keys:**
- `user_role:{userID}` - User roles
- `role_user:{roleID}` - Users by role

**Methods:**
```go
AssignRoleToUser(ctx context.Context, user *models.User, role *models.Role) error
RemoveRoleFromUser(ctx context.Context, user *models.User, role *models.Role) error
ListUserRoles(ctx context.Context, user *models.User) ([]models.Role, error)
GetUsersByRole(ctx context.Context, role *models.Role) ([]models.User, error)
HasRole(ctx context.Context, user *models.User, roleName string) (bool, error)
```

### 8. AuthService (`auth_service.go`)

**Key Features:**
- Authentication token management
- Password hashing and verification
- Email/phone verification
- User authentication state

**Cache Keys:**
- `verification_token:{token}` - Verification tokens
- `user_auth:{userID}` - User authentication state

**Methods:**
```go
GenerateToken() string
CreateVerificationToken(ctx context.Context, userID uint, tokenType string) (string, error)
ValidateVerificationToken(ctx context.Context, token, tokenType string) (*models.VerificationToken, error)
MarkEmailVerified(ctx context.Context, user *models.User) error
MarkPhoneVerified(ctx context.Context, user *models.User) error
HashPassword(password string) (string, error)
VerifyPassword(password, hash string) (bool, error)
ResetUserPassword(ctx context.Context, user *models.User, newPassword string) error
```

### 9. TagService (`tags.go`)

**Key Features:**
- Flexible entity tagging system
- Tag categorization (general, category, feature, system)
- Color-coded tags for UI
- Search and filtering capabilities
- Tag statistics and analytics
- Many-to-many entity relationships

**Cache Keys:**
- `tag:{id}` - Individual tag cache
- `tag:slug:{slug}` - Tag by slug
- `tag:name:{name}` - Tag by name
- `tag:all:{type}` - All tags by type
- `tag:active:list` - Active tags list
- `tag:entity:{entityID}:{entityType}` - Tags by entity
- `tag:search:{query}` - Search results
- `tag:count:total` - Total tag count

**Methods:**
```go
CreateTag(ctx context.Context, tag *Tag) error
GetTagByID(ctx context.Context, id uint) (*Tag, error)
GetTagBySlug(ctx context.Context, slug string) (*Tag, error)
GetTagByName(ctx context.Context, name string) (*Tag, error)
UpdateTag(ctx context.Context, tag *Tag) error
DeleteTag(ctx context.Context, id uint) error
GetAllTags(ctx context.Context, tagType string) ([]Tag, error)
GetActiveTags(ctx context.Context) ([]Tag, error)
GetTagsByEntity(ctx context.Context, entityID uint, entityType string) ([]Tag, error)
AddTagToEntity(ctx context.Context, tagID, entityID uint, entityType string) error
RemoveTagFromEntity(ctx context.Context, tagID, entityID uint, entityType string) error
GetEntitiesByTag(ctx context.Context, tagID uint, entityType string, limit, offset int) ([]map[string]interface{}, int64, error)
SearchTags(ctx context.Context, query string, limit, offset int) ([]Tag, int64, error)
GetTagCount(ctx context.Context) (int64, error)
InvalidateTagCache(ctx context.Context, tagID uint) error
```

### 10. ServiceLogger (`logger.go`)

**Key Features:**
- Structured logging with JSON output
- Multiple log levels (Debug, Info, Warning, Error, Fatal)
- Performance metrics tracking
- Cache operation logging
- Database operation monitoring
- Security event logging
- Audit trail support
- Context-aware logging with trace IDs

**Log Levels:**
```go
LogLevelDebug   // Detailed debugging information
LogLevelInfo    // General information messages
LogLevelWarning // Warning messages
LogLevelError   // Error messages
LogLevelFatal   // Fatal errors that cause application exit
```

**Specialized Logging Methods:**
```go
// Performance logging
logger.Performance(ctx, method, duration, cacheHit, dbOps, metadata)

// Cache operation logging
logger.Cache(ctx, method, operation, cacheKey, success, metadata)

// Database operation logging
logger.Database(ctx, method, operation, table, duration, rowsAffected, err, metadata)

// Security event logging
logger.Security(ctx, method, event, userID, success, metadata)

// Audit trail logging
logger.Audit(ctx, method, action, userID, entityID, entityType, metadata)
```

**Configuration:**
```go
// Initialize logger with specific level and service name
logger := NewServiceLogger(LogLevelInfo, "UserService")

// Initialize global logger
InitializeLogger(LogLevelInfo, "GlobalServices")

// Get logger for specific service
logger := GetLogger("PostService")
```

## Service Manager

The `ServiceManager` provides centralized management of all services:

### Initialization
```go
// Initialize all services with Redis, storage, and logging
manager := NewServiceManager(redisService, storageService, LogLevelInfo)
InitializeServices(redisService, storageService, LogLevelInfo)
```

### Available Services
```go
manager.UserService      // User management
manager.PostService      // Post management
manager.CategoryService  // Category management
manager.CommentService   // Comment management
manager.MediaService     // Media management
manager.RoleService      // Role management
manager.UserRoleService  // User-role management
manager.AuthService      // Authentication
manager.TagService       // Tag management
manager.Logger           // Structured logging
```

### Cache Management
```go
// Warm up frequently accessed data
err := manager.WarmCache(ctx)

// Clear all caches
err := manager.ClearAllCaches(ctx)

// Get cache statistics
stats, err := manager.GetCacheStats(ctx)
```

### Health Checks
```go
// Perform health check on all services
health, err := manager.HealthCheck(ctx)
```

## Tagging System

### Tag Types
- **General**: Default tag type for general categorization
- **Category**: Tags used for content categorization
- **Feature**: Tags for feature flags and system features
- **System**: Internal system tags

### Entity Tagging
```go
// Add tag to post
err := tagService.AddTagToEntity(ctx, tagID, postID, "post")

// Get tags for a post
tags, err := tagService.GetTagsByEntity(ctx, postID, "post")

// Get all posts with a specific tag
posts, total, err := tagService.GetEntitiesByTag(ctx, tagID, "post", 10, 0)

// Search tags
tags, total, err := tagService.SearchTags(ctx, "golang", 10, 0)
```

### Tag Management
```go
// Create a new tag
tag := &Tag{
    Name:        "Golang",
    Slug:        "golang",
    Description: "Go programming language",
    Color:       "#00ADD8",
    Type:        "general",
    IsActive:    true,
}
err := tagService.CreateTag(ctx, tag)

// Get active tags
activeTags, err := tagService.GetActiveTags(ctx)

// Get tags by type
featureTags, err := tagService.GetAllTags(ctx, "feature")
```

## Logging System

### Log Configuration
```go
// Initialize with specific log level
logger := NewServiceLogger(LogLevelDebug, "UserService")

// Log to files and stdout
// - logs/services.log (all logs)
// - logs/services-error.log (error logs only)
```

### Performance Monitoring
```go
start := time.Now()
// ... perform operation
duration := time.Since(start)

logger.Performance(ctx, "GetUserByID", duration, cacheHit, dbOps, map[string]interface{}{
    "user_id": userID,
    "cache_key": cacheKey,
})
```

### Security and Audit Logging
```go
// Log security events
logger.Security(ctx, "Login", "user_login", userID, success, map[string]interface{}{
    "ip_address": clientIP,
    "user_agent": userAgent,
})

// Log audit trail
logger.Audit(ctx, "UpdatePost", "post_updated", userID, postID, "post", map[string]interface{}{
    "changes": changes,
    "old_values": oldValues,
})
```

## Cache Strategy

### TTL (Time To Live)
- **Individual entities**: 30 minutes
- **Lists and pagination**: 15 minutes
- **Statistics and counts**: 10 minutes
- **Verification tokens**: Based on expiration time
- **Tags**: 30 minutes (individual), 15 minutes (lists)

### Cache Invalidation
- **Write operations**: Invalidate related caches
- **Delete operations**: Remove entity and related caches
- **Update operations**: Update entity cache and invalidate related caches
- **Pattern-based invalidation**: For pagination and search results

### Cache Keys
All cache keys follow a consistent naming convention:
- `{entity}:{id}` - Individual entity
- `{entity}:{field}:{value}` - Entity by specific field
- `{entity}:{relation}:{id}` - Related entities
- `{entity}:all` - All entities
- `{entity}:count` - Entity count

## Performance Benefits

1. **Reduced Database Load**: Frequently accessed data is cached
2. **Faster Response Times**: Cache hits provide instant responses
3. **Better Scalability**: Reduced database connections
4. **Improved User Experience**: Faster page loads and API responses

## Error Handling

All services now include:
- Context-aware error handling
- Proper error wrapping with `fmt.Errorf`
- Graceful cache failure handling
- Detailed error messages for debugging

## Usage Examples

### Basic CRUD Operations
```go
// Get user with caching
user, err := userService.GetUserByID(ctx, userID)
if err != nil {
    return err
}

// Create post with cache invalidation
err = postService.CreatePost(ctx, post)
if err != nil {
    return err
}
```

### Search and Filtering
```go
// Search posts with pagination
posts, total, err := postService.SearchPosts(ctx, "golang", 10, 0)
if err != nil {
    return err
}
```

### Tagging Operations
```go
// Create and apply tags
tag := &Tag{
    Name:        "Featured",
    Slug:        "featured",
    Description: "Featured content",
    Color:       "#FFD700",
    Type:        "feature",
    IsActive:    true,
}
err := tagService.CreateTag(ctx, tag)
if err != nil {
    return err
}

// Add tag to post
err = tagService.AddTagToEntity(ctx, tag.ID, postID, "post")
if err != nil {
    return err
}

// Get all featured posts
featuredPosts, total, err := tagService.GetEntitiesByTag(ctx, tag.ID, "post", 10, 0)
if err != nil {
    return err
}
```

### Logging Examples
```go
// Performance logging
start := time.Now()
user, err := userService.GetUserByID(ctx, userID)
duration := time.Since(start)

logger.Performance(ctx, "GetUserByID", duration, cacheHit, 1, map[string]interface{}{
    "user_id": userID,
    "cache_hit": cacheHit,
})

// Security logging
logger.Security(ctx, "Login", "user_login", userID, success, map[string]interface{}{
    "ip_address": clientIP,
    "user_agent": userAgent,
})

// Audit logging
logger.Audit(ctx, "UpdatePost", "post_updated", userID, postID, "post", map[string]interface{}{
    "title_changed": titleChanged,
    "content_changed": contentChanged,
})
```

### Cache Management
```go
// Invalidate specific cache
err = userService.InvalidateUserCache(ctx, userID)

// Get cache statistics
stats, err := serviceManager.GetCacheStats(ctx)
```

## Migration Notes

### Breaking Changes
- All service methods now require `context.Context` as the first parameter
- Service constructors now require Redis service injection
- Global service instances are managed through `ServiceManager`

### Backward Compatibility
- Legacy service instances are still available but deprecated
- New services provide enhanced functionality with caching
- Gradual migration path available

## Configuration

### Redis Configuration
```go
redisService := redis.NewRedisService(redisConfig)
storageService := storage.NewStorageService(storageConfig)
InitializeServices(redisService, storageService, LogLevelInfo)
```

### Logging Configuration
```go
// Set log level based on environment
logLevel := LogLevelInfo
if os.Getenv("ENV") == "development" {
    logLevel = LogLevelDebug
}

// Initialize services with logging
InitializeServices(redisService, storageService, logLevel)
```

### Environment Variables
```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

## Monitoring and Debugging

### Cache Statistics
```go
stats, err := serviceManager.GetCacheStats(ctx)
// Returns cache hit rates, key counts, and performance metrics
```

### Health Checks
```go
health, err := serviceManager.HealthCheck(ctx)
// Returns service status, Redis connectivity, and database health
```

### Logging
All services include comprehensive logging for:
- Cache hits and misses
- Database operations
- Error conditions
- Performance metrics

## Best Practices

1. **Always use context**: Pass context through all service calls
2. **Handle cache failures gracefully**: Don't let cache errors break functionality
3. **Monitor cache performance**: Use cache statistics to optimize TTL values
4. **Invalidate caches appropriately**: Ensure data consistency
5. **Use appropriate TTL values**: Balance performance with data freshness

## Future Enhancements

1. **Distributed caching**: Support for Redis cluster
2. **Cache warming strategies**: Intelligent pre-loading of data
3. **Cache compression**: Reduce memory usage
4. **Advanced invalidation**: More sophisticated cache invalidation patterns
5. **Metrics integration**: Prometheus/Grafana integration for monitoring 