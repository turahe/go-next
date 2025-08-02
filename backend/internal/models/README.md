# Go Next - Database Models Documentation

## Overview

This document describes the optimized database models for the Go Next backend application. All models have been refactored to improve performance, maintainability, and consistency.

## Base Models

### BaseModel
Common fields shared across all models:
- `ID`: Primary key with auto-increment
- `CreatedAt`: Timestamp when record was created (auto-indexed)
- `UpdatedAt`: Timestamp when record was last updated
- `DeletedAt`: Soft delete timestamp (auto-indexed)

### BaseModelWithUser
Extends BaseModel with user tracking fields:
- `CreatedBy`: User ID who created the record (indexed)
- `UpdatedBy`: User ID who last updated the record (indexed)
- `DeletedBy`: User ID who deleted the record (indexed)

### BaseModelWithOrdering
Extends BaseModel with hierarchical ordering support:
- `RecordLeft`: Left boundary for nested set model
- `RecordRight`: Right boundary for nested set model
- `RecordDept`: Depth level in hierarchy
- `RecordOrdering`: Sort order within siblings
- `ParentID`: Parent record ID for hierarchical relationships

## Core Models

### User
**Table**: `users`

Represents system users with authentication and profile information.

**Key Features**:
- Email and phone verification tracking
- Password hashing with bcrypt
- Role-based access control
- Activity status tracking
- Last login tracking

**Optimizations**:
- Unique indexes on username, email, and phone
- Check constraints for username format
- Proper field size limits
- Hidden password hash from JSON responses

**Methods**:
- `CheckPassword(password)`: Verify password against hash
- `HashPassword(password)`: Hash and store password
- `IsEmailVerified()`: Check email verification status
- `MarkEmailVerified()`: Mark email as verified
- `UpdateLastLogin()`: Update last login timestamp

### Role
**Table**: `roles`

Represents user roles for access control.

**Key Features**:
- Role name and description
- Active/inactive status
- Many-to-many relationship with users

**Optimizations**:
- Unique index on name
- Proper field size limits
- Active status tracking

### Post
**Table**: `posts`

Represents blog posts or articles.

**Key Features**:
- Content management with drafts, published, and archived states
- SEO-friendly slugs
- View count tracking
- Public/private visibility
- Category association

**Optimizations**:
- Unique index on slug
- Indexes on status and visibility
- Proper field size limits
- Automatic status management

**Methods**:
- `IncrementViewCount()`: Increase view count
- `IsPublished()`: Check if post is published
- `IsPublic()`: Check if post is publicly visible

### Category
**Table**: `categories`

Represents content categories with hierarchical support.

**Key Features**:
- Hierarchical structure (parent-child relationships)
- SEO-friendly slugs
- Active/inactive status
- Sort ordering

**Optimizations**:
- Unique index on slug
- Indexes on active status and sort order
- Nested set model support for efficient queries

**Methods**:
- `IsRoot()`: Check if category is top-level
- `HasChildren()`: Check if category has subcategories
- `GetDepth()`: Get hierarchy depth level

### Comment
**Table**: `comments`

Represents user comments on posts with threading support.

**Key Features**:
- Approval workflow (pending, approved, rejected)
- Threaded comments (parent-child relationships)
- Public/private visibility
- Media attachments

**Optimizations**:
- Indexes on status and visibility
- Nested set model for efficient threading queries
- Proper foreign key constraints

**Methods**:
- `IsApproved()`: Check approval status
- `IsRejected()`: Check rejection status
- `IsPending()`: Check pending status
- `IsRoot()`: Check if comment is top-level
- `HasChildren()`: Check if comment has replies

### Media
**Table**: `media`

Represents media files with polymorphic relationships.

**Key Features**:
- UUID-based identification
- Multiple storage backends (local, S3, GCS)
- File metadata (size, dimensions, duration)
- Public/private visibility
- Polymorphic relationships

**Optimizations**:
- Unique index on UUID
- Indexes on hash and visibility
- Proper field size limits
- File type detection methods

**Methods**:
- `IsImage()`: Check if file is an image
- `IsVideo()`: Check if file is a video
- `IsAudio()`: Check if file is audio
- `IsDocument()`: Check if file is a document
- `GetFileSize()`: Get human-readable file size
- `GetDimensions()`: Get image dimensions

### Content
**Table**: `contents`

Represents additional content for polymorphic models.

**Key Features**:
- Polymorphic relationships
- Multiple content types (text, HTML, Markdown, JSON)
- Sort ordering

**Optimizations**:
- Indexes on model type and sort order
- Proper field size limits
- Content type validation

**Methods**:
- `IsHTML()`: Check if content is HTML
- `IsMarkdown()`: Check if content is Markdown
- `IsJSON()`: Check if content is JSON

