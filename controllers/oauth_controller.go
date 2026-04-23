package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Haidarr-h/backend-go/config"
	"github.com/Haidarr-h/backend-go/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type GoogleTokenRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

type GoogleUserInfo struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// GoogleMobileSignIn godoc
// @Summary      Sign in with Google
// @Description  Authenticate user via Google ID token (mobile flow) and return a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      GoogleTokenRequest  true  "Google ID Token"
// @Success      200   {object}  map[string]interface{}  "Returns JWT token and user info"
// @Failure      400   {object}  map[string]interface{}  "id_token is required"
// @Failure      401   {object}  map[string]interface{}  "Invalid Google token"
// @Failure      500   {object}  map[string]interface{}  "Internal server error"
// @Router       /auth/google/mobile [post]
func GoogleMobileSignIn(c *gin.Context, cfg *config.Config) {
	// 1. Get the ID token from Flutter (sdks)
	var req GoogleTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id_token is required"})
		return
	}

	// 2. Verify the token with Google + get user info
	googleUser, err := verifyGoogleToken(req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google token"})
		return
	}
	fmt.Printf("Google user: %+v\n", googleUser)

	// 3. Find or create the user in your DB
	var user models.User
	result := cfg.DB.Where("google_id = ?", googleUser.Sub).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		// Check if email already exists
		emailResult := cfg.DB.Where("email = ?", googleUser.Email).First(&user)

		switch emailResult.Error {
		case gorm.ErrRecordNotFound:
			// if email not found = brand new user
			baseUsername := strings.Split(googleUser.Email, "@")[0]
			user = models.User{
				Email:    googleUser.Email,
				FirstName: googleUser.Name,
				LastName: googleUser.Name,
				GoogleID: &googleUser.Sub,
				Picture:  &googleUser.Picture,
				Username: baseUsername,
			}

			if err := cfg.DB.Create(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "failed to create user",
				})
				return
			}
		case nil:
			// Email already exist (signed up manually before) - Link google ID
			if err := cfg.DB.Model(&user).Updates(map[string]interface{}{
				"google_id": googleUser.Sub,
				"picture":   googleUser.Picture,
			}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "failed to link google account",
				})
				return
			}
		default:
			// Unexpected DB error — don't proceed
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
	}

	// 4. create refresh token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": fmt.Sprintf("%v", user.ID),
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	// 5. create access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": fmt.Sprintf("%v", user.ID),
		"exp": time.Now().Add(time.Hour * 12).Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refresh_token": tokenString,
		"access_name": accessTokenString,
		"user": gin.H{
			"id":      user.ID,
			"email":   user.Email,
			"name":    user.FirstName,
			"picture": user.Picture,
		},
	})
}

// Validate the ID Token with google endpoint
func verifyGoogleToken(idToken string) (*GoogleUserInfo, error) {
	url := "https://oauth2.googleapis.com/tokeninfo?id_token=" + idToken

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// check if google accept it (if expired, malformed, fake)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid token")
	}

	// read the body to get user data
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// put the user data to the userInfo
	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	// make sure the token intended for our app
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	var tokenData map[string]interface{}

	json.Unmarshal(body, &tokenData)
	if aud, ok := tokenData["aud"].(string); !ok || aud != clientID {
		return nil, fmt.Errorf("token audience mismatch")
	}

	return &userInfo, nil
}
