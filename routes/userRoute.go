package routes

import (
	"github.com/Haidarr-h/backend-go/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(rg *gin.RouterGroup) {
	user := rg.Group("/users")
	{
		user.GET("/", controllers.GetUsers)
		user.GET("/:id", controllers.GetUser)
		user.POST("/", controllers.CreateExercise)
		user.PATCH("/", controllers.UpdateUser)
		user.DELETE("/:id", controllers.DeleteUser)
	}
}
