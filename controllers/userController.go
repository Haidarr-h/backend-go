package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/Haidarr-h/backend-go/initializers"
	"github.com/Haidarr-h/backend-go/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthRequest struct {
	Email    string `json:"email" example:"haidarGaming69@gmail.com"`
	Password string `json:"password" example:"haidarGaming123"`
}

type SignupRequest struct {
	Email    string `json:"email" binding:"required,email" example:"haidarGaming69@example.com"`
	Password string `json:"password" binding:"required,min=8,max=24" example:"haidarGaming123"`
	Username string `json:"username" binding:"required,min=3,max=24" example:"haidarsebastian99"`
	FullName string `json:"fullName" binding:"required,min=3,max=24" example:"Haidar Maximus Sebastian"`
}

type AuthResponse struct {
	Token string `json:"token" example:"xxxxxxxxxxxxxxx..."`
}

type ErrorResponse struct {
	Error string `json:"error" example:"Invalid Email or Password"`
}

// Signup godoc
// @Summary      Register a new user
// @Description  Create a new user account with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      SignupRequest    true  "Signup credentials"
// @Success      201   {object}  map[string]interface{}
// @Failure      400   {object}  ErrorResponse
// @Router       /signup [post]
func Signup(c *gin.Context) {
	var body SignupRequest

	// read the request body
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Failed to read request body",
			"detail": err.Error(),
		})
		return
	}

	// check if email already exist
	var existingEmail models.User
	errorEmail := initializers.DB.Where("email = ?", body.Email).First(&existingEmail).Error

	if errorEmail == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Failed. Email already exist",
			"detail": errorEmail,
		})
		return
	}

	// check if username already exist
	var existingUsername models.User
	errorUsername := initializers.DB.Where("username = ?", body.Username).First(&existingUsername).Error

	if errorUsername == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Failed. Username already exist",
			"detail": errorUsername,
		})
		return
	}

	// hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	// create user model
	hashedPassword := string(hash)
	user := models.User{
		Email:    body.Email,
		Password: &hashedPassword, // 👈 pass the address
		FullName: body.FullName,
		Username: body.Username,
	}

	// create the user data to database
	result := initializers.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Failed to create user",
			"detail": result.Error,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
	})
}

// Login godoc
// @Summary      Login a user
// @Description  Authenticate user and return a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      AuthRequest   true  "Login credentials"
// @Success      200   {object}  AuthResponse
// @Failure      400   {object}  ErrorResponse
// @Router       /signin [post]
func Login(c *gin.Context) {
	var body AuthRequest

	// read the content type to decides how to parse the body
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// check if user exist
	var user models.User
	initializers.DB.First(&user, "email = ? ", body.Email)

	if user.ID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid Email or Password",
		})
		return
	}

	if user.Password == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "This account uses Google sign in"})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid Email or Password",
		})
		return
	}

	// token generation
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create token",
			"err":   err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
	})
}
