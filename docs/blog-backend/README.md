# Blog Backend Implementation

This document describes the blog backend implementation for the Go-Next project, providing comprehensive blog functionality with public and admin endpoints.

## 🏗️ Architecture Overview

The blog backend extends the existing Go backend with specialized blog functionality:

```
backend/
├── internal/
│   ├── services/
│   │   └── blog_service.go          # Blog-specific business logic
│   ├── http/
│   │   ├── controllers/
│   │   │   └── blog_handlers.go     # Blog API endpoints
│   │   └── requests/
│   │       └── blog_request.go      # Blog request validation
│   └── models/                      # Existing models (Post, Category, etc.)
└── routers/
    └── routes.go                    # Blog routes configuration
```

## 🚀 Features

### Public Blog Features
- **Post Listing**: Paginated public posts with search and filtering
- **Single Post View**: Individual post pages with view count tracking
- **Featured Posts**: Highlighted posts for homepage
- **Popular Posts**: Posts ranked by view count
- **Related Posts**: Similar posts based on categories
- **Search**: Full-text search across posts
- **Categories & Tags**: Organized content browsing
- **Blog Statistics**: Analytics and metrics
- **Monthly Archives**: Time-based post organization

### Admin Features
- **Post Management**: Create, update, delete posts
- **Publishing Workflow**: Draft → Published → Archived
- **Content Organization**: Categories and tags management
- **View Analytics**: Track post popularity
- **Comment Moderation**: Approve/reject comments

### Technical Features
- **RESTful API**: Clean, consistent endpoints
- **Pagination**: Laravel-style pagination responses
- **Search & Filtering**: Advanced content discovery
- **View Tracking**: Automatic view count increments
- **SEO Optimization**: Slug-based URLs
- **Performance**: Optimized database queries

## 📋 API Endpoints

### Public Blog Endpoints

#### Posts
```
GET  /api/v1/blog/posts                    # List public posts
GET  /api/v1/blog/posts/featured           # Get featured posts
GET  /api/v1/blog/posts/popular            # Get popular posts
GET  /api/v1/blog/posts/{slug}             # Get single post
GET  /api/v1/blog/posts/{post_id}/related  # Get related posts
POST /api/v1/blog/posts/{id}/view          # Increment view count
```

#### Search
```
GET  /api/v1/blog/search                   # Search posts
```

#### Categories
```
GET  /api/v1/blog/categories               # List categories
GET  /api/v1/blog/categories/{slug}        # Get category
GET  /api/v1/blog/categories/{slug}/posts  # Get posts by category
```

#### Tags
```
GET  /api/v1/blog/tags                     # List tags
GET  /api/v1/blog/tags/{slug}              # Get tag
GET  /api/v1/blog/tags/{slug}/posts        # Get posts by tag
```

#### Statistics
```
GET  /api/v1/blog/stats                    # Blog statistics
GET  /api/v1/blog/stats/categories         # Category statistics
GET  /api/v1/blog/archive                  # Monthly archive
```

### Admin Blog Endpoints (Require Authentication)

#### Post Management
```
POST   /api/v1/blog/posts                  # Create post
PUT    /api/v1/blog/posts/{id}             # Update post
DELETE /api/v1/blog/posts/{id}             # Delete post
POST   /api/v1/blog/posts/{id}/publish     # Publish post
POST   /api/v1/blog/posts/{id}/unpublish   # Unpublish post
POST   /api/v1/blog/posts/{id}/archive     # Archive post
```

## 🔧 Setup Instructions

### 1. Environment Configuration

The blog backend uses the same environment variables as the main backend. Ensure these are configured:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=admin_panel

# JWT
JWT_SECRET=your-secret-key

