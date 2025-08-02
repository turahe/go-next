package controllers

import (
	"net/http"
	"strconv"

	"go-next/internal/http/responses"
	"go-next/internal/models"
	"go-next/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BlogHandler interface {
	// Public blog endpoints
	GetPublicPosts(c *gin.Context)
	GetPublicPost(c *gin.Context)
	GetFeaturedPosts(c *gin.Context)
	GetRelatedPosts(c *gin.Context)
	GetPopularPosts(c *gin.Context)
	GetPostsByCategory(c *gin.Context)
	GetPostsByTag(c *gin.Context)
	SearchPosts(c *gin.Context)

	// Blog statistics
	GetBlogStats(c *gin.Context)
	GetCategoryStats(c *gin.Context)
	GetMonthlyArchive(c *gin.Context)

	// Categories and tags
	GetPublicCategories(c *gin.Context)
	GetPublicTags(c *gin.Context)
	GetCategoryBySlug(c *gin.Context)
	GetTagBySlug(c *gin.Context)

	// Post management (admin only)
	CreatePost(c *gin.Context)
	UpdatePost(c *gin.Context)
	DeletePost(c *gin.Context)
	PublishPost(c *gin.Context)
	UnpublishPost(c *gin.Context)
	ArchivePost(c *gin.Context)
	IncrementViewCount(c *gin.Context)
}

type blogHandler struct {
	BlogService services.BlogService
}

func NewBlogHandler(blogService services.BlogService) BlogHandler {
	return &blogHandler{BlogService: blogService}
}

// GetPublicPosts godoc
// @Summary      Get public posts
// @Description  Get published posts for public viewing with pagination
// @Tags         blog
// @Produce      json
// @Param        page      query     int    false "Page number" default(1)
// @Param        per_page  query     int    false "Items per page" default(10)
// @Param        search    query     string false "Search term"
// @Param        category  query     string false "Category slug filter"
// @Success      200       {object}  responses.LaravelPaginationResponse
// @Failure      500       {object}  map[string]string
// @Router       /blog/posts [get]
func (h *blogHandler) GetPublicPosts(c *gin.Context) {
	params := responses.ParsePaginationParams(c)
	search := c.Query("search")
	categorySlug := c.Query("category")

	posts, total, err := h.BlogService.GetPublicPosts(params.Page, params.PerPage, search, categorySlug)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch posts")
		return
	}

	responses.SendLaravelPaginationWithMessage(c, "Posts retrieved successfully", posts, total, int64(params.Page), int64(params.PerPage))
}

// GetPublicPost godoc
// @Summary      Get public post
// @Description  Get a published post by slug
// @Tags         blog
// @Produce      json
// @Param        slug   path      string  true  "Post slug"
// @Success      200    {object}  models.Post
// @Failure      404    {object}  map[string]string
// @Router       /blog/posts/{slug} [get]
func (h *blogHandler) GetPublicPost(c *gin.Context) {
	slug := c.Param("slug")

	post, err := h.BlogService.GetPublicPost(slug)
	if err != nil {
		responses.SendError(c, http.StatusNotFound, "Post not found")
		return
	}

	// Increment view count
	if postID, err := uuid.Parse(post.ID.String()); err == nil {
		go h.BlogService.IncrementViewCount(postID)
	}

	c.JSON(http.StatusOK, post)
}

// GetFeaturedPosts godoc
// @Summary      Get featured posts
// @Description  Get featured posts for homepage
// @Tags         blog
// @Produce      json
// @Param        limit query int false "Number of posts" default(5)
// @Success      200   {array}   models.Post
// @Failure      500   {object}  map[string]string
// @Router       /blog/posts/featured [get]
func (h *blogHandler) GetFeaturedPosts(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}

	posts, err := h.BlogService.GetFeaturedPosts(limit)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch featured posts")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Featured posts retrieved successfully",
		"data":    posts,
	})
}

// GetRelatedPosts godoc
// @Summary      Get related posts
// @Description  Get related posts based on category
// @Tags         blog
// @Produce      json
// @Param        post_id path      string true  "Post ID"
// @Param        limit   query     int    false "Number of posts" default(3)
// @Success      200     {array}   models.Post
// @Failure      500     {object}  map[string]string
// @Router       /blog/posts/{slug}/related [get]
func (h *blogHandler) GetRelatedPosts(c *gin.Context) {
	slug := c.Param("slug")

	// First get the post to get its ID
	post, err := h.BlogService.GetPublicPost(slug)
	if err != nil {
		responses.SendError(c, http.StatusNotFound, "Post not found")
		return
	}

	limitStr := c.DefaultQuery("limit", "3")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 3
	}

	posts, err := h.BlogService.GetRelatedPosts(post.ID, limit)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch related posts")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Related posts retrieved successfully",
		"data":    posts,
	})
}

// GetPopularPosts godoc
// @Summary      Get popular posts
// @Description  Get popular posts based on view count
// @Tags         blog
// @Produce      json
// @Param        limit query int false "Number of posts" default(10)
// @Param        days  query int false "Days to look back" default(30)
// @Success      200   {array}   models.Post
// @Failure      500   {object}  map[string]string
// @Router       /blog/posts/popular [get]
func (h *blogHandler) GetPopularPosts(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}

	posts, err := h.BlogService.GetPopularPosts(limit, days)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch popular posts")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Popular posts retrieved successfully",
		"data":    posts,
	})
}

