# 🛠️ Implementation Examples

## 📋 Phase 1: Foundation - Domain Layer

### 1. Domain Entities

```go
// internal/domain/entities/user.go
package entities

import (
    "time"
    "go-next/internal/domain/valueobjects"
)

type User struct {
    ID        valueobjects.UUID `json:"id"`
    Email     valueobjects.Email `json:"email"`
    Username  string            `json:"username"`
    Password  valueobjects.Password `json:"-"`
    Role      Role              `json:"role"`
    Status    UserStatus        `json:"status"`
    CreatedAt time.Time         `json:"created_at"`
    UpdatedAt time.Time         `json:"updated_at"`
    DeletedAt *time.Time        `json:"deleted_at,omitempty"`
}

func NewUser(email, username, password string) (*User, error) {
    emailVO, err := valueobjects.NewEmail(email)
    if err != nil {
        return nil, err
    }
    
    passwordVO, err := valueobjects.NewPassword(password)
    if err != nil {
        return nil, err
    }
    
    return &User{
        ID:        valueobjects.NewUUID(),
        Email:     emailVO,
        Username:  username,
        Password:  passwordVO,
        Role:      RoleUser,
        Status:    UserStatusActive,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }, nil
}

func (u *User) ChangePassword(newPassword string) error {
    passwordVO, err := valueobjects.NewPassword(newPassword)
    if err != nil {
        return err
    }
    u.Password = passwordVO
    u.UpdatedAt = time.Now()
    return nil
}

func (u *User) IsActive() bool {
    return u.Status == UserStatusActive
}

func (u *User) HasRole(role Role) bool {
    return u.Role == role
}
```

### 2. Value Objects

```go
// internal/domain/valueobjects/email.go
package valueobjects

import (
    "errors"
    "strings"
)

type Email struct {
    value string
}

var (
    ErrInvalidEmail = errors.New("invalid email format")
    ErrEmptyEmail   = errors.New("email cannot be empty")
)

func NewEmail(email string) (Email, error) {
    if strings.TrimSpace(email) == "" {
        return Email{}, ErrEmptyEmail
    }
    
    if !isValidEmail(email) {
        return Email{}, ErrInvalidEmail
    }
    
    return Email{value: strings.ToLower(strings.TrimSpace(email))}, nil
}

func (e Email) String() string {
    return e.value
}

func (e Email) Value() string {
    return e.value
}

func isValidEmail(email string) bool {
    // Basic email validation
    return strings.Contains(email, "@") && strings.Contains(email, ".")
}
```

### 3. Repository Interfaces

```go
// internal/domain/repositories/user_repository.go
package repositories

import (
    "context"
    "go-next/internal/domain/entities"
    "go-next/internal/shared/errors"
)

type UserRepository interface {
    Create(ctx context.Context, user *entities.User) error
    GetByID(ctx context.Context, id string) (*entities.User, error)
    GetByEmail(ctx context.Context, email string) (*entities.User, error)
    GetByUsername(ctx context.Context, username string) (*entities.User, error)
    Update(ctx context.Context, user *entities.User) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter UserFilter) ([]*entities.User, error)
    Count(ctx context.Context, filter UserFilter) (int64, error)
}

type UserFilter struct {
    Role     *entities.Role
    Status   *entities.UserStatus
    Search   string
    Limit    int
    Offset   int
    OrderBy  string
    OrderDir string
}
```

## 📋 Phase 2: Application Layer

### 1. Use Cases

