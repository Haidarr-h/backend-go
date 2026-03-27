package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-KEY")
		expected := os.Getenv("CLIENT_API_KEY")

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
