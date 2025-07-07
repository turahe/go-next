package controllers

import (
	"net/http"
	"strconv"
	"wordpress-go-next/backend/internal/http/dto"
	"wordpress-go-next/backend/internal/http/requests"
	"wordpress-go-next/backend/internal/http/responses"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/internal/services"
	"wordpress-go-next/backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

// TagHandler handles tag-related HTTP requests
type TagHandler struct {
	tagService services.TagService
	logger     *logger.ServiceLogger
}

// NewTagHandler creates a new tag handler
func NewTagHandler(tagService services.TagService, logger *logger.ServiceLogger) *TagHandler {
	return &TagHandler{
		tagService: tagService,
		logger:     logger,
	}
}

// CreateTag handles tag creation
// @Summary Create a new tag
// @Description Create a new tag with the provided information
// @Tags tags
// @Accept json
// @Produce json
// @Param tag body requests.CreateTagRequest true "Tag information"
// @Success 201 {object} responses.CommonResponse
// @Failure 400 {object} responses.CommonResponse
// @Failure 500 {object} responses.CommonResponse
// @Router /api/tags [post]
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req requests.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("CreateTag", "Failed to bind request", err)
		c.JSON(http.StatusBadRequest, requests.FormatValidationError(err))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("CreateTag", "Validation failed", err)
		c.JSON(http.StatusBadRequest, responses.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: err.Error()})
		return
	}

	// Convert to model
	tag := req.ToTag()

	// Create tag
	if err := h.tagService.CreateTag(c.Request.Context(), tag); err != nil {
		h.logger.Error("CreateTag", "Failed to create tag", err, map[string]interface{}{
			"tag_name": tag.Name,
		})
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{ResponseCode: http.StatusInternalServerError, ResponseMessage: "Failed to create tag"})
		return
	}

	h.logger.Info("CreateTag", "Tag created successfully", map[string]interface{}{
		"tag_id":   tag.ID,
		"tag_name": tag.Name,
	})

	c.JSON(http.StatusCreated, responses.CommonResponse{
		ResponseCode:    http.StatusCreated,
		ResponseMessage: "Tag created successfully",
		Data:            dto.ToTagDTO(tag),
	})
}

// GetTagByID handles getting a tag by ID
// @Summary Get tag by ID
// @Description Get a tag by its ID
// @Tags tags
// @Produce json
// @Param id path int true "Tag ID"
// @Success 200 {object} responses.CommonResponse
// @Failure 404 {object} responses.CommonResponse
// @Failure 500 {object} responses.CommonResponse
// @Router /api/tags/{id} [get]
func (h *TagHandler) GetTagByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("GetTagByID", "Invalid tag ID", err, map[string]interface{}{
			"tag_id": idStr,
		})
		c.JSON(http.StatusBadRequest, responses.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: "Invalid tag ID"})
		return
	}

	tag, err := h.tagService.GetTagByID(c.Request.Context(), uint64(id))
	if err != nil {
		h.logger.Error("GetTagByID", "Failed to get tag", err, map[string]interface{}{
			"tag_id": id,
		})
		c.JSON(http.StatusNotFound, responses.CommonResponse{ResponseCode: http.StatusNotFound, ResponseMessage: "Tag not found"})
		return
	}

	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Tag fetched successfully",
		Data:            dto.ToTagDTO(tag),
	})
}

// GetTagBySlug handles getting a tag by slug
// @Summary Get tag by slug
// @Description Get a tag by its slug
// @Tags tags
// @Produce json
// @Param slug path string true "Tag slug"
// @Success 200 {object} responses.CommonResponse
// @Failure 404 {object} responses.CommonResponse
// @Failure 500 {object} responses.CommonResponse
// @Router /api/tags/slug/{slug} [get]
func (h *TagHandler) GetTagBySlug(c *gin.Context) {
	slug := c.Param("slug")

	tag, err := h.tagService.GetTagBySlug(c.Request.Context(), slug)
	if err != nil {
		h.logger.Error("GetTagBySlug", "Failed to get tag", err, map[string]interface{}{
			"slug": slug,
		})
		c.JSON(http.StatusNotFound, responses.CommonResponse{ResponseCode: http.StatusNotFound, ResponseMessage: "Tag not found"})
		return
	}

	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Tag fetched successfully",
		Data:            dto.ToTagDTO(tag),
	})
}

