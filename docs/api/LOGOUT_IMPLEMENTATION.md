# 🔐 Logout Implementation with Redis Token Blacklisting

This document describes the logout functionality implemented in the Go-Next Admin Panel, which securely invalidates refresh tokens using Redis blacklisting.

## 📋 Overview

The logout system provides secure session termination by blacklisting refresh tokens in Redis. This ensures that even if a refresh token hasn't expired, it can no longer be used to generate new access tokens once the user logs out.

## 🏗️ Architecture

### Components

1. **AuthService**: Handles logout logic and Redis blacklisting
2. **RedisService**: Manages token blacklist storage
3. **JWT Service**: Validates refresh tokens before blacklisting
4. **AuthHandler**: HTTP endpoint for logout requests
5. **Redis Blacklist**: Stores revoked refresh tokens with expiration

### Flow

```
Logout Request → Validate Refresh Token → Add to Redis Blacklist → Return Success
```

## 🔧 Implementation Details

### 1. Redis Blacklist Strategy

The system uses a dual-key approach for token blacklisting:

```go
// Primary blacklist key for token lookup
blacklistKey := fmt.Sprintf("refresh_token_blacklist:%s", refreshToken)

// User-specific blacklist for potential future use
userBlacklistKey := fmt.Sprintf("user_refresh_tokens:%s", claims.UserID.String())
```

**Features:**
- **Token-specific blacklisting**: Each refresh token gets its own blacklist entry
- **User-specific tracking**: Additional user-based blacklist for analytics
- **Automatic expiration**: Blacklist entries expire with the same TTL as refresh tokens
- **Graceful degradation**: System works even if Redis is unavailable

### 2. Logout Process

The logout process follows these steps:

```go
func (s *authService) Logout(refreshToken string) error {
    // 1. Validate the refresh token first
    claims, err := s.jwt.ValidateToken(refreshToken)
    if err != nil {
        return errors.New("invalid refresh token")
    }

    // 2. If Redis is available, store the token in a blacklist
    if s.redisService != nil {
        ctx := context.Background()
        
        // Create a blacklist key for the refresh token
        blacklistKey := fmt.Sprintf("refresh_token_blacklist:%s", refreshToken)
        
        // Store the token in blacklist with expiration
        expiration := time.Duration(7*24) * time.Hour // 7 days
        err := s.redisService.SetWithExpiration(ctx, blacklistKey, "revoked", expiration)
        
        // Also store user-specific blacklist
        userBlacklistKey := fmt.Sprintf("user_refresh_tokens:%s", claims.UserID.String())
        err = s.redisService.SetWithExpiration(ctx, userBlacklistKey, refreshToken, expiration)
    }

    return nil
}
```

### 3. Refresh Token Validation

The refresh token process now checks the blacklist:

```go
func (s *authService) RefreshToken(refreshToken string) (*dto.AuthDTO, error) {
    // Validate refresh token
    claims, err := s.jwt.ValidateToken(refreshToken)
    if err != nil {
        return nil, err
    }

    // Check if the refresh token is blacklisted in Redis
    if s.redisService != nil {
        ctx := context.Background()
        blacklistKey := fmt.Sprintf("refresh_token_blacklist:%s", refreshToken)
        
        // Check if token exists in blacklist
        exists, err := s.redisService.Exists(ctx, blacklistKey)
        if err != nil {
            logger.Errorf("Error checking refresh token blacklist: %v", err)
            // Continue with refresh if Redis check fails
        } else if exists > 0 {
            return nil, errors.New("refresh token has been revoked")
        }
    }

    // Continue with normal refresh process...
}
```

## 🚀 API Endpoints

### User Logout

**Endpoint:** `POST /api/v1/auth/logout`

