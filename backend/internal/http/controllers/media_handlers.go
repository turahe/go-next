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
