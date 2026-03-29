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

// @title           IronLog Backend Go API
// @version         1.0
// @description     Backend API for IronLog

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
		// version 1 routes
		v1 := api.Group("/v1")
		{
			// auth routes
			authRoutes := v1.Group("/auth")
			authRoutes.Use(middleware.RequireAPIKey())
			authRoutes.Use(middleware.RateLimit(5, time.Minute))
			{
				authRoutes.POST("/signup", controllers.Signup)
				authRoutes.POST("/signin", controllers.Login)
				authRoutes.POST("/google/mobile", controllers.GoogleMobileSignIn)
			}

			// jwt protected routes
			jwtProtectedRoutes := v1.Group("/")
			jwtProtectedRoutes.Use(middleware.RequireAuth())
			{
				// exercises
				exerciseRoutes := jwtProtectedRoutes.Group("/exercises")
				{
					exerciseRoutes.GET("/", controllers.GetExercises)
					exerciseRoutes.GET("/:id", controllers.GetExercise)
					exerciseRoutes.POST("/", controllers.CreateExercise)
					exerciseRoutes.PATCH("/", controllers.UpdateExercises)
					exerciseRoutes.DELETE("/", controllers.DeleteExercises)
				}
			}
		}
	}

	r.Run(":" + port)
}