# Redis (optional)
REDIS_URL=redis://localhost:6379
```

### 2. Database Setup

The blog functionality uses existing models. Run migrations:

```bash
cd backend
go run main.go
```

### 3. Service Initialization

The blog service is automatically initialized in the routes:

```go
// Initialize blog service and handler
blogSvc := services.NewBlogService()
blogHandler := controllers.NewBlogHandler(blogSvc)
```

## 📊 Data Models

### Post Model
```go
type Post struct {
    BaseModelWithUser
    Title       string     `json:"title"`
    Slug        string     `json:"slug"`
    Content     string     `json:"content"`
    Excerpt     string     `json:"excerpt"`
    Status      string     `json:"status"` // draft, published, archived
    Public      bool       `json:"public"`
    PublishedAt *time.Time `json:"published_at"`
    ViewCount   int64      `json:"view_count"`
    CategoryID  *uuid.UUID `json:"category_id"`
    
    // Relationships
    Category *Category `json:"category,omitempty"`
    Comments []Comment `json:"comments,omitempty"`
    Media    []Media   `json:"media,omitempty"`
}
```

### Category Model
```go
type Category struct {
    BaseModelWithOrdering
    Name        string `json:"name"`
    Slug        string `json:"slug"`
    Description string `json:"description"`
    IsActive    bool   `json:"is_active"`
    SortOrder   int    `json:"sort_order"`
    
    // Relationships
    Parent   *Category  `json:"parent,omitempty"`
    Children []Category `json:"children,omitempty"`
    Posts    []Post     `json:"posts,omitempty"`
}
```

## 🎯 Usage Examples

### Get Public Posts
```bash
curl -X GET "http://localhost:8080/api/v1/blog/posts?page=1&per_page=10"
```

Response:
```json
{
  "data": [
    {
      "id": "uuid",
      "title": "Sample Post",
      "slug": "sample-post",
      "content": "Post content...",
      "excerpt": "Post excerpt...",
      "status": "published",
      "public": true,
      "published_at": "2024-01-01T00:00:00Z",
      "view_count": 42,
      "category": {
        "id": "uuid",
        "name": "Technology",
        "slug": "technology"
      }
    }
  ],
  "meta": {
    "current_page": 1,
    "last_page": 5,
    "per_page": 10,
    "total": 50
  },
  "message": "Posts retrieved successfully"
}
```

### Get Single Post
```bash
curl -X GET "http://localhost:8080/api/v1/blog/posts/sample-post"
```

### Search Posts
```bash
curl -X GET "http://localhost:8080/api/v1/blog/search?query=technology&page=1&per_page=10"
```

### Get Blog Statistics
```bash
curl -X GET "http://localhost:8080/api/v1/blog/stats"
```

Response:
```json
{
  "message": "Blog statistics retrieved successfully",
  "data": {
    "total_posts": 100,
    "published_posts": 85,
    "total_views": 15000,
    "total_comments": 250,
    "total_categories": 10,
    "total_tags": 25
  }
}
```

### Create Post (Admin)
```bash
curl -X POST "http://localhost:8080/api/v1/blog/posts" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Post",
    "slug": "new-post",
    "content": "Post content...",
    "excerpt": "Post excerpt...",
    "status": "draft",
    "public": true,
    "category_id": "category-uuid"
  }'
