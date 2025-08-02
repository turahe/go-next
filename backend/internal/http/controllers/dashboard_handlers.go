package controllers

import (
	"net/http"
	"go-next/internal/models"
	"go-next/pkg/database"

	"github.com/gin-gonic/gin"
)

type DashboardHandler interface {
	GetDashboardStats(c *gin.Context)
}

type dashboardHandler struct{}

func NewDashboardHandler() DashboardHandler {
	return &dashboardHandler{}
}

// GetDashboardStats godoc
// @Summary      Get dashboard statistics
// @Description  Get overview statistics for the admin dashboard
// @Tags         dashboard
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /dashboard/stats [get]
func (h *dashboardHandler) GetDashboardStats(c *gin.Context) {
	var totalUsers int64
	var totalPosts int64
	var totalComments int64
	var activeUsers int64

	// Get total users
	if err := database.DB.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count users"})
		return
	}

	// Get total posts
	if err := database.DB.Model(&models.Post{}).Count(&totalPosts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count posts"})
		return
	}

	// Get total comments
	if err := database.DB.Model(&models.Comment{}).Count(&totalComments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count comments"})
		return
	}

	// Get active users (users with recent activity - simplified for now)
	// In a real implementation, you might track user activity timestamps
	if err := database.DB.Model(&models.User{}).Count(&activeUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count active users"})
		return
	}

	// Calculate growth rate (simplified - in real app you'd compare with previous period)
	growthRate := 12.5 // Mock growth rate

	// Calculate revenue (mock data for now)
	revenue := 45678.90

	c.JSON(http.StatusOK, gin.H{
		"totalUsers":    totalUsers,
		"activeUsers":   activeUsers,
		"totalPosts":    totalPosts,
		"totalComments": totalComments,
		"revenue":       revenue,
		"growthRate":    growthRate,
	})
}
