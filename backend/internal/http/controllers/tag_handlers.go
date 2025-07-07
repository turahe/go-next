package controllers

import (
	"net/http"
	"strconv"
	"wordpress-go-next/backend/internal/http/requests"
	"wordpress-go-next/backend/internal/http/responses"
	"wordpress-go-next/backend/internal/services"

	"github.com/gin-gonic/gin"
)

// TagHandler handles tag-related HTTP requests
type TagHandler struct {
	tagService services.TagService
	logger     *services.ServiceLogger
}

// NewTagHandler creates a new tag handler
func NewTagHandler(tagService services.TagService, logger *services.ServiceLogger) *TagHandler {
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
// @Success 201 {object} responses.Response{data=models.Tag}
// @Failure 400 {object} responses.Response
// @Failure 500 {object} responses.Response
// @Router /api/tags [post]
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req requests.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(c.Request.Context(), "CreateTag", "Failed to bind request", err)
		c.JSON(http.StatusBadRequest, requests.FormatValidationError(err))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error(c.Request.Context(), "CreateTag", "Validation failed", err)
		c.JSON(http.StatusBadRequest, responses.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Convert to model
	tag := req.ToTag()

	// Create tag
	if err := h.tagService.CreateTag(c.Request.Context(), tag); err != nil {
		h.logger.Error(c.Request.Context(), "CreateTag", "Failed to create tag", err, map[string]interface{}{
			"tag_name": tag.Name,
		})
		c.JSON(http.StatusInternalServerError, responses.Response{
			Success: false,
			Message: "Failed to create tag",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "CreateTag", "Tag created successfully", map[string]interface{}{
		"tag_id":   tag.ID,
		"tag_name": tag.Name,
	})

	c.JSON(http.StatusCreated, responses.Response{
		Success: true,
		Message: "Tag created successfully",
		Data:    tag,
	})
}

// GetTagByID handles getting a tag by ID
// @Summary Get tag by ID
// @Description Get a tag by its ID
// @Tags tags
// @Produce json
// @Param id path int true "Tag ID"
// @Success 200 {object} responses.Response{data=models.Tag}
// @Failure 404 {object} responses.Response
// @Failure 500 {object} responses.Response
// @Router /api/tags/{id} [get]
func (h *TagHandler) GetTagByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error(c.Request.Context(), "GetTagByID", "Invalid tag ID", err, map[string]interface{}{
			"tag_id": idStr,
		})
		c.JSON(http.StatusBadRequest, responses.Response{
			Success: false,
			Message: "Invalid tag ID",
		})
		return
	}

	tag, err := h.tagService.GetTagByID(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error(c.Request.Context(), "GetTagByID", "Failed to get tag", err, map[string]interface{}{
			"tag_id": id,
		})
		c.JSON(http.StatusNotFound, responses.Response{
			Success: false,
			Message: "Tag not found",
		})
		return
	}

	c.JSON(http.StatusOK, responses.Response{
		Success: true,
		Data:    tag,
	})
}

// GetTagBySlug handles getting a tag by slug
// @Summary Get tag by slug
// @Description Get a tag by its slug
// @Tags tags
// @Produce json
// @Param slug path string true "Tag slug"
// @Success 200 {object} responses.Response{data=models.Tag}
// @Failure 404 {object} responses.Response
// @Failure 500 {object} responses.Response
// @Router /api/tags/slug/{slug} [get]
func (h *TagHandler) GetTagBySlug(c *gin.Context) {
	slug := c.Param("slug")

	tag, err := h.tagService.GetTagBySlug(c.Request.Context(), slug)
	if err != nil {
		h.logger.Error(c.Request.Context(), "GetTagBySlug", "Failed to get tag", err, map[string]interface{}{
			"slug": slug,
		})
		c.JSON(http.StatusNotFound, responses.Response{
			Success: false,
			Message: "Tag not found",
		})
		return
	}

	c.JSON(http.StatusOK, responses.Response{
		Success: true,
		Data:    tag,
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
// @Success 200 {object} responses.Response{data=models.Tag}
// @Failure 400 {object} responses.Response
// @Failure 404 {object} responses.Response
// @Failure 500 {object} responses.Response
// @Router /api/tags/{id} [put]
func (h *TagHandler) UpdateTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error(c.Request.Context(), "UpdateTag", "Invalid tag ID", err, map[string]interface{}{
			"tag_id": idStr,
		})
		c.JSON(http.StatusBadRequest, responses.Response{
			Success: false,
			Message: "Invalid tag ID",
		})
		return
	}

	var req requests.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(c.Request.Context(), "UpdateTag", "Failed to bind request", err)
		c.JSON(http.StatusBadRequest, requests.FormatValidationError(err))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error(c.Request.Context(), "UpdateTag", "Validation failed", err)
		c.JSON(http.StatusBadRequest, responses.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Get existing tag
	existingTag, err := h.tagService.GetTagByID(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error(c.Request.Context(), "UpdateTag", "Tag not found", err, map[string]interface{}{
			"tag_id": id,
		})
		c.JSON(http.StatusNotFound, responses.Response{
			Success: false,
			Message: "Tag not found",
		})
		return
	}

	// Update tag fields
	updatedTag := req.ToTag()
	updatedTag.ID = uint(id)

	// Update tag
	if err := h.tagService.UpdateTag(c.Request.Context(), updatedTag); err != nil {
		h.logger.Error(c.Request.Context(), "UpdateTag", "Failed to update tag", err, map[string]interface{}{
			"tag_id": id,
		})
		c.JSON(http.StatusInternalServerError, responses.Response{
			Success: false,
			Message: "Failed to update tag",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "UpdateTag", "Tag updated successfully", map[string]interface{}{
		"tag_id": id,
	})

	c.JSON(http.StatusOK, responses.Response{
		Success: true,
		Message: "Tag updated successfully",
		Data:    updatedTag,
	})
}

// DeleteTag handles tag deletion
// @Summary Delete a tag
// @Description Delete a tag by its ID
// @Tags tags
// @Produce json
// @Param id path int true "Tag ID"
// @Success 200 {object} responses.Response
// @Failure 400 {object} responses.Response
// @Failure 404 {object} responses.Response
// @Failure 500 {object} responses.Response
// @Router /api/tags/{id} [delete]
func (h *TagHandler) DeleteTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error(c.Request.Context(), "DeleteTag", "Invalid tag ID", err, map[string]interface{}{
			"tag_id": idStr,
		})
		c.JSON(http.StatusBadRequest, responses.Response{
			Success: false,
			Message: "Invalid tag ID",
		})
		return
	}

	if err := h.tagService.DeleteTag(c.Request.Context(), uint(id)); err != nil {
		h.logger.Error(c.Request.Context(), "DeleteTag", "Failed to delete tag", err, map[string]interface{}{
			"tag_id": id,
		})
		c.JSON(http.StatusInternalServerError, responses.Response{
			Success: false,
			Message: "Failed to delete tag",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "DeleteTag", "Tag deleted successfully", map[string]interface{}{
		"tag_id": id,
	})

	c.JSON(http.StatusOK, responses.Response{
		Success: true,
		Message: "Tag deleted successfully",
	})
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
// @Success 200 {object} responses.Response{data=[]models.Tag}
// @Failure 500 {object} responses.Response
// @Router /api/tags [get]
func (h *TagHandler) ListTags(c *gin.Context) {
	var req requests.TagListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error(c.Request.Context(), "ListTags", "Failed to bind query", err)
		c.JSON(http.StatusBadRequest, requests.FormatValidationError(err))
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}

	var tags []services.Tag
	var err error

	if req.Active != nil && *req.Active {
		tags, err = h.tagService.GetActiveTags(c.Request.Context())
	} else {
		tags, err = h.tagService.GetAllTags(c.Request.Context(), req.Type)
	}

	if err != nil {
		h.logger.Error(c.Request.Context(), "ListTags", "Failed to get tags", err)
		c.JSON(http.StatusInternalServerError, responses.Response{
			Success: false,
			Message: "Failed to get tags",
		})
		return
	}

	// Apply pagination
	start := req.Offset
	end := start + req.Limit
	if end > len(tags) {
		end = len(tags)
	}
	if start > len(tags) {
		start = len(tags)
	}

	paginatedTags := tags[start:end]

	c.JSON(http.StatusOK, responses.Response{
		Success: true,
		Data:    paginatedTags,
		Meta: map[string]interface{}{
			"total":  len(tags),
			"limit":  req.Limit,
			"offset": req.Offset,
		},
	})
}

// SearchTags handles tag search
// @Summary Search tags
// @Description Search tags by name or description
// @Tags tags
// @Produce json
// @Param query query string true "Search query"
// @Param type query string false "Tag type filter"
// @Param limit query int false "Limit (default: 10, max: 100)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {object} responses.Response{data=[]models.Tag}
// @Failure 400 {object} responses.Response
// @Failure 500 {object} responses.Response
// @Router /api/tags/search [get]
func (h *TagHandler) SearchTags(c *gin.Context) {
	var req requests.TagSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error(c.Request.Context(), "SearchTags", "Failed to bind query", err)
		c.JSON(http.StatusBadRequest, requests.FormatValidationError(err))
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}

	tags, total, err := h.tagService.SearchTags(c.Request.Context(), req.Query, req.Limit, req.Offset)
	if err != nil {
		h.logger.Error(c.Request.Context(), "SearchTags", "Failed to search tags", err, map[string]interface{}{
			"query": req.Query,
		})
		c.JSON(http.StatusInternalServerError, responses.Response{
			Success: false,
			Message: "Failed to search tags",
		})
		return
	}

	c.JSON(http.StatusOK, responses.Response{
		Success: true,
		Data:    tags,
		Meta: map[string]interface{}{
			"total":  total,
			"limit":  req.Limit,
			"offset": req.Offset,
			"query":  req.Query,
		},
	})
}

// AddTagToEntity handles adding a tag to an entity
// @Summary Add tag to entity
// @Description Add a tag to a specific entity
// @Tags tags
// @Accept json
// @Produce json
// @Param request body requests.AddTagToEntityRequest true "Tag entity request"
// @Success 200 {object} responses.Response
// @Failure 400 {object} responses.Response
// @Failure 500 {object} responses.Response
// @Router /api/tags/entity [post]
func (h *TagHandler) AddTagToEntity(c *gin.Context) {
	var req requests.AddTagToEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(c.Request.Context(), "AddTagToEntity", "Failed to bind request", err)
		c.JSON(http.StatusBadRequest, requests.FormatValidationError(err))
		return
	}

	if err := h.tagService.AddTagToEntity(c.Request.Context(), req.TagID, req.EntityID, req.EntityType); err != nil {
		h.logger.Error(c.Request.Context(), "AddTagToEntity", "Failed to add tag to entity", err, map[string]interface{}{
			"tag_id":      req.TagID,
			"entity_id":   req.EntityID,
			"entity_type": req.EntityType,
		})
		c.JSON(http.StatusInternalServerError, responses.Response{
			Success: false,
			Message: "Failed to add tag to entity",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "AddTagToEntity", "Tag added to entity successfully", map[string]interface{}{
		"tag_id":      req.TagID,
		"entity_id":   req.EntityID,
		"entity_type": req.EntityType,
	})

	c.JSON(http.StatusOK, responses.Response{
		Success: true,
		Message: "Tag added to entity successfully",
	})
}

// RemoveTagFromEntity handles removing a tag from an entity
// @Summary Remove tag from entity
// @Description Remove a tag from a specific entity
// @Tags tags
// @Accept json
// @Produce json
// @Param request body requests.RemoveTagFromEntityRequest true "Remove tag entity request"
// @Success 200 {object} responses.Response
// @Failure 400 {object} responses.Response
// @Failure 500 {object} responses.Response
// @Router /api/tags/entity [delete]
func (h *TagHandler) RemoveTagFromEntity(c *gin.Context) {
	var req requests.RemoveTagFromEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(c.Request.Context(), "RemoveTagFromEntity", "Failed to bind request", err)
		c.JSON(http.StatusBadRequest, requests.FormatValidationError(err))
		return
	}

	if err := h.tagService.RemoveTagFromEntity(c.Request.Context(), req.TagID, req.EntityID, req.EntityType); err != nil {
		h.logger.Error(c.Request.Context(), "RemoveTagFromEntity", "Failed to remove tag from entity", err, map[string]interface{}{
			"tag_id":      req.TagID,
			"entity_id":   req.EntityID,
			"entity_type": req.EntityType,
		})
		c.JSON(http.StatusInternalServerError, responses.Response{
			Success: false,
			Message: "Failed to remove tag from entity",
		})
		return
	}

	h.logger.Info(c.Request.Context(), "RemoveTagFromEntity", "Tag removed from entity successfully", map[string]interface{}{
		"tag_id":      req.TagID,
		"entity_id":   req.EntityID,
		"entity_type": req.EntityType,
	})

	c.JSON(http.StatusOK, responses.Response{
		Success: true,
		Message: "Tag removed from entity successfully",
	})
}

// GetTagsByEntity handles getting tags for a specific entity
// @Summary Get entity tags
// @Description Get all tags for a specific entity
// @Tags tags
// @Produce json
// @Param entity_id query int true "Entity ID"
// @Param entity_type query string true "Entity type"
// @Success 200 {object} responses.Response{data=[]models.Tag}
// @Failure 400 {object} responses.Response
// @Failure 500 {object} responses.Response
// @Router /api/tags/entity [get]
func (h *TagHandler) GetTagsByEntity(c *gin.Context) {
	entityIDStr := c.Query("entity_id")
	entityType := c.Query("entity_type")

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		h.logger.Error(c.Request.Context(), "GetTagsByEntity", "Invalid entity ID", err, map[string]interface{}{
			"entity_id": entityIDStr,
		})
		c.JSON(http.StatusBadRequest, responses.Response{
			Success: false,
			Message: "Invalid entity ID",
		})
		return
	}

	if entityType == "" {
		c.JSON(http.StatusBadRequest, responses.Response{
			Success: false,
			Message: "Entity type is required",
		})
		return
	}

	tags, err := h.tagService.GetTagsByEntity(c.Request.Context(), uint(entityID), entityType)
	if err != nil {
		h.logger.Error(c.Request.Context(), "GetTagsByEntity", "Failed to get entity tags", err, map[string]interface{}{
			"entity_id":   entityID,
			"entity_type": entityType,
		})
		c.JSON(http.StatusInternalServerError, responses.Response{
			Success: false,
			Message: "Failed to get entity tags",
		})
		return
	}

	c.JSON(http.StatusOK, responses.Response{
		Success: true,
		Data:    tags,
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
// @Success 200 {object} responses.Response{data=[]map[string]interface{}}
// @Failure 400 {object} responses.Response
// @Failure 500 {object} responses.Response
// @Router /api/tags/entities [get]
func (h *TagHandler) GetEntitiesByTag(c *gin.Context) {
	var req requests.GetEntitiesByTagRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error(c.Request.Context(), "GetEntitiesByTag", "Failed to bind query", err)
		c.JSON(http.StatusBadRequest, requests.FormatValidationError(err))
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}

	entities, total, err := h.tagService.GetEntitiesByTag(c.Request.Context(), req.TagID, req.EntityType, req.Limit, req.Offset)
	if err != nil {
		h.logger.Error(c.Request.Context(), "GetEntitiesByTag", "Failed to get entities by tag", err, map[string]interface{}{
			"tag_id":      req.TagID,
			"entity_type": req.EntityType,
		})
		c.JSON(http.StatusInternalServerError, responses.Response{
			Success: false,
			Message: "Failed to get entities by tag",
		})
		return
	}

	c.JSON(http.StatusOK, responses.Response{
		Success: true,
		Data:    entities,
		Meta: map[string]interface{}{
			"total":       total,
			"limit":       req.Limit,
			"offset":      req.Offset,
			"tag_id":      req.TagID,
			"entity_type": req.EntityType,
		},
	})
}

// GetTagCount handles getting tag count
// @Summary Get tag count
// @Description Get the total number of tags
// @Tags tags
// @Produce json
// @Success 200 {object} responses.Response{data=int64}
// @Failure 500 {object} responses.Response
// @Router /api/tags/count [get]
func (h *TagHandler) GetTagCount(c *gin.Context) {
	count, err := h.tagService.GetTagCount(c.Request.Context())
	if err != nil {
		h.logger.Error(c.Request.Context(), "GetTagCount", "Failed to get tag count", err)
		c.JSON(http.StatusInternalServerError, responses.Response{
			Success: false,
			Message: "Failed to get tag count",
		})
		return
	}

	c.JSON(http.StatusOK, responses.Response{
		Success: true,
		Data:    count,
	})
}
