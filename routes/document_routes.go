package routes

import (
	"github.com/dipankarupd/text-editor/controllers"
	"github.com/dipankarupd/text-editor/ws"
	"github.com/gin-gonic/gin"
)

func DocumentRoutes(route *gin.Engine) {
	route.POST("/documents", controllers.CreateDocument())
	route.GET("/documents/me", controllers.GetUserDocuments())
	route.GET("/documents/:id", controllers.GetDocumentByID())
	route.PATCH("/documents/:id", controllers.UpdateDocumentTitle()) 
}

// websocket route

func WebSocketRoutes(route *gin.Engine) {
	route.GET("/ws/:docId", ws.WebSocketHandler)
}