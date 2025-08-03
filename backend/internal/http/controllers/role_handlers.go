package controllers

import (
	"go-next/internal/http/requests"
	"go-next/internal/services"
	"go-next/pkg/validation"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoleHandler interface {
	GetRoles(c *gin.Context)
	GetRole(c *gin.Context)
	CreateRole(c *gin.Context)
	UpdateRole(c *gin.Context)
	DeleteRole(c *gin.Context)

	// Menu-related handlers
	AssignMenuToRole(c *gin.Context)
	RemoveMenuFromRole(c *gin.Context)
	GetRoleMenus(c *gin.Context)
	GetMenuRoles(c *gin.Context)
}

type roleHandler struct {
	RoleService services.RoleService
}

func NewRoleHandler(roleService services.RoleService) RoleHandler {
	// Initialize validator with custom error messages
	validator := validation.NewValidator()
	validator.AddCustomMessage("name", "unique_role", "This role name is already taken")

	return &roleHandler{RoleService: roleService}
}

// GetRoles godoc
// @Summary Get all roles
// @Description Retrieve a list of all available roles
// @Tags roles
// @Accept json
// @Produce json
// @Success 200 {array} models.Role "List of roles"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/roles [get]
// @Security BearerAuth
func (h *roleHandler) GetRoles(c *gin.Context) {
	roles, err := h.RoleService.GetAllRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch roles"})
		return
	}
	c.JSON(http.StatusOK, roles)
}

// GetRole godoc
// @Summary Get a specific role
// @Description Retrieve a specific role by its ID
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" format(uuid)
// @Success 200 {object} models.Role "Role details"
// @Failure 404 {object} map[string]interface{} "Role not found"
// @Router /api/v1/roles/{id} [get]
// @Security BearerAuth
func (h *roleHandler) GetRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID format"})
		return
	}
	role, err := h.RoleService.GetRoleByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}
	c.JSON(http.StatusOK, role)
}

// CreateRole godoc
// @Summary Create a new role
// @Description Create a new role with the provided details
// @Tags roles
// @Accept json
// @Produce json
// @Param role body requests.RoleCreateRequest true "Role creation data"
// @Success 201 {object} models.Role "Role created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request data"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/roles [post]
// @Security BearerAuth
func (h *roleHandler) CreateRole(c *gin.Context) {
	var input requests.RoleCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate the request
	validator := validation.NewValidator()
	result := validator.Validate(&input)
	if !result.IsValid {
		// Get the first error message
		var errorMsg string
		for _, errors := range result.Errors {
			if len(errors) > 0 {
				errorMsg = errors[0]
				break
			}
		}
		if errorMsg == "" {
			errorMsg = "Validation failed"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return
	}

	role, err := h.RoleService.CreateRole(input.Name, input.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}
	c.JSON(http.StatusCreated, role)
}

// UpdateRole godoc
// @Summary Update a role
// @Description Update an existing role with new details
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" format(uuid)
// @Param role body requests.RoleCreateRequest true "Role update data"
// @Success 200 {object} models.Role "Role updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request data"
// @Failure 404 {object} map[string]interface{} "Role not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/roles/{id} [put]
// @Security BearerAuth
func (h *roleHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID format"})
		return
	}
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

	// Validate the request
	validator := validation.NewValidator()
	result := validator.Validate(&input)
	if !result.IsValid {
		// Get the first error message
		var errorMsg string
		for _, errors := range result.Errors {
			if len(errors) > 0 {
				errorMsg = errors[0]
				break
			}
		}
		if errorMsg == "" {
			errorMsg = "Validation failed"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return
	}

	role.Name = input.Name
	if err := h.RoleService.UpdateRole(role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}
	c.JSON(http.StatusOK, role)
}

// DeleteRole godoc
// @Summary Delete a role
// @Description Delete a role by its ID
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" format(uuid)
// @Success 204 "Role deleted successfully"
// @Failure 404 {object} map[string]interface{} "Role not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/roles/{id} [delete]
// @Security BearerAuth
func (h *roleHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID format"})
		return
	}
	if err := h.RoleService.DeleteRole(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role"})
		return
	}
	c.Status(http.StatusNoContent)
}

// AssignMenuToRole godoc
// @Summary Assign menu to role
// @Description Assign a menu to a role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" format(uuid)
// @Param menu_id body string true "Menu ID"
// @Success 200 {object} map[string]string "Menu assigned to role successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request data"
// @Failure 404 {object} map[string]interface{} "Role or menu not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/roles/{id}/menus [post]
// @Security BearerAuth
func (h *roleHandler) AssignMenuToRole(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID format"})
		return
	}

	var req struct {
		MenuID uuid.UUID `json:"menu_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := h.RoleService.AssignMenuToRole(roleID, req.MenuID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign menu to role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Menu assigned to role successfully"})
}

// RemoveMenuFromRole godoc
// @Summary Remove menu from role
// @Description Remove a menu from a role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" format(uuid)
// @Param menu_id body string true "Menu ID"
// @Success 200 {object} map[string]string "Menu removed from role successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request data"
// @Failure 404 {object} map[string]interface{} "Role or menu not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/roles/{id}/menus [delete]
// @Security BearerAuth
func (h *roleHandler) RemoveMenuFromRole(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID format"})
		return
	}

	var req struct {
		MenuID uuid.UUID `json:"menu_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := h.RoleService.RemoveMenuFromRole(roleID, req.MenuID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove menu from role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Menu removed from role successfully"})
}

// GetRoleMenus godoc
// @Summary Get role menus
// @Description Get all menus assigned to a role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" format(uuid)
// @Success 200 {array} models.Menu "List of menus assigned to the role"
// @Failure 400 {object} map[string]interface{} "Invalid role ID format"
// @Failure 404 {object} map[string]interface{} "Role not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/roles/{id}/menus [get]
// @Security BearerAuth
func (h *roleHandler) GetRoleMenus(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID format"})
		return
	}

	menus, err := h.RoleService.GetRoleMenus(roleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	c.JSON(http.StatusOK, menus)
}

// GetMenuRoles godoc
// @Summary Get menu roles
// @Description Get all roles assigned to a menu
// @Tags roles
// @Accept json
// @Produce json
// @Param menu_id path string true "Menu ID" format(uuid)
// @Success 200 {array} models.Role "List of roles assigned to the menu"
// @Failure 400 {object} map[string]interface{} "Invalid menu ID format"
// @Failure 404 {object} map[string]interface{} "Menu not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/menus/{menu_id}/roles [get]
// @Security BearerAuth
func (h *roleHandler) GetMenuRoles(c *gin.Context) {
	menuIDStr := c.Param("menu_id")
	menuID, err := uuid.Parse(menuIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID format"})
		return
	}

	roles, err := h.RoleService.GetMenuRoles(menuID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
		return
	}

	c.JSON(http.StatusOK, roles)
}
