# Redis Token Caching Implementation

This document describes the implementation of Redis caching for token models in the Go Next backend application.

## Overview

The token caching system provides Redis-based caching for three main token types:
- **Token** (API tokens)
- **VerificationToken** (email/phone verification, password reset)
- **RefreshToken** (authentication refresh tokens)
- **JWTKey** (JWT signing keys)

## Architecture

### 1. Cache Service (`internal/services/cache_service.go`)

The core caching service provides a unified interface for Redis operations:

```go
type CacheService interface {
    Get(key string, dest interface{}) error
    Set(key string, value interface{}, expiration time.Duration) error
    Delete(key string) error
    DeletePattern(pattern string) error
    Exists(key string) bool
    Increment(key string) (int64, error)
    SetNX(key string, value interface{}, expiration time.Duration) (bool, error)
}
```

**Features:**
- JSON serialization/deserialization for complex objects
- Automatic expiration handling
- Pattern-based cache invalidation
- Null safety with Redis client checks

### 2. Token Cache Service (`internal/services/token_cache_service.go`)

Specialized caching service for token models with the following capabilities:

#### Token Caching
- `GetTokenByID(id uuid.UUID)` - Retrieve token by ID with cache-first strategy
- `GetTokenByValue(token string)` - Retrieve token by value (for validation)
- `GetTokensByUser(userID uuid.UUID)` - Get all tokens for a user
- `CacheToken(token *models.Token)` - Cache individual token
- `InvalidateToken(id uuid.UUID)` - Remove specific token from cache
- `InvalidateUserTokens(userID uuid.UUID)` - Clear all user tokens

#### JWT Key Caching
- `GetJWTKeyByID(id uuid.UUID)` - Retrieve JWT key by ID
- `GetJWTKeyByKeyID(keyID string)` - Retrieve JWT key by key ID
- `GetActiveJWTKeys()` - Get all active JWT keys
- `CacheJWTKey(jwtKey *models.JWTKey)` - Cache JWT key
- `InvalidateJWTKey(id uuid.UUID)` - Remove specific JWT key
- `InvalidateAllJWTKeys()` - Clear all JWT key caches

#### Verification Token Caching
- `GetVerificationTokenByID(id uuid.UUID)` - Retrieve verification token by ID
- `GetVerificationTokenByValue(token string)` - Retrieve by token value
- `GetVerificationTokensByUser(userID, tokenType)` - Get user's tokens by type
- `CacheVerificationToken(token *models.VerificationToken)` - Cache verification token
- `InvalidateVerificationToken(id uuid.UUID)` - Remove specific token
- `InvalidateUserVerificationTokens(userID, tokenType)` - Clear user's tokens by type

#### Refresh Token Caching
- `GetRefreshTokenByID(id uuid.UUID)` - Retrieve refresh token by ID
- `GetRefreshTokenByValue(token string)` - Retrieve by token value
- `GetRefreshTokensByUser(userID uuid.UUID)` - Get all user's refresh tokens
- `CacheRefreshToken(token *models.RefreshToken)` - Cache refresh token
- `InvalidateRefreshToken(id uuid.UUID)` - Remove specific token
- `InvalidateUserRefreshTokens(userID uuid.UUID)` - Clear all user's refresh tokens

## Cache Keys

The system uses consistent key patterns for easy management:

```go
const (
    CacheKeyToken = "token:%s"                    // token:uuid
    CacheKeyTokens = "tokens:user:%s"             // tokens:user:uuid
    CacheKeyJWTKey = "jwt_key:%s"                 // jwt_key:uuid
    CacheKeyJWTKeys = "jwt_keys"                  // jwt_keys (list)
    CacheKeyVerificationToken = "verification_token:%s"           // verification_token:uuid
    CacheKeyVerificationTokens = "verification_tokens:user:%s:type:%s" // verification_tokens:user:uuid:type:email_verification
    CacheKeyRefreshToken = "refresh_token:%s"     // refresh_token:uuid
    CacheKeyRefreshTokens = "refresh_tokens:user:%s" // refresh_tokens:user:uuid
)
```

## Cache Durations

```go
const (
    CacheDurationToken = 1 * time.Hour    // Tokens expire in 1 hour
    CacheDurationJWTKey = 24 * time.Hour  // JWT keys cached for 24 hours
)
```

## Integration Points

### 1. Authentication Service (`internal/services/auth_service.go`)

