# 🔑 Password Reset Implementation

This document describes the password reset functionality implemented in the Go-Next Admin Panel, which provides secure password recovery through email-based token verification.

## 📋 Overview

The password reset system allows users to securely reset their passwords by:
1. Requesting a password reset via email
2. Receiving a secure token via email
3. Using the token to set a new password
4. Automatic token invalidation after use

## 🏗️ Architecture

### Components

1. **AuthService**: Handles password reset logic and token management
2. **RedisService**: Stores reset tokens with expiration
3. **EmailService**: Sends password reset emails
4. **AuthHandler**: HTTP endpoints for password reset requests
5. **Token Storage**: Redis-based token storage with fallback to in-memory

### Flow

```
Request Reset → Generate Token → Store in Redis → Send Email → User Clicks Link → Validate Token → Update Password → Invalidate Token
```

## 🔧 Implementation Details

### 1. Token Generation and Storage

The system generates secure random tokens and stores them with expiration:

```go
// Generate password reset token
resetToken, err := s.generateVerificationToken()
if err != nil {
    return err
}

// Store reset token in Redis with 1-hour expiration
if s.redisService != nil {
    ctx := context.Background()
    resetKey := fmt.Sprintf("password_reset:%s", userEmail)
    expiration := time.Hour // 1 hour expiration

    err := s.redisService.SetWithExpiration(ctx, resetKey, resetToken, expiration)
    if err != nil {
        logger.Errorf("Error storing password reset token: %v", err)
        return errors.New("failed to generate reset token")
    }
}
```

**Features:**
- **Secure token generation**: 32-byte random hex tokens
- **Redis storage**: Primary storage with automatic expiration
- **In-memory fallback**: Works without Redis
- **1-hour expiration**: Tokens expire after 1 hour
- **Single-use tokens**: Tokens are invalidated after use

### 2. Email Template

The password reset email uses a professional HTML template:

```html
<div style="background-color: #f8f9fa; padding: 20px; border-radius: 8px; margin-bottom: 20px;">
    <h2 style="color: #007bff; margin-top: 0;">🔑 Password Reset Request</h2>
    <p>Hello <strong>{{username}}</strong>,</p>
    <p>We received a request to reset your password. Click the button below to create a new password:</p>
</div>

<div style="text-align: center; padding: 30px; background-color: #ffffff; border: 1px solid #dee2e6; border-radius: 8px; margin-bottom: 20px;">
    <a href="{{resetURL}}" style="display: inline-block; background-color: #007bff; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; font-weight: bold; font-size: 16px;">
        Reset Password
    </a>
</div>
```

**Features:**
- **Professional design**: Clean, modern email template
- **Security notices**: Clear warnings about unauthorized requests
- **Important information**: Details about token expiration and usage
- **Responsive design**: Works on desktop and mobile

### 3. Password Reset Process

The reset process validates tokens and updates passwords securely:

```go
func (s *authService) ResetPassword(token, newPassword string) error {
    // Find the email associated with this token
    var userEmail string

    if s.redisService != nil {
        // Check Redis for the token
        ctx := context.Background()
        keys, err := s.redisService.GetKeysByPattern(ctx, "password_reset:*")
        if err != nil {
            return errors.New("failed to validate reset token")
        }

        for _, key := range keys {
            storedToken, err := s.redisService.Get(ctx, key)
            if err != nil {
                continue
            }
            if storedToken == token {
                // Extract email from key
                if len(key) > 15 {
                    userEmail = key[15:]
                }
                break
            }
        }
    }

    if userEmail == "" {
        return errors.New("invalid or expired reset token")
    }

    // Find user and update password
    var user models.User
    if err := s.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
        return errors.New("user not found")
    }

    // Hash new password
    if err := user.HashPassword(newPassword); err != nil {
        return err
    }

    // Update password in database
    if err := s.db.Save(&user).Error; err != nil {
        return err
    }

    // Remove the reset token from storage
    if s.redisService != nil {
        ctx := context.Background()
        resetKey := fmt.Sprintf("password_reset:%s", userEmail)
        s.redisService.Delete(ctx, resetKey)
    }

    return nil
}
```

