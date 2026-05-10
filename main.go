package main

import (
	"fmt"

	"github.com/Haidarr-h/backend-go/config"
	"github.com/Haidarr-h/backend-go/controllers"
	_ "github.com/Haidarr-h/backend-go/docs"
	"github.com/Haidarr-h/backend-go/internal/auth"
	"github.com/Haidarr-h/backend-go/internal/user"
	"github.com/Haidarr-h/backend-go/pkg/logger"
	"github.com/Haidarr-h/backend-go/routes"
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

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func init() {
	config.LoadEnvVariables()
	config.InitDB()
	logger.Init()
}

func main() {
	fmt.Println("Web Server started")

	cfg := config.Load()

	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization", "X-Api-Key"},
		// AllowCredentials: true
	}))

	// ROUTE
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/healthCheck", controllers.HealthCheck)

	// AUTH ROUTE
	userRepo := user.NewUserRepository(cfg.DB)
	refreshRepo := auth.NewRefreshTokenRepository(cfg.DB)
	otpRepo := auth.NewOTPRepository(cfg.DB)
	authService := auth.NewAuthService(userRepo, cfg, refreshRepo, otpRepo)
	authHandler := auth.NewAuthHandler(authService)

	routes.RegisterRoutes(r, cfg, authHandler)

	r.Run(":" + cfg.Port)
}
