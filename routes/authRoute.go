package routes

import (
	"time"

	"github.com/Haidarr-h/backend-go/controllers"
	"github.com/Haidarr-h/backend-go/initializers"
	"github.com/Haidarr-h/backend-go/middleware"
	"github.com/Haidarr-h/backend-go/repositories"
	"github.com/Haidarr-h/backend-go/services"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(rg *gin.RouterGroup) {
	// 1. Wire the dependencies
	authRepo := repositories.NewUserRepository(initializers.DB)
	authService := services.NewAuthService(authRepo)
	authController := controllers.NewAuthController(authService)

	// 2. Define the routes
	auth := rg.Group("/auth")
	auth.Use(middleware.RequireAPIKey())
	auth.Use(middleware.RateLimit(5, time.Minute))
	{
		auth.POST("/signup", authController.SignUp)
		auth.POST("/signin", authController.SignIn)
		auth.POST("/google/mobile", controllers.GoogleMobileSignIn)
	}
}
