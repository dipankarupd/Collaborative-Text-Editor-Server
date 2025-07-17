package routes

import (
	"github.com/dipankarupd/text-editor/controllers"
	"github.com/gin-gonic/gin"
)

func UserRoutes(route *gin.Engine) {
	route.POST("/users/register", controllers.RegisterUser())
	route.GET("/users", controllers.GetUsers())
	route.GET("/users/:id", controllers.GetUser())
	route.POST("/users/login", controllers.Login())
	route.POST("/users/login/google", controllers.LoginWithGoogle())
	route.GET("/refresh", controllers.RefreshHandler())
}


func UserSecureRoutes(route *gin.Engine) {
	route.GET("/users/me", controllers.GetLoggedInUser())
	route.POST("/users/logout", controllers.Logout())
}