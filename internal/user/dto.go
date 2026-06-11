package user

type UserResponse struct {
	Email      string  `json:"email" example:"jane.doe@example.com"`
	FirstName  string  `json:"firstName" example:"Jane"`
	LastName   string  `json:"lastName" example:"Doe"`
	Username   string  `json:"username" example:"janedoe"`
	Picture    *string `json:"picture" example:"https://example.com/avatar.jpg"`
	IsVerified bool    `json:"isVerified" example:"true"`
}