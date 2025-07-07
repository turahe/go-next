package services

import (
	"context"
	"fmt"

	"wordpress-go-next/backend/internal/http/responses"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"
	"wordpress-go-next/backend/pkg/redis"

	"gorm.io/gorm"
)

type PostService interface {
	GetAllPosts(ctx context.Context) ([]models.Post, error)
	GetPostByID(ctx context.Context, id string) (*models.Post, error)
	GetPostBySlug(ctx context.Context, slug string) (*models.Post, error)
	GetPostsWithPagination(ctx context.Context, page, perPage int, search string) (*responses.PaginationResponse, error)
	GetPostsByCategory(ctx context.Context, categoryID string) ([]models.Post, error)
	GetPostsByUser(ctx context.Context, userID string) ([]models.Post, error)
	GetPublishedPosts(ctx context.Context) ([]models.Post, error)
	GetDraftPosts(ctx context.Context) ([]models.Post, error)
	CreatePost(ctx context.Context, post *models.Post) error
	UpdatePost(ctx context.Context, post *models.Post) error
	DeletePost(ctx context.Context, id string) error
	PublishPost(ctx context.Context, id string) error
	ArchivePost(ctx context.Context, id string) error
	SearchPosts(ctx context.Context, query string) ([]models.Post, error)
	GetPostStats(ctx context.Context, postID string) (map[string]interface{}, error)
	GetPostCount(ctx context.Context) (int64, error)
	GetPublishedPostCount(ctx context.Context) (int64, error)
}

type postService struct {
	*BaseService
}

func NewPostService(redisService *redis.RedisService) PostService {
	return &postService{
		BaseService: NewBaseService(redisService),
	}
}

func (s *postService) GetAllPosts(ctx context.Context) ([]models.Post, error) {
	var posts []models.Post
	cacheKey := s.GetListCacheKey(redis.PostCachePrefix)

	err := s.GetAllWithCacheAndPreload(ctx, &posts, cacheKey, redis.DefaultTTL, "User", "Category", "Comments")
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (s *postService) GetPostByID(ctx context.Context, id string) (*models.Post, error) {
	var post models.Post
	cacheKey := s.GetCacheKey(redis.PostCachePrefix, id)

	err := s.GetByIDWithCacheAndPreload(ctx, id, &post, cacheKey, redis.DefaultTTL, "User", "Category", "Comments", "Contents", "Medias")
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *postService) GetPostBySlug(ctx context.Context, slug string) (*models.Post, error) {
	var post models.Post
	cacheKey := fmt.Sprintf("%sslug:%s", redis.PostCachePrefix, slug)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &post); err == nil {
			return &post, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("User").Preload("Category").Preload("Comments").Preload("Contents").Preload("Medias").
		Where("slug = ?", slug).First(&post).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		err := s.Redis.SetCache(ctx, cacheKey, &post, redis.DefaultTTL)
		if err != nil {
			return nil, err
		}
	}

	return &post, nil
}

func (s *postService) GetPostsWithPagination(ctx context.Context, page, perPage int, search string) (*responses.PaginationResponse, error) {
	var posts []models.Post
	cacheKey := s.GetListCacheKey(redis.PostCachePrefix)

	params := PaginationParams{
		Page:    page,
		PerPage: perPage,
	}

	query := database.DB
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("title LIKE ? OR excerpt LIKE ?", like, like)
	}

	result, err := s.PaginateWithCacheQuery(ctx, &models.Post{}, params, &posts, cacheKey, redis.DefaultTTL, query)
	if err != nil {
		return nil, err
	}

	// Preload relationships for each post
	for i := range posts {
		if err := database.DB.Preload("User").Preload("Category").Preload("Comments").Preload("Contents").Preload("Medias").
			First(&posts[i], posts[i].ID).Error; err != nil {
			return nil, err
		}
	}

	result.Data = posts
	return result, nil
}

func (s *postService) GetPostsByCategory(ctx context.Context, categoryID string) ([]models.Post, error) {
	var posts []models.Post
	cacheKey := fmt.Sprintf("%scategory:%s", redis.PostCachePrefix, categoryID)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &posts); err == nil {
			return posts, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("User").Preload("Category").Preload("Comments").
		Where("category_id = ?", categoryID).Find(&posts).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		err := s.Redis.SetCache(ctx, cacheKey, posts, redis.DefaultTTL)
		if err != nil {
			return nil, err
		}
	}

	return posts, nil
}

func (s *postService) GetPostsByUser(ctx context.Context, userID string) ([]models.Post, error) {
	var posts []models.Post
	cacheKey := fmt.Sprintf("%suser:%s", redis.PostCachePrefix, userID)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &posts); err == nil {
			return posts, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("User").Preload("Category").Preload("Comments").
		Where("created_by = ?", userID).Find(&posts).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		err := s.Redis.SetCache(ctx, cacheKey, posts, redis.DefaultTTL)
		if err != nil {
			return nil, err
		}
	}

	return posts, nil
}

func (s *postService) GetPublishedPosts(ctx context.Context) ([]models.Post, error) {
	var posts []models.Post
	cacheKey := fmt.Sprintf("%spublished", redis.PostCachePrefix)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &posts); err == nil {
			return posts, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("User").Preload("Category").Preload("Comments").
		Where("status = ?", "published").Find(&posts).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		err := s.Redis.SetCache(ctx, cacheKey, posts, redis.ShortTTL)
		if err != nil {
			return nil, err
		}
	}

	return posts, nil
}

