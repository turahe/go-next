# 🔄 Migration Script for Project Structure Improvement

## 📋 Phase 1: Foundation Setup

### Step 1: Create New Directory Structure

```bash
# Create new directory structure
mkdir -p backend/internal/domain/{entities,valueobjects,repositories,services}
mkdir -p backend/internal/application/{usecases,dto,interfaces}
mkdir -p backend/internal/infrastructure/{database,cache,storage,messaging,external}
mkdir -p backend/internal/interfaces/{http,grpc,websocket}
mkdir -p backend/internal/shared/{errors,utils,constants,container}
mkdir -p backend/cmd/{server,cli}
mkdir -p backend/api/{openapi,proto}
mkdir -p backend/scripts
mkdir -p backend/docs/{api,architecture,deployment}
mkdir -p backend/deployments/{docker,kubernetes,terraform}
mkdir -p backend/tests/{unit,integration,e2e,fixtures}
mkdir -p backend/tools/{swagger,migrations,codegen}
```

### Step 2: Move Existing Files

```bash
# Move existing files to new structure
# Models -> Domain Entities
mv backend/internal/models/* backend/internal/domain/entities/

# Services -> Domain Services (temporarily)
mv backend/internal/services/* backend/internal/domain/services/

# Controllers -> HTTP Handlers
mv backend/internal/http/controllers/* backend/internal/interfaces/http/handlers/

# Middleware -> HTTP Middleware
mv backend/internal/http/middleware/* backend/internal/interfaces/http/middleware/

# Routes -> HTTP Routes
mv backend/internal/routers/* backend/internal/interfaces/http/routes/

# Requests/Responses -> HTTP layer
mv backend/internal/http/requests/* backend/internal/interfaces/http/requests/
mv backend/internal/http/responses/* backend/internal/interfaces/http/responses/

# Pkg -> Public packages (keep as is)
# Keep backend/pkg/ as is for now
```

### Step 3: Update Import Paths

```bash
# Create a script to update all import paths
cat > update_imports.sh << 'EOF'
#!/bin/bash

# Update import paths in Go files
find . -name "*.go" -type f -exec sed -i 's|go-next/internal/models|go-next/internal/domain/entities|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|go-next/internal/services|go-next/internal/domain/services|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|go-next/internal/http/controllers|go-next/internal/interfaces/http/handlers|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|go-next/internal/http/middleware|go-next/internal/interfaces/http/middleware|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|go-next/internal/routers|go-next/internal/interfaces/http/routes|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|go-next/internal/http/requests|go-next/internal/interfaces/http/requests|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|go-next/internal/http/responses|go-next/internal/interfaces/http/responses|g' {} \;

echo "Import paths updated successfully!"
EOF

chmod +x update_imports.sh
./update_imports.sh
```

## 📋 Phase 2: Domain Layer Implementation

### Step 1: Create Domain Entities

```bash
# Create base entity
cat > backend/internal/domain/entities/base.go << 'EOF'
package entities

import (
    "time"
    "go-next/internal/domain/valueobjects"
)

type BaseEntity struct {
    ID        valueobjects.UUID `json:"id"`
    CreatedAt time.Time         `json:"created_at"`
    UpdatedAt time.Time         `json:"updated_at"`
    DeletedAt *time.Time        `json:"deleted_at,omitempty"`
}

func (e *BaseEntity) SetID(id valueobjects.UUID) {
    e.ID = id
}

func (e *BaseEntity) SetTimestamps() {
    now := time.Now()
    if e.CreatedAt.IsZero() {
        e.CreatedAt = now
    }
    e.UpdatedAt = now
}
EOF
```

### Step 2: Create Value Objects