Updated to use token caching:
- `CreateVerificationToken()` - Caches new verification tokens
- `MarkEmailVerified()` - Invalidates related cached tokens
- `MarkPhoneVerified()` - Invalidates related cached tokens
- `ResetUserPassword()` - Invalidates password reset tokens

### 2. Authentication Handlers (`internal/http/controllers/auth_handler.go`)

Enhanced with caching:
- `RefreshToken()` - Uses cached tokens for validation
- `RequestEmailVerification()` - Caches new verification tokens
- `VerifyEmail()` - Uses cached tokens and invalidates after use
- `RequestPhoneVerification()` - Caches new verification tokens
- `VerifyPhone()` - Uses cached tokens and invalidates after use
- `RequestPasswordReset()` - Caches new reset tokens
- `ResetPassword()` - Uses cached tokens and invalidates after use

### 3. JWT Middleware (`internal/http/middleware/jwt_middleware.go`)

Updated to use cached JWT keys:
- Retrieves active JWT keys from cache first
- Falls back to database on cache miss
- Improves JWT validation performance

### 4. Token Utilities (`pkg/utils/token.go`)

Enhanced `GenerateJWT()` function:
- Uses cached JWT keys for token generation
- Caches new JWT keys on first access
- Reduces database queries for JWT operations

## Cache Strategy

### Cache-First Approach
1. **Cache Hit**: Return data immediately from Redis
2. **Cache Miss**: Query database, cache result, then return
3. **Write-Through**: Update both database and cache on writes
4. **Cache Invalidation**: Remove related cache entries on updates

### Invalidation Patterns
- **Individual Invalidation**: Remove specific token by ID
- **User-Based Invalidation**: Clear all tokens for a specific user
- **Type-Based Invalidation**: Clear tokens of specific type for a user
- **Pattern-Based Invalidation**: Clear multiple related cache entries

## Performance Benefits

1. **Reduced Database Load**: Token lookups served from Redis
2. **Faster Authentication**: JWT validation uses cached keys
3. **Improved Token Validation**: Verification tokens cached for quick access
4. **Scalable Token Management**: User token lists cached efficiently

## Error Handling

The system gracefully handles Redis failures:
- Falls back to database queries when Redis is unavailable
- Continues operation without caching when Redis client is nil
- Provides clear error messages for debugging

## Configuration

The Redis client is initialized in `internal/startup.go`:
```go
func InitRedis() {
    cfg := config.GetConfig()
    RedisClient = redis.NewRedisService(cfg.Redis)
    services.GlobalRedisClient = RedisClient  // Set global cache client
}
```

## Usage Examples

### Creating and Caching a Verification Token
```go
token := models.VerificationToken{
    UserID: userID,
    Token: generateToken(),
    Type: models.EmailVerification,
    ExpiresAt: time.Now().Add(30 * time.Minute),
}

// Save to database
database.DB.Create(&token)

// Cache the token
TokenCacheSvc.CacheVerificationToken(&token)

// Invalidate user's verification tokens cache
TokenCacheSvc.InvalidateUserVerificationTokens(userID, models.EmailVerification)
```

### Retrieving a Cached Token
```go
// Try cache first, fallback to database
token, err := TokenCacheSvc.GetVerificationTokenByValue(tokenValue)
if err != nil {
    // Cache miss - token will be cached automatically
    return err
}
```

### Invalidating Cached Data
```go
// Remove specific token
TokenCacheSvc.InvalidateVerificationToken(tokenID)

// Clear all user's tokens of specific type
TokenCacheSvc.InvalidateUserVerificationTokens(userID, models.PasswordReset)
```

## Monitoring and Maintenance

### Cache Health Checks
- Monitor Redis connection status
- Track cache hit/miss ratios
- Monitor cache memory usage

### Cache Warming
- Pre-load frequently accessed JWT keys
- Cache active user tokens on login
- Warm verification token caches for high-traffic periods

### Cache Cleanup
- Automatic expiration based on token lifetimes
- Manual cleanup of expired tokens
- Pattern-based cleanup for user deletions

## Security Considerations

1. **Token Encryption**: Sensitive token data is not stored in plain text
2. **Access Control**: Cache keys are namespaced by user ID
3. **Expiration**: All cached tokens respect their original expiration times
4. **Invalidation**: Proper cache invalidation prevents stale data access

## Future Enhancements

1. **Distributed Caching**: Support for Redis cluster
2. **Cache Compression**: Compress large token objects
3. **Cache Analytics**: Detailed cache performance metrics
4. **Smart Preloading**: Predictive cache warming based on usage patterns
5. **Cache Persistence**: Backup cache data for disaster recovery 