package services

import "errors"

// auth - sign up
var ErrEmailIsExists = errors.New("email is already exist")
var ErrUsernameIsExists = errors.New("username is already exist")

// auth - sign in
var ErrInvalidCredentials = errors.New("invalid email or username or password")
var ErrUserGoogleSignIn = errors.New("user uses different way to sign in")
var ErrFailedCreateToken = errors.New("failed to create token")

// oauth
var ErrInvalidGoogleIDToken = errors.New("invalid google id token, missmatch")

// otp
var ErrInvalidOTPAttempts = errors.New("too many attempts (5 times), please request a new code")
var ErrOTPExpired = errors.New("verification code has expired, please request a new one")
var ErrInvalidOTPUsed = errors.New("verification code has already been used, please request a new one")
var ErrInvalidOTP = errors.New("invalid otp code")