```

## 🔍 Query Parameters

### Posts Endpoint
- `page` (int): Page number (default: 1)
- `per_page` (int): Items per page (default: 10, max: 100)
- `search` (string): Search term
- `category` (string): Category slug filter

### Featured Posts
- `limit` (int): Number of posts (default: 5)

### Popular Posts
- `limit` (int): Number of posts (default: 10)
- `days` (int): Days to look back (default: 30)

### Related Posts
- `limit` (int): Number of posts (default: 3)

## 🛡️ Security

### Public Endpoints
- No authentication required
- Rate limiting applied
- CORS configured for frontend access

### Admin Endpoints
- JWT authentication required
- Casbin authorization for role-based access
- Admin/Editor roles can manage posts

### Data Validation
- Request validation using struct tags
- Input sanitization
- SQL injection protection via GORM

## 📈 Performance Optimizations

### Database Queries
- Optimized joins for related data
- Indexed fields for fast lookups
- Pagination to limit result sets
- Eager loading of relationships

### Caching Strategy
- Redis caching for frequently accessed data
- Cache invalidation on content updates
- View count caching to reduce database writes

### API Response Optimization
- Laravel-style pagination
- Consistent response format
- Error handling with appropriate HTTP status codes

## 🔧 Configuration

### Blog Service Configuration
```go
type BlogService interface {
    // Public endpoints
    GetPublicPosts(page, perPage int, search, categorySlug string) ([]Post, int64, error)
    GetPublicPost(slug string) (*Post, error)
    GetFeaturedPosts(limit int) ([]Post, error)
    GetRelatedPosts(postID uuid.UUID, limit int) ([]Post, error)
    GetPopularPosts(limit int, days int) ([]Post, error)
    
    // Statistics
    GetBlogStats() (*BlogStats, error)
    GetCategoryStats() ([]CategoryStats, error)
    GetMonthlyArchive() ([]MonthlyArchive, error)
    
    // Admin functions
    CreatePost(post *Post) error
    UpdatePost(post *Post) error
    DeletePost(id string) error
    PublishPost(id string) error
    UnpublishPost(id string) error
    ArchivePost(id string) error
    IncrementViewCount(postID uuid.UUID) error
}
```

## 🧪 Testing

### Unit Tests
```bash
cd backend
go test ./internal/services -v
go test ./internal/http/controllers -v
```

### API Tests
```bash
# Test public endpoints
curl -X GET "http://localhost:8080/api/v1/blog/posts"

# Test admin endpoints (with token)
curl -X POST "http://localhost:8080/api/v1/blog/posts" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","slug":"test","content":"Test content"}'
```

## 🚀 Deployment

### Production Considerations
1. **Database**: Use PostgreSQL for production
2. **Caching**: Configure Redis for performance
3. **Security**: Use HTTPS and secure JWT secrets
4. **Monitoring**: Add logging and metrics
5. **Backup**: Regular database backups

### Environment Variables
```env
# Production settings
DB_HOST=your-db-host
DB_PASSWORD=secure-password
JWT_SECRET=very-secure-secret
REDIS_URL=redis://your-redis-host:6379
```

## 📚 Integration with Frontend

The blog backend is designed to work seamlessly with the blog-frontend:

### Frontend Configuration
```typescript
// blog-frontend/src/services/blogApi.ts
const API_BASE_URL = import.meta.env.VITE_BLOG_API_BASE_URL || 'http://localhost:8080/api/v1/blog';
```

### Environment Variables
```env
# blog-frontend/.env
VITE_BLOG_API_BASE_URL=http://localhost:8080/api/v1/blog
VITE_API_URL=http://localhost:8080
```

## 🔄 Migration from Existing Backend

The blog backend extends the existing backend without breaking changes:

1. **Models**: Uses existing Post, Category, Comment models
2. **Database**: No new migrations required
3. **Authentication**: Uses existing JWT middleware
4. **Authorization**: Uses existing Casbin middleware

## 🆘 Troubleshooting

### Common Issues

1. **Posts not appearing**
   - Check post status is "published"
   - Verify published_at is set
   - Ensure public flag is true

2. **Search not working**
   - Verify database supports ILIKE
   - Check search query length
   - Ensure posts are published

3. **View count not incrementing**
   - Check database permissions
   - Verify post exists
   - Check for database constraints

4. **Categories not loading**
   - Verify is_active is true
   - Check sort_order values
   - Ensure proper relationships

### Debug Mode
Enable debug logging:
```go
// In main.go
gin.SetMode(gin.DebugMode)
```

## 📞 Support

For issues and questions:
1. Check the API documentation
2. Review the error logs
3. Test with curl commands
4. Create an issue with detailed information

---

*Last updated: January 2024* 