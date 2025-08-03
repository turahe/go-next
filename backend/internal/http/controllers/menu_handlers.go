package controllers

import (
	"go-next/internal/http/requests"
	"go-next/internal/http/responses"
	"go-next/internal/rules"
	"net/http"

	"go-next/internal/models"
	"go-next/internal/services"
	"go-next/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MenuHandler interface {
	GetMenus(c *gin.Context)
	GetMenu(c *gin.Context)
	GetMenuTree(c *gin.Context)
	GetMenuByParent(c *gin.Context)
	CreateMenu(c *gin.Context)
	UpdateMenu(c *gin.Context)
	DeleteMenu(c *gin.Context)
	GetMenuDescendants(c *gin.Context)
	GetMenuAncestors(c *gin.Context)
	GetMenuSiblings(c *gin.Context)
	MoveMenu(c *gin.Context)

	// Role-related handlers
	AssignRoleToMenu(c *gin.Context)
	RemoveRoleFromMenu(c *gin.Context)
	GetMenuRoles(c *gin.Context)
	GetRoleMenus(c *gin.Context)
}

type menuHandler struct {
	MenuService services.MenuService
}

func NewMenuHandler(menuService services.MenuService) MenuHandler {
	return &menuHandler{MenuService: menuService}
}

// GetMenus godoc
// @Summary      List menus
// @Description  Get all menus with pagination
// @Tags         menus
// @Produce      json
// @Param        page      query     int    false "Page number"
// @Param        per_page  query     int    false "Items per page"
// @Param        search    query     string false "Search term"
// @Param        parent    query     string false "Parent menu ID filter"
// @Success      200       {object}  responses.LaravelPaginationResponse
// @Failure      500       {object}  map[string]string
// @Router       /menus [get]
func (h *menuHandler) GetMenus(c *gin.Context) {
	params := responses.ParsePaginationParams(c)
	search := c.Query("search")
	parentID := c.Query("parent")

	offset := (params.Page - 1) * params.PerPage

	var menus []models.Menu
	var total int64

	query := database.DB.Model(&models.Menu{})

	// Apply search filter
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Apply parent filter
	if parentID != "" {
		if parsedID, err := uuid.Parse(parentID); err == nil {
			query = query.Where("parent_id = ?", parsedID)
		} else {
			query = query.Where("parent_id IS NULL")
		}
	} else {
		query = query.Where("parent_id IS NULL")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to count menus")
		return
	}

	// Get paginated menus
	if err := query.Offset(offset).Limit(params.PerPage).Order("ordering ASC").Find(&menus).Error; err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch menus")
		return
	}

	// Send Laravel-style pagination response
	responses.SendLaravelPaginationWithMessage(c, "Menus retrieved successfully", menus, total, int64(params.Page), int64(params.PerPage))
}

// GetMenu godoc
// @Summary      Get menu
// @Description  Get a menu by ID
// @Tags         menus
// @Produce      json
// @Param        id   path      string  true  "Menu ID"
// @Success      200  {object}  models.Menu
// @Failure      404  {object}  map[string]string
// @Router       /menus/{id} [get]
func (h *menuHandler) GetMenu(c *gin.Context) {
	id := c.Param("id")
	menuID, err := uuid.Parse(id)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid menu ID")
		return
	}

	menu, err := h.MenuService.GetMenuByID(menuID)
	if err != nil {
		responses.SendError(c, http.StatusNotFound, "Menu not found")
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Menu retrieved successfully", menu)
}

// GetMenuTree godoc
// @Summary      Get menu tree
// @Description  Get all menus in tree structure
// @Tags         menus
// @Produce      json
// @Success      200  {array}   models.Menu
// @Failure      500  {object}  map[string]string
// @Router       /menus/tree [get]
func (h *menuHandler) GetMenuTree(c *gin.Context) {
	menus, err := h.MenuService.GetMenuTree()
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch menu tree")
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Menu tree retrieved successfully", menus)
}

// GetMenuByParent godoc
// @Summary      Get menus by parent
// @Description  Get all menus by parent ID
// @Tags         menus
// @Produce      json
// @Param        parent_id   path      string  true  "Parent Menu ID"
// @Success      200         {array}   models.Menu
// @Failure      500         {object}  map[string]string
// @Router       /menus/parent/{parent_id} [get]
func (h *menuHandler) GetMenuByParent(c *gin.Context) {
	parentID := c.Param("parent_id")
	parsedID, err := uuid.Parse(parentID)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid parent ID")
		return
	}

	menus, err := h.MenuService.GetMenuByParentID(parsedID)
	if err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch menus")
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Menus retrieved successfully", menus)
}