**Request Body:**
```json
{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**
```json
{
    "message": "Logout successful"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid refresh token
- `400 Bad Request`: Invalid JSON format
- `422 Unprocessable Entity`: Validation errors

## 🔧 Configuration

### Environment Variables

```env
# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=redis_password
REDIS_DB=0

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here
```

### Redis Key Patterns

The system uses the following Redis key patterns:

1. **Token Blacklist**: `refresh_token_blacklist:{refresh_token}`
   - Value: `"revoked"`
   - TTL: 7 days (matches refresh token expiry)

2. **User Blacklist**: `user_refresh_tokens:{user_id}`
   - Value: `{refresh_token}`
   - TTL: 7 days
   - Purpose: Track user's revoked tokens

## 🚨 Security Considerations

### Token Revocation

- **Immediate Invalidation**: Tokens are blacklisted immediately upon logout
- **Expiration Handling**: Blacklist entries expire with the same TTL as refresh tokens
- **Redis Persistence**: Blacklist survives Redis restarts (if configured)
- **Graceful Degradation**: System works even if Redis is unavailable

### Security Features

1. **Token Validation**: Refresh tokens are validated before blacklisting
2. **User Tracking**: User-specific blacklist for potential analytics
3. **Error Handling**: Logout succeeds even if Redis operations fail
4. **Logging**: Comprehensive logging for security monitoring

### Best Practices

1. **Always Logout**: Users should explicitly logout to invalidate tokens
2. **Token Rotation**: Consider implementing token rotation for enhanced security
3. **Monitoring**: Monitor Redis blacklist size and performance
4. **Cleanup**: Implement periodic cleanup of expired blacklist entries

## 📊 Redis Operations

### Blacklist Management

```go
// Add token to blacklist
blacklistKey := fmt.Sprintf("refresh_token_blacklist:%s", refreshToken)
err := redisService.SetWithExpiration(ctx, blacklistKey, "revoked", 7*24*time.Hour)

// Check if token is blacklisted
exists, err := redisService.Exists(ctx, blacklistKey)
if exists > 0 {
    // Token is blacklisted
}

// Get user's blacklisted tokens
userKey := fmt.Sprintf("user_refresh_tokens:%s", userID)
tokens, err := redisService.Get(ctx, userKey)
```

### Performance Considerations

- **Fast Lookups**: Redis provides O(1) lookup time for blacklist checks
- **Memory Usage**: Blacklist entries expire automatically
- **Network Latency**: Minimal impact on refresh token validation
- **Scalability**: Redis can handle millions of blacklist entries

## 🧪 Testing

### Unit Tests

```go
func TestLogoutWithRedis(t *testing.T) {
    // Test successful logout with Redis
    // Test logout without Redis (graceful degradation)
    // Test invalid refresh token handling
    // Test blacklist checking in refresh token
}

func TestRefreshTokenWithBlacklist(t *testing.T) {
    // Test refresh with blacklisted token
    // Test refresh with valid token
    // Test refresh when Redis is unavailable
}
```

### Integration Tests

```go
func TestLogoutFlow(t *testing.T) {
    // 1. Login user and get refresh token
    // 2. Logout user with refresh token
    // 3. Try to refresh token (should fail)
    // 4. Verify token is in blacklist
}
```

### Manual Testing

1. **Login and Logout Flow:**
   ```bash
   # 1. Login to get tokens
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"identity": "test@example.com", "password": "password123"}'
   
   # 2. Logout with refresh token
   curl -X POST http://localhost:8080/api/v1/auth/logout \
     -H "Content-Type: application/json" \
     -d '{"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}'
   
   # 3. Try to refresh token (should fail)
   curl -X POST http://localhost:8080/api/v1/auth/refresh \
     -H "Content-Type: application/json" \
     -d '{"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}'
   ```

2. **Redis Verification:**
   ```bash
   # Check Redis for blacklist entries
   redis-cli KEYS "refresh_token_blacklist:*"
   redis-cli KEYS "user_refresh_tokens:*"
   ```

## 📈 Performance Considerations

### Optimizations

1. **Efficient Blacklist Checks:**
   ```go
   exists, err := s.redisService.Exists(ctx, blacklistKey)
   if exists > 0 {
       return nil, errors.New("refresh token has been revoked")
   }
   ```

2. **Graceful Degradation:**
   - System works without Redis
   - Logs warnings when Redis is unavailable
   - Doesn't block logout process

3. **Memory Management:**
   - Automatic expiration of blacklist entries
   - No memory leaks from abandoned tokens
   - Efficient Redis key patterns

### Monitoring

- **Blacklist Size**: Monitor number of blacklisted tokens
- **Redis Performance**: Track Redis operation latency
- **Error Rates**: Monitor blacklist operation failures
- **Memory Usage**: Track Redis memory consumption

## 🔍 Troubleshooting

### Common Issues

1. **Redis Connection Failures:**
   - Check Redis service status
   - Verify Redis configuration
   - Check network connectivity

2. **Token Not Blacklisted:**
   - Verify Redis operations are successful
   - Check Redis key patterns
   - Monitor Redis logs

3. **Performance Issues:**
   - Monitor Redis memory usage
   - Check blacklist size
   - Optimize Redis configuration

### Debugging

1. **Enable Logging:**
   ```go
   logger.Errorf("Error adding refresh token to blacklist: %v", err)
   logger.Infof("Refresh token blacklisted for user %s", claims.UserID.String())
   ```

2. **Redis Debugging:**
   ```bash
   # Check Redis keys
   redis-cli KEYS "*blacklist*"
   
   # Check specific token
   redis-cli GET "refresh_token_blacklist:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
   
   # Monitor Redis operations
   redis-cli MONITOR
   ```

## 📖 Usage Examples

### Complete Logout Flow

```bash
# 1. Login
response=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"identity": "test@example.com", "password": "password123"}')

# Extract refresh token
refresh_token=$(echo $response | jq -r '.refresh_token')

# 2. Logout
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"$refresh_token\"}"

# 3. Try to refresh (should fail)
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"$refresh_token\"}"
```

### Expected Responses

**Successful Logout:**
```json
{
    "message": "Logout successful"
}
```

**Failed Refresh After Logout:**
```json
{
    "error": "refresh token has been revoked"
}
```

## 📚 Additional Resources

- [JWT Authentication Guide](./AUTHENTICATION.md)
- [Redis Configuration](./REDIS_CONFIGURATION.md)
- [Security Best Practices](./SECURITY_GUIDELINES.md)
- [API Documentation](./API_DOCUMENTATION.md) 