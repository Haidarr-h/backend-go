package routes

import (
	"github.com/Haidarr-h/backend-go/config"
	"github.com/Haidarr-h/backend-go/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterExerciseRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	exercise := rg.Group("/exercises")
	{
		exercise.GET("/", func(c *gin.Context) {
			controllers.GetExercises(c, cfg)
		})
		exercise.GET("/:id", func(c *gin.Context) {
			controllers.GetExercise(c, cfg)
		})
		exercise.POST("/", func(c *gin.Context) {
			controllers.CreateExercise(c, cfg)
		})
		exercise.PATCH("/", func(c *gin.Context) {
			controllers.UpdateExercises(c, cfg)
		})
		exercise.DELETE("/", func(c *gin.Context) {
			controllers.DeleteExercises(c, cfg)
		})
	}
}