```bash
# Create UUID value object
cat > backend/internal/domain/valueobjects/uuid.go << 'EOF'
package valueobjects

import (
    "github.com/google/uuid"
)

type UUID struct {
    value string
}

func NewUUID() UUID {
    return UUID{value: uuid.New().String()}
}

func NewUUIDFromString(s string) (UUID, error) {
    _, err := uuid.Parse(s)
    if err != nil {
        return UUID{}, err
    }
    return UUID{value: s}, nil
}

func (u UUID) String() string {
    return u.value
}

func (u UUID) Value() string {
    return u.value
}

func (u UUID) IsZero() bool {
    return u.value == ""
}
EOF

# Create Email value object
cat > backend/internal/domain/valueobjects/email.go << 'EOF'
package valueobjects

import (
    "errors"
    "strings"
    "regexp"
)

type Email struct {
    value string
}

var (
    ErrInvalidEmail = errors.New("invalid email format")
    ErrEmptyEmail   = errors.New("email cannot be empty")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func NewEmail(email string) (Email, error) {
    if strings.TrimSpace(email) == "" {
        return Email{}, ErrEmptyEmail
    }
    
    if !emailRegex.MatchString(email) {
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
EOF
```

### Step 3: Create Repository Interfaces

```bash
# Create base repository interface
cat > backend/internal/domain/repositories/base_repository.go << 'EOF'
package repositories

import (
    "context"
    "go-next/internal/domain/entities"
)

type BaseRepository[T entities.BaseEntity] interface {
    Create(ctx context.Context, entity *T) error
    GetByID(ctx context.Context, id string) (*T, error)
    Update(ctx context.Context, entity *T) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter interface{}) ([]*T, error)
    Count(ctx context.Context, filter interface{}) (int64, error)
}
EOF
```

## 📋 Phase 3: Application Layer Implementation

### Step 1: Create Use Case Base

```bash
# Create base use case
cat > backend/internal/application/usecases/base_usecase.go << 'EOF'
package usecases

import (
    "context"
    "go-next/internal/shared/errors"
    "go-next/pkg/logger"
)

type BaseUseCase struct {
    logger logger.Logger
}

func NewBaseUseCase(logger logger.Logger) BaseUseCase {
    return BaseUseCase{logger: logger}
}

func (uc *BaseUseCase) LogError(ctx context.Context, err error, message string, fields ...map[string]interface{}) {
    uc.logger.Error(message, append([]map[string]interface{}{{"error": err}}, fields...)...)
}

func (uc *BaseUseCase) LogInfo(ctx context.Context, message string, fields ...map[string]interface{}) {
    uc.logger.Info(message, fields...)
}
EOF
```

### Step 2: Create DTO Base

```bash
# Create base DTO
cat > backend/internal/application/dto/base_dto.go << 'EOF'
package dto

import (
    "go-next/pkg/validator"
)

type BaseDTO struct{}

func (d *BaseDTO) Validate() error {
    return validator.Validate(d)
}

type PaginationDTO struct {
    Page     int `json:"page" validate:"min=1"`
    PageSize int `json:"page_size" validate:"min=1,max=100"`
}

func (p *PaginationDTO) GetOffset() int {
    return (p.Page - 1) * p.PageSize
}

func (p *PaginationDTO) GetLimit() int {
    return p.PageSize
}
EOF
```

## 📋 Phase 4: Infrastructure Layer Implementation

### Step 1: Create Database Models

```bash
# Create base database model
cat > backend/internal/infrastructure/database/postgres/models/base.go << 'EOF'
package models

import (
    "time"
    "gorm.io/gorm"
)

type BaseModel struct {
    ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    CreatedAt time.Time      `gorm:"not null"`
    UpdatedAt time.Time      `gorm:"not null"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
    if b.CreatedAt.IsZero() {
        b.CreatedAt = time.Now()
    }
    if b.UpdatedAt.IsZero() {
        b.UpdatedAt = time.Now()
    }
    return nil
}

func (b *BaseModel) BeforeUpdate(tx *gorm.DB) error {
    b.UpdatedAt = time.Now()
    return nil
}
EOF
```

### Step 2: Create Repository Implementation

```bash
# Create base repository implementation
cat > backend/internal/infrastructure/database/postgres/base_repository.go << 'EOF'
package postgres

import (
    "context"
    "database/sql"
    "go-next/internal/shared/errors"
    "go-next/pkg/database"
    "gorm.io/gorm"
)

type BaseRepository[T interface{}] struct {
    db *database.DB
}

