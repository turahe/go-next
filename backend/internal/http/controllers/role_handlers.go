package controllers

import (
	"net/http"
	"wordpress-go-next/backend/internal/http/requests"
	"wordpress-go-next/backend/internal/http/responses"
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

// GetRoles godoc
// @Summary      List roles
// @Description  Get roles with pagination and optional search
// @Tags         roles
// @Produce      json
// @Param        page     query     int     false  "Page number"  default(1)
// @Param        perPage  query     int     false  "Items per page"  default(10)
// @Param        search   query     string  false  "Search keyword"
// @Success      200  {object}  responses.PaginationResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /roles [get]
func (h *roleHandler) GetRoles(c *gin.Context) {
	pagination, err := requests.ParsePaginationFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid pagination parameters",
		})
		return
	}

	result, err := h.RoleService.GetRolesWithPagination(c.Request.Context(), pagination.Page, pagination.PerPage, pagination.Search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to fetch roles",
		})
		return
	}

	roles, ok := result.Data.([]models.Role)
	if !ok {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Invalid data format",
		})
		return
	}
	result.Data = roles

	c.JSON(http.StatusOK, result)
}

func (h *roleHandler) GetRole(c *gin.Context) {
	id := c.Param("id")
	role, err := h.RoleService.GetRoleByID(c.Request.Context(), id)
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
	if err := h.RoleService.CreateRole(c.Request.Context(), &role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}
	c.JSON(http.StatusCreated, role)
}

func (h *roleHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	role, err := h.RoleService.GetRoleByID(c.Request.Context(), id)
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
	if err := h.RoleService.UpdateRole(c.Request.Context(), role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (h *roleHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if err := h.RoleService.DeleteRole(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role"})
		return
	}
	c.Status(http.StatusNoContent)
}
