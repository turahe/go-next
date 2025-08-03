# Database Documentation

This directory contains the Entity Relationship Diagram (ERD) and related documentation for the Go-Next application database.

## 📊 Files Overview

- **`erd.mmd`** - Mermaid diagram source file with complete database schema
- **`erd.svg`** - Vector graphics version of the ERD (generated from .mmd)
- **`convert_diagrams.bat`** - Windows batch script for converting Mermaid to SVG

## 🏗️ Database Architecture

## 🚀 Feature Menu

### Quick Navigation
- [👥 **User Management**](#user-management) - Authentication, RBAC, Organization Membership
- [🏢 **Organization Hierarchy**](#organization-hierarchy) - Nested Set Model, Flexible Structure
- [📝 **Content Management**](#content-management) - Posts, Comments, Categories, Content Blocks
- [📁 **Media Management**](#media-management) - File Storage, Polymorphic Attachments
- [🏷️ **Tagging System**](#tagging-system) - Flexible Tagging, Visual Organization
- [🔔 **Notification System**](#notification-system) - User Notifications, Priority Levels
- [📋 **Menu System**](#menu-system) - Hierarchical Navigation, Dynamic Menus
- [📋 **Menu CRUD Operations**](#menu-crud) - Complete Create, Read, Update, Delete functionality
- [🎯 **Design Patterns**](#design-patterns) - Nested Set, Polymorphic, Junction Tables
- [🔄 **Status Workflows**](#status-workflows) - Publication, Approval, Organization Status
- [📋 **Database Tables**](#database-tables) - Complete table reference
- [🛠️ **Usage & Development**](#usage) - Viewing, Converting, Modifying ERD

---

### Core Features

The Go-Next database is designed as a comprehensive content management system with the following key features:

#### 1. **User Management** {#user-management}
- **Authentication** - Email/phone verification system
- **Role-Based Access Control (RBAC)** - Flexible permission system using Casbin
- **Organization Membership** - Users can belong to multiple organizations with roles

#### 2. **Organization Hierarchy** {#organization-hierarchy}
- **Nested Set Model** - Efficient hierarchical queries for organizations
- **Flexible Structure** - Support for companies, departments, subsidiaries, etc.
- **Role Assignment** - Users have specific roles within each organization

#### 3. **Content Management** {#content-management}
- **Posts** - Main content with publication workflow (draft → published → archived)
- **Comments** - Threaded comments with approval system (pending → approved/rejected)
- **Categories** - Hierarchical content categorization
- **Content Blocks** - Polymorphic content system for flexible post structure

#### 4. **Media Management** {#media-management}
- **File Storage** - Support for multiple storage backends (local, S3, GCS)
- **Polymorphic Attachments** - Media can be attached to any entity
- **Metadata Tracking** - File size, dimensions, duration, hash for deduplication
- **Public/Private Access** - Control over file visibility

#### 5. **Tagging System** {#tagging-system}
- **Flexible Tagging** - Tags can be applied to any entity
- **Tag Categories** - General, category, feature, system tags
- **Visual Organization** - Color-coded tags for UI
- **Grouping** - Tags can be organized into groups

#### 6. **Notification System** {#notification-system}
- **User Notifications** - Personalized notification system
- **Priority Levels** - Low, normal, high, urgent priorities
- **Multiple Types** - Success, error, warning, info notifications
- **Read Status** - Track notification read status

#### 7. **Menu System** {#menu-system}
- **Hierarchical Navigation** - Nested menu structure with nested set model
- **Dynamic Menus** - Active/inactive menu items with status management
- **Icon Support** - Menu icons for visual organization and branding
- **URL Routing** - Navigation URL management and routing
- **CRUD Operations** - Complete Create, Read, Update, Delete functionality
- **Menu Permissions** - Role-based menu access control
- **Menu Ordering** - Sort order management for menu items
- **Parent-Child Relationships** - Hierarchical menu structure support

### 📋 Menu CRUD Operations {#menu-crud}

The menu system provides comprehensive CRUD operations for managing navigation structure:

#### **Create Operations**
```sql
-- Create a new menu item
INSERT INTO menus (
    name, slug, url, icon, parent_id, sort_order, 
    is_active, is_public, record_left, record_right, record_dept
) VALUES (
    'Dashboard', 'dashboard', '/dashboard', 'fas fa-tachometer-alt',
    NULL, 1, true, true, 1, 2, 0
);

-- Create child menu item
INSERT INTO menus (
    name, slug, url, icon, parent_id, sort_order,
    is_active, is_public, record_left, record_right, record_dept
) VALUES (
    'Analytics', 'analytics', '/dashboard/analytics', 'fas fa-chart-bar',
    1, 1, true, true, 2, 3, 1
);
```

#### **Read Operations**
```sql
-- Get all active menus
SELECT * FROM menus WHERE is_active = true ORDER BY sort_order;

-- Get menu hierarchy (nested set model)
SELECT * FROM menus 
WHERE record_left BETWEEN 1 AND 10 
ORDER BY record_left;

-- Get menu with children
SELECT m1.*, m2.name as child_name 
FROM menus m1 
LEFT JOIN menus m2 ON m2.parent_id = m1.id 
WHERE m1.is_active = true;
```

#### **Update Operations**
```sql
-- Update menu item
UPDATE menus SET 
    name = 'New Dashboard Name',
    url = '/new-dashboard',
    icon = 'fas fa-home',
    sort_order = 2,
    is_active = false
WHERE id = 1;

-- Reorder menu items
UPDATE menus SET sort_order = CASE 
    WHEN id = 1 THEN 3
    WHEN id = 2 THEN 1
    WHEN id = 3 THEN 2
END WHERE id IN (1, 2, 3);

-- Move menu to different parent
UPDATE menus SET 
    parent_id = 5,
    record_dept = 2
WHERE id = 3;
```

#### **Delete Operations**
```sql
-- Soft delete menu item
UPDATE menus SET 
    deleted_at = NOW(),
    deleted_by = 'user_id'
WHERE id = 1;

-- Delete menu with all children (cascade)
DELETE FROM menus 
WHERE record_left BETWEEN 1 AND 10;

-- Archive menu instead of delete
UPDATE menus SET 
    is_active = false,
    updated_at = NOW(),
    updated_by = 'user_id'
WHERE id = 1;
```

#### **Advanced Menu Operations**

##### **Menu Permissions**
```sql
-- Assign menu to specific roles
INSERT INTO menu_roles (menu_id, role_id) VALUES (1, 'admin');

-- Check user menu access
SELECT m.* FROM menus m
JOIN menu_roles mr ON m.id = mr.menu_id
JOIN user_roles ur ON mr.role_id = ur.role_id
WHERE ur.user_id = 'user_id' AND m.is_active = true;
```

##### **Menu Hierarchy Management**
```sql
-- Get menu tree structure
WITH RECURSIVE menu_tree AS (
    SELECT id, name, parent_id, record_dept, 0 as level
    FROM menus WHERE parent_id IS NULL AND is_active = true
    UNION ALL
    SELECT m.id, m.name, m.parent_id, m.record_dept, mt.level + 1
    FROM menus m
    JOIN menu_tree mt ON m.parent_id = mt.id
    WHERE m.is_active = true
)
SELECT * FROM menu_tree ORDER BY record_left;
```

##### **Menu Search and Filter**
```sql
-- Search menus by name
SELECT * FROM menus 
WHERE name ILIKE '%dashboard%' AND is_active = true;

-- Filter by menu type
SELECT * FROM menus 
WHERE menu_type = 'main' AND is_public = true;

-- Get menus by parent
SELECT * FROM menus 
WHERE parent_id = 1 AND is_active = true 
ORDER BY sort_order;
```

#### **Menu Validation Rules**
- **Name**: Required, 3-50 characters, unique within parent
- **Slug**: Required, URL-safe, unique within parent
- **URL**: Required, valid URL format
- **Icon**: Optional, valid icon class name
- **Parent ID**: Optional, must reference existing menu
- **Sort Order**: Required, integer >= 0
- **Status**: Boolean, defaults to active

#### **Menu Workflow States**
```
draft → active → inactive → archived
```

## 🎯 Design Patterns {#design-patterns}

### 1. **Nested Set Model**
Used for efficient hierarchical queries in:
- **Organizations** - Company hierarchy
- **Categories** - Content categorization
- **Comments** - Threaded discussions
- **Menus** - Navigation hierarchy

### 2. **Polymorphic Associations**
- **Comments** - Can attach to posts or other comments
- **Media** - Can attach to any entity via mediables
- **Tags** - Can be applied to any entity via tagged_entities
- **Content** - Flexible content blocks for posts

### 3. **Junction Tables**
- **`user_roles`** - Users ↔ Roles (many-to-many)
- **`organization_users`** - Organizations ↔ Users (many-to-many)
- **`mediables`** - Media ↔ Any Entity (polymorphic)
- **`tagged_entities`** - Tags ↔ Any Entity (polymorphic)

### 4. **Enum Types**
- **Status Fields** - Predefined values for workflow states
- **Type Fields** - Categorization for various entities
- **Priority Levels** - Notification priority system

### 5. **Soft Deletes**
- **Data Preservation** - Records are marked as deleted, not physically removed
- **Audit Trail** - Track who deleted what and when
- **Recovery** - Ability to restore deleted records

### 6. **Audit Trail**
- **Change Tracking** - `created_by`, `updated_by`, `deleted_by` fields
- **Timestamps** - `created_at`, `updated_at`, `deleted_at` tracking
- **Compliance** - Full audit trail for regulatory requirements

## 🔄 Status Workflows {#status-workflows}

### Post Publication Workflow
```
draft → published → archived
```

### Comment Approval Workflow
```
pending → approved/rejected
```

### Organization/Category Status
```
active/inactive
```

## 📋 Database Tables {#database-tables}

### Core Tables

| Table | Purpose | Key Features |
|-------|---------|--------------|
| `users` | User accounts and authentication | Email/phone verification, password hashing |
| `roles` | Role definitions for RBAC | Role-based permissions |
| `user_roles` | User-role assignments | Many-to-many relationship |
| `organizations` | Organizational hierarchy | Nested set model, flexible types |
| `organization_users` | Organization membership | Role assignment within organizations |

### Content Tables

| Table | Purpose | Key Features |
|-------|---------|--------------|
| `posts` | Main content articles | Publication workflow, view tracking |
| `comments` | User comments and discussions | Threaded comments, approval system |
| `categories` | Content categorization | Hierarchical structure |
| `contents` | Flexible content blocks | Polymorphic content system |

### Media Tables

| Table | Purpose | Key Features |
|-------|---------|--------------|
| `media` | File storage and metadata | Multiple storage backends, deduplication |
| `mediables` | Media attachments | Polymorphic relationships |

### System Tables

| Table | Purpose | Key Features |
|-------|---------|--------------|
| `tags` | Tag definitions | Color coding, type categorization |
| `tagged_entities` | Tag assignments | Polymorphic tagging |
| `notifications` | User notifications | Priority levels, read tracking |
| `casbin_rule` | RBAC policies | Flexible permission system |
| `menus` | Navigation system | Hierarchical menu structure |

## 🛠️ Usage {#usage}

### Viewing the ERD

1. **SVG File** - Open `erd.svg` in any web browser or image viewer
2. **Mermaid Source** - Edit `erd.mmd` for modifications
3. **Online Viewer** - Use Mermaid Live Editor for interactive viewing

### Converting Diagrams

#### Using the Batch Script
```bash
# Windows
convert_diagrams.bat
```

#### Using npx Directly
```bash
# Convert single file
npx @mermaid-js/mermaid-cli -i erd.mmd -o erd.svg

# Convert multiple files
for file in *.mmd; do
  npx @mermaid-js/mermaid-cli -i "$file" -o "${file%.mmd}.svg"
done
```

### Prerequisites

- **Node.js** - Required for npx commands
- **Internet Connection** - For downloading mermaid-cli package

## 🔧 Development

### Modifying the ERD

1. Edit `erd.mmd` file
2. Add descriptions to fields using quotes: `string field_name "Description"`
3. Regenerate SVG: `npx @mermaid-js/mermaid-cli -i erd.mmd -o erd.svg`

### Adding New Tables

1. Add table definition to `erd.mmd`
2. Include field descriptions for clarity
3. Define relationships with other tables
4. Regenerate the diagram

### Best Practices

- **Descriptive Names** - Use clear, descriptive field names
- **Comments** - Add descriptions to all fields
- **Consistency** - Follow naming conventions
- **Documentation** - Update this README when schema changes

## 📚 Related Documentation

- **API Documentation** - See `/docs/api/` for endpoint documentation
- **Models** - See `/backend/internal/models/` for Go struct definitions
- **Migrations** - Database migration files
- **Enum Types** - See `/backend/internal/models/enums.go` for type definitions

## 🤝 Contributing

When making database changes:

1. **Update ERD** - Modify `erd.mmd` to reflect changes
2. **Regenerate SVG** - Run conversion script
3. **Update README** - Modify this file if needed
4. **Test Changes** - Ensure all relationships are correct
5. **Document** - Add comments explaining new features

## 📞 Support

For questions about the database schema:

1. Check this README for general information
2. Review the ERD diagram for visual understanding
3. Examine the Go models for implementation details
4. Consult the API documentation for usage patterns

---

**Last Updated**: $(date)
**Version**: 1.0.0
**Maintainer**: Development Team 