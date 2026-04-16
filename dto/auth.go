package dto

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

type SignInReq struct {
	Email    string `json:"email" example:"haidar@gmail.com"`
	Password string `json:"password" example:"haidarpassword"`
}

type SignInRes struct {
	Token string `json:"token" example:"xxxxxxxx..."`
}
