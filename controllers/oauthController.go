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

type GoogleTokenRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

type GoogleUserInfo struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type OAuthController struct {
	oAuthService *services.OAuthService
}

func NewOAuthController(oAuthService *services.OAuthService) *OAuthController {
	return &OAuthController{oAuthService: oAuthService}
}

// GoogleMobileSignIn godoc
// @Summary      Sign in with Google
// @Description  Authenticate user via Google ID token (mobile flow) and return a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.GoogleSignInReq  true  "Google ID Token"
// @Success      200   {object}  dto.GoogleSignInRes  "Returns JWT token and user info"
// @Failure      400   {object}  dto.ErrorRes400
// @Failure      500   {object}  dto.ErrorRes500
// @Router       /auth/google/mobile [post]
func (rc *OAuthController) GoogleMobileSignIn(c *gin.Context) {

	// 1. Parse the ID token from client
	var req dto.GoogleSignInReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "bad request", validation.ParseValidationErrors(err))
		return
	}

	// 2. pass to service
	result, err := rc.oAuthService.GoogleSignIn(req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidGoogleIDToken) {
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
