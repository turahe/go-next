# Laravel-Style Validation for Go

This package provides a Laravel-inspired validation system for Go applications, offering familiar validation rules and error messages for HTTP request validation.

## Features

- **Laravel-Style Rules**: Familiar validation rules like `required`, `email`, `min`, `max`, etc.
- **Struct Tag Support**: Define validation rules using struct tags
- **Custom Error Messages**: Customizable error messages for each field and rule
- **Custom Validation Rules**: Add your own validation functions
- **HTTP Integration**: Seamless integration with Gin HTTP framework
- **Comprehensive Rules**: Support for strings, numbers, arrays, dates, URLs, UUIDs, and more

## Installation

Add the required dependencies to your `go.mod`:

```go
require (
    github.com/gin-gonic/gin v1.9.1
    github.com/go-playground/validator/v10 v10.15.5
)
```

## Quick Start

### 1. Define Request Structs

```go
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required|string|min:2|max:50"`
    Email    string `json:"email" validate:"required|email"`
    Password string `json:"password" validate:"required|string|min:8|max:255"`
    Age      int    `json:"age" validate:"numeric|min:18|max:120"`
    Website  string `json:"website" validate:"url"`
    Bio      string `json:"bio" validate:"string|max:500"`
}
```

### 2. Create Validator

```go
validator := validation.NewLaravelValidator()
```

### 3. Validate Request

```go
func CreateUser(c *gin.Context) {
    var request CreateUserRequest
    
    // Bind JSON to request struct
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "message": "Invalid JSON format",
            "error":   err.Error(),
        })
        return
    }
    
    // Validate the request
    result := validator.Validate(&request)
    if !result.IsValid {
        c.JSON(http.StatusUnprocessableEntity, gin.H{
            "message": "Validation failed",
            "errors":  result.Errors,
        })
        return
    }
    
    // Process the valid request
    c.JSON(http.StatusCreated, gin.H{
        "message": "User created successfully",
        "data":    request,
    })
}
```

## Available Validation Rules

### Basic Validation

| Rule | Description | Example |
|------|-------------|---------|
| `required` | Field is required | `validate:"required"` |
| `email` | Valid email address | `validate:"email"` |
| `url` | Valid URL | `validate:"url"` |
| `numeric` | Numeric value | `validate:"numeric"` |
| `integer` | Integer value | `validate:"integer"` |
| `string` | String value | `validate:"string"` |

### String Validation

| Rule | Description | Example |
|------|-------------|---------|
| `min:N` | Minimum length/value | `validate:"min:5"` |
| `max:N` | Maximum length/value | `validate:"max:100"` |
| `size:N` | Exact length/value | `validate:"size:10"` |
| `between:min,max` | Between range | `validate:"between:5,100"` |
| `alpha` | Only letters | `validate:"alpha"` |
| `alpha_num` | Letters and numbers | `validate:"alpha_num"` |
| `alpha_dash` | Letters, numbers, dashes, underscores | `validate:"alpha_dash"` |

### Numeric Validation

| Rule | Description | Example |
|------|-------------|---------|
| `numeric` | Numeric value | `validate:"numeric"` |
| `integer` | Integer value | `validate:"integer"` |
| `min:N` | Minimum value | `validate:"min:0"` |
| `max:N` | Maximum value | `validate:"max:1000"` |
| `between:min,max` | Between range | `validate:"between:0,1000"` |

### Array Validation

| Rule | Description | Example |
|------|-------------|---------|
| `array` | Array value | `validate:"array"` |
| `min:N` | Minimum array length | `validate:"min:1"` |
| `max:N` | Maximum array length | `validate:"max:10"` |
| `size:N` | Exact array length | `validate:"size:5"` |

### Date Validation

| Rule | Description | Example |
|------|-------------|---------|
| `date` | Valid date | `validate:"date"` |
| `date_format:format` | Date with specific format | `validate:"date_format:2006-01-02"` |
| `before:field` | Date before another field | `validate:"before:end_date"` |
| `after:field` | Date after another field | `validate:"after:start_date"` |

### Advanced Validation

| Rule | Description | Example |
|------|-------------|---------|
| `json` | Valid JSON string | `validate:"json"` |
| `ip` | Valid IP address | `validate:"ip"` |
| `ipv4` | Valid IPv4 address | `validate:"ipv4"` |
| `ipv6` | Valid IPv6 address | `validate:"ipv6"` |
| `uuid` | Valid UUID | `validate:"uuid"` |
| `regex:pattern` | Regex pattern match | `validate:"regex:^[0-9]{10}$"` |
| `same:field` | Same as another field | `validate:"same:password_confirm"` |
| `different:field` | Different from another field | `validate:"different:old_password"` |

## Custom Error Messages

### Adding Custom Messages

```go
validator := validation.NewLaravelValidator()

