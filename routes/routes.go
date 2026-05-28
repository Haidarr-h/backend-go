package routes

import (
	"github.com/Haidarr-h/backend-go/internal/auth"
	"github.com/Haidarr-h/backend-go/internal/config"
	"github.com/Haidarr-h/backend-go/internal/exercise"
	"github.com/Haidarr-h/backend-go/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, authHandler *auth.AuthHandler, exerciseHandler *exercise.ExerciseHandler) {
	api := r.Group("/api/v1")

	// PUBLIC ROUTES
	authHandler.RegisterRoutes(api, cfg)

	// JWT PROTECTED ROUTES
	protected := api.Group("/")
	protected.Use(middleware.RequireAuth(cfg))
	{
		exerciseHandler.RegisterRoutes(protected, cfg)
		// RegisterUserRoutes(protected, cfg)
		// RegisterRoutineRoutes(protected, cfg)
		// RegisterExerciseRoutes(protected, cfg)
	}
}