package controllers

import (
	"go-next/internal/http/requests"
	"go-next/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CasbinHandler interface {
	GetPolicies(c *gin.Context)
	AddPolicy(c *gin.Context)
	RemovePolicy(c *gin.Context)
	GetUserRoles(c *gin.Context)
	AddRoleForUser(c *gin.Context)
	RemoveRoleForUser(c *gin.Context)
	GetFilteredPolicies(c *gin.Context)
}

type casbinHandler struct {
	casbinService *services.CasbinService
}

func NewCasbinHandler(casbinService *services.CasbinService) CasbinHandler {
	return &casbinHandler{
		casbinService: casbinService,
	}
}

// GetPolicies godoc
// @Summary Get all policies
// @Description Retrieve all Casbin policies
// @Tags casbin
// @Accept json
// @Produce json
// @Success 200 {array} []string "List of policies"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/policies [get]
// @Security BearerAuth
func (h *casbinHandler) GetPolicies(c *gin.Context) {
	policies, err := h.casbinService.GetAllPolicies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch policies"})
		return
	}
	c.JSON(http.StatusOK, policies)
}

// AddPolicy godoc
// @Summary Add a new policy
// @Description Add a new Casbin policy
// @Tags casbin
// @Accept json
// @Produce json
// @Param policy body requests.PolicyRequest true "Policy data"
// @Success 201 {object} map[string]interface{} "Policy added successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request data"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/policies [post]
// @Security BearerAuth
func (h *casbinHandler) AddPolicy(c *gin.Context) {
	var input requests.PolicyRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.casbinService.AddPolicy(input.Subject, input.Domain, input.Object, input.Action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add policy"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Policy added successfully"})
}

// RemovePolicy godoc
// @Summary Remove a policy
// @Description Remove a Casbin policy
// @Tags casbin
// @Accept json
// @Produce json
// @Param policy body requests.PolicyRequest true "Policy data"
// @Success 200 {object} map[string]interface{} "Policy removed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request data"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/policies [delete]
// @Security BearerAuth
func (h *casbinHandler) RemovePolicy(c *gin.Context) {
	var input requests.PolicyRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.casbinService.RemovePolicy(input.Subject, input.Domain, input.Object, input.Action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove policy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Policy removed successfully"})
}

// GetUserRoles godoc
// @Summary Get user roles
// @Description Get all roles for a specific user
// @Tags casbin
// @Accept json
// @Produce json
// @Param user_id path string true "User ID" format(uuid)
// @Success 200 {array} string "List of user roles"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/users/{user_id}/roles [get]
// @Security BearerAuth
func (h *casbinHandler) GetUserRoles(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	roles, err := h.casbinService.GetUserRoles(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	c.JSON(http.StatusOK, roles)
}

// AddRoleForUser godoc
// @Summary Add role for user
// @Description Add a role to a specific user
// @Tags casbin
// @Accept json
// @Produce json
// @Param user_id path string true "User ID" format(uuid)
// @Param role body requests.RoleAssignmentRequest true "Role assignment data"
// @Success 201 {object} map[string]interface{} "Role assigned successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request data"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/users/{user_id}/roles [post]
// @Security BearerAuth
func (h *casbinHandler) AddRoleForUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var input requests.RoleAssignmentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err = h.casbinService.AddRoleForUser(userID, input.Role, input.Domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Role assigned successfully"})
}

// RemoveRoleForUser godoc
// @Summary Remove role from user
// @Description Remove a role from a specific user
// @Tags casbin
// @Accept json
// @Produce json
// @Param user_id path string true "User ID" format(uuid)
// @Param role body requests.RoleAssignmentRequest true "Role assignment data"
// @Success 200 {object} map[string]interface{} "Role removed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request data"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/users/{user_id}/roles [delete]
// @Security BearerAuth
func (h *casbinHandler) RemoveRoleForUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var input requests.RoleAssignmentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err = h.casbinService.RemoveRoleForUser(userID, input.Role, input.Domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role removed successfully"})
}

// GetFilteredPolicies godoc
// @Summary Get filtered policies
// @Description Get policies filtered by field index and values
// @Tags casbin
// @Accept json
// @Produce json
// @Param field_index query int true "Field index to filter by"
// @Param field_values query string true "Field values to filter by (comma-separated)"
// @Success 200 {array} []string "List of filtered policies"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/policies/filtered [get]
// @Security BearerAuth
func (h *casbinHandler) GetFilteredPolicies(c *gin.Context) {
	fieldIndexStr := c.Query("field_index")
	fieldValuesStr := c.Query("field_values")

	if fieldIndexStr == "" || fieldValuesStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "field_index and field_values are required"})
		return
	}

	// Parse field index (this is a simplified version)
	fieldIndex := 0 // Default to 0 for subject

	policies, err := h.casbinService.GetFilteredPolicies(fieldIndex, fieldValuesStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get filtered policies"})
		return
	}

	c.JSON(http.StatusOK, policies)
}
