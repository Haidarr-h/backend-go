package main

import (
	"fmt"
	"os"

	"github.com/Haidarr-h/backend-go/controllers"
	_ "github.com/Haidarr-h/backend-go/docs" // swag generated docs
	"github.com/Haidarr-h/backend-go/initializers"
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
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// ROUTE
	// Swagger UI route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/healthCheck", controllers.HealthCheck)
	r.POST("/signup", controllers.Signup)
	r.POST("/signin", controllers.Login)
	r.POST("/auth/google/mobile", controllers.GoogleMobileSignIn)

	r.Run(":" + port)
}
