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

type CommentHandler interface {
	GetCommentsByPost(c *gin.Context)
	GetComment(c *gin.Context)
	CreateComment(c *gin.Context)
	UpdateComment(c *gin.Context)
	DeleteComment(c *gin.Context)
	CreateCommentNested(c *gin.Context)
	MoveCommentNested(c *gin.Context)
	DeleteCommentNested(c *gin.Context)
	GetSiblingComments(c *gin.Context)
	GetParentComment(c *gin.Context)
	GetDescendantComments(c *gin.Context)
	GetChildrenComments(c *gin.Context)
}

type commentHandler struct {
	CommentService services.CommentService
}

func NewCommentHandler(commentService services.CommentService) CommentHandler {
	return &commentHandler{CommentService: commentService}
}

// GetCommentsByPost godoc
// @Summary      List comments for a post
// @Description  Get all comments for a specific post with pagination
// @Tags         comments
// @Produce      json
// @Param        post_id   path      string true  "Post ID"
// @Param        page      query     int    false "Page number"
// @Param        per_page  query     int    false "Items per page"
// @Success      200       {object}  responses.LaravelPaginationResponse
// @Failure      500       {object}  map[string]string
// @Router       /posts/{post_id}/comments [get]
func (h *commentHandler) GetCommentsByPost(c *gin.Context) {
	postID := c.Param("post_id")
	params := responses.ParsePaginationParams(c)

	// Parse post ID
	parsedPostID, err := uuid.Parse(postID)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid post ID")
		return
	}

	offset := (params.Page - 1) * params.PerPage

	var comments []models.Comment
	var total int64

	query := database.DB.Model(&models.Comment{}).Where("post_id = ?", parsedPostID).Preload("User")

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to count comments")
		return
	}

	// Get paginated comments
	if err := query.Offset(offset).Limit(params.PerPage).Order("created_at DESC").Find(&comments).Error; err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch comments")
		return
	}

	// Send Laravel-style pagination response
	responses.SendLaravelPaginationWithMessage(c, "Comments retrieved successfully", comments, total, int64(params.Page), int64(params.PerPage))
}

// GetComment godoc
// @Summary      Get comment
// @Description  Get a comment by ID
// @Tags         comments
// @Produce      json
// @Param        id   path      int  true  "Comment ID"
// @Success      200  {object}  models.Comment
// @Failure      404  {object}  map[string]string
// @Router       /comments/{id} [get]
func (h *commentHandler) GetComment(c *gin.Context) {
	id := c.Param("id")
	comment, err := h.CommentService.GetCommentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}
	c.JSON(http.StatusOK, comment)
}

// CreateComment godoc
// @Summary      Create comment
// @Description  Create a new comment
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        comment  body      models.Comment  true  "Comment to create"
// @Success      201   {object}  models.Comment
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /comments [post]
func (h *commentHandler) CreateComment(c *gin.Context) {
	var input requests.CommentCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	comment := models.Comment{
		Content: input.Content,
		UserID:  input.UserID,
		PostID:  input.PostID,
	}
	if err := h.CommentService.CreateComment(&comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}
	c.JSON(http.StatusCreated, comment)
}

// UpdateComment godoc
// @Summary      Update comment
// @Description  Update an existing comment
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id       path      int            true  "Comment ID"
// @Param        comment  body      models.Comment true  "Comment to update"
// @Success      200   {object}  models.Comment
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /comments/{id} [put]
func (h *commentHandler) UpdateComment(c *gin.Context) {
	id := c.Param("id")
	comment, err := h.CommentService.GetCommentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}
	var input requests.CommentUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	comment.Content = input.Content
	comment.UserID = input.UserID
	comment.PostID = input.PostID
	if err := h.CommentService.UpdateComment(comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment"})
		return
	}
	c.JSON(http.StatusOK, comment)
}

// DeleteComment godoc
// @Summary      Delete comment
// @Description  Delete a comment by ID
// @Tags         comments
// @Param        id   path  int  true  "Comment ID"
// @Success      204
// @Failure      500  {object}  map[string]string
// @Router       /comments/{id} [delete]
func (h *commentHandler) DeleteComment(c *gin.Context) {
	id := c.Param("id")
	if err := h.CommentService.DeleteComment(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}
	c.Status(http.StatusNoContent)
}

