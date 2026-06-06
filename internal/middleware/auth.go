package middleware

import (
	"net/http"
	"strings"

	"github.com/Haidarr-h/backend-go/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Grab the header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Missing authorization header",
			})
			return
		}

		// 2. Expect "Bearer Token"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authorization format",
			})
			return
		}

		// 3. Parse and validate the token
		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
			})
			return
		}

		// 4. Pass the user ID downstream to controllers
		claims, _ := token.Claims.(jwt.MapClaims)

		// convert it to uint
		sub, ok := claims["sub"].(float64)

		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid token claims",
			})
			return
		}
		
		c.Set("userID", uint(sub))
		c.Next()
	}
}
