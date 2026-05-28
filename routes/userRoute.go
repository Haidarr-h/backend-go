package routes

// import (
// 	"github.com/Haidarr-h/backend-go/controllers"
// 	"github.com/Haidarr-h/backend-go/internal/config"
// 	"github.com/gin-gonic/gin"
// )

// func RegisterUserRoutes(rg *gin.RouterGroup, cfg *config.Config) {
// 	user := rg.Group("/users")
// 	{
// 		user.GET("/", func(c *gin.Context) {
// 			controllers.GetUsers(c, cfg)
// 		})

// 		user.GET("/:id", func(c *gin.Context) {
// 			controllers.GetUser(c, cfg)
// 		})
// 		user.PATCH("/", func(c *gin.Context) {
// 			controllers.UpdateUser(c, cfg)
// 		})
// 		user.DELETE("/:id", func(c *gin.Context) {
// 			controllers.DeleteUser(c, cfg)
// 		})
// 	}
// }
