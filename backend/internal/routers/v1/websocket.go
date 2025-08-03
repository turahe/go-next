package v1

import (
	"go-next/internal/http/controllers"
	"go-next/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterWebSocketRoutes(api *gin.RouterGroup, wsHandler *controllers.WebSocketHandler) {
	ws := api.Group("/ws")
	{
		ws.GET("/status", wsHandler.GetWebSocketStatus)
		ws.GET("/connect", middleware.JWTMiddleware(), wsHandler.HandleWebSocket)
	}
}
