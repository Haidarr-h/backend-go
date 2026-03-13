package main

import (
	"fmt"
	"os"

	"github.com/Haidarr-h/backend-go/controllers"
	_ "github.com/Haidarr-h/backend-go/docs" // swag generated docs
	"github.com/Haidarr-h/backend-go/initializers"
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

	// Swagger UI route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// test
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/healthCheck", controllers.HealthCheck)
	r.POST("/signup", controllers.Signup)
	r.POST("/signin", controllers.Login)

	r.Run(":" + port)
}