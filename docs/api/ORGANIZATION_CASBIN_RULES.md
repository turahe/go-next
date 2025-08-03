# Organization-Based Casbin Rules Implementation

This document describes the implementation of organization-based access control using Casbin rules in the Go-Next application.

## Overview

The organization-based Casbin rules system extends the existing RBAC (Role-Based Access Control) to support multi-tenant organizations. Each organization can have its own set of policies and users, allowing for fine-grained access control across different organizational contexts.

## Architecture

### Casbin Model

The Casbin model has been extended to support organization context:

```conf
[request_definition]
r = sub, obj, act, org

[policy_definition]
p = sub, obj, act, org

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.org) && r.obj == p.obj && r.act == p.act && r.org == p.org
```

### Key Components

1. **Organization Model**: Defines organizational structure with hierarchical relationships
2. **Casbin Service**: Extended to support organization context in policies
3. **Organization Service**: Manages organizations and user assignments
4. **Organization Middleware**: Provides organization-aware authorization
5. **Organization Handlers**: REST API for organization management

## Database Schema

### Organizations Table

```sql
CREATE TABLE organizations (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description VARCHAR(500),
    code VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(100) NOT NULL,
    parent_id UUID REFERENCES organizations(id),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

### Organization Users Junction Table

```sql
CREATE TABLE organization_users (
    organization_id UUID REFERENCES organizations(id),
    user_id UUID REFERENCES users(id),
    PRIMARY KEY (organization_id, user_id)
);
```

### Casbin Rules Table

```sql
CREATE TABLE casbin_rule (
    id UUID PRIMARY KEY,
    ptype VARCHAR(10),
    v0 VARCHAR(256),
    v1 VARCHAR(256),
    v2 VARCHAR(256),
    v3 VARCHAR(256),
    v4 VARCHAR(256),
    v5 VARCHAR(256),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

## Implementation Details

### 1. Organization Model

The `Organization` model supports hierarchical structures and various organization types:

```go
type Organization struct {
    BaseModelWithOrdering
    Name        string           `json:"name" gorm:"not null;size:100"`
    Slug        string           `json:"slug" gorm:"uniqueIndex;not null;size:100"`
    Description string           `json:"description" gorm:"size:500"`
    Code        string           `json:"code" gorm:"uniqueIndex;not null;size:100"`
    Type        OrganizationType `json:"type" gorm:"not null;size:100"`
    ParentID    uuid.UUID        `json:"parent_id" gorm:"type:uuid;index"`
    Children    []Organization   `json:"children,omitempty" gorm:"foreignKey:ParentID"`
    Users       []User           `json:"users,omitempty" gorm:"many2many:organization_users"`
}
```

### 2. Extended Casbin Service

The Casbin service has been extended to support organization context:

```go
// AddPolicy adds a new policy with organization context
func (cs *CasbinService) AddPolicy(subject, object, action, organization string) error

// AddRoleForUser adds a role for a user with organization context
func (cs *CasbinService) AddRoleForUser(userID uuid.UUID, role, organization string) error

// EnforceWithOrganization checks if a user has permission in a specific organization
func (cs *CasbinService) EnforceWithOrganization(userID uuid.UUID, object, action, organization string) (bool, error)
```

### 3. Organization Service

The organization service manages organizations and integrates with Casbin:

```go
// AddUserToOrganization adds a user to an organization
func (os *OrganizationService) AddUserToOrganization(userID, orgID uuid.UUID) error

// AddOrganizationPolicy adds a policy for an organization
func (os *OrganizationService) AddOrganizationPolicy(role, object, action, orgID string) error

// GetOrganizationPolicies gets all policies for an organization
func (os *OrganizationService) GetOrganizationPolicies(orgID string) ([][]string, error)
```

### 4. Organization-Aware Middleware

Two middleware options are available:

- `CasbinMiddleware`: Standard RBAC without organization context
- `CasbinMiddlewareWithOrganization`: Organization-aware authorization

```go
// Organization-aware middleware
func CasbinMiddlewareWithOrganization(obj string, act string) gin.HandlerFunc
```

## API Endpoints

### Organization Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/organizations` | Get all organizations |
| POST | `/api/organizations` | Create new organization |
| GET | `/api/organizations/:id` | Get organization by ID |
| PUT | `/api/organizations/:id` | Update organization |
| DELETE | `/api/organizations/:id` | Delete organization |

### Organization User Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/organizations/:id/users/:user_id` | Add user to organization |
| DELETE | `/api/organizations/:id/users/:user_id` | Remove user from organization |

### Organization Policies

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/organizations/:id/policies` | Get organization policies |

## Usage Examples

### 1. Creating an Organization

```bash
curl -X POST http://localhost:8080/api/organizations \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "slug": "acme-corp",
    "description": "A technology company",
    "code": "ACME",
    "type": "COMPANY"
  }'