func NewBaseRepository(db *database.DB) *BaseRepository[T] {
    return &BaseRepository[T]{db: db}
}

func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
    if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
        return errors.NewInternalError("failed to create entity", err)
    }
    return nil
}

func (r *BaseRepository[T]) GetByID(ctx context.Context, id string) (*T, error) {
    var entity T
    
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.NewNotFoundError("entity not found")
        }
        return nil, errors.NewInternalError("failed to get entity", err)
    }
    
    return &entity, nil
}

func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
    if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
        return errors.NewInternalError("failed to update entity", err)
    }
    return nil
}

func (r *BaseRepository[T]) Delete(ctx context.Context, id string) error {
    if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&T{}).Error; err != nil {
        return errors.NewInternalError("failed to delete entity", err)
    }
    return nil
}
EOF
```

## 📋 Phase 5: Interface Layer Implementation

### Step 1: Create HTTP Handler Base

```bash
# Create base HTTP handler
cat > backend/internal/interfaces/http/handlers/base_handler.go << 'EOF'
package handlers

import (
    "net/http"
    "go-next/internal/interfaces/http/responses"
    "go-next/internal/shared/errors"
    "go-next/pkg/logger"
)

type BaseHandler struct {
    logger logger.Logger
}

func NewBaseHandler(logger logger.Logger) BaseHandler {
    return BaseHandler{logger: logger}
}

func (h *BaseHandler) HandleError(c *gin.Context, err error) {
    h.logger.Error("handler error", "error", err, "path", c.Request.URL.Path)
    
    switch {
    case errors.IsValidationError(err):
        c.JSON(http.StatusBadRequest, responses.ErrorResponse{
            Error:   "validation_error",
            Message: err.Error(),
        })
    case errors.IsNotFoundError(err):
        c.JSON(http.StatusNotFound, responses.ErrorResponse{
            Error:   "not_found",
            Message: err.Error(),
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
}

func (h *BaseHandler) HandleSuccess(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, responses.SuccessResponse{
        Data: data,
    })
}
EOF
```

### Step 2: Create Response Base

```bash
# Create base response
cat > backend/internal/interfaces/http/responses/base_response.go << 'EOF'
package responses

type SuccessResponse struct {
    Data interface{} `json:"data"`
}

type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
}

type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    Total      int64 `json:"total"`
    TotalPages int   `json:"total_pages"`
}
EOF
```

## 📋 Phase 6: Shared Utilities Implementation

### Step 1: Create Error Handling

```bash
# Create error types
cat > backend/internal/shared/errors/errors.go << 'EOF'
package errors

import (
    "fmt"
    "strings"
)

type ErrorType string

const (
    ErrorTypeValidation   ErrorType = "validation_error"
    ErrorTypeNotFound     ErrorType = "not_found"
    ErrorTypeUnauthorized ErrorType = "unauthorized"
    ErrorTypeForbidden    ErrorType = "forbidden"
    ErrorTypeConflict     ErrorType = "conflict"
    ErrorTypeInternal     ErrorType = "internal_error"
)

