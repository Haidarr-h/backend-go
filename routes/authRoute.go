package routes

import (
	"time"

	"github.com/Haidarr-h/backend-go/controllers"
	"github.com/Haidarr-h/backend-go/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	auth.Use(middleware.RequireAPIKey())
	auth.Use(middleware.RateLimit(5, time.Minute))
	{
		auth.POST("/signup", controllers.Signup)
		auth.POST("/signin", controllers.Login)
		auth.POST("/google/mobile", controllers.GoogleMobileSignIn)
	}
}
