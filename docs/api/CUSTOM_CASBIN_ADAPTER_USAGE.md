# Custom Casbin GORM Adapter Usage

## Overview

This document shows how to use the custom Casbin GORM adapter that utilizes the existing `CasbinRule` model from `backend/internal/models/casbin.go`.

## Key Features

### 1. Uses Existing Model
The custom adapter uses your existing `CasbinRule` model:

```go
type CasbinRule struct {
    BaseModel
    Ptype string `gorm:"column:ptype"`
    V0    string `gorm:"column:v0"`
    V1    string `gorm:"column:v1"`
    V2    string `gorm:"column:v2"`
    V3    string `gorm:"column:v3"`
    V4    string `gorm:"column:v4"`
    V5    string `gorm:"column:v5"`
}
```

### 2. Automatic Database Migration
The `CasbinRule` model is automatically migrated when the application starts:

```go
// In backend/pkg/database/database.go
func AutoMigrate() error {
    err := DB.AutoMigrate(
        &models.User{},
        &models.Post{},
        &models.Comment{},
        &models.Category{},
        &models.Role{},
        &models.UserRole{},
        &models.CasbinRule{}, // Added this line
        &models.Media{},
        &models.Mediable{},
        &models.Content{},
        &models.Notification{},
    )
    return err
}
```

## Usage Examples

### 1. Basic Initialization

```go
// In your service initialization
func InitCasbin() error {
    // RBAC model configuration
    mconf := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

    // Create custom GORM adapter using existing CasbinRule model
    adapter := casbin.NewCustomGormAdapter(database.DB, "casbin_rule")

    // Create enforcer with custom adapter
    e, err := adapter.CreateEnforcer(mconf)
    if err != nil {
        return fmt.Errorf("failed to create Casbin enforcer: %w", err)
    }

    Enforcer = e
    return nil
}
```

### 2. Database Operations

```go
// Get policy count
count, err := adapter.GetPolicyCount()
if err != nil {
    log.Printf("Failed to get policy count: %v", err)
}

// Get role count
roleCount, err := adapter.GetRoleCount()
if err != nil {
    log.Printf("Failed to get role count: %v", err)
}

// Get policies by role
policyCount, err := adapter.GetPolicyCountByRole("admin")
if err != nil {
    log.Printf("Failed to get policy count for role: %v", err)
}
```

### 3. Backup and Restore

```go
// Backup all policies
policies, err := adapter.BackupPolicies()
if err != nil {
    log.Printf("Failed to backup policies: %v", err)
}

// Restore policies from backup
err = adapter.RestorePolicies(policies)
if err != nil {
    log.Printf("Failed to restore policies: %v", err)
}
```

### 4. Policy Management

```go
// Add a new policy
err := adapter.AddPolicy("", "p", []string{"admin", "/api/users", "GET"})
if err != nil {
    log.Printf("Failed to add policy: %v", err)
}

// Remove a policy
err = adapter.RemovePolicy("", "p", []string{"admin", "/api/users", "GET"})
if err != nil {
    log.Printf("Failed to remove policy: %v", err)
}

// Add role for user
err = adapter.AddPolicy("", "g", []string{"user-uuid-1", "admin"})
if err != nil {
    log.Printf("Failed to add role for user: %v", err)
}
```

### 5. Database Statistics

```go
// Get comprehensive database statistics
stats, err := adapter.GetDatabaseStats()
if err != nil {
    log.Printf("Failed to get database stats: %v", err)
}

// Access statistics
totalPolicies := stats["total_policies"].(int64)
totalRoles := stats["total_roles"].(int64)
tableName := stats["table_name"].(string)
tableSize := stats["table_size_bytes"].(int64)
adapterType := stats["adapter_type"].(string)
```

## API Integration

### 1. Handler Usage

```go
// In your handlers
type casbinAdapterHandler struct {
    adapter *casbin.CustomGormAdapter
}

func NewCasbinAdapterHandler(adapter *casbin.CustomGormAdapter) CasbinAdapterHandler {
    return &casbinAdapterHandler{
        adapter: adapter,
    }
}

func (h *casbinAdapterHandler) GetDatabaseStats(c *gin.Context) {
    stats, err := h.adapter.GetDatabaseStats()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database statistics"})
        return
    }
    c.JSON(http.StatusOK, stats)
}
```

