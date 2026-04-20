package routes

import (
	"github.com/Haidarr-h/backend-go/controllers"
	"github.com/Haidarr-h/backend-go/config"
	"github.com/Haidarr-h/backend-go/repositories"
	"github.com/Haidarr-h/backend-go/services"
	"github.com/gin-gonic/gin"
)

func RegisterRoutineRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	// wire dependencies
	routineRepo := repositories.NewRoutineRepository(cfg.DB)
	routineService := services.NewRoutineService(routineRepo)
	routineController := controllers.NewRoutineController(routineService)

	routine := rg.Group("/routines")
	{
		routine.POST("/", routineController.CreateRoutine)
		routine.GET("/", routineController.GetRoutines)
		routine.PUT("/:id", routineController.UpdateRoutine)
		routine.DELETE("/:id", routineController.DeleteRoutine)
	}
}