// Add custom error messages
validator.AddCustomMessage("email", "email", "Please provide a valid email address")
validator.AddCustomMessage("password", "min", "Password must be at least 8 characters long")
validator.AddCustomMessage("age", "min", "You must be at least 18 years old")
```

### Adding Multiple Messages

```go
messages := map[string]string{
    "name.required":     "Please provide your full name",
    "email.required":    "Email address is required",
    "email.email":       "Please provide a valid email address",
    "password.required": "Password is required",
    "password.min":      "Password must be at least 8 characters long",
    "age.min":           "You must be at least 18 years old",
    "age.max":           "Age cannot exceed 120 years",
}

validator.AddCustomMessages(messages)
```

## Custom Validation Rules

### Creating Custom Rules

```go
// Custom validation function
func validateStrongPassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    
    // Check for at least one uppercase letter, lowercase letter, digit, and special character
    hasUpper := false
    hasLower := false
    hasDigit := false
    hasSpecial := false
    
    for _, char := range password {
        switch {
        case char >= 'A' && char <= 'Z':
            hasUpper = true
        case char >= 'a' && char <= 'z':
            hasLower = true
        case char >= '0' && char <= '9':
            hasDigit = true
        case char == '!' || char == '@' || char == '#' || char == '$' || char == '%':
            hasSpecial = true
        }
    }
    
    return hasUpper && hasLower && hasDigit && hasSpecial
}

// Register custom rule
validator := validation.NewLaravelValidator()
validator.validate.RegisterValidation("strong_password", validateStrongPassword)
```

### Using Custom Rules

```go
type UserRequest struct {
    Password string `json:"password" validate:"required|strong_password"`
}
```

## HTTP Handler Integration

### Using with BaseHandler

```go
type UserHandler struct {
    *controllers.BaseHandler
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    var request CreateUserRequest
    
    // Validate request using base handler
    if err := h.ValidateRequestParams(c, &request); err != nil {
        return // Error response already sent by base handler
    }
    
    // Process valid request
    // ...
}
```

### Using with Custom Rules

```go
func (h *UserHandler) UpdateUser(c *gin.Context) {
    var request UpdateUserRequest
    
    // Define custom validation rules
    rules := map[string]string{
        "name":  "string|min:2|max:50",
        "email": "email",
        "age":   "numeric|min:18|max:120",
    }
    
    // Validate with custom rules
    if err := h.ValidateRequestWithRules(c, &request, rules); err != nil {
        return // Error response already sent by base handler
    }
    
    // Process valid request
    // ...
}
```

## Example Request Structs

### User Management

```go
type CreateUserRequest struct {
    Name            string `json:"name" validate:"required|string|min:2|max:50"`
    Email           string `json:"email" validate:"required|email"`
    Password        string `json:"password" validate:"required|string|min:8|max:255"`
    PasswordConfirm string `json:"password_confirm" validate:"required|string|same:password"`
    Age             int    `json:"age" validate:"numeric|min:18|max:120"`
    Website         string `json:"website" validate:"url"`
    Bio             string `json:"bio" validate:"string|max:500"`
    Terms           bool   `json:"terms" validate:"required"`
}

type UpdateUserRequest struct {
    Name     string `json:"name" validate:"string|min:2|max:50"`
    Email    string `json:"email" validate:"email"`
    Password string `json:"password" validate:"string|min:8|max:255"`
    Age      int    `json:"age" validate:"numeric|min:18|max:120"`
    Website  string `json:"website" validate:"url"`
    Bio      string `json:"bio" validate:"string|max:500"`
}
```

### Content Management

```go
type CreatePostRequest struct {
    Title       string    `json:"title" validate:"required|string|min:5|max:255"`
    Content     string    `json:"content" validate:"required|string|min:10"`
    CategoryID  string    `json:"category_id" validate:"required|uuid"`
    Tags        []string  `json:"tags" validate:"array|max:10"`
    PublishedAt *time.Time `json:"published_at" validate:"date"`
    MetaData    string    `json:"meta_data" validate:"json"`
}

type CreateCommentRequest struct {
    Content   string  `json:"content" validate:"required|string|min:1|max:1000"`
    PostID    string  `json:"post_id" validate:"required|uuid"`
    ParentID  *string `json:"parent_id" validate:"uuid"`
    Anonymous bool    `json:"anonymous" validate:"boolean"`
}
```

### Search and Filtering

```go
type SearchRequest struct {
    Query   string `json:"query" validate:"required|string|min:1|max:100"`
    Page    int    `json:"page" validate:"numeric|min:1"`
    PerPage int    `json:"per_page" validate:"numeric|min:1|max:100"`
    SortBy  string `json:"sort_by" validate:"string|oneof:created_at updated_at title name"`
    Order   string `json:"order" validate:"string|oneof:asc desc"`
}