### 2. Route Registration

```go
// In your routes
func RegisterCasbinAdapterRoutes(api *gin.RouterGroup, adapterHandler controllers.CasbinAdapterHandler) {
    adapter := api.Group("/casbin/adapter")
    {
        adapter.GET("/stats", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/adapter", "GET"), adapterHandler.GetDatabaseStats)
        adapter.GET("/policies/count", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/adapter", "GET"), adapterHandler.GetPolicyCount)
        adapter.GET("/roles/count", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/adapter", "GET"), adapterHandler.GetRoleCount)
        // ... more routes
    }
}
```

## Database Schema

The custom adapter creates and uses the following table structure:

```sql
CREATE TABLE casbin_rule (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    ptype VARCHAR(10) NOT NULL,
    v0 VARCHAR(256),
    v1 VARCHAR(256),
    v2 VARCHAR(256),
    v3 VARCHAR(256),
    v4 VARCHAR(256),
    v5 VARCHAR(256)
);
```

## Field Descriptions

| Field | Type | Description | Usage |
|-------|------|-------------|-------|
| `id` | BIGSERIAL | Primary key | Auto-generated |
| `created_at` | TIMESTAMP | Creation time | GORM soft delete |
| `updated_at` | TIMESTAMP | Update time | GORM auto-update |
| `deleted_at` | TIMESTAMP | Deletion time | GORM soft delete |
| `ptype` | VARCHAR(10) | Policy type | 'p' for policies, 'g' for roles |
| `v0` | VARCHAR(256) | Subject/Role | Role name or user ID |
| `v1` | VARCHAR(256) | Object/User | Resource path or user ID |
| `v2` | VARCHAR(256) | Action | HTTP method or action |
| `v3-v5` | VARCHAR(256) | Additional fields | Reserved for future use |

## Benefits

### 1. Consistency
- Uses your existing model structure
- Maintains consistency with your codebase
- Leverages existing GORM patterns

### 2. Integration
- Seamlessly integrates with existing database migrations
- Uses your existing BaseModel for timestamps and soft deletes
- Compatible with your existing GORM setup

### 3. Flexibility
- Custom table name support
- Full CRUD operations
- Backup and restore functionality
- Database statistics and monitoring

### 4. Performance
- Direct GORM queries
- Optimized for your database schema
- Efficient policy loading and saving

## Error Handling

```go
// Handle adapter creation errors
adapter := casbin.NewCustomGormAdapter(database.DB, "casbin_rule")
if adapter == nil {
    return fmt.Errorf("failed to create custom adapter")
}

// Handle enforcer creation errors
e, err := adapter.CreateEnforcer(modelConfig)
if err != nil {
    return fmt.Errorf("failed to create enforcer: %w", err)
}

// Handle policy operation errors
err = adapter.AddPolicy("", "p", []string{"admin", "/api/users", "GET"})
if err != nil {
    log.Printf("Policy operation failed: %v", err)
}
```

## Testing

```go
// Test adapter creation
func TestCustomAdapter(t *testing.T) {
    adapter := casbin.NewCustomGormAdapter(testDB, "casbin_rule")
    assert.NotNil(t, adapter)
}

// Test enforcer creation
func TestEnforcerCreation(t *testing.T) {
    adapter := casbin.NewCustomGormAdapter(testDB, "casbin_rule")
    e, err := adapter.CreateEnforcer(modelConfig)
    assert.NoError(t, err)
    assert.NotNil(t, e)
}

// Test policy operations
func TestPolicyOperations(t *testing.T) {
    adapter := casbin.NewCustomGormAdapter(testDB, "casbin_rule")
    
    // Add policy
    err := adapter.AddPolicy("", "p", []string{"admin", "/api/users", "GET"})
    assert.NoError(t, err)
    
    // Get policy count
    count, err := adapter.GetPolicyCount()
    assert.NoError(t, err)
    assert.Equal(t, int64(1), count)
}
```

## Conclusion

The custom Casbin GORM adapter provides a seamless integration with your existing `CasbinRule` model while offering all the functionality of the standard Casbin adapter. It maintains consistency with your codebase architecture and provides additional features for policy management and monitoring. 