// GetPostsByCategory godoc
// @Summary      Get posts by category
// @Description  Get posts filtered by category slug
// @Tags         blog
// @Produce      json
// @Param        slug path      string true  "Category slug"
// @Param        page          query     int    false "Page number" default(1)
// @Param        per_page      query     int    false "Items per page" default(10)
// @Success      200           {object}  responses.LaravelPaginationResponse
// @Failure      500           {object}  map[string]string
// @Router       /blog/categories/{slug}/posts [get]
func (h *blogHandler) GetPostsByCategory(c *gin.Context) {
	categorySlug := c.Param("slug")
	params := responses.ParsePaginationParams(c)

	posts, total, err := h.BlogService.GetPostsByCategory(categorySlug, params.Page, params.PerPage)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch posts")
		return
	}

	responses.SendLaravelPaginationWithMessage(c, "Posts retrieved successfully", posts, total, int64(params.Page), int64(params.PerPage))
}

// GetPostsByTag godoc
// @Summary      Get posts by tag
// @Description  Get posts filtered by tag slug
// @Tags         blog
// @Produce      json
// @Param        slug path      string true  "Tag slug"
// @Param        page      query     int    false "Page number" default(1)
// @Param        per_page  query     int    false "Items per page" default(10)
// @Success      200       {object}  responses.LaravelPaginationResponse
// @Failure      500       {object}  map[string]string
// @Router       /blog/tags/{slug}/posts [get]
func (h *blogHandler) GetPostsByTag(c *gin.Context) {
	tagSlug := c.Param("slug")
	params := responses.ParsePaginationParams(c)

	posts, total, err := h.BlogService.GetPostsByTag(tagSlug, params.Page, params.PerPage)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch posts")
		return
	}

	responses.SendLaravelPaginationWithMessage(c, "Posts retrieved successfully", posts, total, int64(params.Page), int64(params.PerPage))
}

// SearchPosts godoc
// @Summary      Search posts
// @Description  Search posts by query
// @Tags         blog
// @Produce      json
// @Param        query     query     string true  "Search query"
// @Param        page      query     int    false "Page number" default(1)
// @Param        per_page  query     int    false "Items per page" default(10)
// @Success      200       {object}  responses.LaravelPaginationResponse
// @Failure      500       {object}  map[string]string
// @Router       /blog/search [get]
func (h *blogHandler) SearchPosts(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		responses.SendError(c, http.StatusBadRequest, "Search query is required")
		return
	}

	params := responses.ParsePaginationParams(c)

	posts, total, err := h.BlogService.SearchPosts(query, params.Page, params.PerPage)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to search posts")
		return
	}

	responses.SendLaravelPaginationWithMessage(c, "Search results retrieved successfully", posts, total, int64(params.Page), int64(params.PerPage))
}

// GetBlogStats godoc
// @Summary      Get blog statistics
// @Description  Get overall blog statistics
// @Tags         blog
// @Produce      json
// @Success      200  {object}  services.BlogStats
// @Failure      500  {object}  map[string]string
// @Router       /blog/stats [get]
func (h *blogHandler) GetBlogStats(c *gin.Context) {
	stats, err := h.BlogService.GetBlogStats()
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch blog statistics")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blog statistics retrieved successfully",
		"data":    stats,
	})
}

// GetCategoryStats godoc
// @Summary      Get category statistics
// @Description  Get statistics for each category
// @Tags         blog
// @Produce      json
// @Success      200  {array}   services.CategoryStats
// @Failure      500  {object}  map[string]string
// @Router       /blog/stats/categories [get]
func (h *blogHandler) GetCategoryStats(c *gin.Context) {
	stats, err := h.BlogService.GetCategoryStats()
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch category statistics")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category statistics retrieved successfully",
		"data":    stats,
	})
}

// GetMonthlyArchive godoc
// @Summary      Get monthly archive
// @Description  Get monthly post counts
// @Tags         blog
// @Produce      json
// @Success      200  {array}   services.MonthlyArchive
// @Failure      500  {object}  map[string]string
// @Router       /blog/archive [get]
func (h *blogHandler) GetMonthlyArchive(c *gin.Context) {
	archives, err := h.BlogService.GetMonthlyArchive()
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch monthly archive")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Monthly archive retrieved successfully",
		"data":    archives,
	})
}

// GetPublicCategories godoc
// @Summary      Get public categories
// @Description  Get active categories for public viewing
// @Tags         blog
// @Produce      json
// @Success      200  {array}   models.Category
// @Failure      500  {object}  map[string]string
// @Router       /blog/categories [get]
func (h *blogHandler) GetPublicCategories(c *gin.Context) {
	categories, err := h.BlogService.GetPublicCategories()
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Categories retrieved successfully",
		"data":    categories,
	})
}