```go
// internal/application/usecases/auth/login.go
package auth

import (
    "context"
    "time"
    
    "go-next/internal/domain/entities"
    "go-next/internal/domain/repositories"
    "go-next/internal/application/dto"
    "go-next/internal/shared/errors"
    "go-next/pkg/crypto"
    "go-next/pkg/jwt"
)

type LoginUseCase struct {
    userRepo repositories.UserRepository
    jwtService jwt.Service
    cryptoService crypto.Service
}

func NewLoginUseCase(
    userRepo repositories.UserRepository,
    jwtService jwt.Service,
    cryptoService crypto.Service,
) *LoginUseCase {
    return &LoginUseCase{
        userRepo: userRepo,
        jwtService: jwtService,
        cryptoService: cryptoService,
    }
}

func (uc *LoginUseCase) Execute(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, errors.NewValidationError("invalid login request", err)
    }
    
    // Find user by email
    user, err := uc.userRepo.GetByEmail(ctx, req.Email)
    if err != nil {
        return nil, errors.NewNotFoundError("user not found")
    }
    
    // Check if user is active
    if !user.IsActive() {
        return nil, errors.NewUnauthorizedError("account is not active")
    }
    
    // Verify password
    if !uc.cryptoService.VerifyPassword(req.Password, user.Password.Value()) {
        return nil, errors.NewUnauthorizedError("invalid credentials")
    }
    
    // Generate JWT token
    token, err := uc.jwtService.GenerateToken(user.ID.Value(), user.Role.String())
    if err != nil {
        return nil, errors.NewInternalError("failed to generate token", err)
    }
    
    // Generate refresh token
    refreshToken, err := uc.jwtService.GenerateRefreshToken(user.ID.Value())
    if err != nil {
        return nil, errors.NewInternalError("failed to generate refresh token", err)
    }
    
    return &dto.LoginResponse{
        User: dto.UserDTO{
            ID:       user.ID.Value(),
            Email:    user.Email.Value(),
            Username: user.Username,
            Role:     user.Role.String(),
        },
        Token:        token,
        RefreshToken: refreshToken,
        ExpiresIn:    time.Hour * 24, // 24 hours
    }, nil
}
```

### 2. DTOs (Data Transfer Objects)

```go
// internal/application/dto/auth_dto.go
package dto

import (
    "time"
    "go-next/pkg/validator"
)

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}

func (r *LoginRequest) Validate() error {
    return validator.Validate(r)
}

type LoginResponse struct {
    User         UserDTO `json:"user"`
    Token        string  `json:"token"`
    RefreshToken string  `json:"refresh_token"`
    ExpiresIn    time.Duration `json:"expires_in"`
}

type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Username string `json:"username" validate:"required,alphanum,min=3,max=30"`
    Password string `json:"password" validate:"required,min=8"`
}

func (r *RegisterRequest) Validate() error {
    return validator.Validate(r)
}

type UserDTO struct {
    ID       string `json:"id"`
    Email    string `json:"email"`
    Username string `json:"username"`
    Role     string `json:"role"`
}
```

## 📋 Phase 3: Infrastructure Layer

### 1. Repository Implementation

```go
// internal/infrastructure/database/postgres/user_repository.go
package postgres

import (
    "context"
    "database/sql"
    
    "go-next/internal/domain/entities"
    "go-next/internal/domain/repositories"
    "go-next/internal/infrastructure/database/postgres/models"
    "go-next/internal/shared/errors"
    "go-next/pkg/database"
)

type userRepository struct {
    db *database.DB
}

func NewUserRepository(db *database.DB) repositories.UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
    model := models.User{
        ID:        user.ID.Value(),
        Email:     user.Email.Value(),
        Username:  user.Username,
        Password:  user.Password.Value(),
        Role:      user.Role.String(),
        Status:    user.Status.String(),
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
    }
    
    if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
        return errors.NewInternalError("failed to create user", err)
    }
    
    return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
    var model models.User
    
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.NewNotFoundError("user not found")
        }
        return errors.NewInternalError("failed to get user", err)
    }
    
    return r.modelToEntity(&model), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
    var model models.User
    
    if err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error; err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.NewNotFoundError("user not found")
        }
        return errors.NewInternalError("failed to get user", err)
    }
    
    return r.modelToEntity(&model), nil
}

func (r *userRepository) modelToEntity(model *models.User) *entities.User {
    email, _ := entities.NewEmail(model.Email)
    password, _ := entities.NewPassword(model.Password)
    
    return &entities.User{
        ID:        entities.UUID{Value: model.ID},
        Email:     email,
        Username:  model.Username,
        Password:  password,
        Role:      entities.RoleFromString(model.Role),
        Status:    entities.UserStatusFromString(model.Status),
        CreatedAt: model.CreatedAt,
        UpdatedAt: model.UpdatedAt,
        DeletedAt: model.DeletedAt,
    }
}
```

### 2. Database Models

```go
// internal/infrastructure/database/postgres/models/user.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Email     string         `gorm:"uniqueIndex;not null"`
    Username  string         `gorm:"uniqueIndex;not null"`
    Password  string         `gorm:"not null"`
    Role      string         `gorm:"not null;default:'user'"`
    Status    string         `gorm:"not null;default:'active'"`
    CreatedAt time.Time      `gorm:"not null"`
    UpdatedAt time.Time      `gorm:"not null"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (User) TableName() string {
    return "users"
}
```

## 📋 Phase 4: Interface Layer

### 1. HTTP Handlers

```go
// internal/interfaces/http/handlers/auth_handler.go
package handlers

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "go-next/internal/application/usecases/auth"
    "go-next/internal/application/dto"
    "go-next/internal/interfaces/http/requests"
    "go-next/internal/interfaces/http/responses"
    "go-next/internal/shared/errors"
    "go-next/pkg/logger"
)

