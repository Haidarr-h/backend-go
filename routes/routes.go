package routes

import (
	"github.com/Haidarr-h/backend-go/internal/auth"
	"github.com/Haidarr-h/backend-go/internal/config"
	"github.com/Haidarr-h/backend-go/internal/exercise"
	"github.com/Haidarr-h/backend-go/internal/middleware"
	"github.com/Haidarr-h/backend-go/internal/routine"
	"github.com/Haidarr-h/backend-go/internal/user"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, authHandler *auth.AuthHandler, exerciseHandler *exercise.ExerciseHandler, routineHandler *routine.RoutineHandler, userHandler *user.UserHandler) {
	api := r.Group("/api/v1")

	// PUBLIC ROUTES
	authHandler.RegisterRoutes(api, cfg)

	// JWT PROTECTED ROUTES
	protected := api.Group("/")
	protected.Use(middleware.RequireAuth(cfg))
	{
		userHandler.RegisterUserRoutes(protected)
		exerciseHandler.RegisterRoutes(protected, cfg)
		routineHandler.RegisterRoutes(protected, cfg)
	}
}
