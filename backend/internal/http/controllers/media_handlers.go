package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"wordpress-go-next/backend/internal/http/dto"
	"wordpress-go-next/backend/internal/http/requests"
	"wordpress-go-next/backend/internal/http/responses"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type MediaHandler interface {
	UploadMedia(c *gin.Context)
	AssociateMedia(c *gin.Context)
	CreateMediaNested(c *gin.Context)
	MoveMediaNested(c *gin.Context)
	DeleteMediaNested(c *gin.Context)
	GetSiblingMedia(c *gin.Context)
	GetParentMedia(c *gin.Context)
	GetDescendantMedia(c *gin.Context)
	GetChildrenMedia(c *gin.Context)
}

type mediaHandler struct {
	MediaService services.MediaService
}

func NewMediaHandler(mediaService services.MediaService) MediaHandler {
	return &mediaHandler{MediaService: mediaService}
}

// UploadMedia godoc
// @Summary      Upload media
// @Description  Upload a media file
// @Tags         media
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "Media file"
// @Success      201   {object}  dto.MediaDTO
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /media/upload [post]
func (h *mediaHandler) UploadMedia(c *gin.Context) {
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "File is required",
		})
		return
	}
	defer file.Close()
	var createdBy *int64
	userID, exists := c.Get("user_id")
	if exists {
		id := int64(userID.(uint))
		createdBy = &id
	}
	media, err := h.MediaService.UploadAndSaveMedia(c.Request.Context(), file, fileHeader, createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, responses.CommonResponse{
		ResponseCode:    http.StatusCreated,
		ResponseMessage: "Media uploaded successfully",
		Data:            dto.ToMediaDTO(media),
	})
}

// AssociateMedia godoc
// @Summary      Associate media
// @Description  Associate a media file with another entity
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        id            path      int    true  "Media ID"
// @Param        association   body      object true  "Association info"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /media/{id}/associate [post]
func (h *mediaHandler) AssociateMedia(c *gin.Context) {
	mediaID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input requests.MediaAssociationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := h.MediaService.AssociateMedia(c.Request.Context(), mediaID, input.MediableID, input.MediableType, input.Group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Media associated"})
}

// CreateMediaNested godoc
// @Summary      Create media (nested)
// @Description  Create a new media as root or as a child (nested set)
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        media     body      models.Media  true  "Media to create"
// @Param        parent_id query     int           false "Parent media ID"
// @Success      201   {object}  dto.MediaDTO
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /media/nested [post]
func (h *mediaHandler) CreateMediaNested(c *gin.Context) {
	var input requests.MediaCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid request",
		})
		return
	}
	media := models.Media{
		Name: input.Name,
		Size: int64(input.Size),
	}
	var parentID *uint64
	if pid := c.Query("parent_id"); pid != "" {
		var parsed uint64
		if _, err := fmt.Sscan(pid, &parsed); err == nil {
			parentID = &parsed
		}
	}
	if err := h.MediaService.CreateNested(c.Request.Context(), &media, parentID); err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to create media (nested)",
		})
		return
	}
	c.JSON(http.StatusCreated, responses.CommonResponse{
		ResponseCode:    http.StatusCreated,
		ResponseMessage: "Media created successfully",
		Data:            dto.ToMediaDTO(&media),
	})
}

// MoveMediaNested godoc
// @Summary      Move media (nested)
// @Description  Move a media to a new parent (nested set)
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        id        path      int   true  "Media ID"
// @Param        parent_id query     int   false "New parent media ID"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /media/{id}/move [post]
func (h *mediaHandler) MoveMediaNested(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}
	var parentID *uint64
	if pid := c.Query("parent_id"); pid != "" {
		var parsed uint64
		if _, err := fmt.Sscan(pid, &parsed); err == nil {
			parentID = &parsed
		}
	}
	if err := h.MediaService.MoveNested(c.Request.Context(), id, parentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move media (nested)"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Media moved"})
}

// DeleteMediaNested godoc
// @Summary      Delete media (nested)
// @Description  Delete a media and its subtree (nested set)
// @Tags         media
// @Param        id   path  int  true  "Media ID"
// @Success      204
// @Failure      500  {object}  map[string]string
// @Router       /media/{id}/nested [delete]
func (h *mediaHandler) DeleteMediaNested(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}
	if err := h.MediaService.DeleteNested(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete media (nested)"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetSiblingMedia godoc
// @Summary      Get sibling media
// @Description  Get sibling media of a media item
// @Tags         media
// @Produce      json
// @Param        id   path      int  true  "Media ID"
// @Success      200  {array}   dto.MediaDTO
// @Failure      404  {object}  map[string]string
// @Router       /media/{id}/siblings [get]
func (h *mediaHandler) GetSiblingMedia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid media ID",
		})
		return
	}
	siblings, err := h.MediaService.GetSiblingMedia(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to fetch siblings",
		})
		return
	}
	dtos := make([]*dto.MediaDTO, len(siblings))
	for i, m := range siblings {
		dtos[i] = dto.ToMediaDTO(&m)
	}
	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Siblings fetched successfully",
		Data:            dtos,
	})
}

// GetParentMedia godoc
// @Summary      Get parent media
// @Description  Get parent of a media item
// @Tags         media
// @Produce      json
// @Param        id   path      int  true  "Media ID"
// @Success      200  {object}   dto.MediaDTO
// @Failure      404  {object}  map[string]string
// @Router       /media/{id}/parent [get]
func (h *mediaHandler) GetParentMedia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid media ID",
		})
		return
	}
	parent, err := h.MediaService.GetParentMedia(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to fetch parent",
		})
		return
	}
	if parent == nil {
		c.JSON(http.StatusNotFound, responses.CommonResponse{
			ResponseCode:    http.StatusNotFound,
			ResponseMessage: "No parent (root)",
		})
		return
	}
	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Parent fetched successfully",
		Data:            dto.ToMediaDTO(parent),
	})
}

// GetDescendantMedia godoc
// @Summary      Get descendant media
// @Description  Get all descendant media of a media item
// @Tags         media
// @Produce      json
// @Param        id   path      int  true  "Media ID"
// @Success      200  {array}   dto.MediaDTO
// @Failure      404  {object}  map[string]string
// @Router       /media/{id}/descendants [get]
func (h *mediaHandler) GetDescendantMedia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid media ID",
		})
		return
	}
	descendants, err := h.MediaService.GetDescendantMedia(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to fetch descendants",
		})
		return
	}
	dtos := make([]*dto.MediaDTO, len(descendants))
	for i, m := range descendants {
		dtos[i] = dto.ToMediaDTO(&m)
	}
	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Descendants fetched successfully",
		Data:            dtos,
	})
}

// GetChildrenMedia godoc
// @Summary      Get children media
// @Description  Get direct children of a media item
// @Tags         media
// @Produce      json
// @Param        id   path      int  true  "Media ID"
// @Success      200  {array}   dto.MediaDTO
// @Failure      404  {object}  map[string]string
// @Router       /media/{id}/children [get]
func (h *mediaHandler) GetChildrenMedia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid media ID",
		})
		return
	}
	children, err := h.MediaService.GetChildrenMedia(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to fetch children",
		})
		return
	}
	dtos := make([]*dto.MediaDTO, len(children))
	for i, m := range children {
		dtos[i] = dto.ToMediaDTO(&m)
	}
	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Children fetched successfully",
		Data:            dtos,
	})
}