// UpdateTag handles tag updates
// @Summary Update a tag
// @Description Update an existing tag
// @Tags tags
// @Accept json
// @Produce json
// @Param id path int true "Tag ID"
// @Param tag body requests.UpdateTagRequest true "Updated tag information"
// @Success 200 {object} responses.CommonResponse
// @Failure 400 {object} responses.CommonResponse
// @Failure 404 {object} responses.CommonResponse
// @Failure 500 {object} responses.CommonResponse
// @Router /api/tags/{id} [put]
func (h *TagHandler) UpdateTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("UpdateTag", "Invalid tag ID", err, map[string]interface{}{
			"tag_id": idStr,
		})
		c.JSON(http.StatusBadRequest, responses.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: "Invalid tag ID"})
		return
	}

	var req requests.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("UpdateTag", "Failed to bind request", err)
		c.JSON(http.StatusBadRequest, requests.FormatValidationError(err))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("UpdateTag", "Validation failed", err)
		c.JSON(http.StatusBadRequest, responses.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: err.Error()})
		return
	}

	// Get existing tag
	_, err = h.tagService.GetTagByID(c.Request.Context(), uint64(id))
	if err != nil {
		h.logger.Error("UpdateTag", "Tag not found", err, map[string]interface{}{
			"tag_id": id,
		})
		c.JSON(http.StatusNotFound, responses.CommonResponse{ResponseCode: http.StatusNotFound, ResponseMessage: "Tag not found"})
		return
	}

	// Update tag fields
	updatedTag := req.ToTag()
	updatedTag.ID = uint64(id)

	// Update tag
	if err := h.tagService.UpdateTag(c.Request.Context(), updatedTag); err != nil {
		h.logger.Error("UpdateTag", "Failed to update tag", err, map[string]interface{}{
			"tag_id": id,
		})
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{ResponseCode: http.StatusInternalServerError, ResponseMessage: "Failed to update tag"})
		return
	}

	h.logger.Info("UpdateTag", "Tag updated successfully", map[string]interface{}{
		"tag_id": id,
	})

	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Tag updated successfully",
		Data:            dto.ToTagDTO(updatedTag),
	})
}

// DeleteTag handles tag deletion
// @Summary Delete a tag
// @Description Delete a tag by its ID
// @Tags tags
// @Produce json
// @Param id path int true "Tag ID"
// @Success 200 {object} responses.CommonResponse
// @Failure 400 {object} responses.CommonResponse
// @Failure 404 {object} responses.CommonResponse
// @Failure 500 {object} responses.CommonResponse
// @Router /api/tags/{id} [delete]
func (h *TagHandler) DeleteTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("DeleteTag", "Invalid tag ID", err, map[string]interface{}{
			"tag_id": idStr,
		})
		c.JSON(http.StatusBadRequest, responses.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: "Invalid tag ID"})
		return
	}

	if err := h.tagService.DeleteTag(c.Request.Context(), uint64(id)); err != nil {
		h.logger.Error("DeleteTag", "Failed to delete tag", err, map[string]interface{}{
			"tag_id": id,
		})
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{ResponseCode: http.StatusInternalServerError, ResponseMessage: "Failed to delete tag"})
		return
	}

	h.logger.Info("DeleteTag", "Tag deleted successfully", map[string]interface{}{
		"tag_id": id,
	})

	c.JSON(http.StatusOK, responses.CommonResponse{ResponseCode: http.StatusOK, ResponseMessage: "Tag deleted successfully"})
}

// ListTags handles listing tags with filters
// @Summary List tags
// @Description Get a list of tags with optional filtering
// @Tags tags
// @Produce json
// @Param type query string false "Tag type filter"
// @Param active query bool false "Active status filter"
// @Param limit query int false "Limit (default: 10, max: 100)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {object} responses.CommonResponse
// @Failure 500 {object} responses.CommonResponse
// @Router /api/tags [get]
// ListTags godoc
// @Summary      List tags
// @Description  Get tags with pagination and optional search
// @Tags         tags
// @Produce      json
// @Param        page     query     int     false  "Page number"  default(1)
// @Param        perPage  query     int     false  "Items per page"  default(10)
// @Param        type     query     string  false "Tag type filter"
// @Param        search   query     string  false  "Search keyword"
// @Success      200  {object}  responses.PaginationResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/tags [get]
func (h *TagHandler) ListTags(c *gin.Context) {
	pagination, err := requests.ParsePaginationFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: "Invalid pagination parameters"})
		return
	}

	tagType := c.Query("type")
	result, err := h.tagService.GetTagsWithPagination(c.Request.Context(), pagination.Page, pagination.PerPage, pagination.Search, tagType)
	if err != nil {
		h.logger.Error("ListTags", "Failed to get tags", err)
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{ResponseCode: http.StatusInternalServerError, ResponseMessage: "Failed to get tags"})
		return
	}

	tags, ok := result.Data.([]models.Tag)
	if !ok {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{ResponseCode: http.StatusInternalServerError, ResponseMessage: "Invalid data format"})
		return
	}
	result.Data = dto.ToTagDTOs(tags)

	c.JSON(http.StatusOK, result)
}

