package controllers

import (
	"net/http"

	"go-next/internal/http/requests"
	"go-next/internal/http/responses"
	"go-next/internal/models"
	"go-next/internal/services"
	"go-next/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PostHandler interface {
	GetPosts(c *gin.Context)
	GetPost(c *gin.Context)
	CreatePost(c *gin.Context)
	UpdatePost(c *gin.Context)
	DeletePost(c *gin.Context)
}

type postHandler struct {
	PostService services.PostService
}

func NewPostHandler(postService services.PostService) PostHandler {
	return &postHandler{PostService: postService}
}

// GetPosts godoc
// @Summary      List posts
// @Description  Get all posts with pagination
// @Tags         posts
// @Produce      json
// @Param        page      query     int    false "Page number"
// @Param        per_page  query     int    false "Items per page"
// @Param        search    query     string false "Search term"
// @Param        category  query     string false "Category ID filter"
// @Success      200       {object}  responses.LaravelPaginationResponse
// @Failure      500       {object}  map[string]string
// @Router       /posts [get]
func (h *postHandler) GetPosts(c *gin.Context) {
	params := responses.ParsePaginationParams(c)
	search := c.Query("search")
	categoryID := c.Query("category")

	offset := (params.Page - 1) * params.PerPage

	var posts []models.Post
	var total int64

	query := database.DB.Model(&models.Post{}).Preload("Category").Preload("User")

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
		responses.SendError(c, http.StatusInternalServerError, "Failed to count posts")
		return
	}

	// Get paginated posts
	if err := query.Offset(offset).Limit(params.PerPage).Order("created_at DESC").Find(&posts).Error; err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch posts")
		return
	}

	// Send Laravel-style pagination response
	responses.SendLaravelPaginationWithMessage(c, "Posts retrieved successfully", posts, total, int64(params.Page), int64(params.PerPage))
}

// GetPost godoc
// @Summary      Get post
// @Description  Get a post by ID
// @Tags         posts
// @Produce      json
// @Param        id   path      int  true  "Post ID"
// @Success      200  {object}  models.Post
// @Failure      404  {object}  map[string]string
// @Router       /posts/{id} [get]
func (h *postHandler) GetPost(c *gin.Context) {
	id := c.Param("id")
	post, err := h.PostService.GetPostByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}
	c.JSON(http.StatusOK, post)
}

// CreatePost godoc
// @Summary      Create post
// @Description  Create a new post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        post  body      models.Post  true  "Post to create"
// @Success      201   {object}  models.Post
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /posts [post]
func (h *postHandler) CreatePost(c *gin.Context) {
	var input requests.PostCreateRequest
	if !requests.ValidateRequest(c, &input) {
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}
	post := models.Post{
		Title:      input.Title,
		Content:    input.Content,
		CategoryID: &input.CategoryID,
		BaseModelWithUser: models.BaseModelWithUser{
			CreatedBy: &uid,
			UpdatedBy: &uid,
		},
	}
	if err := h.PostService.CreatePost(&post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}
	c.JSON(http.StatusCreated, post)
}

// UpdatePost godoc
// @Summary      Update post
// @Description  Update an existing post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id    path      int         true  "Post ID"
// @Param        post  body      models.Post true  "Post to update"
// @Success      200   {object}  models.Post
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /posts/{id} [put]
func (h *postHandler) UpdatePost(c *gin.Context) {
	id := c.Param("id")
	post, err := h.PostService.GetPostByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}
	var input requests.PostUpdateRequest
	if !requests.ValidateRequest(c, &input) {
		return
	}
	post.Title = input.Title
	post.Content = input.Content
	post.CategoryID = &input.CategoryID
	if err := h.PostService.UpdatePost(post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}
	c.JSON(http.StatusOK, post)
}

// DeletePost godoc
// @Summary      Delete post
// @Description  Delete a post by ID
// @Tags         posts
// @Param        id   path  int  true  "Post ID"
// @Success      204
// @Failure      500  {object}  map[string]string
// @Router       /posts/{id} [delete]
func (h *postHandler) DeletePost(c *gin.Context) {
	id := c.Param("id")
	if err := h.PostService.DeletePost(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}
	c.Status(http.StatusNoContent)
}
