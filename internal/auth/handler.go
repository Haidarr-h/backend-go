package auth

import (
	"errors"
	"time"

	"github.com/Haidarr-h/backend-go/config"
	"github.com/Haidarr-h/backend-go/middleware"
	"github.com/Haidarr-h/backend-go/pkg/logger"
	"github.com/Haidarr-h/backend-go/pkg/response"
	"github.com/Haidarr-h/backend-go/pkg/validation"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService Service
}

func NewAuthHandler(authService Service) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (rc *AuthHandler) RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config) {


	// 2. Define the routes
	auth := rg.Group("/auth")
	auth.Use(middleware.RequireAPIKey(cfg))
	auth.Use(middleware.RateLimit(5, time.Minute))
	{
		auth.POST("/signup", rc.SignUp)
		auth.POST("/signin", rc.SignIn)
		auth.POST("/refresh", rc.Refresh)
		auth.POST("/signout", rc.SignOut)
		auth.POST("/verifyOTP", rc.VerifyOtp)
		auth.POST("/resendOTP", rc.ResendOTP)
		auth.POST("/google/mobile", rc.GoogleMobileSignIn)
	}
}

// Signup godoc
// @Summary      Sign Up
// @Description  Create a new user account with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      SignUpRequest    true  "Signup credentials"
// @Param        x-api-key  header      string    true  "api key"
// @Security     ApiKeyAuth
// @Success      201   {object}  SignUpResponse
// @Failure      400   {object}  ErrorRes400
// @Failure      409   {object}  ErrorRes400
// @Failure      500   {object}  ErrorRes500
// @Router       /api/v1/auth/signup [post]
func (rc *AuthHandler) SignUp(c *gin.Context) {

	// 1. Parse the body request
	var body SignUpRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Log.Debug("wrong body request", "body", body)
		response.BadRequest(c, "Bad request", validation.ParseValidationErrors(err))
		return
	}

	// 2. Pass to the service
	result, err := rc.authService.SignUp(body)

	if err != nil {
		if errors.Is(err, ErrEmailIsExists) || errors.Is(err, ErrUsernameIsExists) {
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
// @Param        body  body      SignInReq   true  "Login credentials"
// @Param        x-api-key  header      string    true  "api key"
// @Security     ApiKeyAuth
// @Success      200   {object}  SignInRes
// @Failure      400   {object}  ErrorRes400
// @Failure      401   {object}  ErrorRes400
// @Failure      500   {object}  ErrorRes500
// @Router       /api/v1/auth/signin [post]
func (rc *AuthHandler) SignIn(c *gin.Context) {

	// 1. parse the body
	var body SignInReq

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "bad request", validation.ParseValidationErrors(err))
		return
	}

	// 2. send to service
	token, err := rc.authService.SignIn(body)

	if err != nil {

		// not found or wrong creds
		if errors.Is(err, ErrInvalidCredentials) {
			response.Unauthorized(c, "invalid credentials")
			return
		}

		// sign in with different method (ex: google)
		if errors.Is(err, ErrUserGoogleSignIn) {
			response.BadRequest(c, "failed to sign in", err.Error())
			return
		}

		// failed to create token
		if errors.Is(err, ErrFailedCreateToken) {
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
// @Param        body  body      RefreshTokenReq   true  "Login credentials"
// @Param        x-api-key  header      string    true  "api key"
// @Security     ApiKeyAuth
// @Success      200   {object}  RefreshTokenRes
// @Failure      400   {object}  any
// @Router       /api/v1/auth/refresh [post]
func (rc *AuthHandler) Refresh(c *gin.Context) {

	// 1. parse the body
	var req RefreshTokenReq

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Bad request", err.Error())
		return
	}

	// 2. perform refresh
	result, err := rc.authService.Refresh(req)
	if err != nil {
		
		if errors.Is(err, ErrExpiredToken) {
			response.BadRequest(c, "Bad Request", err.Error())
			return
		}
		
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
// @Param        body  body      RefreshTokenReq   true  "Login credentials"
// @Param        x-api-key  header      string    true  "api key"
// @Security     ApiKeyAuth
// @Success      200   {object}  any
// @Failure      400   {object}  any
// @Router       /api/v1/auth/signout [post]
func (rc *AuthHandler) SignOut(c *gin.Context) {
	// 1. parse body
	var req RefreshTokenReq

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
// @Description  verify the otp that send to email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      VerifyOTPreq   true  "Login credentials"
// @Param        x-api-key  header      string    true  "api key"
// @Security     ApiKeyAuth
// @Success      200   {object}  any
// @Failure      400   {object}  ErrorRes400
// @Failure      500   {object}  ErrorRes500
// @Router       /api/v1/auth/verifyOTP [post]
func (rc *AuthHandler) VerifyOtp(c *gin.Context) {

	// 1. parse body
	var req VerifyOTPreq

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "bad request", validation.ParseValidationErrors(err))
		logger.Log.Info("wrong request body at verify OTP")
		return
	}

	// 2. send the req body to service
	_, err := rc.authService.VerifyOTP(req)
	if err != nil {

		// too many attempts
		if errors.Is(err, ErrInvalidOTPAttempts) {
			response.BadRequest(c, "bad request", err.Error())
			return
		}

		// otp expired
		if errors.Is(err, ErrOTPExpired) {
			response.Gone(c, "OTP Expired", err.Error())
			return
		}

		// already used
		if errors.Is(err, ErrInvalidOTPUsed) {
			response.BadRequest(c, "bad request - OTP already used", err.Error())
			return
		}

		// invalid otp
		if errors.Is(err, ErrInvalidOTP) {
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
// @Param        body  body      ResendOTPreq   true  "Login credentials"
// @Param        x-api-key  header      string    true  "api key"
// @Security     ApiKeyAuth
// @Success      200   {object}  any
// @Failure      400   {object}  ErrorRes400
// @Failure      500   {object}  ErrorRes500
// @Router       /api/v1/auth/resendOTP [post]
func (rc *AuthHandler) ResendOTP(c *gin.Context) {

	// 1. parsing
	var req ResendOTPreq

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

// GoogleMobileSignIn godoc
// @Summary      Sign in with Google
// @Description  Authenticate user via Google ID token (mobile flow) and return a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      GoogleSignInReq  true  "Google ID Token"
// @Success      200   {object}  GoogleSignInRes  "Returns JWT token and user info"
// @Failure      400   {object}  ErrorRes400
// @Failure      500   {object}  ErrorRes500
// @Router       /auth/google/mobile [post]
func (rc *AuthHandler) GoogleMobileSignIn(c *gin.Context) {

	// 1. Parse the ID token from client
	var req GoogleSignInReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "bad request", validation.ParseValidationErrors(err))
		return
	}

	// 2. pass to service
	result, err := rc.authService.GoogleSignIn(req)
	if err != nil {
		if errors.Is(err, ErrInvalidGoogleIDToken) {
			response.BadRequest(c, "bad request - failed id token verification", err.Error())
			return
		}

		logger.Log.Error("invalid sign in by google", "error", err)
		response.InternalError(c, "internal server error", "internal server error")
		return
	}

	// 3. success
	response.OK(c, "sign in by google successfull", result)
}