// CreateMenu godoc
// @Summary      Create menu
// @Description  Create a new menu
// @Tags         menus
// @Accept       json
// @Produce      json
// @Param        menu  body      requests.MenuCreateRequest  true  "Menu data"
// @Success      201   {object}  models.Menu
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /menus [post]
func (h *menuHandler) CreateMenu(c *gin.Context) {
	var req requests.MenuCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	// Validate request
	if err := rules.Validate.Struct(req); err != nil {
		responses.SendValidationError(c, err)
		return
	}

	menu := &models.Menu{
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		URL:         req.URL,
		ParentID:    uuid.Nil,
	}
	menu.RecordOrdering = req.Ordering

	if req.ParentID != nil {
		menu.ParentID = *req.ParentID
	}

	if err := h.MenuService.CreateMenu(menu); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to create menu")
		return
	}

	responses.SendSuccess(c, http.StatusCreated, "Menu created successfully", menu)
}

// UpdateMenu godoc
// @Summary      Update menu
// @Description  Update an existing menu
// @Tags         menus
// @Accept       json
// @Produce      json
// @Param        id    path      string                    true  "Menu ID"
// @Param        menu  body      requests.MenuUpdateRequest true  "Menu data"
// @Success      200   {object}  models.Menu
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /menus/{id} [put]
func (h *menuHandler) UpdateMenu(c *gin.Context) {
	id := c.Param("id")
	menuID, err := uuid.Parse(id)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid menu ID")
		return
	}

	var req requests.MenuUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	// Validate request
	if err := rules.Validate.Struct(req); err != nil {
		responses.SendValidationError(c, err)
		return
	}

	// Get existing menu
	existingMenu, err := h.MenuService.GetMenuByID(menuID)
	if err != nil {
		responses.SendError(c, http.StatusNotFound, "Menu not found")
		return
	}

	// Update menu fields
	existingMenu.Name = req.Name
	existingMenu.Description = req.Description
	existingMenu.Icon = req.Icon
	existingMenu.URL = req.URL
	existingMenu.RecordOrdering = req.Ordering

	if req.ParentID != nil {
		existingMenu.ParentID = *req.ParentID
	} else {
		existingMenu.ParentID = uuid.Nil
	}

	if err := h.MenuService.UpdateMenu(existingMenu); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to update menu")
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Menu updated successfully", existingMenu)
}

// DeleteMenu godoc
// @Summary      Delete menu
// @Description  Delete a menu by ID
// @Tags         menus
// @Produce      json
// @Param        id   path      string  true  "Menu ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /menus/{id} [delete]
func (h *menuHandler) DeleteMenu(c *gin.Context) {
	id := c.Param("id")
	menuID, err := uuid.Parse(id)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid menu ID")
		return
	}

	if err := h.MenuService.DeleteMenu(menuID); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to delete menu")
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Menu deleted successfully", nil)
}

// GetMenuDescendants godoc
// @Summary      Get menu descendants
// @Description  Get all descendants of a menu using nested set model
// @Tags         menus
// @Produce      json
// @Param        id   path      string  true  "Menu ID"
// @Success      200  {array}   models.Menu
// @Failure      404  {object}  map[string]string
// @Router       /menus/{id}/descendants [get]
func (h *menuHandler) GetMenuDescendants(c *gin.Context) {
	id := c.Param("id")
	menuID, err := uuid.Parse(id)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid menu ID")
		return
	}

	descendants, err := h.MenuService.GetMenuDescendants(menuID)
	if err != nil {
		responses.SendError(c, http.StatusNotFound, "Menu not found")
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Menu descendants retrieved successfully", descendants)
}

// GetMenuAncestors godoc
// @Summary      Get menu ancestors
// @Description  Get all ancestors of a menu using nested set model
// @Tags         menus
// @Produce      json
// @Param        id   path      string  true  "Menu ID"
// @Success      200  {array}   models.Menu
// @Failure      404  {object}  map[string]string
// @Router       /menus/{id}/ancestors [get]
func (h *menuHandler) GetMenuAncestors(c *gin.Context) {
	id := c.Param("id")
	menuID, err := uuid.Parse(id)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid menu ID")
		return
	}

	ancestors, err := h.MenuService.GetMenuAncestors(menuID)
	if err != nil {
		responses.SendError(c, http.StatusNotFound, "Menu not found")
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Menu ancestors retrieved successfully", ancestors)
}

