# Laravel-Style Validation Guide

This guide explains how to use the new Laravel-style validation system in the Go Next backend.

## Overview

The validation system provides Laravel-style error responses with:
- HTTP status code 422 (Unprocessable Entity) for validation errors
- Structured error messages with field names and values
- Request tracking with request IDs and timestamps
- Comprehensive validation rules and helpers

## Response Format

### Validation Error Response
```json
{
  "message": "The given data was invalid.",
  "errors": [
    {
      "field": "email",
      "message": "The email must be a valid email address.",
      "value": "invalid-email"
    },
    {
      "field": "password",
      "message": "The password must be at least 8 characters.",
      "value": "123"
    }
  ],
  "status": 422,
  "timestamp": "2024-01-15T10:30:00Z",
  "path": "/api/auth/register",
  "request_id": "req_123456789"
}
```

### General Error Response
```json
{
  "message": "User not found",
  "status": 404,
  "timestamp": "2024-01-15T10:30:00Z",
  "path": "/api/users/123",
  "request_id": "req_123456789",
  "details": "Optional error details"
}
```

### Success Response
```json
{
  "message": "User created successfully",
  "data": {
    "id": "uuid-here",
    "username": "john_doe",
    "email": "john@example.com"
  },
  "status": 201,
  "timestamp": "2024-01-15T10:30:00Z",
  "path": "/api/users",
  "request_id": "req_123456789"
}
```

## Usage Examples

### 1. Basic Request Validation

```go
func (h *userHandler) CreateUser(c *gin.Context) {
    var req requests.UserCreateRequest
    
    // Use the new validation helper
    if !requests.ValidateRequest(c, &req) {
        return // Validation error already sent
    }
    
    // Continue with business logic...
    user, err := h.UserService.CreateUser(&req)
    if err != nil {
        responses.SendError(c, 500, "Failed to create user", err.Error())
        return
    }
    
    responses.SendSuccess(c, 201, "User created successfully", user)
}
```

### 2. Partial Validation

```go
func (h *authHandler) Login(c *gin.Context) {
    var req requests.AuthRequest
    
    // Validate only email and password fields
    if !requests.ValidateRequestPartial(c, &req, "Email", "Password") {
        return
    }
    
    // Continue with authentication...
}
```

### 3. Query Parameter Validation

```go
type PaginationRequest struct {
    Page  int `form:"page" validate:"min=1"`
    Limit int `form:"limit" validate:"min=1,max=100"`
}

func (h *postHandler) GetPosts(c *gin.Context) {
    var req PaginationRequest
    
    if !requests.ValidateQuery(c, &req) {
        return
    }
    
    // Continue with pagination...
}
```

### 4. Form Data Validation

```go
type FileUploadRequest struct {
    Title string `form:"title" validate:"required,min=3,max=255"`
    File  *multipart.FileHeader `form:"file" validate:"required"`
}

func (h *mediaHandler) UploadFile(c *gin.Context) {
    var req FileUploadRequest
    
    if !requests.ValidateForm(c, &req) {
        return
    }
    
    // Validate file separately
    if !requests.ValidateFile(c, "file", 10*1024*1024, "image/jpeg", "image/png") {
        return
    }
    
    // Continue with file upload...
}
```

### 5. UUID Parameter Validation

```go
func (h *postHandler) GetPost(c *gin.Context) {
    if !requests.ValidateUUID(c, "id") {
        return
    }
    
    id := c.Param("id")
    post, err := h.PostService.GetPostByID(id)
    if err != nil {
        responses.SendError(c, 404, "Post not found")
        return
    }
    
    responses.SendSuccess(c, 200, "Post retrieved successfully", post)
}
```

### 6. Conditional Validation

```go
type UserUpdateRequest struct {
    Username string `json:"username" validate:"omitempty,min=3,max=50"`
    Email    string `json:"email" validate:"omitempty,email"`
    Password string `json:"password" validate:"omitempty,min=8"`
}

func (h *userHandler) UpdateUser(c *gin.Context) {
    var req UserUpdateRequest
    
    if !requests.ValidateRequest(c, &req) {
        return
    }
    
    // Conditional validation based on business logic
    if req.Password != "" {
        if !requests.ValidateVar(c, req.Password, "min=8,containsany=!@#$%^&*") {
            return
        }
    }
    
    // Continue with update...
}
```

### 7. Nested Struct Validation