type AppError struct {
    Type    ErrorType              `json:"type"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
    Cause   error                  `json:"-"`
}

func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Cause
}

func NewValidationError(message string, cause error) *AppError {
    return &AppError{
        Type:    ErrorTypeValidation,
        Message: message,
        Cause:   cause,
    }
}

func NewNotFoundError(message string) *AppError {
    return &AppError{
        Type:    ErrorTypeNotFound,
        Message: message,
    }
}

func NewUnauthorizedError(message string) *AppError {
    return &AppError{
        Type:    ErrorTypeUnauthorized,
        Message: message,
    }
}

func NewForbiddenError(message string) *AppError {
    return &AppError{
        Type:    ErrorTypeForbidden,
        Message: message,
    }
}

func NewConflictError(message string) *AppError {
    return &AppError{
        Type:    ErrorTypeConflict,
        Message: message,
    }
}

func NewInternalError(message string, cause error) *AppError {
    return &AppError{
        Type:    ErrorTypeInternal,
        Message: message,
        Cause:   cause,
    }
}

// Helper functions to check error types
func IsValidationError(err error) bool {
    return isErrorType(err, ErrorTypeValidation)
}

func IsNotFoundError(err error) bool {
    return isErrorType(err, ErrorTypeNotFound)
}

func IsUnauthorizedError(err error) bool {
    return isErrorType(err, ErrorTypeUnauthorized)
}

func IsForbiddenError(err error) bool {
    return isErrorType(err, ErrorTypeForbidden)
}

func IsConflictError(err error) bool {
    return isErrorType(err, ErrorTypeConflict)
}

func IsInternalError(err error) bool {
    return isErrorType(err, ErrorTypeInternal)
}

func isErrorType(err error, errorType ErrorType) bool {
    if err == nil {
        return false
    }
    
    if appErr, ok := err.(*AppError); ok {
        return appErr.Type == errorType
    }
    
    // Check wrapped errors
    return strings.Contains(err.Error(), string(errorType))
}
EOF
```

### Step 2: Create Constants

```bash
# Create application constants
cat > backend/internal/shared/constants/roles.go << 'EOF'
package constants

type Role string

const (
    RoleAdmin    Role = "admin"
    RoleEditor   Role = "editor"
    RoleModerator Role = "moderator"
    RoleUser     Role = "user"
)

func (r Role) String() string {
    return string(r)
}

func (r Role) IsValid() bool {
    switch r {
    case RoleAdmin, RoleEditor, RoleModerator, RoleUser:
        return true
    default:
        return false
    }
}

func RoleFromString(s string) Role {
    role := Role(s)
    if role.IsValid() {
        return role
    }
    return RoleUser // default role
}
EOF

cat > backend/internal/shared/constants/status.go << 'EOF'
package constants

type Status string

const (
    StatusActive   Status = "active"
    StatusInactive Status = "inactive"
    StatusPending  Status = "pending"
    StatusBanned   Status = "banned"
)

func (s Status) String() string {
    return string(s)
}

func (s Status) IsValid() bool {
    switch s {
    case StatusActive, StatusInactive, StatusPending, StatusBanned:
        return true
    default:
        return false
    }
}

func StatusFromString(s string) Status {
    status := Status(s)
    if status.IsValid() {
        return status
    }
    return StatusActive // default status
}
EOF
```

## 📋 Phase 7: Testing Setup

### Step 1: Create Test Utilities

```bash
# Create test utilities
cat > backend/tests/utils/test_utils.go << 'EOF'
package utils

import (
    "context"
    "testing"
    "time"
    
    "go-next/pkg/database"
    "go-next/internal/shared/container"
)

type TestSuite struct {
    DB        *database.DB
    Container *container.Container
    Ctx       context.Context
}

func NewTestSuite(t *testing.T) *TestSuite {
    // Setup test database
    db := setupTestDB(t)
    
    // Setup container
    container := container.NewContainer(db)
    
    return &TestSuite{
        DB:        db,
        Container: container,
        Ctx:       context.Background(),
    }
}

func (ts *TestSuite) Cleanup() {
    if ts.DB != nil {
        cleanupTestDB(ts.DB)
    }
}

func setupTestDB(t *testing.T) *database.DB {
    // Implementation for test database setup
    // This would typically use a test database or in-memory database
    return nil // Placeholder
}

func cleanupTestDB(db *database.DB) {
    // Implementation for test database cleanup
}

func WaitForCondition(condition func() bool, timeout time.Duration) bool {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if condition() {
            return true
        }
        time.Sleep(10 * time.Millisecond)
    }
    return false
}
EOF
```

## 📋 Phase 8: Build and Test

### Step 1: Create Build Script

```bash
# Create build script
cat > backend/scripts/build.sh << 'EOF'
#!/bin/bash

set -e

echo "Building Go-Next Backend..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

# Set build variables
VERSION=$(git describe --tags --always --dirty)
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH=$(git rev-parse HEAD)

# Build flags
LDFLAGS="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.CommitHash=$COMMIT_HASH"

# Build for different platforms
echo "Building for current platform..."
go build -ldflags "$LDFLAGS" -o bin/server cmd/server/main.go

echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/server-linux-amd64 cmd/server/main.go

echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/server-windows-amd64.exe cmd/server/main.go

echo "Build completed successfully!"
echo "Binaries created in bin/ directory"
EOF

chmod +x backend/scripts/build.sh
```

### Step 2: Create Test Script

```bash
# Create test script
cat > backend/scripts/test.sh << 'EOF'
#!/bin/bash

set -e

echo "Running Go-Next Backend Tests..."

# Run unit tests
echo "Running unit tests..."
go test -v ./tests/unit/...

# Run integration tests
echo "Running integration tests..."
go test -v ./tests/integration/...

# Run e2e tests
echo "Running e2e tests..."
go test -v ./tests/e2e/...

# Generate coverage report
echo "Generating coverage report..."
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

echo "Tests completed successfully!"
echo "Coverage report: coverage.html"
EOF

chmod +x backend/scripts/test.sh
```

## 📋 Phase 9: Documentation

### Step 1: Create API Documentation

```bash
# Create API documentation template
cat > backend/docs/api/README.md << 'EOF'
# API Documentation

## Overview

This document describes the REST API endpoints for the Go-Next backend.

## Base URL

- Development: `http://localhost:8080`
- Production: `https://api.go-next.com`

## Authentication

Most endpoints require authentication using JWT tokens.

### Headers

```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

## Endpoints

### Authentication

#### POST /api/auth/login

Login with email and password.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "username": "username",
      "role": "user"
    },
    "token": "jwt_token",
    "refresh_token": "refresh_token",
    "expires_in": 86400
  }
}
```

### Users

#### GET /api/v1/users

Get list of users (admin only).

**Headers:**
```
Authorization: Bearer <admin_token>
```

**Query Parameters:**
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 10, max: 100)
- `search`: Search term
- `role`: Filter by role
- `status`: Filter by status

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "email": "user@example.com",
      "username": "username",
      "role": "user",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 100,
    "total_pages": 10
  }
}
```

## Error Responses

All endpoints may return error responses in the following format:

```json
{
  "error": "error_type",
  "message": "Human readable error message"
}
```

### Error Types

- `validation_error`: Request validation failed
- `not_found`: Resource not found
- `unauthorized`: Authentication required
- `forbidden`: Insufficient permissions
- `conflict`: Resource conflict
- `internal_error`: Server error

## Rate Limiting

API requests are rate limited to:
- 100 requests per minute for authenticated users
- 10 requests per minute for unauthenticated users

## Pagination

List endpoints support pagination with the following parameters:
- `page`: Page number (1-based)
- `page_size`: Items per page (1-100)

## Filtering

List endpoints support filtering with query parameters specific to each resource.

## Sorting

List endpoints support sorting with the `sort` parameter:
- `sort=field_name` (ascending)
- `sort=-field_name` (descending)
EOF
```

