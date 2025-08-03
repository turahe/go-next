# Organization-User Many-to-Many Relationship

## Overview

This document describes the implementation of a many-to-many relationship between organizations and users, allowing users to belong to multiple organizations and organizations to have multiple users.

## Database Schema

### OrganizationUser Join Table

The `organization_users` table serves as the join table for the many-to-many relationship:

```sql
CREATE TABLE organization_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role VARCHAR(50) DEFAULT 'member',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    UNIQUE(organization_id, user_id)
);
```

### Model Relationships

#### Organization Model
```go
type Organization struct {
    BaseModelWithOrdering
    Name        string           `json:"name"`
    Slug        string           `json:"slug"`
    Description string           `json:"description"`
    Code        string           `json:"code"`
    Type        OrganizationType `json:"type"`
    ParentID    uuid.UUID        `json:"parent_id"`
    Children    []Organization   `json:"children,omitempty"`
    Users       []User           `json:"users,omitempty" gorm:"many2many:organization_users;constraint:OnDelete:CASCADE"`
    // ... other fields
}
```

#### User Model
```go
type User struct {
    BaseModel
    Username      string         `json:"username"`
    Email         string         `json:"email"`
    Password      string         `json:"-"`
    Phone         string         `json:"phone"`
    // ... other fields
    Organizations []Organization `json:"organizations,omitempty" gorm:"many2many:organization_users;constraint:OnDelete:CASCADE"`
}
```

#### OrganizationUser Join Model
```go
type OrganizationUser struct {
    BaseModel
    OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index"`
    UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
    Role           string    `json:"role" gorm:"size:50;default:'member'"`
    IsActive       bool      `json:"is_active" gorm:"default:true"`
    
    // Relationships
    Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
    User         User         `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
```

## Service Layer

### OrganizationService Methods

The `OrganizationService` provides comprehensive methods for managing the many-to-many relationship:

#### Core Relationship Methods
- `AddUserToOrganization(userID, orgID uuid.UUID) error` - Add user to organization
- `RemoveUserFromOrganization(userID, orgID uuid.UUID) error` - Remove user from organization
- `GetUsersInOrganization(orgID uuid.UUID) ([]models.User, error)` - Get all users in organization
- `GetOrganizationsForUser(userID uuid.UUID) ([]models.Organization, error)` - Get all organizations for user

#### Role Management Methods
- `GetUserRoleInOrganization(userID, orgID uuid.UUID) (string, error)` - Get user's role in organization
- `UpdateUserRoleInOrganization(userID, orgID uuid.UUID, role string) error` - Update user's role
- `DeactivateUserInOrganization(userID, orgID uuid.UUID) error` - Soft delete user from organization
- `GetOrganizationUsersWithRoles(orgID uuid.UUID) ([]models.OrganizationUser, error)` - Get users with roles

## API Endpoints

### Organization User Management

#### Add User to Organization
```http
POST /api/organizations/{id}/users/{user_id}
Authorization: Bearer <token>
```

#### Remove User from Organization
```http
DELETE /api/organizations/{id}/users/{user_id}
Authorization: Bearer <token>
```

#### Get User Role in Organization
```http
GET /api/organizations/{id}/users/{user_id}/role
Authorization: Bearer <token>
```

#### Update User Role in Organization
```http
PUT /api/organizations/{id}/users/{user_id}/role
Authorization: Bearer <token>
Content-Type: application/json

{
    "role": "admin"
}
```

#### Get Organization Users with Roles
```http
GET /api/organizations/{id}/users-with-roles
Authorization: Bearer <token>
```

### Response Examples

#### Get Organization Users with Roles
```json
{
    "message": "Organization users with roles retrieved successfully",
    "data": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440000",
            "organization_id": "550e8400-e29b-41d4-a716-446655440001",
            "user_id": "550e8400-e29b-41d4-a716-446655440002",
            "role": "admin",
            "is_active": true,
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-01T00:00:00Z",
            "user": {
                "id": "550e8400-e29b-41d4-a716-446655440002",
                "username": "john_doe",
                "email": "john@example.com",
                "phone": "+1234567890"
            }
        }
    ],
    "status": 200,
    "timestamp": "2024-01-01T00:00:00Z"
}
```

## Features

### 1. Role-Based Access Control
- Each user can have different roles within different organizations
- Default role is "member"
- Roles can be updated dynamically
- Supports organization-specific permissions via Casbin

### 2. Soft Delete Support
- Users can be deactivated in organizations without permanent deletion
- `is_active` field controls visibility
- Maintains historical data

### 3. Duplicate Prevention
- Unique constraint on `(organization_id, user_id)`
- Service layer checks for existing relationships before adding

### 4. Cascading Operations
- When an organization is deleted, all user relationships are removed
- When a user is deleted, all organization relationships are removed

### 5. Integration with Casbin
- Automatic role assignment in Casbin when users are added to organizations
- Organization-aware authorization checks
- Policy management per organization

## Usage Examples

### Adding a User to an Organization
```go
orgService := services.NewOrganizationService()
err := orgService.AddUserToOrganization(userID, orgID)
if err != nil {
    // Handle error
}
```

### Getting User's Organizations
```go
organizations, err := orgService.GetOrganizationsForUser(userID)
if err != nil {
    // Handle error
}
for _, org := range organizations {
    fmt.Printf("User belongs to: %s\n", org.Name)
}
```

### Updating User Role
```go
err := orgService.UpdateUserRoleInOrganization(userID, orgID, "admin")
if err != nil {
    // Handle error
}
```

### Getting Organization Members with Roles
```go
orgUsers, err := orgService.GetOrganizationUsersWithRoles(orgID)
if err != nil {
    // Handle error
}
for _, orgUser := range orgUsers {
    fmt.Printf("User: %s, Role: %s\n", orgUser.User.Username, orgUser.Role)
}
```

## Database Migrations

The `OrganizationUser` model is automatically migrated when the application starts:

```go
func AutoMigrate() error {
    err := DB.AutoMigrate(
        &models.User{},
        &models.Organization{},
        &models.OrganizationUser{}, // Join table
        // ... other models
    )
    return err
}
```

## Security Considerations

1. **Authorization**: All endpoints require JWT authentication
2. **Role Validation**: Ensure only valid roles can be assigned
3. **Organization Access**: Users should only access organizations they belong to
4. **Audit Trail**: All changes are timestamped and tracked

## Performance Optimizations

1. **Indexes**: Proper indexing on `organization_id` and `user_id`
2. **Eager Loading**: Use `Preload()` for related data
3. **Pagination**: For large organizations, implement pagination
4. **Caching**: Consider caching frequently accessed organization data

## Future Enhancements

1. **Bulk Operations**: Add endpoints for bulk user management
2. **Role Templates**: Predefined role sets for organizations
3. **Hierarchical Roles**: Support for role inheritance
4. **Audit Logging**: Detailed audit trail for all operations
5. **Notification System**: Notify users when added/removed from organizations 