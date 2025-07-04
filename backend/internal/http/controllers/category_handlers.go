package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"wordpress-go-next/backend/internal/http/requests"

	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type CategoryHandler interface {
	GetCategories(c *gin.Context)
	GetCategory(c *gin.Context)
	GetChildrenCategories(c *gin.Context)
	CreateCategory(c *gin.Context)
	UpdateCategory(c *gin.Context)
	DeleteCategory(c *gin.Context)
	CreateCategoryNested(c *gin.Context)
	MoveCategoryNested(c *gin.Context)
	DeleteCategoryNested(c *gin.Context)
}

type categoryHandler struct {
	CategoryService services.CategoryService
	MediaService    services.MediaService
}

func NewCategoryHandler(categoryService services.CategoryService, mediaService services.MediaService) CategoryHandler {
	return &categoryHandler{CategoryService: categoryService, MediaService: mediaService}
}

// GetCategories godoc
// @Summary      List categories
// @Description  Get all categories
// @Tags         categories
// @Produce      json
// @Success      200  {array}   models.Category
// @Failure      500  {object}  map[string]string
// @Router       /categories [get]
func (h *categoryHandler) GetCategories(c *gin.Context) {
	categories, err := h.CategoryService.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// GetCategory godoc
// @Summary      Get category
// @Description  Get a category by ID
// @Tags         categories
// @Produce      json
// @Param        id   path      int  true  "Category ID"
// @Success      200  {object}  models.Category
// @Failure      404  {object}  map[string]string
// @Router       /categories/{id} [get]
func (h *categoryHandler) GetCategory(c *gin.Context) {
	id := c.Param("id")
	category, err := h.CategoryService.GetCategoryByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}
	c.JSON(http.StatusOK, category)
}

// CreateCategory godoc
// @Summary      Create category
// @Description  Create a new category
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        category  body      models.Category  true  "Category to create"
// @Success      201   {object}  models.Category
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /categories [post]
func (h *categoryHandler) CreateCategory(c *gin.Context) {
	var reqParams requests.CategoryCreateRequest
	if err := c.ShouldBind(&reqParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	category := models.Category{
		Name:        reqParams.Name,
		Description: reqParams.Description,
	}
	if reqParams.ParentID > 0 {
		category.ParentID = &reqParams.ParentID
	}
	file, fileHeader, err := c.Request.FormFile("image")
	if err == nil && file != nil && fileHeader != nil {
		media, err := h.MediaService.UploadAndSaveMedia(file, fileHeader, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
			return
		}
		if err := h.CategoryService.CreateCategory(&category); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
			return
		}
		if err := h.MediaService.AssociateMedia(media.ID, category.ID, "categories", "image"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate image with category"})
			return
		}
		c.JSON(http.StatusCreated, category)
		return
	}
	if err := h.CategoryService.CreateCategory(&category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}
	c.JSON(http.StatusCreated, category)
}

// UpdateCategory godoc
// @Summary      Update category
// @Description  Update an existing category
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id        path      int             true  "Category ID"
// @Param        category  body      models.Category true  "Category to update"
// @Success      200   {object}  models.Category
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /categories/{id} [put]
func (h *categoryHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var requestBody requests.CategoryUpdateRequest
	category, err := h.CategoryService.GetCategoryByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}
	if err := c.ShouldBindJSON(requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := h.CategoryService.UpdateCategory(category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}
	c.JSON(http.StatusOK, category)
}

// DeleteCategory godoc
// @Summary      Delete category
// @Description  Delete a category by ID
// @Tags         categories
// @Param        id   path  int  true  "Category ID"
// @Success      204
// @Failure      500  {object}  map[string]string
// @Router       /categories/{id} [delete]
func (h *categoryHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if err := h.CategoryService.DeleteCategory(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}
	c.Status(http.StatusNoContent)
}

// CreateCategoryNested godoc
// @Summary      Create category (nested)
// @Description  Create a new category as root or as a child (nested set)
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        category  body      models.Category  true  "Category to create"
// @Param        parent_id query     int              false "Parent category ID"
// @Success      201   {object}  models.Category
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /categories/nested [post]
func (h *categoryHandler) CreateCategoryNested(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	var parentID *int64
	if pid := c.Query("parent_id"); pid != "" {
		var parsed int64
		if _, err := fmt.Sscan(pid, &parsed); err == nil {
			parentID = &parsed
		}
	}
	if err := h.CategoryService.CreateNested(&category, parentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category (nested)"})
		return
	}
	c.JSON(http.StatusCreated, category)
}

// MoveCategoryNested godoc
// @Summary      Move category (nested)
// @Description  Move a category to a new parent (nested set)
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id        path      int   true  "Category ID"
// @Param        parent_id query     int   false "New parent category ID"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /categories/{id}/move [post]
func (h *categoryHandler) MoveCategoryNested(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}
	var parentID *int64
	if pid := c.Query("parent_id"); pid != "" {
		var parsed int64
		if _, err := fmt.Sscan(pid, &parsed); err == nil {
			parentID = &parsed
		}
	}
	if err := h.CategoryService.MoveNested(uint(id), parentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move category (nested)"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Category moved"})
}

// DeleteCategoryNested godoc
// @Summary      Delete category (nested)
// @Description  Delete a category and its subtree (nested set)
// @Tags         categories
// @Param        id   path  int  true  "Category ID"
// @Success      204
// @Failure      500  {object}  map[string]string
// @Router       /categories/{id}/nested [delete]
func (h *categoryHandler) DeleteCategoryNested(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}
	if err := h.CategoryService.DeleteNested(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category (nested)"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetChildrenCategories godoc
// @Summary      Get children categories
// @Description  Get direct children of a category
// @Tags         categories
// @Produce      json
// @Param        id   path      int  true  "Category ID"
// @Success      200  {array}   models.Category
// @Failure      404  {object}  map[string]string
// @Router       /categories/{id}/children [get]
func (h *categoryHandler) GetChildrenCategories(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}
	children, err := h.CategoryService.GetChildrenCategories(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch children"})
		return
	}
	c.JSON(http.StatusOK, children)
}