// GetMenuSiblings godoc
// @Summary      Get menu siblings
// @Description  Get all siblings of a menu using nested set model
// @Tags         menus
// @Produce      json
// @Param        id   path      string  true  "Menu ID"
// @Success      200  {array}   models.Menu
// @Failure      404  {object}  map[string]string
// @Router       /menus/{id}/siblings [get]
func (h *menuHandler) GetMenuSiblings(c *gin.Context) {
	id := c.Param("id")
	menuID, err := uuid.Parse(id)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid menu ID")
		return
	}

	siblings, err := h.MenuService.GetMenuSiblings(menuID)
	if err != nil {
		responses.SendError(c, http.StatusNotFound, "Menu not found")
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Menu siblings retrieved successfully", siblings)
}

// MoveMenu godoc
// @Summary      Move menu
// @Description  Move a menu and its subtree to a new parent using nested set model
// @Tags         menus
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Menu ID"
// @Param        new_parent_id body     string true "New parent ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /menus/{id}/move [post]
func (h *menuHandler) MoveMenu(c *gin.Context) {
	id := c.Param("id")
	menuID, err := uuid.Parse(id)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid menu ID")
		return
	}

	var req struct {
		NewParentID uuid.UUID `json:"new_parent_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	if err := h.MenuService.MoveMenu(menuID, req.NewParentID); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to move menu")
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Menu moved successfully", nil)
}

// AssignRoleToMenu godoc
// @Summary      Assign role to menu
// @Description  Assign a role to a menu
// @Tags         menus
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Menu ID"
// @Param        role_id body   string  true  "Role ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /menus/{id}/roles [post]
func (h *menuHandler) AssignRoleToMenu(c *gin.Context) {
	id := c.Param("id")
	menuID, err := uuid.Parse(id)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid menu ID")
		return
	}

	var req struct {
		RoleID uuid.UUID `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	if err := h.MenuService.AssignRoleToMenu(menuID, req.RoleID); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to assign role to menu")
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Role assigned to menu successfully", nil)
}

// RemoveRoleFromMenu godoc
// @Summary      Remove role from menu
// @Description  Remove a role from a menu
// @Tags         menus
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Menu ID"
// @Param        role_id body   string  true  "Role ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /menus/{id}/roles [delete]
func (h *menuHandler) RemoveRoleFromMenu(c *gin.Context) {
	id := c.Param("id")
	menuID, err := uuid.Parse(id)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid menu ID")
		return
	}

	var req struct {
		RoleID uuid.UUID `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	if err := h.MenuService.RemoveRoleFromMenu(menuID, req.RoleID); err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to remove role from menu")
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Role removed from menu successfully", nil)
}

// GetMenuRoles godoc
// @Summary      Get menu roles
// @Description  Get all roles assigned to a menu
// @Tags         menus
// @Produce      json
// @Param        id   path      string  true  "Menu ID"
// @Success      200  {array}   models.Role
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /menus/{id}/roles [get]
func (h *menuHandler) GetMenuRoles(c *gin.Context) {
	id := c.Param("id")
	menuID, err := uuid.Parse(id)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid menu ID")
		return
	}

	roles, err := h.MenuService.GetMenuRoles(menuID)
	if err != nil {
		responses.SendError(c, http.StatusNotFound, "Menu not found")
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Menu roles retrieved successfully", roles)
}

// GetRoleMenus godoc
// @Summary      Get role menus
// @Description  Get all menus assigned to a role
// @Tags         menus
// @Produce      json
// @Param        role_id path   string  true  "Role ID"
// @Success      200  {array}   models.Menu
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /roles/{role_id}/menus [get]
func (h *menuHandler) GetRoleMenus(c *gin.Context) {
	roleID := c.Param("role_id")
	parsedRoleID, err := uuid.Parse(roleID)
	if err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid role ID")
		return
	}

	menus, err := h.MenuService.GetRoleMenus(parsedRoleID)
	if err != nil {
		responses.SendError(c, http.StatusNotFound, "Role not found")
		return
	}

	responses.SendSuccess(c, http.StatusOK, "Role menus retrieved successfully", menus)
}