type FilterRequest struct {
    CategoryID string   `json:"category_id" validate:"uuid"`
    Tags       []string `json:"tags" validate:"array|max:10"`
    DateFrom   string   `json:"date_from" validate:"date"`
    DateTo     string   `json:"date_to" validate:"date"`
    AuthorID   string   `json:"author_id" validate:"uuid"`
    Status     string   `json:"status" validate:"string|oneof:draft published archived"`
}
```

## Error Response Format

When validation fails, the response will be in this format:

```json
{
    "message": "Validation failed",
    "errors": {
        "name": [
            "The name field is required."
        ],
        "email": [
            "The email field must be a valid email address."
        ],
        "password": [
            "The password field must be at least 8 characters."
        ],
        "age": [
            "The age field must be at least 18."
        ]
    }
}
```

## Best Practices

### 1. Define Clear Validation Rules

```go
// Good: Clear and specific rules
type UserRequest struct {
    Name     string `json:"name" validate:"required|string|min:2|max:50"`
    Email    string `json:"email" validate:"required|email"`
    Password string `json:"password" validate:"required|string|min:8|max:255"`
}

// Avoid: Vague or missing rules
type UserRequest struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password"`
}
```

### 2. Use Custom Error Messages

```go
// Add meaningful error messages
validator.AddCustomMessage("email", "email", "Please provide a valid email address")
validator.AddCustomMessage("password", "min", "Password must be at least 8 characters long")
```

### 3. Group Related Validations

```go
// Group related fields in separate structs
type UserProfile struct {
    Name     string `json:"name" validate:"required|string|min:2|max:50"`
    Bio      string `json:"bio" validate:"string|max:500"`
    Website  string `json:"website" validate:"url"`
}

type UserCredentials struct {
    Email    string `json:"email" validate:"required|email"`
    Password string `json:"password" validate:"required|string|min:8|max:255"`
}
```

### 4. Use Conditional Validation

```go
// Use custom validation for complex rules
type UserRequest struct {
    Email    string `json:"email" validate:"required|email"`
    Password string `json:"password" validate:"required|string|min:8"`
    Role     string `json:"role" validate:"required|string|oneof:user admin moderator"`
}
```

### 5. Validate at Multiple Levels

```go
// Validate at both client and server level
// Client-side: JavaScript validation for immediate feedback
// Server-side: Go validation for security and data integrity
```

## Performance Considerations

- **Reuse Validators**: Create validator instances once and reuse them
- **Cache Validation Results**: For complex validations, consider caching results
- **Lazy Loading**: Load custom validation rules only when needed
- **Batch Validation**: Validate multiple requests in batches when possible

## Testing

### Unit Testing Validation

```go
func TestUserValidation(t *testing.T) {
    validator := validation.NewLaravelValidator()
    
    // Test valid request
    validRequest := &CreateUserRequest{
        Name:     "John Doe",
        Email:    "john@example.com",
        Password: "password123",
        Age:      25,
    }
    
    result := validator.Validate(validRequest)
    assert.True(t, result.IsValid)
    
    // Test invalid request
    invalidRequest := &CreateUserRequest{
        Name:     "J", // Too short
        Email:    "invalid-email", // Invalid email
        Password: "123", // Too short
        Age:      15, // Too young
    }
    
    result = validator.Validate(invalidRequest)
    assert.False(t, result.IsValid)
    assert.Contains(t, result.Errors, "name")
    assert.Contains(t, result.Errors, "email")
    assert.Contains(t, result.Errors, "password")
    assert.Contains(t, result.Errors, "age")
}
```

## Migration from Other Validation Libraries

### From go-playground/validator

```go
// Before: Using go-playground/validator directly
type User struct {
    Name  string `validate:"required,min=2,max=50"`
    Email string `validate:"required,email"`
}

// After: Using Laravel-style validation
type User struct {
    Name  string `json:"name" validate:"required|string|min:2|max:50"`
    Email string `json:"email" validate:"required|email"`
}
```

### From Laravel (PHP)

```php
// Laravel (PHP)
$request->validate([
    'name' => 'required|string|min:2|max:50',
    'email' => 'required|email',
    'password' => 'required|string|min:8|max:255',
]);
```

```go
// Go with Laravel-style validation
type UserRequest struct {
    Name     string `json:"name" validate:"required|string|min:2|max:50"`
    Email    string `json:"email" validate:"required|email"`
    Password string `json:"password" validate:"required|string|min:8|max:255"`
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project is licensed under the MIT License. 