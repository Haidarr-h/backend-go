package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Haidarr-h/backend-go/controllers"
	_ "github.com/Haidarr-h/backend-go/docs" // swag generated docs
	"github.com/Haidarr-h/backend-go/initializers"
	"github.com/Haidarr-h/backend-go/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Backend Go API
// @version         1.0
// @description     My backend API built with Go and Gin

// @host      api-staging.liftlogs.my.id
// @schemes   https
// @BasePath  /

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
	initializers.SyncDatabase()
}

func main() {
	fmt.Println("Web Server started")
	port := os.Getenv("PORT")

	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
		// AllowCredentials: true
	}))

	// ROUTE
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/healthCheck", controllers.HealthCheck)

	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			authRoutes := v1.Group("/auth")
			authRoutes.Use(middleware.RequireAPIKey())
			authRoutes.Use(middleware.RateLimit(5, time.Minute))
			{
				authRoutes.POST("/signup", controllers.Signup)
				authRoutes.POST("/signin", controllers.Login)
				authRoutes.POST("/google/mobile", controllers.GoogleMobileSignIn)
			}

			exerciseRoutes := v1.Group("/exercises")
			{
				exerciseRoutes.GET("/", controllers.GetExercises)
				exerciseRoutes.GET("/:id", controllers.GetExercise)
				exerciseRoutes.POST("/", controllers.CreateExercise)
			}
		}
	}

	r.Run(":" + port)
}
