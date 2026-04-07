package routes

import (
	"github.com/Haidarr-h/backend-go/controllers"
	"github.com/Haidarr-h/backend-go/initializers"
	"github.com/Haidarr-h/backend-go/repositories"
	"github.com/Haidarr-h/backend-go/services"
	"github.com/gin-gonic/gin"
)

func RegisterRoutineRoutes(rg *gin.RouterGroup) {
	// wire dependencies
	routineRepo := repositories.NewRoutineRepository(initializers.DB)
	routineService := services.NewRoutineService(routineRepo)
	routineController := controllers.NewRoutineController(routineService)

	routine := rg.Group("/routine")
	{
		routine.POST("/", routineController.CreateRoutine)
	}

}
