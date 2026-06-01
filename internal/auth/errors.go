package auth

import "errors"

// auth - sign up
var ErrEmailIsExists = errors.New("email already exists")
var ErrUsernameIsExists = errors.New("username already exists")

// auth - sign in
var ErrInvalidCredentials = errors.New("invalid email or username or password")
var ErrUserGoogleSignIn = errors.New("user uses different way to sign in")
var ErrFailedCreateToken = errors.New("failed to create token")

// oauth
var ErrInvalidGoogleIDToken = errors.New("invalid google id token, mismatch")

// otp
var ErrInvalidOTPAttempts = errors.New("too many attempts (5 times), please request a new code")
var ErrOTPExpired = errors.New("verification code has expired, please request a new one")
var ErrInvalidOTPUsed = errors.New("verification code has already been used, please request a new one")
var ErrInvalidOTP = errors.New("invalid otp code")
var ErrEmailNotFound = errors.New("email not found")
var ErrOTPUpdate = errors.New("failed to update otp record")
var ErrPendingRegistrationNotFound = errors.New("no pending registration found for this email, please sign up first")


// refresh token repo
var ErrGenerateToken = errors.New("failed to insert refresh token data")
var ErrExpiredToken = errors.New("refresh token is expired")

// error
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