```

### 2. Adding a User to an Organization

```bash
curl -X POST http://localhost:8080/api/organizations/{org_id}/users/{user_id} \
  -H "Authorization: Bearer <token>"
```

### 3. Adding Organization-Specific Policies

```go
// Add organization-specific policy
organizationService.AddOrganizationPolicy("editor", "/api/posts", "GET", orgID.String())
organizationService.AddOrganizationPolicy("editor", "/api/posts", "POST", orgID.String())
```

### 4. Using Organization-Aware Authorization

```go
// In your routes
r.GET("/api/posts", middleware.CasbinMiddlewareWithOrganization("/api/posts", "GET"), handler)
```

## Default Policies

The system initializes with default policies that include organization context:

### Admin Policies (Global)
```go
{"admin", "/api/users", "GET", "*"}
{"admin", "/api/roles", "POST", "*"}
{"admin", "/api/organizations", "GET", "*"}
```

### Editor Policies (Organization-Specific)
```go
{"editor", "/api/posts", "GET", "*"}
{"editor", "/api/posts", "POST", "*"}
{"editor", "/api/comments", "GET", "*"}
```

## Organization Types

The system supports various organization types:

- `COMPANY`: Standard company
- `COMPANY_HOLDING`: Holding company
- `COMPANY_SUBSIDIARY`: Subsidiary company
- `OUTLET`: Retail outlet
- `STORE`: Store
- `DEPARTMENT`: Department
- `DIVISION`: Division
- `INSTITUTION`: Educational or research institution
- `FOUNDATION`: Non-profit foundation
- `PARTNER`: Business partner

## Security Considerations

### 1. Organization Isolation

- Users can only access resources within their assigned organizations
- Policies are scoped to specific organizations
- Cross-organization access requires explicit permissions

### 2. Role Inheritance

- Global roles (admin) have access to all organizations
- Organization-specific roles only work within their assigned organization
- Role hierarchy can be implemented within organizations

### 3. Policy Management

- Organization policies are managed separately from global policies
- Policy changes are scoped to the specific organization
- Audit trails track policy modifications

## Integration with Existing Systems

### 1. User Registration

When users register, they can be automatically assigned to a default organization:

```go
// In auth_service.go Register method
if defaultOrgID != nil {
    organizationService.AddUserToOrganization(user.ID, *defaultOrgID)
}
```

### 2. Existing RBAC

The organization system extends rather than replaces the existing RBAC:

- Global roles continue to work across all organizations
- Organization-specific roles provide additional granularity
- Backward compatibility is maintained

### 3. API Compatibility

Existing APIs continue to work:

- Non-organization-aware endpoints use global policies
- Organization-aware endpoints use organization-specific policies
- Middleware can be mixed based on requirements

## Monitoring and Logging

### 1. Policy Changes

All policy modifications are logged:

```go
logger.Infof("Added policy for organization %s: %s", orgID, policy)
```

### 2. Access Attempts

Organization-aware access attempts are logged:

```go
logger.Infof("Access check for user %s in organization %s: %s", userID, orgID, result)
```

### 3. Organization Statistics

The system tracks organization usage:

- Number of users per organization
- Policy count per organization
- Access patterns per organization

## Best Practices

### 1. Organization Design

- Use hierarchical organization structures for complex enterprises
- Implement clear organization boundaries
- Consider organization size and complexity

### 2. Policy Management

- Start with broad policies and refine over time
- Use organization-specific policies for sensitive operations
- Regularly audit organization policies

### 3. User Management

- Assign users to appropriate organizations
- Implement organization transfer workflows
- Monitor organization membership changes

### 4. Performance

- Cache organization policies for frequently accessed resources
- Use database indexes on organization-related queries
- Monitor policy evaluation performance

## Troubleshooting

### Common Issues

1. **User not found in organization**: Check organization membership
2. **Policy not found**: Verify organization-specific policies exist
3. **Access denied**: Check both global and organization-specific policies

### Debug Commands

```go
// Check user's organizations
organizations, err := organizationService.GetOrganizationsForUser(userID)

// Check organization policies
policies, err := organizationService.GetOrganizationPolicies(orgID.String())

// Check user roles in organization
roles, err := casbinService.GetUserRolesInOrganization(userID, orgID.String())
```

## Future Enhancements

1. **Organization Hierarchies**: Support for nested organization permissions
2. **Cross-Organization Policies**: Policies that span multiple organizations
3. **Organization Templates**: Predefined organization structures
4. **Advanced Analytics**: Organization usage analytics and reporting
5. **API Rate Limiting**: Organization-specific rate limiting
6. **Audit Logging**: Comprehensive audit trails for organization changes

## Conclusion

The organization-based Casbin rules implementation provides a robust foundation for multi-tenant access control. It extends the existing RBAC system while maintaining backward compatibility and providing the flexibility needed for complex organizational structures.

The system is designed to be scalable, secure, and maintainable, with clear separation of concerns and comprehensive documentation for all components. 