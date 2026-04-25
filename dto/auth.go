package dto

// SIGN UP
type SignUpRequest struct {
	Email     string `json:"email" binding:"required,email" example:"haidar@gmail.com"`
	Password  string `json:"password" binding:"required,min=8,max=24" example:"password123"`
	Username  string `json:"username" binding:"required,alphanum,min=3,max=24" example:"haidarIron"`
	FirstName string `json:"firstName" binding:"required,alpha,min=3,max=24" example:"Haidar"`
	LastName  string `json:"lastName" binding:"required,alpha,min=3,max=24" example:"Hanif"`
}

type SignUpResponse struct {
	ID        uint   `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
}

// SIGN IN
type SignInReq struct {
	Identifier string `json:"identifier" binding:"required,min=3,max=24" example:"haidar@gmail.com"`
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
