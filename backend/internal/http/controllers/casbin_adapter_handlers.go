package controllers

import (
	"go-next/pkg/casbin"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CasbinAdapterHandler interface {
	GetDatabaseStats(c *gin.Context)
	GetPolicyCount(c *gin.Context)
	GetRoleCount(c *gin.Context)
	GetPolicyCountByRole(c *gin.Context)
	GetUserRoleCount(c *gin.Context)
	ClearAllPolicies(c *gin.Context)
	ClearPolicies(c *gin.Context)
	ClearRoles(c *gin.Context)
	BackupPolicies(c *gin.Context)
	RestorePolicies(c *gin.Context)
	GetAdapterConfig(c *gin.Context)
}

type casbinAdapterHandler struct {
	adapter *casbin.CasbinAdapter
}

func NewCasbinAdapterHandler(adapter *casbin.CasbinAdapter) CasbinAdapterHandler {
	return &casbinAdapterHandler{
		adapter: adapter,
	}
}

// GetDatabaseStats godoc
// @Summary Get Casbin database statistics
// @Description Retrieve comprehensive statistics about the Casbin database
// @Tags casbin-adapter
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Database statistics"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/adapter/stats [get]
// @Security BearerAuth
func (h *casbinAdapterHandler) GetDatabaseStats(c *gin.Context) {
	stats, err := h.adapter.GetDatabaseStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database statistics"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetPolicyCount godoc
// @Summary Get total policy count
// @Description Get the total number of policies in the database
// @Tags casbin-adapter
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Policy count"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/adapter/policies/count [get]
// @Security BearerAuth
func (h *casbinAdapterHandler) GetPolicyCount(c *gin.Context) {
	count, err := h.adapter.GetPolicyCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get policy count"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"policy_count": count})
}

// GetRoleCount godoc
// @Summary Get total role count
// @Description Get the total number of role assignments in the database
// @Tags casbin-adapter
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Role count"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/adapter/roles/count [get]
// @Security BearerAuth
func (h *casbinAdapterHandler) GetRoleCount(c *gin.Context) {
	count, err := h.adapter.GetRoleCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role count"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"role_count": count})
}

// GetPolicyCountByRole godoc
// @Summary Get policy count by role
// @Description Get the number of policies for a specific role
// @Tags casbin-adapter
// @Accept json
// @Produce json
// @Param role query string true "Role name"
// @Success 200 {object} map[string]interface{} "Policy count for role"
// @Failure 400 {object} map[string]interface{} "Invalid role parameter"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/adapter/policies/count/role [get]
// @Security BearerAuth
func (h *casbinAdapterHandler) GetPolicyCountByRole(c *gin.Context) {
	role := c.Query("role")
	if role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role parameter is required"})
		return
	}

	count, err := h.adapter.GetPolicyCountByRole(role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get policy count for role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"role": role, "policy_count": count})
}

// GetUserRoleCount godoc
// @Summary Get user role count
// @Description Get the number of role assignments for a specific user
// @Tags casbin-adapter
// @Accept json
// @Produce json
// @Param user_id query string true "User ID"
// @Success 200 {object} map[string]interface{} "Role count for user"
// @Failure 400 {object} map[string]interface{} "Invalid user ID parameter"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/adapter/users/roles/count [get]
// @Security BearerAuth
func (h *casbinAdapterHandler) GetUserRoleCount(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID parameter is required"})
		return
	}

	count, err := h.adapter.GetUserRoleCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role count for user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_id": userID, "role_count": count})
}

// ClearAllPolicies godoc
// @Summary Clear all policies and roles
// @Description Remove all policies and role assignments from the database
// @Tags casbin-adapter
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "All policies cleared successfully"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/adapter/clear/all [delete]
// @Security BearerAuth
func (h *casbinAdapterHandler) ClearAllPolicies(c *gin.Context) {
	err := h.adapter.ClearAllPolicies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear all policies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "All policies and roles cleared successfully"})
}

// ClearPolicies godoc
// @Summary Clear all policies
// @Description Remove all policies but keep role assignments
// @Tags casbin-adapter
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Policies cleared successfully"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/adapter/clear/policies [delete]
// @Security BearerAuth
func (h *casbinAdapterHandler) ClearPolicies(c *gin.Context) {
	err := h.adapter.ClearPolicies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear policies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Policies cleared successfully"})
}

// ClearRoles godoc
// @Summary Clear all role assignments
// @Description Remove all role assignments but keep policies
// @Tags casbin-adapter
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Roles cleared successfully"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/adapter/clear/roles [delete]
// @Security BearerAuth
func (h *casbinAdapterHandler) ClearRoles(c *gin.Context) {
	err := h.adapter.ClearRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear roles"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Roles cleared successfully"})
}

// BackupPolicies godoc
// @Summary Backup all policies
// @Description Create a backup of all policies and role assignments
// @Tags casbin-adapter
// @Accept json
// @Produce json
// @Success 200 {array} []string "Backup of all policies"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/adapter/backup [get]
// @Security BearerAuth
func (h *casbinAdapterHandler) BackupPolicies(c *gin.Context) {
	policies, err := h.adapter.BackupPolicies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to backup policies"})
		return
	}
	c.JSON(http.StatusOK, policies)
}

// RestorePolicies godoc
// @Summary Restore policies from backup
// @Description Restore all policies and role assignments from backup
// @Tags casbin-adapter
// @Accept json
// @Produce json
// @Param policies body [][]string true "Backup policies to restore"
// @Success 200 {object} map[string]interface{} "Policies restored successfully"
// @Failure 400 {object} map[string]interface{} "Invalid backup data"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/adapter/restore [post]
// @Security BearerAuth
func (h *casbinAdapterHandler) RestorePolicies(c *gin.Context) {
	var policies [][]string
	if err := c.ShouldBindJSON(&policies); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backup data"})
		return
	}

	err := h.adapter.RestorePolicies(policies)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore policies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Policies restored successfully"})
}

// GetAdapterConfig godoc
// @Summary Get adapter configuration
// @Description Get the current configuration of the Casbin adapter
// @Tags casbin-adapter
// @Accept json
// @Produce json
// @Success 200 {object} casbin.AdapterConfig "Adapter configuration"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/casbin/adapter/config [get]
// @Security BearerAuth
func (h *casbinAdapterHandler) GetAdapterConfig(c *gin.Context) {
	config := h.adapter.GetAdapterConfig()
	c.JSON(http.StatusOK, config)
}
