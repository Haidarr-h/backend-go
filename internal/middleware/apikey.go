package middleware

import (
	"net/http"

	"github.com/Haidarr-h/backend-go/internal/config"
	"github.com/gin-gonic/gin"
)

func RequireAPIKey(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-KEY")
		expected := cfg.ClientAPIKey

		if key == "" || key != expected {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Missing or invalid API key",
			})
			return
		}
		c.Next()
	}
}