type AuthHandler struct {
    loginUseCase    *auth.LoginUseCase
    registerUseCase *auth.RegisterUseCase
    logger          logger.Logger
}

func NewAuthHandler(
    loginUseCase *auth.LoginUseCase,
    registerUseCase *auth.RegisterUseCase,
    logger logger.Logger,
) *AuthHandler {
    return &AuthHandler{
        loginUseCase:    loginUseCase,
        registerUseCase: registerUseCase,
        logger:          logger,
    }
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req requests.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Error("failed to bind login request", "error", err)
        c.JSON(http.StatusBadRequest, responses.ErrorResponse{
            Error:   "invalid request",
            Message: err.Error(),
        })
        return
    }
    
    // Convert to application DTO
    loginReq := dto.LoginRequest{
        Email:    req.Email,
        Password: req.Password,
    }
    
    // Execute use case
    result, err := h.loginUseCase.Execute(c.Request.Context(), loginReq)
    if err != nil {
        h.logger.Error("login failed", "error", err, "email", req.Email)
        
        switch {
        case errors.IsValidationError(err):
            c.JSON(http.StatusBadRequest, responses.ErrorResponse{
                Error:   "validation_error",
                Message: err.Error(),
            })
        case errors.IsNotFoundError(err):
            c.JSON(http.StatusNotFound, responses.ErrorResponse{
                Error:   "not_found",
                Message: "user not found",
            })
        case errors.IsUnauthorizedError(err):
            c.JSON(http.StatusUnauthorized, responses.ErrorResponse{
                Error:   "unauthorized",
                Message: err.Error(),
            })
        default:
            c.JSON(http.StatusInternalServerError, responses.ErrorResponse{
                Error:   "internal_error",
                Message: "an unexpected error occurred",
            })
        }
        return
    }
    
    // Convert to response DTO
    response := responses.LoginResponse{
        User: responses.UserResponse{
            ID:       result.User.ID,
            Email:    result.User.Email,
            Username: result.User.Username,
            Role:     result.User.Role,
        },
        Token:        result.Token,
        RefreshToken: result.RefreshToken,
        ExpiresIn:    result.ExpiresIn.Seconds(),
    }
    
    h.logger.Info("user logged in successfully", "user_id", result.User.ID)
    c.JSON(http.StatusOK, response)
}
```

### 2. Request/Response DTOs

```go
// internal/interfaces/http/requests/auth_requests.go
package requests

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Username string `json:"username" binding:"required,alphanum,min=3,max=30"`
    Password string `json:"password" binding:"required,min=8"`
}

// internal/interfaces/http/responses/auth_responses.go
package responses

import "time"

type LoginResponse struct {
    User         UserResponse `json:"user"`
    Token        string       `json:"token"`
    RefreshToken string       `json:"refresh_token"`
    ExpiresIn    int64        `json:"expires_in"`
}

type UserResponse struct {
    ID       string `json:"id"`
    Email    string `json:"email"`
    Username string `json:"username"`
    Role     string `json:"role"`
}

type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
}
```

## 📋 Phase 5: Dependency Injection

### 1. Container Setup

```go
// internal/shared/container/container.go
package container

import (
    "go-next/internal/domain/repositories"
    "go-next/internal/infrastructure/database/postgres"
    "go-next/internal/application/usecases/auth"
    "go-next/internal/interfaces/http/handlers"
    "go-next/pkg/database"
    "go-next/pkg/jwt"
    "go-next/pkg/crypto"
    "go-next/pkg/logger"
)

type Container struct {
    // Repositories
    UserRepository repositories.UserRepository
    
    // Use Cases
    LoginUseCase    *auth.LoginUseCase
    RegisterUseCase *auth.RegisterUseCase
    
    // Handlers
    AuthHandler *handlers.AuthHandler
    
    // Services
    JWTService    jwt.Service
    CryptoService crypto.Service
    Logger        logger.Logger
}