### Mediable
**Table**: `mediables`

Polymorphic relationship table between Media and other models.

**Key Features**:
- Polymorphic associations
- Grouping support
- Sort ordering

**Optimizations**:
- Composite primary key
- Indexes on all key fields
- Proper field size limits

**Methods**:
- `IsDefaultGroup()`: Check if in default group
- `GetGroup()`: Get group name

## Authentication Models

### Token
**Table**: `tokens`

Represents access tokens for API authentication.

**Key Features**:
- Token and refresh token storage
- Expiration tracking
- Client secret for additional security
- Usage tracking

**Optimizations**:
- Unique index on token
- Indexes on user ID and expiration
- Proper field size limits
- IPv6-compatible IP address storage

**Methods**:
- `IsExpired()`: Check token expiration
- `IsValid()`: Check if token is active and not expired
- `UpdateLastUsed()`: Update last usage timestamp

### JWTKey
**Table**: `jwt_keys`

Stores per-user/client JWT signing keys.

**Key Features**:
- Client-specific keys
- Configurable token expiration
- Active/inactive status

**Optimizations**:
- Unique index on client key
- Index on user ID
- Proper field size limits
- Expiration validation

**Methods**:
- `IsValid()`: Check if key is active

### RefreshToken
**Table**: `refresh_tokens`

Legacy refresh token storage (maintained for backward compatibility).

**Key Features**:
- Refresh token storage
- Expiration tracking
- Usage tracking

**Optimizations**:
- Unique index on token
- Indexes on user ID and expiration
- IPv6-compatible IP address storage

**Methods**:
- `IsExpired()`: Check token expiration
- `IsValid()`: Check if token is active and not expired

### VerificationToken
**Table**: `verification_tokens`

Represents verification tokens for user actions.

**Key Features**:
- Multiple token types (email, phone, password reset)
- Expiration tracking
- Usage tracking
- IP and user agent logging

**Optimizations**:
- Unique index on token
- Indexes on user ID, type, and expiration
- Proper field size limits
- Type validation

**Methods**:
- `IsExpired()`: Check token expiration
- `IsValid()`: Check if token is valid and not used
- `MarkAsUsed()`: Mark token as used
- `IsEmailVerification()`: Check if email verification token
- `IsPhoneVerification()`: Check if phone verification token
- `IsPasswordReset()`: Check if password reset token

## Notification Models

### Notification
**Table**: `notifications`

Represents user notifications.

**Key Features**:
- Multiple notification types (success, error, warning, info)
- Priority levels (normal, high, urgent)
- Read/unread status
- JSON data for additional context

**Optimizations**:
- Indexes on user ID, type, read status, and priority
- Proper field size limits
- Type and priority validation

**Methods**:
- `MarkAsRead()`: Mark notification as read
- `MarkAsUnread()`: Mark notification as unread
- `IsRead()`: Check read status
- `IsHighPriority()`: Check if high priority
- `IsUrgent()`: Check if urgent priority

## Key Optimizations Made

### 1. Base Model Inheritance
- Eliminated code duplication
- Consistent timestamp handling
- Standardized soft delete support

### 2. Database Indexes
- Added strategic indexes for performance
- Unique constraints where appropriate
- Composite indexes for common queries

### 3. Field Constraints
- Proper field size limits
- Check constraints for data validation
- Not null constraints where required

### 4. Validation Tags
- Added comprehensive validation rules
- Type checking for enums
- Length and format validation

### 5. JSON Tags
- Consistent JSON field naming
- Hidden sensitive fields (e.g., password hash)
- Optional field handling

### 6. Relationship Constraints
- Proper foreign key constraints
- Cascade delete rules
- Restrict delete rules where appropriate

### 7. Helper Methods
- Business logic encapsulation
- Status checking methods
- Utility functions for common operations

### 8. Type Safety
- Strong typing for enums
- Consistent data types across models
- Proper pointer usage for optional fields

## Migration Notes

When migrating from the old models to these optimized versions:

1. **Database Schema Changes**:
   - New indexes will be created
   - Field size constraints may be added
   - New fields may be added with defaults

2. **Code Changes**:
   - Update import paths if needed
   - Review validation rules in handlers
   - Update any hardcoded field references

3. **Performance Improvements**:
   - Query performance should improve with new indexes
   - Reduced memory usage with proper field sizes
   - Better caching opportunities with consistent structures

## Best Practices

1. **Always use the base models** for new entities
2. **Add indexes** for fields used in WHERE clauses
3. **Use validation tags** for data integrity
4. **Implement proper relationships** with constraints
5. **Add helper methods** for common business logic
6. **Use soft deletes** for data preservation
7. **Log user actions** with user tracking fields
8. **Implement proper error handling** in hooks 