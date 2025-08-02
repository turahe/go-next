package v1

import (
	"wordpress-go-next/backend/internal/http/controllers"
	"wordpress-go-next/backend/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterMediaRoutes(api *gin.RouterGroup, mediaHandler controllers.MediaHandler) {
	media := api.Group("/media")
	{
		media.POST("/upload", middleware.JWTMiddleware(), mediaHandler.UploadMedia)
		media.POST(":id/associate", middleware.JWTMiddleware(), mediaHandler.AssociateMedia)
		media.GET(":id/siblings", mediaHandler.GetSiblingMedia)
		media.GET(":id/parent", mediaHandler.GetParentMedia)
		media.GET(":id/descendants", mediaHandler.GetDescendantMedia)
		media.GET(":id/children", mediaHandler.GetChildrenMedia)
		media.POST("/nested", middleware.JWTMiddleware(), mediaHandler.CreateMediaNested)
		media.POST(":id/move", middleware.JWTMiddleware(), mediaHandler.MoveMediaNested)
		media.DELETE(":id/nested", middleware.JWTMiddleware(), mediaHandler.DeleteMediaNested)
	}
}
