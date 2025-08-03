# RBAC Model Configuration

This document explains the RBAC (Role-Based Access Control) model configuration used in the application.

## File Location

The RBAC model configuration is stored in:
```
backend/config/rbac_model.conf
```

## Model Structure

The RBAC model uses a domain-based approach with the following structure:

### Request Definition
```
[request_definition]
r = sub, dom, obj, act
```

**Parameters:**
- `sub` (subject): The user or role making the request
- `dom` (domain): The domain/context where the request is made
- `obj` (object): The resource being accessed
- `act` (action): The action being performed (GET, POST, PUT, DELETE, etc.)

### Policy Definition
```
[policy_definition]
p = sub, dom, obj, act
```

**Parameters:**
- `sub` (subject): The role that has permission
- `dom` (domain): The domain where the permission applies
- `obj` (object): The resource that can be accessed
- `act` (action): The allowed action

### Role Definition
```
[role_definition]
g = _, _, _
```

**Parameters:**
- First parameter: User ID
- Second parameter: Role name
- Third parameter: Domain

### Policy Effect
```
[policy_effect]
e = some(where (p.eft == allow))
```

This means that if any policy allows the action, the request is permitted.

### Matchers
```
[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && r.obj == p.obj && r.act == p.act
```

**Logic:**
1. `g(r.sub, p.sub, r.dom)`: Check if the user has the required role in the domain
2. `r.dom == p.dom`: Ensure the domain matches
3. `r.obj == p.obj`: Ensure the object/resource matches
4. `r.act == p.act`: Ensure the action matches

## Domain-Based Access Control

### Global Access
- Use `"*"` as the domain for global policies
- These policies apply across all domains

### Domain-Specific Access
- Use specific domain identifiers (e.g., organization IDs)
- Policies are scoped to specific domains
- Users can have different roles in different domains

## Example Policies

### Global Admin Policy
```
["admin", "*", "/api/users", "GET"]
```
- Role: admin
- Domain: * (global)
- Object: /api/users
- Action: GET

### Organization-Specific Policy
```
["editor", "org-123", "/api/posts", "POST"]
```
- Role: editor
- Domain: org-123 (specific organization)
- Object: /api/posts
- Action: POST

### Role Assignment
```
["user-456", "editor", "org-123"]
```
- User: user-456
- Role: editor
- Domain: org-123

## Usage in Code

### Loading the Model
```go
// Load RBAC model configuration from file
modelPath := filepath.Join("config", "rbac_model.conf")
mconf, err := os.ReadFile(modelPath)
if err != nil {
    return fmt.Errorf("failed to read RBAC model configuration: %w", err)
}

// Create enforcer with the model
e, err := casbin.NewEnforcer(string(mconf), adapter)
```

### Adding Policies
```go
// Add domain-specific policy
casbinService.AddPolicy("editor", "org-123", "/api/posts", "POST")

// Add global policy
casbinService.AddPolicy("admin", "*", "/api/users", "GET")
```

### Assigning Roles
```go
// Assign role in specific domain
casbinService.AddRoleForUser(userID, "editor", "org-123")

// Assign global role
casbinService.AddRoleForUser(userID, "admin", "*")
```

### Checking Permissions
```go
// Check domain-specific permission
allowed, err := casbinService.EnforceWithDomain(userID, "/api/posts", "POST", "org-123")

// Check global permission
allowed, err := casbinService.Enforce(userID, "/api/users", "GET")
```

## Benefits

1. **Multi-Tenancy**: Each organization can have isolated policies
2. **Role Flexibility**: Users can have different roles in different contexts
3. **Security**: Domain isolation prevents cross-domain access
4. **Scalability**: Easy to add new domains without affecting existing ones
5. **Maintainability**: Model configuration is externalized and easily modifiable

## Configuration Management

### Environment-Specific Models
You can create different model files for different environments:

- `config/rbac_model.conf` (default)
- `config/rbac_model_prod.conf` (production)
- `config/rbac_model_dev.conf` (development)

### Model Validation
The model configuration is validated when the application starts. Common validation checks:

1. File exists and is readable
2. Required sections are present
3. Syntax is correct
4. Matcher logic is valid

## Troubleshooting

### Common Issues

1. **File Not Found**: Ensure the model file exists in the correct location
2. **Syntax Errors**: Check the model file for proper formatting
3. **Permission Denied**: Verify file permissions
4. **Invalid Matcher**: Ensure the matcher logic is correct

### Debugging

Enable debug logging to see detailed information about policy evaluation:

```go
logger.SetLevel(logger.DebugLevel)
```

## Related Files

- `backend/internal/services/casbin.go`: Casbin service implementation
- `backend/internal/http/middleware/casbin_middleware.go`: Authorization middleware
- `backend/internal/http/controllers/casbin_handlers.go`: Policy management API
- `docs/api/CASBIN_RBAC_IMPLEMENTATION.md`: General RBAC documentation 