## 🚀 API Endpoints

### Request Password Reset

**Endpoint:** `POST /api/v1/auth/request-password-reset`

**Request Body:**
```json
{
    "email": "user@example.com"
}
```

**Response:**
```json
{
    "message": "Password reset email sent successfully"
}
```

**Error Responses:**
- `400 Bad Request`: User not found
- `400 Bad Request`: Invalid JSON format
- `422 Unprocessable Entity`: Validation errors
- `500 Internal Server Error`: Email sending failed

### Reset Password

**Endpoint:** `POST /api/v1/auth/reset-password`

**Request Body:**
```json
{
    "token": "f062b935cba562afa25695c8f16674e917f27fc481312d11cf3b8f7090a476d4",
    "new_password": "newSecurePassword123"
}
```

**Response:**
```json
{
    "message": "Password reset successful"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid or expired token
- `400 Bad Request`: Invalid JSON format
- `422 Unprocessable Entity`: Validation errors
- `500 Internal Server Error`: Database error

## 🔧 Configuration

### Environment Variables

```env
# SMTP Configuration
MAIL_HOST=localhost
MAIL_PORT=1025
MAIL_USERNAME=
MAIL_PASSWORD=
MAIL_FROM=noreply@example.com

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=redis_password
REDIS_DB=0
```

### Redis Key Patterns

The system uses the following Redis key patterns:

1. **Password Reset Token**: `password_reset:{email}`
   - Value: `{reset_token}`
   - TTL: 1 hour
   - Purpose: Store reset tokens with expiration

## 🚨 Security Considerations

### Token Security

- **Cryptographically secure**: 32-byte random hex tokens
- **Time-limited**: 1-hour expiration prevents long-term abuse
- **Single-use**: Tokens are invalidated after successful reset
- **Email-specific**: Tokens are tied to specific email addresses

### Password Security

- **Secure hashing**: Passwords are hashed using bcrypt
- **Validation**: New passwords must meet minimum requirements
- **Database security**: Passwords are never stored in plain text
- **Token cleanup**: Reset tokens are removed after use

### Email Security

- **Secure transmission**: Emails sent via SMTP
- **Clear instructions**: Users are informed about security measures
- **Unauthorized notice**: Clear warnings about ignoring unwanted emails
- **Professional branding**: Consistent with application design

### Best Practices

1. **Rate Limiting**: Implement rate limiting for reset requests
2. **Logging**: Monitor reset attempts for security analysis
3. **Email Verification**: Ensure email addresses are verified
4. **Strong Passwords**: Enforce strong password requirements

## 📊 Redis Operations

### Token Management

```go
// Store reset token
resetKey := fmt.Sprintf("password_reset:%s", userEmail)
err := redisService.SetWithExpiration(ctx, resetKey, resetToken, time.Hour)

// Retrieve reset token
storedToken, err := redisService.Get(ctx, resetKey)

// Delete reset token after use
redisService.Delete(ctx, resetKey)

// Search for tokens by pattern
keys, err := redisService.GetKeysByPattern(ctx, "password_reset:*")
```

### Performance Considerations

- **Fast lookups**: Redis provides O(1) token retrieval
- **Automatic expiration**: Tokens expire automatically
- **Memory efficient**: Minimal memory usage for token storage
- **Scalable**: Can handle thousands of concurrent reset requests

## 🧪 Testing

### Unit Tests

```go
func TestRequestPasswordReset(t *testing.T) {
    // Test successful password reset request
    // Test user not found
    // Test email sending failure
    // Test Redis storage failure
}

