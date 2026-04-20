package routes

import (
	"github.com/Haidarr-h/backend-go/config"
	"github.com/Haidarr-h/backend-go/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config) {
	api := r.Group("/api/v1")

	// PUBLIC ROUTES
	RegisterAuthRoutes(api, cfg)

	// JWT PROTECTED ROUTES
	protected := api.Group("/")
	protected.Use(middleware.RequireAuth(cfg))
	{
		RegisterUserRoutes(protected, cfg)
		RegisterRoutineRoutes(protected, cfg)
		RegisterExerciseRoutes(protected, cfg)
	}
}