func NewContainer(db *database.DB) *Container {
    // Initialize services
    jwtService := jwt.NewService()
    cryptoService := crypto.NewService()
    logger := logger.NewLogger()
    
    // Initialize repositories
    userRepo := postgres.NewUserRepository(db)
    
    // Initialize use cases
    loginUseCase := auth.NewLoginUseCase(userRepo, jwtService, cryptoService)
    registerUseCase := auth.NewRegisterUseCase(userRepo, cryptoService)
    
    // Initialize handlers
    authHandler := handlers.NewAuthHandler(loginUseCase, registerUseCase, logger)
    
    return &Container{
        UserRepository: userRepo,
        LoginUseCase:   loginUseCase,
        RegisterUseCase: registerUseCase,
        AuthHandler:    authHandler,
        JWTService:     jwtService,
        CryptoService:  cryptoService,
        Logger:         logger,
    }
}
```

## 📋 Phase 6: Testing

### 1. Unit Tests

```go
// tests/unit/usecases/auth/login_test.go
package auth

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    "go-next/internal/application/usecases/auth"
    "go-next/internal/application/dto"
    "go-next/internal/domain/entities"
    "go-next/tests/mocks"
)

func TestLoginUseCase_Execute(t *testing.T) {
    // Arrange
    mockUserRepo := &mocks.MockUserRepository{}
    mockJWTService := &mocks.MockJWTService{}
    mockCryptoService := &mocks.MockCryptoService{}
    
    useCase := auth.NewLoginUseCase(mockUserRepo, mockJWTService, mockCryptoService)
    
    // Create test user
    user, _ := entities.NewUser("test@example.com", "testuser", "password123")
    
    // Setup mocks
    mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
    mockCryptoService.On("VerifyPassword", "password123", user.Password.Value()).Return(true)
    mockJWTService.On("GenerateToken", user.ID.Value(), user.Role.String()).Return("jwt-token", nil)
    mockJWTService.On("GenerateRefreshToken", user.ID.Value()).Return("refresh-token", nil)
    
    // Act
    req := dto.LoginRequest{
        Email:    "test@example.com",
        Password: "password123",
    }
    
    result, err := useCase.Execute(context.Background(), req)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "jwt-token", result.Token)
    assert.Equal(t, "refresh-token", result.RefreshToken)
    assert.Equal(t, user.ID.Value(), result.User.ID)
    assert.Equal(t, user.Email.Value(), result.User.Email)
    
    mockUserRepo.AssertExpectations(t)
    mockJWTService.AssertExpectations(t)
    mockCryptoService.AssertExpectations(t)
}
```

### 2. Integration Tests

```go
// tests/integration/auth_test.go
package integration

import (
    "context"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    
    "go-next/internal/shared/container"
    "go-next/tests/fixtures"
)

func TestAuthIntegration(t *testing.T) {
    // Setup test database
    db := fixtures.SetupTestDB(t)
    defer fixtures.CleanupTestDB(t, db)
    
    // Setup container
    container := container.NewContainer(db)
    
    // Setup router
    gin.SetMode(gin.TestMode)
    router := gin.New()
    
    // Register routes
    authRoutes := router.Group("/api/auth")
    {
        authRoutes.POST("/login", container.AuthHandler.Login)
        authRoutes.POST("/register", container.AuthHandler.Register)
    }
    
    // Test login
    t.Run("login_success", func(t *testing.T) {
        // Create test user first
        fixtures.CreateTestUser(t, db, "test@example.com", "password123")
        
        // Make request
        reqBody := `{"email":"test@example.com","password":"password123"}`
        req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(reqBody))
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        // Assert response
        assert.Equal(t, http.StatusOK, w.Code)
        
        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        
        assert.Contains(t, response, "token")
        assert.Contains(t, response, "user")
    })
}
```

## 🚀 Benefits of This Structure

### ✅ **Clean Architecture**
- **Separation of Concerns**: Each layer has a specific responsibility
- **Dependency Inversion**: High-level modules don't depend on low-level modules
- **Testability**: Easy to mock dependencies and test in isolation

### ✅ **Domain-Driven Design**
- **Rich Domain Models**: Business logic encapsulated in domain entities
- **Value Objects**: Immutable objects representing domain concepts
- **Repository Pattern**: Abstract data access layer

### ✅ **Scalability**
- **Modular Design**: Easy to add new features without affecting existing code
- **Microservices Ready**: Structure supports future microservices migration
- **Performance**: Optimized for high-performance applications

### ✅ **Maintainability**
- **Clear Dependencies**: Easy to understand and modify
- **Consistent Patterns**: Standardized approach across the codebase
- **Documentation**: Self-documenting code structure

This implementation provides a solid foundation for a scalable, maintainable, and testable Go application following clean architecture principles. 