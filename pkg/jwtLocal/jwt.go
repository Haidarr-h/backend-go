package jwtlocal

import (
	"time"

	"github.com/Haidarr-h/backend-go/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

// Create token component function
func CreateToken(hour int, userID uint, isRefresh bool, cfg *config.Config) (string, error) {

	// 1. convert the hour to time.duration
	duration := time.Duration(hour) * time.Hour

	// 2. Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(duration).Unix(),
	})

	// 3. secret adjustments
	var secret = []byte(cfg.JWTSecret)

	if isRefresh {
		secret = []byte(cfg.RefreshSecret)
	}

	// 3. signed the token
	tokenString, err := token.SignedString(secret)

	// 4. error checks
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