## 📋 Phase 10: Deployment

### Step 1: Create Docker Configuration

```bash
# Create production Dockerfile
cat > backend/deployments/docker/Dockerfile.prod << 'EOF'
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy configuration files
COPY --from=builder /app/.env.example .env

# Create necessary directories
RUN mkdir -p log storage data

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
EOF

# Create docker-compose for production
cat > backend/deployments/docker/docker-compose.prod.yml << 'EOF'
version: '3.8'

services:
  backend:
    build:
      context: ../..
      dockerfile: deployments/docker/Dockerfile.prod
    container_name: go-next-backend-prod
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=go_next_prod
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    volumes:
      - ./logs:/root/log
      - ./storage:/root/storage
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped
    networks:
      - backend-network

  postgres:
    image: postgres:15-alpine
    container_name: go-next-postgres-prod
    environment:
      - POSTGRES_DB=go_next_prod
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres -d go_next_prod"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
    networks:
      - backend-network

  redis:
    image: redis:7-alpine
    container_name: go-next-redis-prod
    command: redis-server --requirepass ${REDIS_PASSWORD}
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
    networks:
      - backend-network

volumes:
  postgres_data:
  redis_data:

networks:
  backend-network:
    driver: bridge
EOF
```

## 🚀 Final Steps

### Step 1: Update Main Application

