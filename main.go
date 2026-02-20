package main

import (
	"fmt"

	"github.com/Haidarr-h/backend-go/controllers"
	"github.com/Haidarr-h/backend-go/initializers"
	_ "github.com/Haidarr-h/backend-go/docs" // swag generated docs
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/gin-gonic/gin"
)

// @title           Backend Go API
// @version         1.0
// @description     My backend API built with Go and Gin

// @host      103.150.100.6:8080
// @BasePath  /

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
	initializers.SyncDatabase()
}

func main() {
	fmt.Println("Web Server started")

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

	r.Run()
}