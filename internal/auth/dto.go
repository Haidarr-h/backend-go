package auth

// SIGN UP
type SignUpRequest struct {
	Email     string `json:"email" binding:"required,email" example:"haidar@gmail.com"`
	Password  string `json:"password" binding:"required,min=8,max=24" example:"password123"`
	Username  string `json:"username" binding:"required,alphanum,min=3,max=24" example:"haidarIron"`
	FirstName string `json:"firstName" binding:"required,min=3,max=24" example:"Haidar"`
	LastName  string `json:"lastName" binding:"required,alpha,min=3,max=24" example:"Hanif"`
}

type SignUpResponse struct {
	ID         uint   `json:"id"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Username   string `json:"username"`
	IsVerified bool   `json:"isVerified"`
}

// SIGN IN
type SignInReq struct {
	Identifier string `json:"identifier" binding:"required,min=3,max=320" example:"haidar@gmail.com"`
	Password   string `json:"password" binding:"required,min=8,max=24" example:"password123"`
}

type SignInRes struct {
	AccessToken  string `json:"accessToken" example:"xxxxx"`
	RefreshToken string `json:"refreshToken" example:"yyyyyy"`
}

// REFRESH TOKEN
type RefreshTokenReq struct {
	RefreshToken string `json:"refreshToken" binding:"required" example:"yyyyy..."`
}

type RefreshTokenRes struct {
	AccessToken  string `json:"accessToken" example:"xxx..."`
	RefreshToken string `json:"refreshToken" example:"yyyyyy"`
}

// JWT CREATION
type Tokens struct {
	AccessToken  string `json:"accessToken" example:"xxx..."`
	RefreshToken string `json:"refreshToken" example:"yyyyyy"`
}

// OTP
type VerifyOTPreq struct {
	Email   string `json:"email" binding:"required,email" example:"haidar@gmail.com"`
	OtpCode string `json:"otpCode" binding:"required" example:"111111"`
}

type ResendOTPreq struct {
	Email string `json:"email" binding:"required,email" example:"haidar@gmail.com"`
}

// OAUTH
type GoogleSignInReq struct {
	IDToken string `json:"id_token" binding:"required"`
}

type GoogleSignInRes struct {
	AccessToken  string `json:"accessToken" example:"xxx..."`
	RefreshToken string `json:"refreshToken" example:"yyyyyy"`
	ID           uint   `json:"id"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Username     string `json:"username"`
}

type GoogleUserInfo struct {
	Sub        string `json:"sub"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Picture    string `json:"picture"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
}
