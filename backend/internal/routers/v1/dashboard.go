package v1

import (
	"go-next/internal/http/controllers"
	"go-next/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterDashboardRoutes(api *gin.RouterGroup, dashboardHandler controllers.DashboardHandler) {
	dashboard := api.Group("/dashboard")
	{
		dashboard.GET("/stats", middleware.JWTMiddleware(), dashboardHandler.GetDashboardStats)
	}
}
