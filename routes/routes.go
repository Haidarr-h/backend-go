package routes

import (
	"github.com/Haidarr-h/backend-go/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")

	// PUBLIC ROUTES
	RegisterAuthRoutes(api)

	// JWT PROTECTED ROUTES
	protected := api.Group("/")
	protected.Use(middleware.RequireAuth())
	{
		RegisterUserRoutes(protected)
		RegisterRoutineRoutes(protected)
	}
}