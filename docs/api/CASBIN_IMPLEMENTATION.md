# Casbin RBAC Implementation with Database Adapter

## Overview

This document describes the implementation of Role-Based Access Control (RBAC) using Casbin with a GORM database adapter in the Go-Next application.

## Architecture

### Components

1. **CasbinService** (`backend/internal/services/casbin.go`)
   - Manages Casbin enforcer and policies
   - Provides CRUD operations for policies and roles
   - Handles user-role assignments

2. **CasbinMiddleware** (`backend/internal/http/middleware/casbin_middleware.go`)
   - Validates user permissions for protected routes
   - Integrates with JWT authentication
   - Provides granular access control

3. **CasbinController** (`backend/internal/http/controllers/casbin_handlers.go`)
   - RESTful API for policy management
   - User role assignment endpoints
   - Policy filtering and querying

4. **Database Adapter** (`github.com/casbin/gorm-adapter/v3`)
   - Stores policies and role assignments in PostgreSQL
   - Provides persistence across application restarts
   - Enables policy management via database operations

## RBAC Model

### Policy Structure
```
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
```

### Default Roles and Permissions

#### Admin Role
- **Full Access**: All endpoints and operations
- **Policies**: Complete CRUD operations on users, roles, posts, categories, comments, media
- **Casbin Management**: Full access to Casbin policy management

#### Editor Role
- **Content Management**: Posts, categories, comments, media
- **Policies**: Create, read, update, delete content
- **No Access**: User management, role management

#### Moderator Role
- **Comment Moderation**: Read, create, update, delete comments
- **Content Reading**: Read posts and categories
- **No Access**: Content creation, user management

#### User Role
- **Basic Access**: Read content, create comments
- **Policies**: Read posts, categories, media; create comments
- **No Access**: Content management, user management

#### Guest Role
- **Read-Only Access**: View public content
- **Policies**: Read posts and categories
- **No Access**: Any write operations

## API Endpoints

### Policy Management

#### Get All Policies
```http
GET /api/v1/casbin/policies
Authorization: Bearer <jwt_token>
```

#### Add Policy
```http
POST /api/v1/casbin/policies
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "subject": "admin",
  "object": "/api/users",
  "action": "GET"
}
```

#### Remove Policy
```http
DELETE /api/v1/casbin/policies
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "subject": "admin",
  "object": "/api/users",
  "action": "GET"
}
```

#### Get Filtered Policies
```http
GET /api/v1/casbin/policies/filtered?field_index=0&field_values=admin
Authorization: Bearer <jwt_token>
```

### User Role Management

#### Get User Roles
```http
GET /api/v1/casbin/users/{user_id}/roles
Authorization: Bearer <jwt_token>
```

#### Assign Role to User
```http
POST /api/v1/casbin/users/{user_id}/roles
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "role": "admin"
}
```

#### Remove Role from User
```http
DELETE /api/v1/casbin/users/{user_id}/roles
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "role": "admin"
}
```

## Database Schema

### Casbin Tables (Auto-generated)

The GORM adapter automatically creates the following tables:

#### `casbin_rule`
Stores individual policies and role assignments.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGINT | Primary key |
| `ptype` | VARCHAR | Policy type (p for policy, g for role) |
| `v0` | VARCHAR | Subject/Role |
| `v1` | VARCHAR | Object/User |
| `v2` | VARCHAR | Action |
| `v3` | VARCHAR | Additional field (unused) |
| `v4` | VARCHAR | Additional field (unused) |
| `v5` | VARCHAR | Additional field (unused) |

### Example Data

#### Policies (ptype = 'p')
```
admin, /api/users, GET
admin, /api/users, POST
admin, /api/users, PUT
admin, /api/users, DELETE
editor, /api/posts, GET
editor, /api/posts, POST
user, /api/posts, GET
guest, /api/posts, GET
```

#### Role Assignments (ptype = 'g')
```
user-uuid-1, admin
user-uuid-2, editor
user-uuid-3, user
```

## Integration with Authentication

### JWT Integration
- JWT middleware extracts user ID from token
- Casbin middleware uses user ID to check permissions
- User roles are retrieved from Casbin storage

### Automatic Role Assignment
- New users are automatically assigned the "user" role
- Role assignment happens in both database and Casbin
- Transaction ensures data consistency

## Usage Examples

### Protecting Routes

```go
// In route definition
roles.GET("", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/roles", "GET"), roleHandler.GetRoles)
```

### Checking Permissions in Code

```go
casbinService := services.NewCasbinService()
allowed, err := casbinService.Enforce(userID, "/api/users", "GET")
if err != nil {
    // Handle error
}
if !allowed {
    // Handle access denied
}
```

### Adding Custom Policies

```go
casbinService := services.NewCasbinService()
err := casbinService.AddPolicy("custom_role", "/api/custom", "POST")
```

## Security Considerations

### Policy Validation
- All policies are validated before storage
- Malicious policy injection is prevented
- Input sanitization is implemented

### Role Hierarchy
- Admin role has highest privileges
- Role inheritance can be implemented
- Granular permissions per endpoint

### Audit Trail
- All policy changes are logged
- User role assignments are tracked
- Failed access attempts are monitored

## Performance Optimization

### Caching
- Casbin enforcer caches policies in memory
- Database queries are minimized
- Policy evaluation is optimized

### Database Optimization
- Indexes on frequently queried columns
- Connection pooling for database adapter
- Efficient policy storage format

## Monitoring and Debugging

### Logging
- Policy changes are logged with details
- Access denied events are recorded
- Performance metrics are tracked

### Debugging Tools
- Policy inspection endpoints
- Role assignment verification
- Permission testing utilities

## Migration and Deployment

### Database Migration
- Casbin tables are auto-created
- Default policies are initialized
- Existing data is preserved

### Environment Configuration
- Database connection for Casbin adapter
- Policy initialization on startup
- Role assignment during user registration

## Future Enhancements

### Planned Features
- Role hierarchy implementation
- Dynamic policy loading
- Policy templates
- Advanced permission models

### Integration Opportunities
- LDAP integration for enterprise
- OAuth2 role mapping
- Multi-tenant role isolation
- Policy versioning and rollback

## Troubleshooting

### Common Issues

1. **Permission Denied**
   - Check user role assignment
   - Verify policy exists for endpoint
   - Ensure JWT token is valid

2. **Database Connection Issues**
   - Verify database connectivity
   - Check Casbin adapter configuration
   - Review database permissions

3. **Policy Not Applied**
   - Restart application to reload policies
   - Check policy format and syntax
   - Verify database storage

### Debug Commands

```bash
# Check Casbin policies
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/casbin/policies

# Check user roles
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/casbin/users/{user_id}/roles
```

## Conclusion

The Casbin RBAC implementation provides a robust, scalable, and secure access control system for the Go-Next application. With database persistence, comprehensive API management, and seamless integration with JWT authentication, it offers enterprise-grade authorization capabilities while maintaining simplicity and performance. 