// CreateCommentNested godoc
// @Summary      Create comment (nested)
// @Description  Create a new comment as root or as a child (nested set)
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        comment   body      models.Comment  true  "Comment to create"
// @Param        parent_id query     int             false "Parent comment ID"
// @Success      201   {object}  models.Comment
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /comments/nested [post]
func (h *commentHandler) CreateCommentNested(c *gin.Context) {
	var comment models.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	var parentID *uuid.UUID
	if pid := c.Query("parent_id"); pid != "" {
		if parsed, err := uuid.Parse(pid); err == nil {
			parentID = &parsed
		}
	}
	if err := h.CommentService.CreateNested(&comment, parentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment (nested)"})
		return
	}
	c.JSON(http.StatusCreated, comment)
}

// MoveCommentNested godoc
// @Summary      Move comment (nested)
// @Description  Move a comment to a new parent (nested set)
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id        path      int   true  "Comment ID"
// @Param        parent_id query     int   false "New parent comment ID"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /comments/{id}/move [post]
func (h *commentHandler) MoveCommentNested(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}
	var parentID *uuid.UUID
	if pid := c.Query("parent_id"); pid != "" {
		parsed, err := uuid.Parse(pid)
		if err == nil {
			parentID = &parsed
		}
	}
	if err := h.CommentService.MoveNested(id, parentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move comment (nested)"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Comment moved"})
}

// DeleteCommentNested godoc
// @Summary      Delete comment (nested)
// @Description  Delete a comment and its subtree (nested set)
// @Tags         comments
// @Param        id   path  int  true  "Comment ID"
// @Success      204
// @Failure      500  {object}  map[string]string
// @Router       /comments/{id}/nested [delete]
func (h *commentHandler) DeleteCommentNested(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}
	if err := h.CommentService.DeleteNested(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment (nested)"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetSiblingComments godoc
// @Summary      Get sibling comments
// @Description  Get sibling comments of a comment
// @Tags         comments
// @Produce      json
// @Param        id   path      int  true  "Comment ID"
// @Success      200  {array}   models.Comment
// @Failure      404  {object}  map[string]string
// @Router       /comments/{id}/siblings [get]
func (h *commentHandler) GetSiblingComments(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}
	siblings, err := h.CommentService.GetSiblingComments(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch siblings"})
		return
	}
	c.JSON(http.StatusOK, siblings)
}

// GetParentComment godoc
// @Summary      Get parent comment
// @Description  Get parent of a comment
// @Tags         comments
// @Produce      json
// @Param        id   path      int  true  "Comment ID"
// @Success      200  {object}   models.Comment
// @Failure      404  {object}  map[string]string
// @Router       /comments/{id}/parent [get]
func (h *commentHandler) GetParentComment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}
	parent, err := h.CommentService.GetParentComment(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch parent"})
		return
	}
	if parent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No parent (root)"})
		return
	}
	c.JSON(http.StatusOK, parent)
}

// GetDescendantComments godoc
// @Summary      Get descendant comments
// @Description  Get all descendant comments of a comment
// @Tags         comments
// @Produce      json
// @Param        id   path      int  true  "Comment ID"
// @Success      200  {array}   models.Comment
// @Failure      404  {object}  map[string]string
// @Router       /comments/{id}/descendants [get]
func (h *commentHandler) GetDescendantComments(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}
	descendants, err := h.CommentService.GetDescendantComments(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch descendants"})
		return
	}
	c.JSON(http.StatusOK, descendants)
}

// GetChildrenComments godoc
// @Summary      Get children comments
// @Description  Get direct children of a comment
// @Tags         comments
// @Produce      json
// @Param        id   path      int  true  "Comment ID"
// @Success      200  {array}   models.Comment
// @Failure      404  {object}  map[string]string
// @Router       /comments/{id}/children [get]
func (h *commentHandler) GetChildrenComments(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}
	children, err := h.CommentService.GetChildrenComments(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch children"})
		return
	}
	c.JSON(http.StatusOK, children)
}