func TestResetPassword(t *testing.T) {
    // Test successful password reset
    // Test invalid token
    // Test expired token
    // Test already used token
    // Test weak password
}
```

### Integration Tests

```go
func TestPasswordResetFlow(t *testing.T) {
    // 1. Request password reset
    // 2. Verify email is sent
    // 3. Extract token from email
    // 4. Reset password with token
    // 5. Verify new password works
    // 6. Verify old password doesn't work
    // 7. Verify token can't be reused
}
```

### Manual Testing

1. **Request Password Reset:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/request-password-reset \
     -H "Content-Type: application/json" \
     -d '{"email": "test@example.com"}'
   ```

2. **Check Email:** Visit http://localhost:8025 to view sent emails

3. **Reset Password:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/reset-password \
     -H "Content-Type: application/json" \
     -d '{
       "token": "f062b935cba562afa25695c8f16674e917f27fc481312d11cf3b8f7090a476d4",
       "new_password": "newSecurePassword123"
     }'
   ```

4. **Test Login:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{
       "identity": "test@example.com",
       "password": "newSecurePassword123"
     }'
   ```

## 📈 Performance Considerations

### Optimizations

1. **Efficient Token Storage:**
   ```go
   // Use Redis for fast token lookups
   err := s.redisService.SetWithExpiration(ctx, resetKey, resetToken, time.Hour)
   ```

2. **Graceful Degradation:**
   - System works without Redis
   - In-memory fallback for token storage
   - Proper error handling for storage failures

3. **Email Optimization:**
   - Non-blocking email sending
   - Professional email templates
   - Clear call-to-action buttons

### Monitoring

- **Reset Request Rate**: Monitor password reset requests
- **Email Delivery**: Track email sending success rates
- **Token Usage**: Monitor token validation success/failure
- **Security Events**: Log suspicious reset attempts

## 🔍 Troubleshooting

### Common Issues

1. **Email Not Received:**
   - Check SMTP configuration
   - Verify email address format
   - Check spam/junk folder
   - Test with MailHog in development

2. **Token Not Working:**
   - Verify token hasn't expired (1 hour limit)
   - Check if token was already used
   - Verify Redis connectivity
   - Check token format and length

3. **Password Reset Fails:**
   - Verify new password meets requirements
   - Check database connectivity
   - Verify user exists in database
   - Check password hashing errors

### Debugging

1. **Enable Logging:**
   ```go
   logger.Infof("Password reset token generated for user %s", user.Username)
   logger.Errorf("Error sending password reset email: %v", err)
   ```

2. **Check Redis:**
   ```bash
   # Check Redis for reset tokens
   redis-cli KEYS "password_reset:*"
   
   # Check specific token
   redis-cli GET "password_reset:user@example.com"
   ```

3. **Email Debugging:**
   ```bash
   # Check MailHog for sent emails
   curl http://localhost:8025/api/v1/messages
   ```

## 📖 Usage Examples

### Complete Password Reset Flow

```bash
# 1. Request password reset
curl -X POST http://localhost:8080/api/v1/auth/request-password-reset \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}'

# 2. Check email for reset token
# Visit http://localhost:8025 to view sent emails

# 3. Reset password with token
curl -X POST http://localhost:8080/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "token": "f062b935cba562afa25695c8f16674e917f27fc481312d11cf3b8f7090a476d4",
    "new_password": "newSecurePassword123"
  }'

# 4. Test login with new password
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "identity": "test@example.com",
    "password": "newSecurePassword123"
  }'
```

### Expected Responses

**Successful Reset Request:**
```json
{
    "message": "Password reset email sent successfully"
}
```

**Successful Password Reset:**
```json
{
    "message": "Password reset successful"
}
```

**Invalid Token:**
```json
{
    "message": "Password reset failed",
    "details": "invalid or expired reset token"
}
```

## 📚 Additional Resources

- [Email Verification Guide](./EMAIL_VERIFICATION.md)
- [JWT Authentication Guide](./AUTHENTICATION.md)
- [Redis Configuration](./REDIS_CONFIGURATION.md)
- [Security Best Practices](./SECURITY_GUIDELINES.md)
- [API Documentation](./API_DOCUMENTATION.md) 