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
	var createdBy *int64
	userID, exists := c.Get("user_id")
	if exists {
		id := int64(userID.(uint))
		createdBy = &id
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
	mediaID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input requests.MediaAssociationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := h.MediaService.AssociateMedia(uint(mediaID), input.MediableID, input.MediableType, input.Group); err != nil {
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
// @Success      201   {object}  models.Media
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /media/nested [post]
func (h *mediaHandler) CreateMediaNested(c *gin.Context) {
	var input requests.MediaCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	media := models.Media{
		Name: input.Name,
		Size: int(input.Size),
		// Map other fields as appropriate if they exist in the request and model
	}
	var parentID *int64
	if pid := c.Query("parent_id"); pid != "" {
		var parsed int64
		if _, err := fmt.Sscan(pid, &parsed); err == nil {
			parentID = &parsed
		}
	}
	if err := h.MediaService.CreateNested(&media, parentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create media (nested)"})
		return
	}
	c.JSON(http.StatusCreated, media)
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
	var parentID *int64
	if pid := c.Query("parent_id"); pid != "" {
		var parsed int64
		if _, err := fmt.Sscan(pid, &parsed); err == nil {
			parentID = &parsed
		}
	}
	if err := h.MediaService.MoveNested(uint(id), parentID); err != nil {
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
	if err := h.MediaService.DeleteNested(uint(id)); err != nil {
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
// @Success      200  {array}   models.Media
// @Failure      404  {object}  map[string]string
// @Router       /media/{id}/siblings [get]
func (h *mediaHandler) GetSiblingMedia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}
	siblings, err := h.MediaService.GetSiblingMedia(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch siblings"})
		return
	}
	c.JSON(http.StatusOK, siblings)
}

// GetParentMedia godoc
// @Summary      Get parent media
// @Description  Get parent of a media item
// @Tags         media
// @Produce      json
// @Param        id   path      int  true  "Media ID"
// @Success      200  {object}   models.Media
// @Failure      404  {object}  map[string]string
// @Router       /media/{id}/parent [get]
func (h *mediaHandler) GetParentMedia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}
	parent, err := h.MediaService.GetParentMedia(uint(id))
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

// GetDescendantMedia godoc
// @Summary      Get descendant media
// @Description  Get all descendant media of a media item
// @Tags         media
// @Produce      json
// @Param        id   path      int  true  "Media ID"
// @Success      200  {array}   models.Media
// @Failure      404  {object}  map[string]string
// @Router       /media/{id}/descendants [get]
func (h *mediaHandler) GetDescendantMedia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}
	descendants, err := h.MediaService.GetDescendantMedia(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch descendants"})
		return
	}
	c.JSON(http.StatusOK, descendants)
}

// GetChildrenMedia godoc
// @Summary      Get children media
// @Description  Get direct children of a media item
// @Tags         media
// @Produce      json
// @Param        id   path      int  true  "Media ID"
// @Success      200  {array}   models.Media
// @Failure      404  {object}  map[string]string
// @Router       /media/{id}/children [get]
func (h *mediaHandler) GetChildrenMedia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
		return
	}
	children, err := h.MediaService.GetChildrenMedia(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch children"})
		return
	}
	c.JSON(http.StatusOK, children)
}
