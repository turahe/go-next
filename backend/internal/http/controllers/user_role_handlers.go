package controllers

import (
	"go-next/internal/http/requests"
	"go-next/internal/models"
	"go-next/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	var input requests.UserRoleAssignmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	user := &models.User{BaseModel: models.BaseModel{ID: uuid.Nil}}
	role := &models.Role{BaseModel: models.BaseModel{ID: input.RoleID}}
	if err := h.UserRoleService.AssignRoleToUser(user, role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role assigned"})
}

func (h *userRoleHandler) RemoveRoleFromUser(c *gin.Context) {
	user := &models.User{BaseModel: models.BaseModel{ID: uuid.Nil}}
	role := &models.Role{BaseModel: models.BaseModel{ID: uuid.Nil}}
	if err := h.UserRoleService.RemoveRoleFromUser(user, role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role removed"})
}

func (h *userRoleHandler) ListUserRoles(c *gin.Context) {
	user := &models.User{BaseModel: models.BaseModel{ID: uuid.Nil}}
	roles, err := h.UserRoleService.ListUserRoles(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list roles"})
		return
	}
	c.JSON(http.StatusOK, roles)
}