// SearchTags handles tag search
// @Summary Search tags
// @Description Search tags by name or description with pagination
// @Tags tags
// @Produce json
// @Param query query string true "Search query"
// @Param type query string false "Tag type filter"
// @Param page query int false "Page number" default(1)
// @Param perPage query int false "Items per page" default(10)
// @Success 200 {object} responses.PaginationResponse
// @Failure 400 {object} responses.CommonResponse
// @Failure 500 {object} responses.CommonResponse
// @Router /api/tags/search [get]
func (h *TagHandler) SearchTags(c *gin.Context) {
	pagination, err := requests.ParsePaginationFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: "Invalid pagination parameters"})
		return
	}

	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: "Search query is required"})
		return
	}

	tags, total, err := h.tagService.SearchTags(c.Request.Context(), query, pagination.PerPage, pagination.GetOffset())
	if err != nil {
		h.logger.Error("SearchTags", "Failed to search tags", err, map[string]interface{}{
			"query": query,
		})
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{ResponseCode: http.StatusInternalServerError, ResponseMessage: "Failed to search tags"})
		return
	}

	// Create pagination response
	totalPages := (int(total) + pagination.PerPage - 1) / pagination.PerPage
	result := &responses.PaginationResponse{
		Data:         dto.ToTagDTOs(tags),
		TotalCount:   total,
		TotalPage:    int64(totalPages),
		CurrentPage:  int64(pagination.Page),
		LastPage:     int64(totalPages),
		PerPage:      int64(pagination.PerPage),
		NextPage:     int64(pagination.Page + 1),
		PreviousPage: int64(pagination.Page - 1),
	}

	// Adjust next/previous page
	if pagination.Page >= totalPages {
		result.NextPage = 0
	}
	if pagination.Page <= 1 {
		result.PreviousPage = 0
	}

	c.JSON(http.StatusOK, result)
}

// AddTagToEntity handles adding a tag to an entity
// @Summary Add tag to entity
// @Description Add a tag to a specific entity
// @Tags tags
// @Accept json
// @Produce json
// @Param request body requests.AddTagToEntityRequest true "Tag entity request"
// @Success 200 {object} responses.CommonResponse
// @Failure 400 {object} responses.CommonResponse
// @Failure 500 {object} responses.CommonResponse
// @Router /api/tags/entity [post]
func (h *TagHandler) AddTagToEntity(c *gin.Context) {
	var req requests.AddTagToEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("AddTagToEntity", "Failed to bind request", err)
		c.JSON(http.StatusBadRequest, requests.FormatValidationError(err))
		return
	}

	if err := h.tagService.AddTagToEntity(c.Request.Context(), req.TagID, req.EntityID, req.EntityType); err != nil {
		h.logger.Error("AddTagToEntity", "Failed to add tag to entity", err, map[string]interface{}{
			"tag_id":      req.TagID,
			"entity_id":   req.EntityID,
			"entity_type": req.EntityType,
		})
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{ResponseCode: http.StatusInternalServerError, ResponseMessage: "Failed to add tag to entity"})
		return
	}

	h.logger.Info("AddTagToEntity", "Tag added to entity successfully", map[string]interface{}{
		"tag_id":      req.TagID,
		"entity_id":   req.EntityID,
		"entity_type": req.EntityType,
	})

	c.JSON(http.StatusOK, responses.CommonResponse{ResponseCode: http.StatusOK, ResponseMessage: "Tag added to entity successfully"})
}

