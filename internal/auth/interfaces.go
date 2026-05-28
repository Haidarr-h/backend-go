package auth

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
	Create(otp OtpVerification) (OtpVerification, error)
	FindByEmail(email string) (OtpVerification, error)
	Update(otp OtpVerification) (OtpVerification, error)
}

type RefreshRepo interface {
	Create(token *RefreshToken) (*RefreshToken, error)
	FindByToken(token string) (*RefreshToken, error)
	Update(token *RefreshToken) (*RefreshToken, error)
	DeleteByToken(token string) error
}
