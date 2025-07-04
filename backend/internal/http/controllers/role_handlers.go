package controllers

import (
	"net/http"
	"wordpress-go-next/backend/internal/http/requests"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type RoleHandler interface {
	GetRoles(c *gin.Context)
	GetRole(c *gin.Context)
	CreateRole(c *gin.Context)
	UpdateRole(c *gin.Context)
	DeleteRole(c *gin.Context)
}

type roleHandler struct {
	RoleService services.RoleService
}

func NewRoleHandler(roleService services.RoleService) RoleHandler {
	return &roleHandler{RoleService: roleService}
}

func (h *roleHandler) GetRoles(c *gin.Context) {
	roles, err := h.RoleService.GetAllRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch roles"})
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (h *roleHandler) GetRole(c *gin.Context) {
	id := c.Param("id")
	role, err := h.RoleService.GetRoleByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (h *roleHandler) CreateRole(c *gin.Context) {
	var input requests.RoleCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	role := models.Role{Name: input.Name}
	if err := h.RoleService.CreateRole(&role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}
	c.JSON(http.StatusCreated, role)
}

func (h *roleHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	role, err := h.RoleService.GetRoleByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}
	var input requests.RoleCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	role.Name = input.Name
	if err := h.RoleService.UpdateRole(role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (h *roleHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if err := h.RoleService.DeleteRole(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role"})
		return
	}
	c.Status(http.StatusNoContent)
}
