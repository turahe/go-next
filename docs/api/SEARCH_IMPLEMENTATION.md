# 🔍 Search Implementation with Meilisearch

This document describes the implementation of search functionality using Meilisearch in the Go-Next Admin Panel.

## 📋 Overview

The search implementation provides fast, typo-tolerant search across multiple content types:
- **Posts** - Blog posts and articles
- **Users** - User accounts and profiles
- **Categories** - Content categories
- **Media** - Files and media assets

## 🏗️ Architecture

### Components

1. **SearchService** (`internal/services/search_service.go`)
   - Core search functionality
   - Index management
   - Document indexing
   - Search operations

2. **SearchHandler** (`internal/http/controllers/search_handlers.go`)
   - HTTP endpoints
   - Request validation
   - Response formatting

3. **Search Routes** (`internal/routers/v1/search.go`)
   - Route registration
   - Middleware integration

4. **Search Config** (`pkg/config/search.go`)
   - Configuration management
   - Environment variables

## 🔧 Configuration

### Environment Variables

```bash
# Meilisearch Configuration
MEILISEARCH_HOST=http://localhost:7700
MEILISEARCH_API_KEY=your-api-key-here
```

### Default Settings

```go
SearchSettings{
    DefaultLimit:     20,
    MaxLimit:         100,
    DefaultPage:      1,
    HighlightEnabled: true,
    SuggestionsLimit: 10,
}
```

## 📊 Index Configuration

### Posts Index
- **Searchable**: title, excerpt, description, slug
- **Filterable**: status, public, category_id, created_by, created_at, published_at
- **Sortable**: created_at, updated_at, published_at, view_count

### Users Index
- **Searchable**: username, email, phone
- **Filterable**: email_verified_at, phone_verified_at, created_at
- **Sortable**: created_at, updated_at

### Categories Index
- **Searchable**: name, slug, description
- **Filterable**: is_active, parent_id, created_at
- **Sortable**: sort_order, created_at, updated_at

### Media Index
- **Searchable**: file_name, original_name, mime_type
- **Filterable**: is_public, disk, mime_type, created_at
- **Sortable**: created_at, updated_at, size

## 🚀 API Endpoints

### Public Search Endpoints

#### 1. General Search
```http
GET /api/v1/search?query=search_term&page=1&limit=20
```

**Parameters:**
- `query` (required): Search term
- `page` (optional): Page number (default: 1)
- `limit` (optional): Results per page (default: 20)
- `indexes` (optional): Comma-separated list of indexes
- `filters` (optional): JSON string of filters
- `sort_by` (optional): Comma-separated sort fields
- `highlight` (optional): Enable highlighting (true/false)

#### 2. Posts Search
```http
GET /api/v1/search/posts?query=search_term&status=published&public=true
```

**Parameters:**
- `query` (required): Search term
- `status` (optional): Filter by post status
- `public` (optional): Filter by public status
- `category_id` (optional): Filter by category ID
- `created_by` (optional): Filter by author ID

#### 3. Users Search
```http
GET /api/v1/search/users?query=username&email_verified=true
```

**Parameters:**
- `query` (required): Search term
- `email_verified` (optional): Filter by email verification
- `phone_verified` (optional): Filter by phone verification

#### 4. Categories Search
```http
GET /api/v1/search/categories?query=category_name&is_active=true
```

**Parameters:**
- `query` (required): Search term
- `is_active` (optional): Filter by active status
- `parent_id` (optional): Filter by parent category

#### 5. Media Search
```http
GET /api/v1/search/media?query=filename&is_public=true&mime_type=image
```

**Parameters:**
- `query` (required): Search term
- `is_public` (optional): Filter by public status
- `disk` (optional): Filter by storage disk
- `mime_type` (optional): Filter by MIME type

#### 6. Search Suggestions
```http
GET /api/v1/search/suggestions?query=search&index=posts
```

**Parameters:**
- `query` (required): Search term (minimum 2 characters)
- `index` (optional): Index to search (default: posts)

#### 7. Search Statistics
```http
GET /api/v1/search/stats?index=posts
```

**Parameters:**
- `index` (optional): Specific index to get stats for

#### 8. Health Check
```http
GET /api/v1/search/health
```

### Admin Endpoints (Require Authentication)

#### 1. Reindex All Data
```http
POST /api/v1/search/reindex
```

#### 2. Initialize Indexes
```http
POST /api/v1/search/init
```

## 📝 Usage Examples

### Basic Search
```bash
curl "http://localhost:8080/api/v1/search?query=go&page=1&limit=10"
```

### Filtered Search
```bash
curl "http://localhost:8080/api/v1/search/posts?query=tutorial&status=published&public=true"
```

### Search with Filters
```bash
curl "http://localhost:8080/api/v1/search?query=admin&filters={\"type\":\"user\",\"is_active\":true}"
```

### Get Suggestions
```bash
curl "http://localhost:8080/api/v1/search/suggestions?query=go&index=posts"
```

