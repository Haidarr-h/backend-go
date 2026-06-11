package main

import (
	"fmt"

	_ "github.com/Haidarr-h/backend-go/docs"
	"github.com/Haidarr-h/backend-go/internal/auth"
	"github.com/Haidarr-h/backend-go/internal/config"
	"github.com/Haidarr-h/backend-go/internal/exercise"
	"github.com/Haidarr-h/backend-go/internal/health"
	"github.com/Haidarr-h/backend-go/internal/migrate"
	"github.com/Haidarr-h/backend-go/internal/routine"
	"github.com/Haidarr-h/backend-go/internal/session"
	"github.com/Haidarr-h/backend-go/internal/user"
	"github.com/Haidarr-h/backend-go/pkg/cache"
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
	r.GET("/healthCheck", health.HealthCheck)

	// AUTH ROUTE
	userRepo := user.NewUserRepository(cfg.DB)
	refreshRepo := auth.NewRefreshTokenRepository(cfg.DB)
	registrationCache := cache.NewRegistrationCache()
	registrationCache.StartCleanup(make(chan struct{}))
	authService := auth.NewAuthService(userRepo, cfg, refreshRepo, registrationCache)
	authHandler := auth.NewAuthHandler(authService)

	// EXERCISE ROUTE
	exerciseRepo := exercise.NewExerciseRepository(cfg.DB)
	exerciseService := exercise.NewExerciseService(exerciseRepo, cfg)
	exerciseHandler := exercise.NewExerciseHandler(exerciseService)

	// ROUTINE ROUTE
	routineRepo := routine.NewRoutineRepository(cfg.DB)
	routineService := routine.NewRoutineService(routineRepo, exerciseRepo)
	routineHandler := routine.NewRoutineHandler(routineService)

	// USER ROUTE
	userService := user.NewUserService(userRepo)
	userHandler := user.NewUserHandler(userService)

	// SESSION ROUTE
	sessionRepo := session.NewSessionRepository(cfg.DB)
	sessionService := session.NewSessionService(sessionRepo, routineRepo)
	sessionHandler := session.NewSessionHandler(sessionService)

	// ORCHESTRATE
	routes.RegisterRoutes(r, cfg, authHandler, exerciseHandler, routineHandler, userHandler, sessionHandler)

	migrate.Run(cfg.DB)
	r.Run(":" + cfg.Port)
	
}
