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
func Login2(c *gin.Context) {
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
