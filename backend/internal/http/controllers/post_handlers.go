package controllers

import (
	"net/http"

	"wordpress-go-next/backend/internal/http/requests"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/internal/services"

	"github.com/gin-gonic/gin"
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
// @Description  Get all posts
// @Tags         posts
// @Produce      json
// @Success      200  {array}   models.Post
// @Failure      500  {object}  map[string]string
// @Router       /posts [get]
func (h *postHandler) GetPosts(c *gin.Context) {
	posts, err := h.PostService.GetAllPosts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}
	c.JSON(http.StatusOK, posts)
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
	post, err := h.PostService.GetPostByID(c.Request.Context(), id)
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
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	post := models.Post{
		Title:      input.Title,
		CategoryID: uint(input.CategoryID),
	}
	for _, c := range input.Contents {
		post.Contents = append(post.Contents, models.Content{
			Content:   c.Content,
			Type:      c.Type,
			ModelType: "post",
		})
	}
	if err := h.PostService.CreatePost(c.Request.Context(), &post); err != nil {
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
	post, err := h.PostService.GetPostByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}
	var input requests.PostUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	post.Title = input.Title
	post.CategoryID = uint(input.CategoryID)
	post.Contents = nil
	for _, c := range input.Contents {
		post.Contents = append(post.Contents, models.Content{
			Content:   c.Content,
			Type:      c.Type,
			ModelType: "post",
		})
	}
	if err := h.PostService.UpdatePost(c.Request.Context(), post); err != nil {
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
	if err := h.PostService.DeletePost(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}
	c.Status(http.StatusNoContent)
}
