# Post-Content Relationship

This document explains the relationship between Posts and Content in the application.

## Overview

Posts can have multiple content blocks through a polymorphic relationship. This allows for flexible content management where a single post can contain different types of content (HTML, Markdown, JSON, text) organized in a specific order.

## Database Schema

### Post Model
```go
type Post struct {
    BaseModelWithUser
    Title       string     `json:"title"`
    Slug        string     `json:"slug"`
    Content     string     `json:"content"` // Main content field
    // ... other fields
    
    // Relationships
    Contents []Content `json:"contents,omitempty" gorm:"polymorphic:Model;polymorphicValue:post;constraint:OnDelete:CASCADE"`
}
```

### Content Model
```go
type Content struct {
    BaseModel
    ModelID   uuid.UUID `json:"model_id"`   // References the Post ID
    ModelType string    `json:"model_type"`  // Always "post" for posts
    Type      string    `json:"type"`        // html, markdown, json, text
    Content   string    `json:"content"`     // The actual content
    SortOrder int       `json:"sort_order"`  // Order of display
}
```

## Relationship Details

### Polymorphic Relationship
- **Model**: The polymorphic field that references the parent model
- **ModelType**: The type of the parent model ("post")
- **ModelID**: The ID of the parent model (Post ID)

### Content Types
- `html`: HTML content
- `markdown`: Markdown content
- `json`: JSON content
- `text`: Plain text content

### Sort Order
- Content blocks are ordered by the `sort_order` field
- Lower numbers appear first
- Default sort order is 0

## Usage Examples

### Creating a Post with Content
```go
post := &models.Post{
    Title: "My Post",
    Slug:  "my-post",
    Content: "Main content",
    Contents: []models.Content{
        {
            Type:      "html",
            Content:   "<h1>Introduction</h1><p>Welcome to my post.</p>",
            SortOrder: 0,
        },
        {
            Type:      "markdown",
            Content:   "## Section 1\nThis is markdown content.",
            SortOrder: 1,
        },
        {
            Type:      "json",
            Content:   `{"key": "value", "data": [1, 2, 3]}`,
            SortOrder: 2,
        },
    },
}

// Create post with content
err := postService.CreatePost(post)
```

### Adding Content to Existing Post
```go
// Add HTML content
content, err := postService.AddContentToPost("post-id", "html", "<h2>New Section</h2>", 3)

// Add Markdown content
content, err := postService.AddContentToPost("post-id", "markdown", "## Another Section", 4)
```

### Updating Content
```go
err := postService.UpdatePostContent("post-id", "content-id", "html", "<h2>Updated Section</h2>", 1)
```

### Removing Content
```go
err := postService.RemoveContentFromPost("post-id", "content-id")
```

### Getting Post with Content
```go
post, err := postService.GetPostByID("post-id")
// post.Contents will contain all content blocks ordered by sort_order
```

### Getting Content by Type
```go
htmlContents, err := postService.GetPostContentsByType("post-id", "html")
markdownContents, err := postService.GetPostContentsByType("post-id", "markdown")
```

### Reordering Content
```go
contentOrder := []string{"content-id-1", "content-id-3", "content-id-2"}
err := postService.ReorderPostContents("post-id", contentOrder)
```

## Model Methods

### Post Model Methods
```go
// Add content block
post.AddContent("html", "<p>New content</p>", 5)

// Get content by type
htmlContent := post.GetContentByType("html")

// Get sorted content
sortedContent := post.GetSortedContents()

// Remove content
post.RemoveContent(contentID)

// Update content
post.UpdateContent(contentID, "markdown", "# Updated", 2)

// Check if post has content
hasContent := post.HasContent()

// Get content count
count := post.GetContentCount()
```

## API Endpoints

### Content Management Endpoints
```
POST   /api/posts/{id}/contents          # Add content to post
PUT    /api/posts/{id}/contents/{content_id}  # Update content
DELETE /api/posts/{id}/contents/{content_id}  # Remove content
GET    /api/posts/{id}/contents          # Get all content
GET    /api/posts/{id}/contents?type=html     # Get content by type
PUT    /api/posts/{id}/contents/reorder  # Reorder content
```

## Benefits

1. **Flexibility**: Posts can contain multiple types of content
2. **Ordering**: Content blocks can be arranged in any order
3. **Type Safety**: Different content types are clearly defined
4. **Scalability**: Easy to add new content types
5. **Performance**: Content can be loaded separately or together
6. **Polymorphic**: The Content model can be used by other models

## Best Practices

1. **Content Types**: Use appropriate content types for your data
2. **Sort Order**: Plan your content ordering carefully
3. **Validation**: Validate content before saving
4. **Performance**: Use pagination for posts with many content blocks
5. **Caching**: Consider caching frequently accessed content

## Database Queries

### Get Post with Content
```sql
SELECT * FROM posts p
LEFT JOIN contents c ON c.model_id = p.id AND c.model_type = 'post'
WHERE p.id = ?
ORDER BY c.sort_order ASC;
```

### Get Content by Type
```sql
SELECT * FROM contents
WHERE model_id = ? AND model_type = 'post' AND type = ?
ORDER BY sort_order ASC;
```

### Update Sort Order
```sql
UPDATE contents 
SET sort_order = ? 
WHERE id = ? AND model_id = ? AND model_type = 'post';
```

## Related Files

- `backend/internal/models/post.go`: Post model with content relationship
- `backend/internal/models/content.go`: Content model
- `backend/internal/services/post_service.go`: Post service with content methods
- `backend/internal/http/controllers/post_handlers.go`: Post API handlers 