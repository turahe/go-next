package controllers

import (
	"net/http"

	"wordpress-go-next/backend/internal/http/dto"
	"wordpress-go-next/backend/internal/http/requests"
	"wordpress-go-next/backend/internal/http/responses"
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
// @Description  Get posts with pagination and optional search
// @Tags         posts
// @Produce      json
// @Param        page     query     int     false  "Page number"  default(1)
// @Param        perPage  query     int     false  "Items per page"  default(10)
// @Param        search   query     string  false  "Search keyword"
// @Success      200  {object}  responses.PaginationResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /posts [get]
func (h *postHandler) GetPosts(c *gin.Context) {
	pagination, err := requests.ParsePaginationFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid pagination parameters",
		})
		return
	}

	result, err := h.PostService.GetPostsWithPagination(c.Request.Context(), pagination.Page, pagination.PerPage, pagination.Search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to fetch posts",
		})
		return
	}

	posts, ok := result.Data.([]models.Post)
	if !ok {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Invalid data format",
		})
		return
	}
	dtos := dto.ToPostDTOs(posts)
	result.Data = dtos

	c.JSON(http.StatusOK, result)
}

// GetPost godoc
// @Summary      Get post
// @Description  Get a post by ID
// @Tags         posts
// @Produce      json
// @Param        id   path      int  true  "Post ID"
// @Success      200  {object}  dto.PostDTO
// @Failure      404  {object}  map[string]string
// @Router       /posts/{id} [get]
func (h *postHandler) GetPost(c *gin.Context) {
	id := c.Param("id")
	post, err := h.PostService.GetPostByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, responses.CommonResponse{
			ResponseCode:    http.StatusNotFound,
			ResponseMessage: "Post not found",
		})
		return
	}
	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Post fetched successfully",
		Data:            dto.ToPostDTO(post),
	})
}

// CreatePost godoc
// @Summary      Create post
// @Description  Create a new post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        post  body      models.Post  true  "Post to create"
// @Success      201   {object}  dto.PostDTO
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /posts [post]
func (h *postHandler) CreatePost(c *gin.Context) {
	var input requests.PostCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid request",
		})
		return
	}
	post := models.Post{
		Title:      input.Title,
		CategoryID: uint(input.CategoryID),
	}
	for _, ctn := range input.Contents {
		post.Contents = append(post.Contents, models.Content{
			Content:   ctn.Content,
			Type:      ctn.Type,
			ModelType: "post",
		})
	}
	if err := h.PostService.CreatePost(c.Request.Context(), &post); err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to create post",
		})
		return
	}
	c.JSON(http.StatusCreated, responses.CommonResponse{
		ResponseCode:    http.StatusCreated,
		ResponseMessage: "Post created successfully",
		Data:            dto.ToPostDTO(&post),
	})
}

// UpdatePost godoc
// @Summary      Update post
// @Description  Update an existing post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id    path      int         true  "Post ID"
// @Param        post  body      models.Post true  "Post to update"
// @Success      200   {object}  dto.PostDTO
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /posts/{id} [put]
func (h *postHandler) UpdatePost(c *gin.Context) {
	id := c.Param("id")
	post, err := h.PostService.GetPostByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, responses.CommonResponse{
			ResponseCode:    http.StatusNotFound,
			ResponseMessage: "Post not found",
		})
		return
	}
	var input requests.PostUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid request",
		})
		return
	}
	post.Title = input.Title
	post.CategoryID = uint(input.CategoryID)
	post.Contents = nil
	for _, ctn := range input.Contents {
		post.Contents = append(post.Contents, models.Content{
			Content:   ctn.Content,
			Type:      ctn.Type,
			ModelType: "post",
		})
	}
	if err := h.PostService.UpdatePost(c.Request.Context(), post); err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to update post",
		})
		return
	}
	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Post updated successfully",
		Data:            dto.ToPostDTO(post),
	})
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
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to delete post",
		})
		return
	}
	c.JSON(http.StatusNoContent, responses.CommonResponse{
		ResponseCode:    http.StatusNoContent,
		ResponseMessage: "Post deleted successfully",
	})
}
