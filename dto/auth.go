package dto

// SIGN UP
type SignUpRequest struct {
	Email    string `json:"email" binding:"required,email" example:"haidar@gmail.com"`
	Password string `json:"password" binding:"required,min=8,max=24" example:"haidarpassword"`
	Username string `json:"username" binding:"required,min=3,max=24" example:"haidarIron"`
	FullName string `json:"fullName" binding:"required,min=3,max=24" example:"Haidar Sebastian"`
}

type SignUpResponse struct {
	ID       uint   `json:"id"`
	FullName string `json:"fullName"`
	Username string `json:"username"`
}

// SIGN IN
type SignInReq struct {
	Email    string `json:"email" example:"haidar@gmail.com"`
	Password string `json:"password" example:"haidarpassword"`
}

type SignInRes struct {
	AccessToken string `json:"accessToken" example:"xxxxx"`
	RefreshToken string `json:"refreshToken" example:"yyyyyy"`
}

// REFRESH TOKEN
type RefreshTokenReq struct {
	RefreshToken string `json:"refreshToken" binding:"required" example:"yyyyy..."`
}

type RefreshTokenRes struct {
	AccessToken string `json:"accessToken" example:"xxx..."`
	RefreshToken string `json:"refreshToken" example:"yyyyyy"`
}