## 🔄 Indexing

### Automatic Indexing

The search service automatically indexes documents when they are created, updated, or deleted:

```go
// Index a post
searchService.IndexPost(post)

// Index a user
searchService.IndexUser(user)

// Index a category
searchService.IndexCategory(category)

// Index media
searchService.IndexMedia(media)
```

### Manual Reindexing

To reindex all data from the database:

```bash
curl -X POST "http://localhost:8080/api/v1/search/reindex" \
  -H "Authorization: Bearer your-token"
```

## 🛠️ Integration with Existing Services

### Post Service Integration

```go
// In post_service.go
func (s *PostService) Create(post *models.Post) error {
    // ... existing logic ...
    
    // Index the post for search
    if err := s.searchService.IndexPost(post); err != nil {
        log.Printf("Failed to index post: %v", err)
    }
    
    return nil
}
```

### User Service Integration

```go
// In user_service.go
func (s *UserService) Create(user *models.User) error {
    // ... existing logic ...
    
    // Index the user for search
    if err := s.searchService.IndexUser(user); err != nil {
        log.Printf("Failed to index user: %v", err)
    }
    
    return nil
}
```

## 🔍 Search Features

### Typo Tolerance
Meilisearch provides built-in typo tolerance, handling common spelling mistakes.

### Faceted Search
Search results can be filtered by various attributes:
- Post status (draft, published, archived)
- User verification status
- Category active status
- Media public status

### Sorting
Results can be sorted by:
- Creation date
- Update date
- View count (posts)
- Sort order (categories)
- File size (media)

### Highlighting
Search terms are highlighted in results when the `highlight` parameter is enabled.

### Pagination
Results are paginated with configurable page size and page number.

## 📈 Performance

### Index Optimization
- Searchable attributes are optimized for each content type
- Filterable attributes are indexed for fast filtering
- Sortable attributes are indexed for efficient sorting

### Caching
- Meilisearch provides built-in caching
- Search results are cached for improved performance

### Monitoring
- Index statistics are available via `/api/v1/search/stats`
- Health checks via `/api/v1/search/health`

## 🔒 Security

### Authentication
- Public search endpoints are accessible without authentication
- Admin endpoints require JWT authentication
- RBAC middleware enforces permissions

### Authorization
- Search management requires `search:manage` permission
- Individual search operations are public

## 🚨 Error Handling

### Common Errors

1. **Meilisearch Connection Error**
   ```json
   {
     "message": "Search service is not healthy: connection refused",
     "status": 500
   }
   ```

2. **Invalid Query**
   ```json
   {
     "message": "Query parameter is required",
     "status": 400
   }
   ```

3. **Invalid Index**
   ```json
   {
     "message": "Invalid index. Must be one of: posts, users, categories, media",
     "status": 400
   }
   ```

## 🧪 Testing

### Health Check
```bash
curl "http://localhost:8080/api/v1/search/health"
```

### Index Statistics
```bash
curl "http://localhost:8080/api/v1/search/stats"
```

### Test Search
```bash
curl "http://localhost:8080/api/v1/search?query=test&limit=5"
```

## 📚 Dependencies

- **Meilisearch Go Client**: `github.com/meilisearch/meilisearch-go`
- **Gin Framework**: For HTTP routing
- **GORM**: For database operations

## 🔄 Migration

### Adding New Searchable Content

1. **Update SearchService**
   ```go
   func (s *SearchService) IndexNewContent(content *models.NewContent) error {
       doc := map[string]interface{}{
           "id": content.ID.String(),
           "type": "new_content",
           // ... other fields
       }
       
       _, err := s.client.Index("new_content").AddDocuments([]map[string]interface{}{doc})
       return err
   }
   ```

2. **Update Index Configuration**
   ```go
   case "new_content":
       searchableAttributes = []string{"title", "description"}
       filterableAttributes = []string{"status", "created_at"}
       sortableAttributes = []string{"created_at", "updated_at"}
   ```

3. **Add Routes**
   ```go
   search.GET("/new-content", searchHandler.SearchNewContent)
   ```

## 🎯 Best Practices

1. **Index Management**
   - Initialize indexes on application startup
   - Reindex data after schema changes
   - Monitor index health regularly

2. **Search Optimization**
   - Use appropriate filters to narrow results
   - Limit result size for better performance
   - Enable highlighting only when needed

3. **Error Handling**
   - Gracefully handle Meilisearch connection issues
   - Log indexing failures but don't block operations
   - Provide fallback search when needed

4. **Security**
   - Validate search parameters
   - Sanitize user input
   - Implement rate limiting for search endpoints

## 📖 Additional Resources

- [Meilisearch Documentation](https://docs.meilisearch.com/)
- [Meilisearch Go Client](https://github.com/meilisearch/meilisearch-go)
- [Search API Reference](./docs.go)

---

*Last updated: December 2024* 