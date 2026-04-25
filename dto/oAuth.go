package dto

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
