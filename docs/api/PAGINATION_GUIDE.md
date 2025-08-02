# Laravel-Style Pagination Guide

This guide explains the Laravel-style pagination system implemented in the Go Next backend.

## Overview

The pagination system follows Laravel's pagination structure, providing a consistent and familiar API for frontend developers. It includes:

- **Data**: The actual items being paginated
- **Links**: Navigation links (first, last, prev, next)
- **Meta**: Metadata about the pagination (current_page, from, last_page, per_page, to, total)

## Response Structure

### LaravelPaginationResponse

```go
type LaravelPaginationResponse struct {
    Data  interface{}      `json:"data"`
    Links PaginationLinks  `json:"links"`
    Meta  PaginationMeta   `json:"meta"`
}
```

### PaginationLinks

```go
type PaginationLinks struct {
    First string `json:"first"`  // URL to first page
    Last  string `json:"last"`   // URL to last page
    Prev  string `json:"prev"`   // URL to previous page (empty if on first page)
    Next  string `json:"next"`   // URL to next page (empty if on last page)
}
```

### PaginationMeta

```go
type PaginationMeta struct {
    CurrentPage int64 `json:"current_page"` // Current page number
    From        int64 `json:"from"`         // First item number on current page
    LastPage    int64 `json:"last_page"`    // Last page number
    PerPage     int64 `json:"per_page"`     // Items per page
    To          int64 `json:"to"`           // Last item number on current page
    Total       int64 `json:"total"`        // Total number of items
}
```

## Example Response

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Sample Post",
      "content": "This is a sample post content...",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "links": {
    "first": "http://localhost:8080/api/v1/posts?page=1",
    "last": "http://localhost:8080/api/v1/posts?page=5",
    "prev": "http://localhost:8080/api/v1/posts?page=2",
    "next": "http://localhost:8080/api/v1/posts?page=4"
  },
  "meta": {
    "current_page": 3,
    "from": 21,
    "last_page": 5,
    "per_page": 10,
    "to": 30,
    "total": 50
  }
}
```

## Usage in Handlers

### Basic Pagination

```go
func (h *handler) GetItems(c *gin.Context) {
    params := responses.ParsePaginationParams(c)
    
    offset := (params.Page - 1) * params.PerPage
    
    var items []models.Item
    var total int64
    
    query := database.DB.Model(&models.Item{})
    
    // Get total count
    if err := query.Count(&total).Error; err != nil {
        responses.SendError(c, http.StatusInternalServerError, "Failed to count items")
        return
    }
    
    // Get paginated items
    if err := query.Offset(offset).Limit(params.PerPage).Find(&items).Error; err != nil {
        responses.SendError(c, http.StatusInternalServerError, "Failed to fetch items")
        return
    }
    
    // Send Laravel-style pagination response
    responses.SendLaravelPaginationWithMessage(c, "Items retrieved successfully", items, total, int64(params.Page), int64(params.PerPage))
}
```

### With Search and Filters

```go
func (h *handler) GetItems(c *gin.Context) {
    params := responses.ParsePaginationParams(c)
    search := c.Query("search")
    categoryID := c.Query("category")
    
    offset := (params.Page - 1) * params.PerPage
    
    var items []models.Item
    var total int64
    
    query := database.DB.Model(&models.Item{})
    
    // Apply search filter
    if search != "" {
        query = query.Where("title ILIKE ? OR content ILIKE ?", "%"+search+"%", "%"+search+"%")
    }
    
    // Apply category filter
    if categoryID != "" {
        if parsedID, err := uuid.Parse(categoryID); err == nil {
            query = query.Where("category_id = ?", parsedID)
        }
    }
    
    // Get total count
    if err := query.Count(&total).Error; err != nil {
        responses.SendError(c, http.StatusInternalServerError, "Failed to count items")
        return
    }
    
    // Get paginated items
    if err := query.Offset(offset).Limit(params.PerPage).Order("created_at DESC").Find(&items).Error; err != nil {
        responses.SendError(c, http.StatusInternalServerError, "Failed to fetch items")
        return
    }
    
    // Send Laravel-style pagination response
    responses.SendLaravelPaginationWithMessage(c, "Items retrieved successfully", items, total, int64(params.Page), int64(params.PerPage))
}
```

## Available Functions

### ParsePaginationParams

Parses pagination parameters from the request with validation:

```go
params := responses.ParsePaginationParams(c)
// Returns PaginationParams{Page: 1, PerPage: 15} with defaults
// Validates: page >= 1, per_page >= 1, per_page <= 100
```

### SendLaravelPagination

Sends a basic Laravel-style pagination response:

```go
responses.SendLaravelPagination(c, data, total, currentPage, perPage)
```

### SendLaravelPaginationWithMessage

Sends a Laravel-style pagination response with a custom message:

```go
responses.SendLaravelPaginationWithMessage(c, "Items retrieved successfully", data, total, currentPage, perPage)
```

### CreateLaravelPaginationResponse

Creates a Laravel-style pagination response object:

```go
response := responses.CreateLaravelPaginationResponse(c, data, total, currentPage, perPage)
```

## Query Parameters

### Standard Parameters

- `page`: Page number (default: 1)
- `per_page`: Items per page (default: 15, max: 100)

### Example URLs

```
GET /api/v1/posts?page=2&per_page=20
GET /api/v1/users?page=1&per_page=10&search=john
GET /api/v1/categories?page=3&per_page=25&parent=550e8400-e29b-41d4-a716-446655440000
```

## Frontend Integration

### React/TypeScript Example

```typescript
interface PaginationResponse<T> {
  data: T[];
  links: {
    first: string;
    last: string;
    prev: string | null;
    next: string | null;
  };
  meta: {
    current_page: number;
    from: number;
    last_page: number;
    per_page: number;
    to: number;
    total: number;
  };
}