// RemoveTagFromEntity handles removing a tag from an entity
// @Summary Remove tag from entity
// @Description Remove a tag from a specific entity
// @Tags tags
// @Accept json
// @Produce json
// @Param request body requests.RemoveTagFromEntityRequest true "Remove tag entity request"
// @Success 200 {object} responses.CommonResponse
// @Failure 400 {object} responses.CommonResponse
// @Failure 500 {object} responses.CommonResponse
// @Router /api/tags/entity [delete]
func (h *TagHandler) RemoveTagFromEntity(c *gin.Context) {
	var req requests.RemoveTagFromEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("RemoveTagFromEntity", "Failed to bind request", err)
		c.JSON(http.StatusBadRequest, requests.FormatValidationError(err))
		return
	}

	if err := h.tagService.RemoveTagFromEntity(c.Request.Context(), req.TagID, req.EntityID, req.EntityType); err != nil {
		h.logger.Error("RemoveTagFromEntity", "Failed to remove tag from entity", err, map[string]interface{}{
			"tag_id":      req.TagID,
			"entity_id":   req.EntityID,
			"entity_type": req.EntityType,
		})
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{ResponseCode: http.StatusInternalServerError, ResponseMessage: "Failed to remove tag from entity"})
		return
	}

	h.logger.Info("RemoveTagFromEntity", "Tag removed from entity successfully", map[string]interface{}{
		"tag_id":      req.TagID,
		"entity_id":   req.EntityID,
		"entity_type": req.EntityType,
	})

	c.JSON(http.StatusOK, responses.CommonResponse{ResponseCode: http.StatusOK, ResponseMessage: "Tag removed from entity successfully"})
}

// GetTagsByEntity handles getting tags for a specific entity
// @Summary Get entity tags
// @Description Get all tags for a specific entity
// @Tags tags
// @Produce json
// @Param entity_id query int true "Entity ID"
// @Param entity_type query string true "Entity type"
// @Success 200 {object} responses.CommonResponse
// @Failure 400 {object} responses.CommonResponse
// @Failure 500 {object} responses.CommonResponse
// @Router /api/tags/entity [get]
func (h *TagHandler) GetTagsByEntity(c *gin.Context) {
	entityIDStr := c.Query("entity_id")
	entityType := c.Query("entity_type")

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		h.logger.Error("GetTagsByEntity", "Invalid entity ID", err, map[string]interface{}{
			"entity_id": entityIDStr,
		})
		c.JSON(http.StatusBadRequest, responses.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: "Invalid entity ID"})
		return
	}

	if entityType == "" {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: "Entity type is required"})
		return
	}

	tags, err := h.tagService.GetTagsByEntity(c.Request.Context(), uint64(entityID), entityType)
	if err != nil {
		h.logger.Error("GetTagsByEntity", "Failed to get entity tags", err, map[string]interface{}{
			"entity_id":   entityID,
			"entity_type": entityType,
		})
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{ResponseCode: http.StatusInternalServerError, ResponseMessage: "Failed to get entity tags"})
		return
	}

	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Tags fetched successfully",
		Data:            dto.ToTagDTOs(tags),
	})
}

// GetEntitiesByTag handles getting entities by tag
// @Summary Get entities by tag
// @Description Get all entities of a specific type that have a particular tag
// @Tags tags
// @Produce json
// @Param tag_id query int true "Tag ID"
// @Param entity_type query string true "Entity type"
// @Param limit query int false "Limit (default: 10, max: 100)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {object} responses.CommonResponse
// @Failure 400 {object} responses.CommonResponse
// @Failure 500 {object} responses.CommonResponse
// @Router /api/tags/entities [get]
func (h *TagHandler) GetEntitiesByTag(c *gin.Context) {
	var req requests.GetEntitiesByTagRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("GetEntitiesByTag", "Failed to bind query", err)
		c.JSON(http.StatusBadRequest, requests.FormatValidationError(err))
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}

	entities, _, err := h.tagService.GetEntitiesByTag(c.Request.Context(), uint64(req.TagID), req.EntityType, req.Limit, req.Offset)
	if err != nil {
		h.logger.Error("GetEntitiesByTag", "Failed to get entities by tag", err, map[string]interface{}{
			"tag_id":      req.TagID,
			"entity_type": req.EntityType,
		})
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{ResponseCode: http.StatusInternalServerError, ResponseMessage: "Failed to get entities by tag"})
		return
	}

	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Tags fetched successfully",
		Data:            entities,
	})
}

// GetTagCount handles getting tag count
// @Summary Get tag count
// @Description Get the total number of tags
// @Tags tags
// @Produce json
// @Success 200 {object} responses.CommonResponse
// @Failure 500 {object} responses.CommonResponse
// @Router /api/tags/count [get]
func (h *TagHandler) GetTagCount(c *gin.Context) {
	count, err := h.tagService.GetTagCount(c.Request.Context())
	if err != nil {
		h.logger.Error("GetTagCount", "Failed to get tag count", err)
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{ResponseCode: http.StatusInternalServerError, ResponseMessage: "Failed to get tag count"})
		return
	}

	c.JSON(http.StatusOK, responses.CommonResponse{ResponseCode: http.StatusOK, Data: count})
}
