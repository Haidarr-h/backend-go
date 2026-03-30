package controllers

import (
	"os"
	"time"

	"github.com/Haidarr-h/backend-go/initializers"
	"github.com/Haidarr-h/backend-go/models"
	"github.com/Haidarr-h/backend-go/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthRequest struct {
	Email    string `json:"email" example:"haidar@gmail.com"`
	Password string `json:"password" example:"haidarpassword"`
}

type SignupRequest struct {
	Email    string `json:"email" binding:"required,email" example:"haidar@gmail.com"`
	Password string `json:"password" binding:"required,min=8,max=24" example:"haidarpassword"`
	Username string `json:"username" binding:"required,min=3,max=24" example:"haidarIron"`
	FullName string `json:"fullName" binding:"required,min=3,max=24" example:"Haidar Sebastian"`
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
// @Router       /api/v1/auth/signup [post]
func Signup(c *gin.Context) {
	var body SignupRequest

	// read the request body
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "Failed to read request body", err.Error())
		return
	}

	// check if email already exist
	var existingEmail models.User
	errorEmail := initializers.DB.Where("email = ?", body.Email).First(&existingEmail).Error

	if errorEmail == nil {
		response.BadRequest(c, "Failed. Email already exist", "Email already exist")
		return
	}

	// check if username already exist
	var existingUsername models.User
	errorUsername := initializers.DB.Where("username = ?", body.Username).First(&existingUsername).Error

	if errorUsername == nil {
		response.BadRequest(c, "Failed. Username already exist", "Username already exist")
		return
	}

	// hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		response.InternalError(c, "Failed to hash password", err.Error())
		return
	}

	// create user model
	hashedPassword := string(hash)
	user := models.User{
		Email:    body.Email,
		Password: &hashedPassword, 
		FullName: body.FullName,
		Username: body.Username,
	}

	// create the user data to database
	result := initializers.DB.Create(&user)
	if result.Error != nil {
		response.InternalError(c, "Failed to create user", result.Error.Error())
		return
	}

	response.Created(c, "User Created Succesfully", user)
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
// @Router       /api/v1/auth/signin [post]
func Login(c *gin.Context) {
	var body AuthRequest

	// read the content type to decides how to parse the body
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "Failed to read request body", err.Error())
		return
	}

	// check if user exist
	var user models.User
	initializers.DB.First(&user, "email = ? ", body.Email)

	if user.ID == 0 {
		response.Unauthorized(c, "Invalid Email or Password")
		return
	}

	if user.Password == nil {
		response.Unauthorized(c, "This account uses Google sign in")
		return
	}

	// checks passwords
	err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(body.Password))

	if err != nil {
		response.BadRequest(c, "Invalid Email or Password", err.Error())
		return
	}

	// token generation
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		response.BadRequest(c, "Failed to create token", err.Error())
		return
	}

	response.OK(c, "Login Successfull", gin.H{
		"token": tokenString,
	})
}
