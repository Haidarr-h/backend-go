package user

type UserResponse struct {
	Email      string  `json:"email"`
	FirstName  string  `json:"firstName"`
	LastName   string  `json:"lastName"`
	Username   string  `json:"username"`
	Picture    *string `json:"picture"`
	IsVerified bool    `json:"isVerified"`
}

type updateReq struct {
	Id       int    `json:"id" binding:"required,min=1"`
	FullName string `json:"fullName"`
	Username string `json:"username"`
}