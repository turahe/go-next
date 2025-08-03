# Casbin GORM Adapter Implementation

## Overview

This document describes the enhanced Casbin GORM adapter implementation that provides advanced database management capabilities for Casbin policies and role assignments.

## Architecture

### Components

1. **CasbinAdapter** (`backend/pkg/casbin/adapter.go`)
   - Enhanced GORM adapter with configuration options
   - Database statistics and monitoring
   - Backup and restore functionality
   - Policy validation and management

2. **CasbinAdapterHandler** (`backend/internal/http/controllers/casbin_adapter_handlers.go`)
   - RESTful API for adapter management
   - Database statistics endpoints
   - Backup and restore operations
   - Configuration management

3. **CasbinAdapterRoutes** (`backend/internal/routers/v1/casbin_adapter.go`)
   - Route definitions for adapter management
   - Protected endpoints with JWT and Casbin middleware

## Configuration

### AdapterConfig Structure

```go
type AdapterConfig struct {
    TableName    string // Custom table name (default: "casbin_rule")
    AutoMigrate  bool   // Whether to auto-migrate the table
    BatchSize    int    // Batch size for operations (default: 1000)
    MaxRetries   int    // Maximum retries for database operations
    EnableCache  bool   // Enable adapter-level caching
    CacheTimeout int    // Cache timeout in seconds
}
```

### Default Configuration

```go
func DefaultAdapterConfig() *AdapterConfig {
    return &AdapterConfig{
        TableName:    "casbin_rule",
        AutoMigrate:  true,
        BatchSize:    1000,
        MaxRetries:   3,
        EnableCache:  true,
        CacheTimeout: 300, // 5 minutes
    }
}
```

## Database Schema

### CasbinRule Table Structure

The GORM adapter automatically creates the `casbin_rule` table with the following structure:

```sql
CREATE TABLE casbin_rule (
    id BIGSERIAL PRIMARY KEY,
    ptype VARCHAR(10) NOT NULL,
    v0 VARCHAR(256),
    v1 VARCHAR(256),
    v2 VARCHAR(256),
    v3 VARCHAR(256),
    v4 VARCHAR(256),
    v5 VARCHAR(256)
);
```

### Field Descriptions

| Field | Type | Description | Usage |
|-------|------|-------------|-------|
| `id` | BIGSERIAL | Primary key | Auto-generated |
| `ptype` | VARCHAR(10) | Policy type | 'p' for policies, 'g' for roles |
| `v0` | VARCHAR(256) | Subject/Role | Role name or user ID |
| `v1` | VARCHAR(256) | Object/User | Resource path or user ID |
| `v2` | VARCHAR(256) | Action | HTTP method or action |
| `v3-v5` | VARCHAR(256) | Additional fields | Reserved for future use |

## API Endpoints

### Database Statistics

#### Get Database Stats
```http
GET /api/v1/casbin/adapter/stats
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "total_policies": 45,
  "total_roles": 12,
  "table_name": "casbin_rule",
  "table_size_bytes": 8192,
  "auto_migrate": true,
  "batch_size": 1000,
  "max_retries": 3,
  "cache_enabled": true,
  "cache_timeout": 300
}
```

#### Get Policy Count
```http
GET /api/v1/casbin/adapter/policies/count
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "policy_count": 45
}
```

#### Get Role Count
```http
GET /api/v1/casbin/adapter/roles/count
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "role_count": 12
}
```

#### Get Policy Count by Role
```http
GET /api/v1/casbin/adapter/policies/count/role?role=admin
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "role": "admin",
  "policy_count": 15
}
```

#### Get User Role Count
```http
GET /api/v1/casbin/adapter/users/roles/count?user_id=user-uuid-1
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "user_id": "user-uuid-1",
  "role_count": 2
}
```

### Configuration Management

#### Get Adapter Configuration
```http
GET /api/v1/casbin/adapter/config
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "table_name": "casbin_rule",
  "auto_migrate": true,
  "batch_size": 1000,
  "max_retries": 3,
  "enable_cache": true,
  "cache_timeout": 300
}
```

### Backup and Restore

#### Backup Policies
```http
GET /api/v1/casbin/adapter/backup
Authorization: Bearer <jwt_token>
```

**Response:**
```json
[
  ["admin", "/api/users", "GET"],
  ["admin", "/api/users", "POST"],
  ["admin", "/api/users", "PUT"],
  ["admin", "/api/users", "DELETE"],
  ["user-uuid-1", "admin"],
  ["user-uuid-2", "editor"]
]
```

#### Restore Policies
```http
POST /api/v1/casbin/adapter/restore
Authorization: Bearer <jwt_token>
Content-Type: application/json

[
  ["admin", "/api/users", "GET"],
  ["admin", "/api/users", "POST"],
  ["user-uuid-1", "admin"]
]
```

**Response:**
```json
{
  "message": "Policies restored successfully"
}
```

### Clear Operations (Admin Only)

#### Clear All Policies and Roles
```http
DELETE /api/v1/casbin/adapter/clear/all
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "message": "All policies and roles cleared successfully"
}
```

#### Clear Policies Only
```http
DELETE /api/v1/casbin/adapter/clear/policies
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "message": "Policies cleared successfully"
}
```

#### Clear Roles Only
```http
DELETE /api/v1/casbin/adapter/clear/roles
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "message": "Roles cleared successfully"
}
```

## Usage Examples

### Creating an Adapter

