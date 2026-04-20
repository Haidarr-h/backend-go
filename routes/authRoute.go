package routes

import (
	"time"

	"github.com/Haidarr-h/backend-go/controllers"
	"github.com/Haidarr-h/backend-go/config"
	"github.com/Haidarr-h/backend-go/middleware"
	"github.com/Haidarr-h/backend-go/repositories"
	"github.com/Haidarr-h/backend-go/services"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	// 1. Wire the dependencies
	authRepo := repositories.NewUserRepository(cfg.DB)
	authService := services.NewAuthService(authRepo, cfg)
	authController := controllers.NewAuthController(authService)

	// 2. Define the routes
	auth := rg.Group("/auth")
	auth.Use(middleware.RequireAPIKey(cfg))
	auth.Use(middleware.RateLimit(5, time.Minute))
	{
		auth.POST("/signup", authController.SignUp)
		auth.POST("/signin", authController.SignIn)
		auth.POST("/google/mobile", func(c *gin.Context) {
			controllers.GoogleMobileSignIn(c, cfg)
		})
	}
}
