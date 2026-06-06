package user

type UserResponse struct {
	Email      string  `json:"email" example:"jane.doe@example.com"`
	FirstName  string  `json:"firstName" example:"Jane"`
	LastName   string  `json:"lastName" example:"Doe"`
	Username   string  `json:"username" example:"janedoe"`
	Picture    *string `json:"picture" example:"https://example.com/avatar.jpg"`
	IsVerified bool    `json:"isVerified" example:"true"`
}

type updateReq struct {
	Id       int    `json:"id" binding:"required,min=1"`
	FullName string `json:"fullName"`
	Username string `json:"username"`
}