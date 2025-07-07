package controllers

import (
	"net/http"
	"wordpress-go-next/backend/internal/http/requests"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type UserRoleHandler interface {
	AssignRoleToUser(c *gin.Context)
	RemoveRoleFromUser(c *gin.Context)
	ListUserRoles(c *gin.Context)
}

type userRoleHandler struct {
	UserRoleService services.UserRoleService
}

func NewUserRoleHandler(userRoleService services.UserRoleService) UserRoleHandler {
	return &userRoleHandler{UserRoleService: userRoleService}
}

func (h *userRoleHandler) AssignRoleToUser(c *gin.Context) {
	var input requests.UserRoleAssignmentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	user := &models.User{}
	role := &models.Role{Base: models.Base{ID: input.RoleID}}
	if err := h.UserRoleService.AssignRoleToUser(c.Request.Context(), user, role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role assigned"})
}

func (h *userRoleHandler) RemoveRoleFromUser(c *gin.Context) {
	user := &models.User{}
	role := &models.Role{}
	if err := h.UserRoleService.RemoveRoleFromUser(c.Request.Context(), user, role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role removed"})
}

func (h *userRoleHandler) ListUserRoles(c *gin.Context) {
	user := &models.User{}
	roles, err := h.UserRoleService.ListUserRoles(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list roles"})
		return
	}
	c.JSON(http.StatusOK, roles)
}
