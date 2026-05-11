package auth

import "github.com/Haidarr-h/backend-go/models"

// consumed by handler
type Service interface {
    SignUp(req SignUpRequest) (SignUpResponse, error)
    SignIn(req SignInReq) (SignInRes, error)
    Refresh(req RefreshTokenReq) (RefreshTokenRes, error)
    DeleteToken(req RefreshTokenReq) error
    VerifyOTP(req VerifyOTPreq) (bool, error)
    ResendOTP(req ResendOTPreq) error
    GoogleSignIn(req GoogleSignInReq) (GoogleSignInRes, error)
}

// consumed by service
type OtpRepo interface {
    Create(otp models.OtpVerification) (models.OtpVerification, error)
    FindByEmail(email string) (models.OtpVerification, error)
    Update(otp models.OtpVerification) (models.OtpVerification, error)
}

type RefreshRepo interface {
    Create(token *models.RefreshToken) (*models.RefreshToken, error)
    FindByToken(token string) (*models.RefreshToken, error)
    Update(token *models.RefreshToken) (*models.RefreshToken, error)
    DeleteByToken(token string) error
}