package controllers

import (
	"net/http"
	"time"

	"go-next/internal/http/responses"
	"go-next/internal/models"
	"go-next/internal/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OptimizedBlogHandler demonstrates the optimized controller pattern
type OptimizedBlogHandler struct {
	*BaseHandler
	BlogService services.BlogService
}

// NewOptimizedBlogHandler creates a new optimized blog handler
func NewOptimizedBlogHandler(blogService services.BlogService, logger *zap.Logger) *OptimizedBlogHandler {
	return &OptimizedBlogHandler{
		BaseHandler: NewBaseHandler(logger),
		BlogService: blogService,
	}
}

// GetPublicPosts demonstrates optimized request handling
func (h *OptimizedBlogHandler) GetPublicPosts(c *gin.Context) {
	startTime := time.Now()
	h.LogRequest(c, "GetPublicPosts")

	// Use base handler utilities for parameter extraction
	page := h.GetQueryAsInt(c, "page", 1)
	perPage := h.GetQueryAsInt(c, "per_page", 10)
	search := c.Query("search")
	categorySlug := c.Query("category")

	posts, total, err := h.BlogService.GetPublicPosts(page, perPage, search, categorySlug)
	if err != nil {
		h.HandleServiceError(c, err, "GetPublicPosts")
		return
	}

	// Convert to responses using DTOs
	postResponses := responses.ToBlogPostSimpleResponses(posts)
	responses.SendLaravelPaginationWithMessage(c, "Posts retrieved successfully", postResponses, total, int64(page), int64(perPage))

	h.LogResponse(c, "GetPublicPosts", http.StatusOK, time.Since(startTime).Milliseconds())
}

// GetPublicPost demonstrates optimized single resource handling
func (h *OptimizedBlogHandler) GetPublicPost(c *gin.Context) {
	startTime := time.Now()
	h.LogRequest(c, "GetPublicPost")

	// Use base handler for parameter validation
	slug := c.Param("slug")
	if slug == "" {
		responses.SendError(c, http.StatusBadRequest, "Missing slug parameter")
		return
	}

	post, err := h.BlogService.GetPublicPost(slug)
	if err != nil {
		h.HandleServiceError(c, err, "GetPublicPost")
		return
	}

	// Convert to response DTO
	postResponse := responses.ToBlogPostResponse(post)
	responses.SendSuccess(c, http.StatusOK, "Post retrieved successfully", postResponse)

	h.LogResponse(c, "GetPublicPost", http.StatusOK, time.Since(startTime).Milliseconds())
}

// GetFeaturedPosts demonstrates optimized list handling
func (h *OptimizedBlogHandler) GetFeaturedPosts(c *gin.Context) {
	startTime := time.Now()
	h.LogRequest(c, "GetFeaturedPosts")

	limit := h.GetQueryAsInt(c, "limit", 5)

	posts, err := h.BlogService.GetFeaturedPosts(limit)
	if err != nil {
		h.HandleServiceError(c, err, "GetFeaturedPosts")
		return
	}

	postResponses := responses.ToBlogPostSimpleResponses(posts)
	responses.SendSuccess(c, http.StatusOK, "Featured posts retrieved successfully", postResponses)

	h.LogResponse(c, "GetFeaturedPosts", http.StatusOK, time.Since(startTime).Milliseconds())
}

// GetRelatedPosts demonstrates optimized related resource handling
func (h *OptimizedBlogHandler) GetRelatedPosts(c *gin.Context) {
	startTime := time.Now()
	h.LogRequest(c, "GetRelatedPosts")

	slug := c.Param("slug")
	if slug == "" {
		responses.SendError(c, http.StatusBadRequest, "Missing slug parameter")
		return
	}

	limit := h.GetQueryAsInt(c, "limit", 3)

	// First get the post to get its ID
	post, err := h.BlogService.GetPublicPost(slug)
	if err != nil {
		h.HandleServiceError(c, err, "GetRelatedPosts")
		return
	}

	relatedPosts, err := h.BlogService.GetRelatedPosts(post.ID, limit)
	if err != nil {
		h.HandleServiceError(c, err, "GetRelatedPosts")
		return
	}

	postResponses := responses.ToBlogPostSimpleResponses(relatedPosts)
	responses.SendSuccess(c, http.StatusOK, "Related posts retrieved successfully", postResponses)

	h.LogResponse(c, "GetRelatedPosts", http.StatusOK, time.Since(startTime).Milliseconds())
}

// GetPopularPosts demonstrates optimized query handling
func (h *OptimizedBlogHandler) GetPopularPosts(c *gin.Context) {
	startTime := time.Now()
	h.LogRequest(c, "GetPopularPosts")

	limit := h.GetQueryAsInt(c, "limit", 10)
	days := h.GetQueryAsInt(c, "days", 30)

	posts, err := h.BlogService.GetPopularPosts(limit, days)
	if err != nil {
		h.HandleServiceError(c, err, "GetPopularPosts")
		return
	}

	postResponses := responses.ToBlogPostSimpleResponses(posts)
	responses.SendSuccess(c, http.StatusOK, "Popular posts retrieved successfully", postResponses)

	h.LogResponse(c, "GetPopularPosts", http.StatusOK, time.Since(startTime).Milliseconds())
}

// GetPostsByCategory demonstrates optimized filtered list handling
func (h *OptimizedBlogHandler) GetPostsByCategory(c *gin.Context) {
	startTime := time.Now()
	h.LogRequest(c, "GetPostsByCategory")

	slug := c.Param("slug")
	if slug == "" {
		responses.SendError(c, http.StatusBadRequest, "Missing slug parameter")
		return
	}

	page := h.GetQueryAsInt(c, "page", 1)
	perPage := h.GetQueryAsInt(c, "per_page", 10)

	posts, total, err := h.BlogService.GetPostsByCategory(slug, page, perPage)
	if err != nil {
		h.HandleServiceError(c, err, "GetPostsByCategory")
		return
	}

	postResponses := responses.ToBlogPostSimpleResponses(posts)
	responses.SendLaravelPaginationWithMessage(c, "Posts retrieved successfully", postResponses, total, int64(page), int64(perPage))

	h.LogResponse(c, "GetPostsByCategory", http.StatusOK, time.Since(startTime).Milliseconds())
}

// GetPostsByTag demonstrates optimized tag filtering
func (h *OptimizedBlogHandler) GetPostsByTag(c *gin.Context) {
	startTime := time.Now()
	h.LogRequest(c, "GetPostsByTag")

	slug := c.Param("slug")
	if slug == "" {
		responses.SendError(c, http.StatusBadRequest, "Missing slug parameter")
		return
	}

	page := h.GetQueryAsInt(c, "page", 1)
	perPage := h.GetQueryAsInt(c, "per_page", 10)

	posts, total, err := h.BlogService.GetPostsByTag(slug, page, perPage)
	if err != nil {
		h.HandleServiceError(c, err, "GetPostsByTag")
		return
	}

	postResponses := responses.ToBlogPostSimpleResponses(posts)
	responses.SendLaravelPaginationWithMessage(c, "Posts retrieved successfully", postResponses, total, int64(page), int64(perPage))

	h.LogResponse(c, "GetPostsByTag", http.StatusOK, time.Since(startTime).Milliseconds())
}

// SearchPosts demonstrates optimized search handling
func (h *OptimizedBlogHandler) SearchPosts(c *gin.Context) {
	startTime := time.Now()
	h.LogRequest(c, "SearchPosts")

	query := c.Query("q")
	if query == "" {
		responses.SendError(c, http.StatusBadRequest, "Search query is required")
		return
	}

	page := h.GetQueryAsInt(c, "page", 1)
	perPage := h.GetQueryAsInt(c, "per_page", 10)

	posts, total, err := h.BlogService.SearchPosts(query, page, perPage)
	if err != nil {
		h.HandleServiceError(c, err, "SearchPosts")
		return
	}

	postResponses := responses.ToBlogPostSimpleResponses(posts)
	responses.SendLaravelPaginationWithMessage(c, "Search results retrieved successfully", postResponses, total, int64(page), int64(perPage))

	h.LogResponse(c, "SearchPosts", http.StatusOK, time.Since(startTime).Milliseconds())
}

// GetBlogStats demonstrates optimized statistics handling
func (h *OptimizedBlogHandler) GetBlogStats(c *gin.Context) {
	startTime := time.Now()
	h.LogRequest(c, "GetBlogStats")

	stats, err := h.BlogService.GetBlogStats()
	if err != nil {
		h.HandleServiceError(c, err, "GetBlogStats")
		return
	}

	// Convert stats to response format
	statsResponse := map[string]interface{}{
		"total_posts":      stats.TotalPosts,
		"published_posts":  stats.PublishedPosts,
		"total_views":      stats.TotalViews,
		"total_comments":   stats.TotalComments,
		"total_categories": stats.TotalCategories,
		"total_tags":       stats.TotalTags,
	}

	responses.SendSuccess(c, http.StatusOK, "Blog statistics retrieved successfully", statsResponse)

	h.LogResponse(c, "GetBlogStats", http.StatusOK, time.Since(startTime).Milliseconds())
}

// GetPublicCategories demonstrates optimized category handling
func (h *OptimizedBlogHandler) GetPublicCategories(c *gin.Context) {
	startTime := time.Now()
	h.LogRequest(c, "GetPublicCategories")

	categories, err := h.BlogService.GetPublicCategories()
	if err != nil {
		h.HandleServiceError(c, err, "GetPublicCategories")
		return
	}

	// Convert to response format
	categoryResponses := make([]map[string]interface{}, len(categories))
	for i, category := range categories {
		categoryResponses[i] = map[string]interface{}{
			"id":   category.ID,
			"name": category.Name,
			"slug": category.Slug,
		}
	}

	responses.SendSuccess(c, http.StatusOK, "Categories retrieved successfully", categoryResponses)

	h.LogResponse(c, "GetPublicCategories", http.StatusOK, time.Since(startTime).Milliseconds())
}

// GetPublicTags demonstrates optimized tag handling
func (h *OptimizedBlogHandler) GetPublicTags(c *gin.Context) {
	startTime := time.Now()
	h.LogRequest(c, "GetPublicTags")

	tags, err := h.BlogService.GetPublicTags()
	if err != nil {
		h.HandleServiceError(c, err, "GetPublicTags")
		return
	}

	// Convert to response format
	tagResponses := make([]map[string]interface{}, len(tags))
	for i, tag := range tags {
		tagResponses[i] = map[string]interface{}{
			"id":   tag.ID,
			"name": tag.Name,
			"slug": tag.Slug,
		}
	}

	responses.SendSuccess(c, http.StatusOK, "Tags retrieved successfully", tagResponses)

	h.LogResponse(c, "GetPublicTags", http.StatusOK, time.Since(startTime).Milliseconds())
}

// IncrementViewCount demonstrates optimized background processing
func (h *OptimizedBlogHandler) IncrementViewCount(c *gin.Context) {
	startTime := time.Now()
	h.LogRequest(c, "IncrementViewCount")

	slug := c.Param("slug")
	if slug == "" {
		responses.SendError(c, http.StatusBadRequest, "Missing slug parameter")
		return
	}

	// Process view count increment in background for better performance
	go func() {
		// Note: This would need to be implemented in the service layer
		// For now, we'll just log the request
		h.logger.Info("View count increment requested",
			zap.String("slug", slug),
		)
	}()

	responses.SendSuccess(c, http.StatusOK, "View count update queued", nil)

	h.LogResponse(c, "IncrementViewCount", http.StatusOK, time.Since(startTime).Milliseconds())
}

// CreatePost demonstrates optimized resource creation with validation
func (h *OptimizedBlogHandler) CreatePost(c *gin.Context) {
	startTime := time.Now()
	h.LogRequest(c, "CreatePost")

	// Validate user permissions
	if err := h.ValidatePermission(c, "post:create"); err != nil {
		responses.SendError(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	var post models.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Get current user ID
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	post.CreatedBy = &userID
	post.UpdatedBy = &userID

	if err := h.BlogService.CreatePost(&post); err != nil {
		h.HandleServiceError(c, err, "CreatePost")
		return
	}

	postResponse := responses.ToBlogPostResponse(&post)
	responses.SendSuccess(c, http.StatusCreated, "Post created successfully", postResponse)

	h.LogResponse(c, "CreatePost", http.StatusCreated, time.Since(startTime).Milliseconds())
}

// UpdatePost demonstrates optimized resource updating
func (h *OptimizedBlogHandler) UpdatePost(c *gin.Context) {
	startTime := time.Now()
	h.LogRequest(c, "UpdatePost")

	// Validate user permissions
	if err := h.ValidatePermission(c, "post:update"); err != nil {
		responses.SendError(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	postID, err := h.GetParamAsUUID(c, "id")
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid post ID")
		return
	}

	var post models.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	post.ID = postID

	// Get current user ID for audit trail
	userID, err := h.GetUserIDFromContext(c)
	if err != nil {
		responses.SendError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	post.UpdatedBy = &userID

	if err := h.BlogService.UpdatePost(&post); err != nil {
		h.HandleServiceError(c, err, "UpdatePost")
		return
	}

	postResponse := responses.ToBlogPostResponse(&post)
	responses.SendSuccess(c, http.StatusOK, "Post updated successfully", postResponse)

	h.LogResponse(c, "UpdatePost", http.StatusOK, time.Since(startTime).Milliseconds())
}

// DeletePost demonstrates optimized resource deletion
func (h *OptimizedBlogHandler) DeletePost(c *gin.Context) {
	startTime := time.Now()
	h.LogRequest(c, "DeletePost")

	// Validate user permissions
	if err := h.ValidatePermission(c, "post:delete"); err != nil {
		responses.SendError(c, http.StatusForbidden, "Insufficient permissions")
		return
	}

	postID, err := h.GetParamAsUUID(c, "id")
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid post ID")
		return
	}

	if err := h.BlogService.DeletePost(postID.String()); err != nil {
		h.HandleServiceError(c, err, "DeletePost")
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Post deleted successfully", nil)

	h.LogResponse(c, "DeletePost", http.StatusOK, time.Since(startTime).Milliseconds())
}