// API call
const getPosts = async (page: number = 1, perPage: number = 15): Promise<PaginationResponse<Post>> => {
  const response = await fetch(`/api/v1/posts?page=${page}&per_page=${perPage}`);
  return response.json();
};

// Usage in component
const [pagination, setPagination] = useState<PaginationResponse<Post> | null>(null);

useEffect(() => {
  const fetchPosts = async () => {
    const result = await getPosts(1, 10);
    setPagination(result);
  };
  fetchPosts();
}, []);
```

## Migration from Old Pagination

### Before (Old Format)

```go
c.JSON(http.StatusOK, gin.H{
    "users": users,
    "total": total,
    "page":  page,
    "limit": limit,
    "pages": (int(total) + limit - 1) / limit,
})
```

### After (Laravel Style)

```go
responses.SendLaravelPaginationWithMessage(c, "Users retrieved successfully", users, total, int64(params.Page), int64(params.PerPage))
```

## Error Handling

The pagination system integrates with the existing error response system:

```go
if err := query.Count(&total).Error; err != nil {
    responses.SendError(c, http.StatusInternalServerError, "Failed to count items")
    return
}
```

## Best Practices

1. **Always validate pagination parameters** using `ParsePaginationParams`
2. **Use consistent ordering** (e.g., `Order("created_at DESC")`)
3. **Apply filters before counting** to ensure accurate totals
4. **Handle empty results gracefully** - the system handles zero items correctly
5. **Use meaningful success messages** in `SendLaravelPaginationWithMessage`
6. **Preserve query parameters** in pagination links for filters and search

## Supported Endpoints

The following endpoints now support Laravel-style pagination:

- `GET /api/v1/users` - List users with search
- `GET /api/v1/posts` - List posts with search and category filter
- `GET /api/v1/categories` - List categories with search and parent filter
- `GET /api/v1/posts/{post_id}/comments` - List comments for a post

## Testing

Test pagination with various scenarios:

```bash
# Basic pagination
curl "http://localhost:8080/api/v1/posts?page=1&per_page=10"

# With search
curl "http://localhost:8080/api/v1/posts?page=2&per_page=5&search=test"

# With filters
curl "http://localhost:8080/api/v1/posts?page=1&per_page=20&category=550e8400-e29b-41d4-a716-446655440000"

# Edge cases
curl "http://localhost:8080/api/v1/posts?page=0&per_page=1000"  # Should use defaults
curl "http://localhost:8080/api/v1/posts?page=999&per_page=10"  # Should return empty data
``` 