func (s *postService) GetDraftPosts(ctx context.Context) ([]models.Post, error) {
	var posts []models.Post
	cacheKey := fmt.Sprintf("%sdraft", redis.PostCachePrefix)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &posts); err == nil {
			return posts, nil
		}
	}

	// Get from database
	if err := database.DB.Preload("User").Preload("Category").
		Where("status = ?", "draft").Find(&posts).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		err := s.Redis.SetCache(ctx, cacheKey, posts, redis.ShortTTL)
		if err != nil {
			return nil, err
		}
	}

	return posts, nil
}

func (s *postService) CreatePost(ctx context.Context, post *models.Post) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(post).Error; err != nil {
			return err
		}

		// Create related contents
		for i := range post.Contents {
			post.Contents[i].ModelID = post.ID
			post.Contents[i].ModelType = "post"
			if err := tx.Create(&post.Contents[i]).Error; err != nil {
				return err
			}
		}

		// Invalidate caches
		if s.Redis != nil {
			err := s.Redis.DeletePattern(ctx, fmt.Sprintf("%s*", redis.PostCachePrefix))
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *postService) UpdatePost(ctx context.Context, post *models.Post) error {
	return s.UpdateWithCache(ctx, post, redis.PostCachePrefix)
}

func (s *postService) DeletePost(ctx context.Context, id string) error {
	post := &models.Post{}
	if err := database.DB.First(post, id).Error; err != nil {
		return err
	}

	return s.DeleteWithCache(ctx, post, redis.PostCachePrefix)
}

func (s *postService) PublishPost(ctx context.Context, id string) error {
	post := &models.Post{}
	if err := database.DB.First(post, id).Error; err != nil {
		return err
	}

	post.Publish()

	// Update database
	if err := database.DB.Save(post).Error; err != nil {
		return err
	}

	// Invalidate caches
	if s.Redis != nil {
		err := s.Redis.InvalidatePostCache(ctx, id)
		if err != nil {
			return err
		}
		err = s.Redis.DeletePattern(ctx, fmt.Sprintf("%s*", redis.PostCachePrefix))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *postService) ArchivePost(ctx context.Context, id string) error {
	post := &models.Post{}
	if err := database.DB.First(post, id).Error; err != nil {
		return err
	}

	post.Archive()

	// Update database
	if err := database.DB.Save(post).Error; err != nil {
		return err
	}

	// Invalidate caches
	if s.Redis != nil {
		err := s.Redis.InvalidatePostCache(ctx, id)
		if err != nil {
			return err
		}
		err = s.Redis.DeletePattern(ctx, fmt.Sprintf("%s*", redis.PostCachePrefix))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *postService) SearchPosts(ctx context.Context, query string) ([]models.Post, error) {
	var posts []models.Post
	cacheKey := s.GetSearchCacheKey(redis.PostCachePrefix, query)

	// Try cache first
	if s.Redis != nil {
		if err := s.Redis.GetCache(ctx, cacheKey, &posts); err == nil {
			return posts, nil
		}
	}

	// Search in database
	if err := database.DB.Preload("User").Preload("Category").
		Where("title LIKE ? OR content LIKE ? OR excerpt LIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%").
		Find(&posts).Error; err != nil {
		return nil, err
	}

	// Cache the result
	if s.Redis != nil {
		err := s.Redis.SetCache(ctx, cacheKey, posts, redis.ShortTTL)
		if err != nil {
			return nil, err
		}
	}

	return posts, nil
}

func (s *postService) GetPostStats(ctx context.Context, postID string) (map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("%sstats:%s", redis.PostCachePrefix, postID)

	// Try cache first
	if s.Redis != nil {
		var cachedStats map[string]interface{}
		if err := s.Redis.GetCache(ctx, cacheKey, &cachedStats); err == nil {
			return cachedStats, nil
		}
	}

	// Calculate stats from database
	var commentCount int64
	var post models.Post

	if err := database.DB.First(&post, postID).Error; err != nil {
		return nil, err
	}

	if err := database.DB.Model(&models.Comment{}).Where("post_id = ?", postID).Count(&commentCount).Error; err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"post_id":       postID,
		"comment_count": commentCount,
		"status":        post.Status,
		"created_at":    post.CreatedAt,
		"updated_at":    post.UpdatedAt,
		"word_count":    len(post.Contents),
	}

	// Cache the stats
	if s.Redis != nil {
		err := s.Redis.SetCache(ctx, cacheKey, stats, redis.ShortTTL)
		if err != nil {
			return nil, err
		}
	}

	return stats, nil
}

func (s *postService) GetPostCount(ctx context.Context) (int64, error) {
	cacheKey := fmt.Sprintf("%scount", redis.PostCachePrefix)

	// Try cache first
	if s.Redis != nil {
		var count int64
		if err := s.Redis.GetCache(ctx, cacheKey, &count); err == nil {
			return count, nil
		}
	}

	// Get from database
	var count int64
	if err := database.DB.Model(&models.Post{}).Count(&count).Error; err != nil {
		return 0, err
	}

	// Cache the result
	if s.Redis != nil {
		err := s.Redis.SetCache(ctx, cacheKey, count, redis.LongTTL)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (s *postService) GetPublishedPostCount(ctx context.Context) (int64, error) {
	cacheKey := fmt.Sprintf("%spublished_count", redis.PostCachePrefix)

	// Try cache first
	if s.Redis != nil {
		var count int64
		if err := s.Redis.GetCache(ctx, cacheKey, &count); err == nil {
			return count, nil
		}
	}

	// Get from database
	var count int64
	if err := database.DB.Model(&models.Post{}).Where("status = ?", "published").Count(&count).Error; err != nil {
		return 0, err
	}

	// Cache the result
	if s.Redis != nil {
		err := s.Redis.SetCache(ctx, cacheKey, count, redis.LongTTL)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

// Global service instance
var PostSvc PostService