```go
type Address struct {
    Street  string `json:"street" validate:"required"`
    City    string `json:"city" validate:"required"`
    Country string `json:"country" validate:"required"`
}

type UserCreateRequest struct {
    Username string  `json:"username" validate:"required,min=3,max=50"`
    Email    string  `json:"email" validate:"required,email"`
    Address  Address `json:"address" validate:"required"`
}

func (h *userHandler) CreateUser(c *gin.Context) {
    var req UserCreateRequest
    
    // Validate nested structs
    if !requests.ValidateNested(c, &req) {
        return
    }
    
    // Continue with user creation...
}
```

## Available Validation Rules

### Basic Rules
- `required` - Field is required
- `email` - Must be a valid email address
- `min` - Minimum length/value
- `max` - Maximum length/value
- `numeric` - Must be a number
- `alpha` - Must contain only letters
- `alphanum` - Must contain only letters and numbers
- `url` - Must be a valid URL
- `uuid` - Must be a valid UUID

### Comparison Rules
- `gt` - Greater than
- `gte` - Greater than or equal to
- `lt` - Less than
- `lte` - Less than or equal to
- `eq` - Equal to
- `ne` - Not equal to

### String Rules
- `contains` - Must contain substring
- `containsany` - Must contain any of the characters
- `startswith` - Must start with
- `endswith` - Must end with
- `oneof` - Must be one of the specified values

### Date Rules
- `datetime` - Must be a valid date/time
- `after` - Must be after specified date
- `before` - Must be before specified date

### File Rules
- `file` - Must be a file upload
- `image` - Must be an image file
- `mimes` - Must be one of the specified MIME types
- `size` - File size validation

## Custom Validation Messages

The system automatically provides Laravel-style validation messages. You can customize them by modifying the `getValidationMessage` function in `responses/response.go`.

## Helper Functions

### Request Validation
- `ValidateRequest(c, request)` - Validate JSON request
- `ValidateRequestPartial(c, request, fields...)` - Validate specific fields
- `ValidateQuery(c, request)` - Validate query parameters
- `ValidateForm(c, request)` - Validate form data

### Field Validation
- `ValidateVar(c, value, tag)` - Validate single variable
- `ValidateSlice(c, slice, tag)` - Validate slice
- `ValidateMap(c, map, tag)` - Validate map
- `ValidateStruct(c, obj)` - Validate struct
- `ValidateStructPartial(c, obj, fields...)` - Validate specific struct fields

### Special Validation
- `ValidateUUID(c, paramName)` - Validate UUID parameter
- `ValidateFile(c, fieldName, maxSize, allowedTypes...)` - Validate file upload
- `ValidateRequiredFields(c, fields...)` - Validate required fields
- `ValidateNested(c, obj)` - Validate nested structs
- `ValidateConditional(c, obj, condition, fields...)` - Conditional validation

### Pagination and Sorting
- `ValidatePagination(c)` - Validate pagination parameters
- `ValidateSort(c, allowedFields...)` - Validate sorting parameters
- `ValidateDateRange(c)` - Validate date range parameters
- `ValidateSearch(c)` - Validate search parameters

## Error Handling

The validation system automatically handles:
- JSON parsing errors
- Validation rule violations
- Missing required fields
- Invalid data types
- File upload errors

All errors are returned with appropriate HTTP status codes and structured error messages.

## Best Practices

1. **Use the helper functions** instead of manual validation
2. **Validate early** in your handler functions
3. **Use appropriate validation rules** for your data
4. **Handle business logic errors** separately from validation errors
5. **Use conditional validation** when fields depend on each other
6. **Validate file uploads** with size and type restrictions
7. **Use nested validation** for complex data structures

## Migration from Old System

To migrate from the old validation system:

1. Replace manual `c.ShouldBindJSON()` and `rules.Validate.Struct()` calls with `requests.ValidateRequest()`
2. Replace `c.JSON(400, gin.H{"error": "message"})` with `responses.SendError()`
3. Replace `c.JSON(200, data)` with `responses.SendSuccess()`
4. Update your request structs to use the new validation tags
5. Remove manual validation error formatting

## Example Migration

### Before
```go
func (h *userHandler) CreateUser(c *gin.Context) {
    var req UserCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    if err := rules.Validate.Struct(req); err != nil {
        c.JSON(400, requests.FormatValidationError(err))
        return
    }
    // ... rest of the function
}
```

### After
```go
func (h *userHandler) CreateUser(c *gin.Context) {
    var req UserCreateRequest
    if !requests.ValidateRequest(c, &req) {
        return
    }
    // ... rest of the function
}
```

This new system provides a more consistent, maintainable, and user-friendly validation experience that matches Laravel's validation patterns. 