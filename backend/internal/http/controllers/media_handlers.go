package controllers

import (
	"go-next/internal/http/requests"
	"go-next/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MediaHandler interface {
	UploadMedia(c *gin.Context)
	AssociateMedia(c *gin.Context)
	GetSiblingMedia(c *gin.Context)
	GetParentMedia(c *gin.Context)
	GetDescendantMedia(c *gin.Context)
	GetChildrenMedia(c *gin.Context)
	CreateMediaNested(c *gin.Context)
	MoveMediaNested(c *gin.Context)
	DeleteMediaNested(c *gin.Context)
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
// @Success      201   {object}  models.Media
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /media/upload [post]
func (h *mediaHandler) UploadMedia(c *gin.Context) {
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	defer file.Close()
	var createdBy *uuid.UUID
	userID, exists := c.Get("user_id")
	if exists {
		if id, ok := userID.(uuid.UUID); ok {
			createdBy = &id
		}
	}
	media, err := h.MediaService.UploadAndSaveMedia(file, fileHeader, createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, media)
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
	mediaIDStr := c.Param("id")
	mediaID, err := uuid.Parse(mediaIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}
	var input requests.MediaAssociationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := h.MediaService.AssociateMedia(mediaID, input.MediableID, input.MediableType, input.Group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Media associated"})
}

// GetSiblingMedia godoc
// @Summary      Get sibling media
// @Description  Get sibling media files
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Media ID"
// @Success      200  {object}  []models.Media
// @Failure      400  {object}  map[string]string
// @Router       /media/{id}/siblings [get]
func (h *mediaHandler) GetSiblingMedia(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

// GetParentMedia godoc
// @Summary      Get parent media
// @Description  Get parent media file
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Media ID"
// @Success      200  {object}  models.Media
// @Failure      400  {object}  map[string]string
// @Router       /media/{id}/parent [get]
func (h *mediaHandler) GetParentMedia(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

// GetDescendantMedia godoc
// @Summary      Get descendant media
// @Description  Get descendant media files
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Media ID"
// @Success      200  {object}  []models.Media
// @Failure      400  {object}  map[string]string
// @Router       /media/{id}/descendants [get]
func (h *mediaHandler) GetDescendantMedia(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

// GetChildrenMedia godoc
// @Summary      Get children media
// @Description  Get children media files
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Media ID"
// @Success      200  {object}  []models.Media
// @Failure      400  {object}  map[string]string
// @Router       /media/{id}/children [get]
func (h *mediaHandler) GetChildrenMedia(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

// CreateMediaNested godoc
// @Summary      Create nested media
// @Description  Create nested media structure
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        media body      object true  "Media data"
// @Success      201   {object}  models.Media
// @Failure      400   {object}  map[string]string
// @Router       /media/nested [post]
func (h *mediaHandler) CreateMediaNested(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

// MoveMediaNested godoc
// @Summary      Move nested media
// @Description  Move nested media structure
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        id    path      int    true  "Media ID"
// @Param        move  body      object true  "Move data"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Router       /media/{id}/move [post]
func (h *mediaHandler) MoveMediaNested(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

// DeleteMediaNested godoc
// @Summary      Delete nested media
// @Description  Delete nested media structure
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Media ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /media/{id}/nested [delete]
func (h *mediaHandler) DeleteMediaNested(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}
