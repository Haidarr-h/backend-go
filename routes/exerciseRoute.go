package routes

import (
	"github.com/Haidarr-h/backend-go/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterExerciseRoutes(rg *gin.RouterGroup) {
	exercise := rg.Group("/exercises")
	{
		exercise.GET("/", controllers.GetExercises)
		exercise.GET("/:id", controllers.GetExercise)
		exercise.POST("/", controllers.CreateExercise)
		exercise.PATCH("/", controllers.UpdateExercises)
		exercise.DELETE("/", controllers.DeleteExercises)
	}
}