// GetPublicTags godoc
// @Summary      Get public tags
// @Description  Get active tags for public viewing
// @Tags         blog
// @Produce      json
// @Success      200  {array}   models.Tag
// @Failure      500  {object}  map[string]string
// @Router       /blog/tags [get]
func (h *blogHandler) GetPublicTags(c *gin.Context) {
	tags, err := h.BlogService.GetPublicTags()
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch tags")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tags retrieved successfully",
		"data":    tags,
	})
}

// GetCategoryBySlug godoc
// @Summary      Get category by slug
// @Description  Get a category by its slug
// @Tags         blog
// @Produce      json
// @Param        slug path      string true "Category slug"
// @Success      200  {object}  models.Category
// @Failure      404  {object}  map[string]string
// @Router       /blog/categories/{slug} [get]
func (h *blogHandler) GetCategoryBySlug(c *gin.Context) {
	slug := c.Param("slug")

	category, err := h.BlogService.GetCategoryBySlug(slug)
	if err != nil {
		responses.SendError(c, http.StatusNotFound, "Category not found")
		return
	}

	c.JSON(http.StatusOK, category)
}

// GetTagBySlug godoc
// @Summary      Get tag by slug
// @Description  Get a tag by its slug
// @Tags         blog
// @Produce      json
// @Param        slug path      string true "Tag slug"
// @Success      200  {object}  models.Tag
// @Failure      404  {object}  map[string]string
// @Router       /blog/tags/{slug} [get]
func (h *blogHandler) GetTagBySlug(c *gin.Context) {
	slug := c.Param("slug")

	tag, err := h.BlogService.GetTagBySlug(slug)
	if err != nil {
		responses.SendError(c, http.StatusNotFound, "Tag not found")
		return
	}

	c.JSON(http.StatusOK, tag)
}

// CreatePost godoc
// @Summary      Create post (Admin only)
// @Description  Create a new blog post
// @Tags         blog
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        post body models.Post true "Post to create"
// @Success      201   {object}  models.Post
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /blog/posts [post]
func (h *blogHandler) CreatePost(c *gin.Context) {
	var post models.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Set user ID from context
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			post.CreatedBy = &uid
			post.UpdatedBy = &uid
		}
	}

	if err := h.BlogService.CreatePost(&post); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to create post")
		return
	}

	c.JSON(http.StatusCreated, post)
}

// UpdatePost godoc
// @Summary      Update post (Admin only)
// @Description  Update an existing blog post
// @Tags         blog
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string      true  "Post ID"
// @Param        post body      models.Post true  "Post to update"
// @Success      200   {object}  models.Post
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /blog/posts/{id} [put]
func (h *blogHandler) UpdatePost(c *gin.Context) {
	_ = c.Param("id") // ID is not used in this implementation

	var post models.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Set user ID from context
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			post.UpdatedBy = &uid
		}
	}

	if err := h.BlogService.UpdatePost(&post); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to update post")
		return
	}

	c.JSON(http.StatusOK, post)
}

// DeletePost godoc
// @Summary      Delete post (Admin only)
// @Description  Delete a blog post
// @Tags         blog
// @Security     BearerAuth
// @Param        id   path  string true "Post ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /blog/posts/{id} [delete]
func (h *blogHandler) DeletePost(c *gin.Context) {
	id := c.Param("id")

	if err := h.BlogService.DeletePost(id); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to delete post")
		return
	}

	c.Status(http.StatusNoContent)
}

// PublishPost godoc
// @Summary      Publish post (Admin only)
// @Description  Publish a blog post
// @Tags         blog
// @Security     BearerAuth
// @Param        id   path  string true "Post ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /blog/posts/{id}/publish [post]
func (h *blogHandler) PublishPost(c *gin.Context) {
	id := c.Param("id")

	if err := h.BlogService.PublishPost(id); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to publish post")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post published successfully"})
}

// UnpublishPost godoc
// @Summary      Unpublish post (Admin only)
// @Description  Unpublish a blog post
// @Tags         blog
// @Security     BearerAuth
// @Param        id   path  string true "Post ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /blog/posts/{id}/unpublish [post]
func (h *blogHandler) UnpublishPost(c *gin.Context) {
	id := c.Param("id")

	if err := h.BlogService.UnpublishPost(id); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to unpublish post")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post unpublished successfully"})
}

// ArchivePost godoc
// @Summary      Archive post (Admin only)
// @Description  Archive a blog post
// @Tags         blog
// @Security     BearerAuth
// @Param        id   path  string true "Post ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /blog/posts/{id}/archive [post]
func (h *blogHandler) ArchivePost(c *gin.Context) {
	id := c.Param("id")

	if err := h.BlogService.ArchivePost(id); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to archive post")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post archived successfully"})
}

// IncrementViewCount godoc
// @Summary      Increment view count
// @Description  Increment the view count for a post
// @Tags         blog
// @Param        id   path  string true "Post ID"
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /blog/posts/{id}/view [post]
func (h *blogHandler) IncrementViewCount(c *gin.Context) {
	id := c.Param("id")
	postID, err := uuid.Parse(id)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid post ID")
		return
	}

	if err := h.BlogService.IncrementViewCount(postID); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to increment view count")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "View count incremented successfully"})
}
