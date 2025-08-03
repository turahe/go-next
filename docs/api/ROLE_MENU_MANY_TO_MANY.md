# Role-Menu Many-to-Many Relationship

This document describes the implementation of the many-to-many relationship between roles and menus in the system.

## Overview

The system implements a many-to-many relationship between roles and menus, allowing:
- Multiple roles to be assigned to a single menu
- Multiple menus to be assigned to a single role
- Flexible permission management for menu access

## Database Schema

### Junction Table
The relationship is managed through a junction table called `role_menus` with the following structure:
- `role_id` (UUID) - Foreign key to roles table
- `menu_id` (UUID) - Foreign key to menus table

### Model Relationships

#### Role Model
```go
type Role struct {
    BaseModel
    Name        string `json:"name" gorm:"uniqueIndex;not null;size:50"`
    Description string `json:"description" gorm:"size:255"`
    Users       []User `json:"users,omitempty" gorm:"many2many:user_roles;constraint:OnDelete:CASCADE"`
    Menus       []Menu `json:"menus,omitempty" gorm:"many2many:role_menus;constraint:OnDelete:CASCADE"`
}
```

#### Menu Model
```go
type Menu struct {
    BaseModelWithOrdering
    Name        string    `json:"name" gorm:"not null;size:50"`
    Description string    `json:"description" gorm:"size:255"`
    Icon        string    `json:"icon" gorm:"size:50"`
    URL         string    `json:"url" gorm:"size:255"`
    ParentID    uuid.UUID `json:"parent_id" gorm:"type:uuid;index"`
    Children    []Menu    `json:"children,omitempty" gorm:"foreignKey:ParentID"`
    Roles       []Role    `json:"roles,omitempty" gorm:"many2many:role_menus;constraint:OnDelete:CASCADE"`
}
```

## Service Layer

### RoleService Methods
- `AssignMenuToRole(roleID, menuID uuid.UUID) error` - Assign a menu to a role
- `RemoveMenuFromRole(roleID, menuID uuid.UUID) error` - Remove a menu from a role
- `GetRoleMenus(roleID uuid.UUID) ([]models.Menu, error)` - Get all menus for a role
- `GetMenuRoles(menuID uuid.UUID) ([]models.Role, error)` - Get all roles for a menu

### MenuService Methods
- `AssignRoleToMenu(menuID, roleID uuid.UUID) error` - Assign a role to a menu
- `RemoveRoleFromMenu(menuID, roleID uuid.UUID) error` - Remove a role from a menu
- `GetMenuRoles(menuID uuid.UUID) ([]models.Role, error)` - Get all roles for a menu
- `GetRoleMenus(roleID uuid.UUID) ([]models.Menu, error)` - Get all menus for a role

## API Endpoints

### Menu-Role Management

#### Get Menu Roles
```
GET /api/v1/menus/{id}/roles
```
Returns all roles assigned to a specific menu.

#### Assign Role to Menu
```
POST /api/v1/menus/{id}/roles
Content-Type: application/json

{
    "role_id": "uuid-of-role"
}
```
Assigns a role to a menu.

#### Remove Role from Menu
```
DELETE /api/v1/menus/{id}/roles
Content-Type: application/json

{
    "role_id": "uuid-of-role"
}
```
Removes a role from a menu.

### Role-Menu Management

#### Get Role Menus
```
GET /api/v1/roles/{id}/menus
```
Returns all menus assigned to a specific role.

#### Assign Menu to Role
```
POST /api/v1/roles/{id}/menus
Content-Type: application/json

{
    "menu_id": "uuid-of-menu"
}
```
Assigns a menu to a role.

#### Remove Menu from Role
```
DELETE /api/v1/roles/{id}/menus
Content-Type: application/json

{
    "menu_id": "uuid-of-menu"
}
```
Removes a menu from a role.

## Usage Examples

### Assigning a Role to a Menu
```go
// Using the service directly
err := menuService.AssignRoleToMenu(menuID, roleID)
if err != nil {
    // Handle error
}
```

### Getting All Menus for a Role
```go
menus, err := roleService.GetRoleMenus(roleID)
if err != nil {
    // Handle error
}
for _, menu := range menus {
    fmt.Printf("Menu: %s\n", menu.Name)
}
```

### Getting All Roles for a Menu
```go
roles, err := menuService.GetMenuRoles(menuID)
if err != nil {
    // Handle error
}
for _, role := range roles {
    fmt.Printf("Role: %s\n", role.Name)
}
```

## Database Migration

The many-to-many relationship is automatically created by GORM when the models are migrated. The junction table `role_menus` will be created with the appropriate foreign key constraints.

## Security Considerations

1. **Authorization**: All role-menu assignment endpoints require JWT authentication
2. **Validation**: Both role and menu existence are validated before assignment
3. **Duplicate Prevention**: The system checks for existing assignments to prevent duplicates
4. **Cascade Deletion**: When a role or menu is deleted, the relationships are automatically removed

## Error Handling

The system handles various error scenarios:
- Invalid UUID formats
- Non-existent roles or menus
- Database constraint violations
- Duplicate assignments

## Testing

Unit tests are provided in `backend/internal/models/role_menu_test.go` to verify:
- Model structure integrity
- Association field presence
- Table name correctness

Integration tests should be added to verify the complete workflow including database operations. 