package controllers

import (
	"errors"

	"github.com/Haidarr-h/backend-go/dto"
	"github.com/Haidarr-h/backend-go/pkg/logger"
	"github.com/Haidarr-h/backend-go/pkg/response"
	"github.com/Haidarr-h/backend-go/pkg/validation"
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
// @Summary      Sign Up
// @Description  Create a new user account with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.SignUpRequest    true  "Signup credentials"
// @Param        x-api-key  header      string    true  "api key"
// @Security     ApiKeyAuth
// @Success      201   {object}  dto.SignUpResponse
// @Failure      400   {object}  dto.ErrorRes400
// @Failure      409   {object}  dto.ErrorRes400
// @Failure      500   {object}  dto.ErrorRes500
// @Router       /api/v1/auth/signup [post]
func (rc *AuthController) SignUp(c *gin.Context) {

	// 1. Parse the body request
	var body dto.SignUpRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Log.Debug("wrong body request", "body", body)
		response.BadRequest(c, "Bad request", validation.ParseValidationErrors(err))
		return
	}

	// 2. Pass to the service
	result, err := rc.authService.SignUp(body)

	if err != nil {
		if errors.Is(err, services.ErrEmailIsExists) || errors.Is(err, services.ErrUsernameIsExists) {
			response.Conflict(c, "failed to sign up", err.Error())
			return
		}

		logger.Log.Error("sign up internal server error", "error", err)
		response.InternalError(c, "failed to sign up", "internal server error")
		return
	}

	// 3. Success
	response.OK(c, "sign up successful, check email for verification", result)
}

// Login godoc
// @Summary      Sign in
// @Description  Authenticate user and return a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.SignInReq   true  "Login credentials"
// @Param        x-api-key  header      string    true  "api key"
// @Security     ApiKeyAuth
// @Success      200   {object}  dto.SignInRes
// @Failure      400   {object}  dto.ErrorRes400
// @Failure      401   {object}  dto.ErrorRes400
// @Failure      500   {object}  dto.ErrorRes500
// @Router       /api/v1/auth/signin [post]
func (rc *AuthController) SignIn(c *gin.Context) {

	// 1. parse the body
	var body dto.SignInReq

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "bad request", validation.ParseValidationErrors(err))
		return
	}

	// 2. send to service
	token, err := rc.authService.SignIn(body)

	if err != nil {

		// not found or wrong creds
		if errors.Is(err, services.ErrInvalidCredentials) {
			response.Unauthorized(c, "invalid credentials")
			return
		}

		// sign in with different method (ex: google)
		if errors.Is(err, services.ErrUserGoogleSignIn) {
			response.BadRequest(c, "failed to sign in", err.Error())
			return
		}

		// failed to create token
		if errors.Is(err, services.ErrFailedCreateToken) {
			response.InternalError(c, "failed to sign in", err.Error())
			return
		}

		response.InternalError(c, "failed to perform login service", err.Error())
		return
	}

	// 3. return the token
	response.OK(c, "sign in successful", token)
}

// Refresh godoc
// @Summary      Refresh Token
// @Description  refresh access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RefreshTokenReq   true  "Login credentials"
// @Param        x-api-key  header      string    true  "api key"
// @Security     ApiKeyAuth
// @Success      200   {object}  dto.RefreshTokenRes
// @Failure      400   {object}  any
// @Router       /api/v1/auth/refresh [post]
func (rc *AuthController) Refresh(c *gin.Context) {

	// 1. parse the body
	var req dto.RefreshTokenReq

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Bad request", err.Error())
		return
	}

	// 2. perform refresh
	result, err := rc.authService.Refresh(req)
	if err != nil {
		response.InternalError(c, "error", err.Error())
		return
	}

	response.OK(c, "refresh token successfull", result)
}

// SignOut godoc
// @Summary      Sign Out
// @Description  signout to delete the refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RefreshTokenReq   true  "Login credentials"
// @Param        x-api-key  header      string    true  "api key"
// @Security     ApiKeyAuth
// @Success      200   {object}  any
// @Failure      400   {object}  any
// @Router       /api/v1/auth/signout [post]
func (rc *AuthController) SignOut(c *gin.Context) {
	// 1. parse body
	var req dto.RefreshTokenReq

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "bad request", err.Error())
		return
	}

	// 2. delete the refresh token
	if err := rc.authService.DeleteToken(req); err != nil {
		response.InternalError(c, "failed to sign out", err.Error())
		return
	}

	response.OK(c, "sign out successful", nil)
}

// VerifyOtp godoc
// @Summary      Verify OTP
// @Description  verify the otp that sends to email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.VerifyOTPreq   true  "Login credentials"
// @Param        x-api-key  header      string    true  "api key"
// @Security     ApiKeyAuth
// @Success      200   {object}  any
// @Failure      400   {object}  dto.ErrorRes400
// @Failure      500   {object}  dto.ErrorRes500
// @Router       /api/v1/auth/verifyOTP [post]
func (rc *AuthController) VerifyOtp(c *gin.Context) {

	// 1. parse body
	var req dto.VerifyOTPreq

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "bad request", validation.ParseValidationErrors(err))
		logger.Log.Info("wrong request body at verify OTP")
		return
	}

	// 2. send the req body to service
	_, err := rc.authService.VerifyOTP(req)
	if err != nil {

		// too many attempts
		if errors.Is(err, services.ErrInvalidOTPAttempts) {
			response.BadRequest(c, "bad request", err.Error())
			return
		}

		// otp expired
		if errors.Is(err, services.ErrOTPExpired) {
			response.Gone(c, "OTP Expired", err.Error())
			return
		}

		// already used
		if errors.Is(err, services.ErrInvalidOTPUsed) {
			response.BadRequest(c, "bad request - OTP already used", err.Error())
			return
		}

		// invalid otp
		if errors.Is(err, services.ErrInvalidOTP) {
			response.BadRequest(c, "bad request - otp not match", err.Error())
			return
		}

		logger.Log.Error("internal error", "error", err)
		response.InternalError(c, "internal server error", "internal server error")
		return
	}

	response.OK(c, "otp verified successfully", "otp verified successfully")
}

// ResendOTP godoc
// @Summary      Resend OTP
// @Description  Resend OTP to the email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.ResendOTPreq   true  "Login credentials"
// @Param        x-api-key  header      string    true  "api key"
// @Security     ApiKeyAuth
// @Success      200   {object}  any
// @Failure      400   {object}  dto.ErrorRes400
// @Failure      500   {object}  dto.ErrorRes500
// @Router       /api/v1/auth/resendOTP [post]
func (rc *AuthController) ResendOTP(c *gin.Context) {

	// 1. parsing
	var req dto.ResendOTPreq

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "bad request", validation.ParseValidationErrors(err))
		return
	}

	// 2. goes to service
	resultErr := rc.authService.ResendOTP(req)
	if resultErr != nil {
		logger.Log.Error("failed to perform resend otp", "error", resultErr)
		response.InternalError(c, "internal server error", "internal server error")
		return
	}

	// 3. return
	response.OK(c, "successfully perform resend otp", "otp has been sent to email")
}
