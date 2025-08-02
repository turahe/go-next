package controllers

import (
	"net/http"
	"go-next/internal/services"

	"github.com/gin-gonic/gin"
)

type WebSocketHandler struct {
	hub *services.Hub
}

func NewWebSocketHandler(hub *services.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

// HandleWebSocket handles WebSocket connections with authentication
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Extract user ID from JWT token
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	services.ServeWs(h.hub, c.Writer, c.Request, userID)
}

// GetWebSocketStatus returns the current WebSocket hub status
func (h *WebSocketHandler) GetWebSocketStatus(c *gin.Context) {
	// This endpoint can be used to check if WebSocket service is running
	c.JSON(http.StatusOK, gin.H{
		"status":  "running",
		"message": "WebSocket service is active",
	})
}
