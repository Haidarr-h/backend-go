package dto

type ErrorRes400 struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"failed to process..."`
	Error   string `json:"error,omitempty" example:"bad request"`
}

type ErrorRes500 struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"failed to process..."`
	Error   string `json:"error,omitempty" example:"internal server error"`
}