```go
// Create adapter with default configuration
adapter, err := casbin.NewCasbinAdapter(nil)
if err != nil {
    log.Fatal(err)
}

// Create adapter with custom configuration
config := &casbin.AdapterConfig{
    TableName:    "custom_casbin_rules",
    AutoMigrate:  true,
    BatchSize:    500,
    MaxRetries:   5,
    EnableCache:  true,
    CacheTimeout: 600,
}

adapter, err := casbin.NewCasbinAdapter(config)
if err != nil {
    log.Fatal(err)
}
```

### Creating an Enforcer

```go
// Create enforcer with adapter
modelConfig := `
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

enforcer, err := adapter.CreateEnforcer(modelConfig)
if err != nil {
    log.Fatal(err)
}
```

### Database Operations

```go
// Get database statistics
stats, err := adapter.GetDatabaseStats()
if err != nil {
    log.Printf("Failed to get stats: %v", err)
}

// Get policy count
count, err := adapter.GetPolicyCount()
if err != nil {
    log.Printf("Failed to get policy count: %v", err)
}

// Backup policies
policies, err := adapter.BackupPolicies()
if err != nil {
    log.Printf("Failed to backup policies: %v", err)
}

// Restore policies
err = adapter.RestorePolicies(policies)
if err != nil {
    log.Printf("Failed to restore policies: %v", err)
}
```

### Policy Validation

```go
// Validate a policy before insertion
policy := []string{"admin", "/api/users", "GET"}
err := adapter.ValidatePolicy(policy)
if err != nil {
    log.Printf("Invalid policy: %v", err)
}
```

## Performance Optimization

### Caching

The adapter supports configurable caching to improve performance:

```go
config := &casbin.AdapterConfig{
    EnableCache:  true,
    CacheTimeout: 300, // 5 minutes
}
```

### Batch Operations

Configure batch size for bulk operations:

```go
config := &casbin.AdapterConfig{
    BatchSize: 1000, // Process 1000 records at a time
}
```

### Retry Logic

Configure retry attempts for database operations:

```go
config := &casbin.AdapterConfig{
    MaxRetries: 5, // Retry failed operations up to 5 times
}
```

## Security Considerations

### Access Control

All adapter endpoints are protected with:
- JWT authentication
- Casbin authorization
- Admin-only access for destructive operations

### Data Validation

- Policy validation before insertion
- Input sanitization
- SQL injection prevention through GORM

### Backup Security

- Backup data includes all policies and roles
- Restore operations clear existing data first
- Validation of backup data format

## Monitoring and Debugging

### Database Statistics

Monitor database performance and usage:

```bash
# Get comprehensive statistics
curl -H "Authorization: Bearer <token>" \
     http://localhost:8080/api/v1/casbin/adapter/stats

# Get policy count
curl -H "Authorization: Bearer <token>" \
     http://localhost:8080/api/v1/casbin/adapter/policies/count

# Get role count
curl -H "Authorization: Bearer <token>" \
     http://localhost:8080/api/v1/casbin/adapter/roles/count
```

### Backup and Recovery

```bash
# Create backup
curl -H "Authorization: Bearer <token>" \
     http://localhost:8080/api/v1/casbin/adapter/backup \
     -o casbin_backup.json

# Restore from backup
curl -H "Authorization: Bearer <token>" \
     -H "Content-Type: application/json" \
     -d @casbin_backup.json \
     http://localhost:8080/api/v1/casbin/adapter/restore
```

## Error Handling

### Common Errors

1. **Database Connection Issues**
   ```go
   // Check database connectivity
   if err := adapter.GetDatabaseStats(); err != nil {
       log.Printf("Database connection issue: %v", err)
   }
   ```

2. **Policy Validation Errors**
   ```go
   // Validate policies before insertion
   if err := adapter.ValidatePolicy(policy); err != nil {
       log.Printf("Policy validation failed: %v", err)
   }
   ```

3. **Backup/Restore Errors**
   ```go
   // Handle backup errors
   policies, err := adapter.BackupPolicies()
   if err != nil {
       log.Printf("Backup failed: %v", err)
   }
   ```

## Integration with Existing Casbin Service

The enhanced adapter integrates seamlessly with the existing Casbin service:

```go
// In your existing Casbin service
func InitCasbin() error {
    // Create enhanced adapter
    adapter, err := casbin.NewCasbinAdapter(nil)
    if err != nil {
        return err
    }

    // Create enforcer with adapter
    enforcer, err := adapter.CreateEnforcer(modelConfig)
    if err != nil {
        return err
    }

    // Use enforcer as before
    Enforcer = enforcer
    return nil
}
```

## Best Practices

### Configuration

1. **Use Custom Table Names** for multi-tenant applications
2. **Enable Auto-Migration** for development environments
3. **Configure Appropriate Batch Sizes** based on data volume
4. **Set Reasonable Retry Limits** for production environments

### Monitoring

1. **Regular Statistics Checks** to monitor database growth
2. **Backup Before Major Changes** to ensure data safety
3. **Monitor Policy Counts** to detect unusual activity
4. **Track Role Assignments** for security auditing

### Performance

1. **Enable Caching** for frequently accessed policies
2. **Use Appropriate Batch Sizes** for bulk operations
3. **Monitor Database Performance** regularly
4. **Optimize Query Patterns** for large datasets

## Conclusion

The enhanced Casbin GORM adapter provides comprehensive database management capabilities for Casbin policies and role assignments. With features like statistics monitoring, backup and restore functionality, and configurable performance options, it offers enterprise-grade capabilities while maintaining simplicity and ease of use.

The adapter seamlessly integrates with existing Casbin implementations while providing additional management and monitoring capabilities through a RESTful API. 