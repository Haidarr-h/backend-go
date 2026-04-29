package jwt

import (
	"time"

	"github.com/Haidarr-h/backend-go/config"
	"github.com/Haidarr-h/backend-go/dto"
	"github.com/Haidarr-h/backend-go/models"
	"github.com/Haidarr-h/backend-go/repositories"
	"github.com/golang-jwt/jwt/v5"
)

// CreateAccessRefreshToken generates access and refresh tokens
// used for: sign in manual, sign in by google, and refresh 
func CreateAccessRefreshToken(userID uint, cfg *config.Config) (dto.Tokens, error) {

	s := repositories.NewRefreshTokenRepository(cfg.DB)

	// 1. Generate the refresh token
	refreshTokenString, refreshErr := CreateToken(24*30, userID, true, cfg)

	if refreshErr != nil {
		return dto.Tokens{}, refreshErr
	}

	// 2. create new row in the refresh token table
	refreshTokenModel := models.RefreshToken{
		UserID:    userID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
	}

	if _, createErr := s.Create(&refreshTokenModel); createErr != nil {
		return dto.Tokens{}, createErr
	}

	// 3. create the access token
	accessTokenString, accessErr := CreateToken(12, userID, false, cfg)

	if accessErr != nil {
		return dto.Tokens{}, refreshErr
	}

	return dto.Tokens{AccessToken: accessTokenString, RefreshToken: refreshTokenString}, nil
}

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