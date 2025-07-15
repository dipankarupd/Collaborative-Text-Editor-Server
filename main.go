package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dipankarupd/text-editor/controllers"
	"github.com/dipankarupd/text-editor/db"
	"github.com/dipankarupd/text-editor/middlewares"
	"github.com/dipankarupd/text-editor/routes"
	"github.com/dipankarupd/text-editor/ws"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)
func init() {
	// Load .env only if it exists (typically for local dev)
	err := godotenv.Load()
	if err != nil {
		fmt.Println("⚠️ No .env file found, proceeding with environment variables")
	}
}


func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	database := db.ConnectDB()
	db.InitRedis()

	controllers.InitControllers(database)
	ws.InitDb(database)

	config := cors.Config{
		AllowOrigins:     []string{"https://collaborative-text-edito-92724.web.app"}, // frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders: []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	

	router := gin.New()

	router.Use(cors.New(config))

	router.Use(gin.Logger())

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"ok": "success"})

	})

	routes.UserRoutes(router)
	routes.WebSocketRoutes(router)
		

	router.Use(middlewares.Authentication())

	routes.UserSecureRoutes(router)

	routes.DocumentRoutes(router)
	



	

	// start the app:
	fmt.Println("Starting the app. Running in port: " + port)
	router.Run(":" + port)
}
