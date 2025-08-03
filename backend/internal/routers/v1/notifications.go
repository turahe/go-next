package v1

import (
	"go-next/internal/http/controllers"
	"go-next/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterNotificationRoutes(api *gin.RouterGroup, notificationHandler *controllers.NotificationHandler) {
	notifications := api.Group("/notifications")
	{
		notifications.GET("", middleware.JWTMiddleware(), notificationHandler.GetUserNotifications)
		notifications.GET("/unread-count", middleware.JWTMiddleware(), notificationHandler.GetUnreadCount)
		notifications.PUT(":id/read", middleware.JWTMiddleware(), notificationHandler.MarkAsRead)
		notifications.PUT("/mark-all-read", middleware.JWTMiddleware(), notificationHandler.MarkAllAsRead)
		notifications.DELETE(":id", middleware.JWTMiddleware(), notificationHandler.DeleteNotification)
		notifications.DELETE("", middleware.JWTMiddleware(), notificationHandler.DeleteAllNotifications)
	}
}

func RegisterAdminNotificationRoutes(api *gin.RouterGroup, notificationHandler *controllers.NotificationHandler) {
	admin := api.Group("/admin")
	{
		admin.POST("/notifications", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/admin/notifications", "POST"), notificationHandler.CreateNotification)
	}
}
