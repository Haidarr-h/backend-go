package routes

import (
	"github.com/Haidarr-h/backend-go/config"
	"github.com/Haidarr-h/backend-go/internal/auth"
	"github.com/Haidarr-h/backend-go/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, authHandler *auth.AuthHandler) {
	api := r.Group("/api/v1")

	// PUBLIC ROUTES
	authHandler.RegisterRoutes(api, cfg)
	RegisterUserRoutes(api, cfg)

	// JWT PROTECTED ROUTES
	protected := api.Group("/")
	protected.Use(middleware.RequireAuth(cfg))
	{
		RegisterRoutineRoutes(protected, cfg)
		RegisterExerciseRoutes(protected, cfg)
	}
}