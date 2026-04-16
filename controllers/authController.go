package controllers

import (
	"github.com/Haidarr-h/backend-go/dto"
	"github.com/Haidarr-h/backend-go/pkg/response"
	"github.com/Haidarr-h/backend-go/services"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// Signup godoc
// @Summary      Register a new user
// @Description  Create a new user account with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.SignUpRequest    true  "Signup credentials"
// @Success      201   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Router       /api/v1/auth/signup [post]
func (rc *AuthController) SignUp(c *gin.Context) {
	
	// 1. Parse the body request
	var body dto.SignUpRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "Bad request", err)
		return
	}

	// 2. Pass to the service
	respon, err := rc.authService.SignUp(body)

	if err != nil {
		response.InternalError(c, "failed to perform auth service", err.Error())
		return
	}

	// 3. Success
	response.OK(c, "Succesfully signed up", respon)
}

// Login godoc
// @Summary      Login a user
// @Description  Authenticate user and return a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.SignInReq   true  "Login credentials"
// @Success      200   {object}  dto.SignInRes
// @Failure      400   {object}  any
// @Router       /api/v1/auth/signin [post]
func (rc *AuthController) SignIn(c *gin.Context) {

	// 1. parse the body
	var body dto.SignInReq

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "Bad request", err.Error())
		return
	}

	// 2. send to service
	token, err := rc.authService.SignIn(body)

	if err != nil {
		response.InternalError(c, "failed to perform login service", err.Error())
		return
	}

	// 3. return the token
	response.OK(c, "Sign In Successful", token)
}