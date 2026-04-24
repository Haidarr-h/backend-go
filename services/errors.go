package services

import "errors"

// auth - sign up
var ErrEmailIsExists = errors.New("email is already exist")
var ErrUsernameIsExists = errors.New("username is already exist")

// auth - sign in
var ErrInvalidCredentials = errors.New("invalid email or username or password")
var ErrUserGoogleSignIn = errors.New("user uses different way to sign in")
var ErrFailedCreateToken = errors.New("failed to create token")