```bash
# Update main.go to use new structure
cat > backend/cmd/server/main.go << 'EOF'
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "go-next/internal/shared/container"
    "go-next/pkg/config"
    "go-next/pkg/database"
    "go-next/pkg/logger"
)

var (
    Version     = "dev"
    BuildTime   = "unknown"
    CommitHash  = "unknown"
)

func main() {
    // Load configuration
    cfg := config.GetConfig()
    
    // Initialize logger
    logger := logger.NewLogger()
    logger.Info("Starting Go-Next Backend", 
        "version", Version,
        "build_time", BuildTime,
        "commit_hash", CommitHash,
    )
    
    // Initialize database
    db, err := database.NewConnection(cfg.Database)
    if err != nil {
        logger.Fatal("Failed to connect to database", "error", err)
    }
    defer db.Close()
    
    // Initialize container
    container := container.NewContainer(db)
    
    // Start server
    server := NewServer(cfg, container, logger)
    
    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        if err := server.Start(); err != nil {
            logger.Fatal("Server failed to start", "error", err)
        }
    }()
    
    <-quit
    logger.Info("Shutting down server...")
    
    if err := server.Shutdown(); err != nil {
        logger.Error("Error during server shutdown", "error", err)
    }
    
    logger.Info("Server stopped")
}
EOF
```

### Step 2: Run Migration

```bash
# Execute the migration
echo "Starting project structure migration..."

# Run all migration steps
cd backend

# Phase 1: Create directories
echo "Phase 1: Creating directory structure..."
mkdir -p internal/domain/{entities,valueobjects,repositories,services}
mkdir -p internal/application/{usecases,dto,interfaces}
mkdir -p internal/infrastructure/{database,cache,storage,messaging,external}
mkdir -p internal/interfaces/{http,grpc,websocket}
mkdir -p internal/shared/{errors,utils,constants,container}
mkdir -p cmd/{server,cli}
mkdir -p api/{openapi,proto}
mkdir -p scripts
mkdir -p docs/{api,architecture,deployment}
mkdir -p deployments/{docker,kubernetes,terraform}
mkdir -p tests/{unit,integration,e2e,fixtures}
mkdir -p tools/{swagger,migrations,codegen}

echo "Directory structure created successfully!"

# Phase 2: Move files
echo "Phase 2: Moving existing files..."
# (Execute the move commands from Step 2)

# Phase 3: Update imports
echo "Phase 3: Updating import paths..."
# (Execute the import update script)

# Phase 4: Build and test
echo "Phase 4: Building and testing..."
go mod tidy
go build ./...
go test ./...

echo "Migration completed successfully!"
echo "New project structure is ready!"
```

## 📊 Migration Checklist

- [ ] **Phase 1**: Create new directory structure
- [ ] **Phase 2**: Move existing files to new locations
- [ ] **Phase 3**: Update import paths
- [ ] **Phase 4**: Create domain layer (entities, value objects, repositories)
- [ ] **Phase 5**: Create application layer (use cases, DTOs)
- [ ] **Phase 6**: Create infrastructure layer (database, external services)
- [ ] **Phase 7**: Create interface layer (HTTP handlers, middleware)
- [ ] **Phase 8**: Create shared utilities (errors, constants, container)
- [ ] **Phase 9**: Set up testing framework
- [ ] **Phase 10**: Create build and deployment scripts
- [ ] **Phase 11**: Update documentation
- [ ] **Phase 12**: Run tests and validate

## 🎯 Success Metrics

After migration, you should have:

✅ **Clean Architecture**: Clear separation of concerns
✅ **Domain-Driven Design**: Rich domain models with business logic
✅ **Testability**: Easy to write unit and integration tests
✅ **Scalability**: Modular design ready for microservices
✅ **Maintainability**: Consistent patterns and clear dependencies
✅ **Performance**: Optimized for high-performance applications

This migration script provides a comprehensive approach to restructuring your Go-Next project following clean architecture principles and modern Go best practices. 